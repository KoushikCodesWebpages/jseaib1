package appuser

import (
    "RAAS/internal/models"

	"context"
	"net/http"
    "time"
    "fmt"
    "strconv"


	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type ApplicationTrackerHandler struct{}

func NewApplicationTrackerHandler() *ApplicationTrackerHandler {
    return &ApplicationTrackerHandler{}
}

type ApplicationTrackerResponse struct {
	JobID        string 	`json:"job_id"`
	Title        string 	`json:"title"`
	Company      string 	`json:"company"`
	Location     string 	`json:"location"`
	JobTitle     string 	`json:"job_title"`
	JobDesc      string 	`json:"job_desc"`
	Skills       string 	`json:"skills"`
	KeySkills    []string 	`json:"key_skills"`
	MatchScore 	 float64   	`bson:"match_score" json:"match_score"` 
	Status       string 	`json:"status"`
	Source       string 	`json:"source"`
    SelectedDate time.Time  `json:"selected_date"`
}
func (h *ApplicationTrackerHandler) GetApplicationTracker(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    userID := c.MustGet("userID").(string)

    // 1️⃣ Parse pagination params
    page := 1
    size := 10
    if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
        page = p
    }
    if s, err := strconv.Atoi(c.DefaultQuery("size", "10")); err == nil && s > 0 {
        size = s
    }
    skip := (page - 1) * size

    // 1b️⃣ Optional status filter
    statusParam, statusProvided := c.GetQuery("status") // e.g. "pending", "applied", "interview"

    selColl := db.Collection("selected_job_applications")
    seekerColl := db.Collection("seekers")
    matchScoreColl := db.Collection("match_scores")
    internalJobColl := db.Collection("jobs")
    externalJobColl := db.Collection("external_jobs")

    // 2️⃣ Build Mongo filter
    filter := bson.M{
        "auth_user_id":           userID,
        "cv_generated":           true,
        "cover_letter_generated": true,
        "view_link":              true,
        "status":                 bson.M{"$ne": "deleted"}, // always exclude deleted
    }
    if statusProvided {
        filter["status"] = statusParam // only include the requested status
    }

    // 🧮 Count total
    total, err := selColl.CountDocuments(context.TODO(), filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count applications"})
        return
    }

    // 3️⃣ Fetch paginated selection documents
    cursor, err := selColl.Find(
        context.TODO(), filter,
        options.Find().SetSort(bson.D{{"selected_date", -1}}),
        options.Find().SetSkip(int64(skip)),
        options.Find().SetLimit(int64(size)),
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
        return
    }
    defer cursor.Close(context.TODO())

    var apps []models.SelectedJobApplication
    if err := cursor.All(context.TODO(), &apps); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode applications"})
        return
    }

    // 4️⃣ Only transition "pending" → "applied"
    for i := range apps {
        if apps[i].Status == "pending" {
            _, _ = selColl.UpdateOne(context.TODO(),
                bson.M{"auth_user_id": userID, "job_id": apps[i].JobID},
                bson.M{"$set": bson.M{"status": "applied"}},
            )
            apps[i].Status = "applied"
        }
    }

    // 5️⃣ Fetch seeker's key skills once
    var seeker models.Seeker
    _ = seekerColl.FindOne(context.TODO(), bson.M{"auth_user_id": userID}).Decode(&seeker)

    // 6️⃣ Build the response slice
    resp := make([]ApplicationTrackerResponse, 0, len(apps))
    for _, app := range apps {
        var title, company, location, jobTitle, jobDesc, skills string

        if app.Source == "external" {
            var ext struct {
                Title, Company, Description, Location, JobTitle, Skills string
            }
            if err := externalJobColl.FindOne(context.TODO(), bson.M{"job_id": app.JobID}).Decode(&ext); err != nil {
                continue
            }
            title, company, jobDesc, location, jobTitle, skills =
                ext.Title, ext.Company, ext.Description, ext.Location, ext.JobTitle, ext.Skills
        } else {
            var intJob struct {
                Title, Company, Location, JobTitle, JobDescription, Skills string
            }
            if err := internalJobColl.FindOne(context.TODO(), bson.M{"job_id": app.JobID}).Decode(&intJob); err != nil {
                continue
            }
            title, company, jobDesc, location, jobTitle, skills =
                intJob.Title, intJob.Company, intJob.JobDescription, intJob.Location, intJob.JobTitle, intJob.Skills
        }

        var match struct{ MatchScore float64 `bson:"match_score"` }
        _ = matchScoreColl.FindOne(context.TODO(),
            bson.M{"auth_user_id": userID, "job_id": app.JobID}).Decode(&match)

        resp = append(resp, ApplicationTrackerResponse{
            JobID:        app.JobID,
            Title:        title,
            Company:      company,
            Location:     location,
            JobTitle:     jobTitle,
            JobDesc:      jobDesc,
            Skills:       skills,
            KeySkills:    seeker.KeySkills,
            MatchScore:   match.MatchScore,
            Status:       app.Status,
            Source:       app.Source,
            SelectedDate: app.SelectedDate,
        })
    }

    // 7️⃣ Create pagination metadata
    totalPages := (int(total) + size - 1) / size
    next, prev := "", ""
    if page < totalPages {
        next = fmt.Sprintf("?page=%d&size=%d", page+1, size)
    }
    if page > 1 {
        prev = fmt.Sprintf("?page=%d&size=%d", page-1, size)
    }

    // 8️⃣ Send final payload
    c.JSON(http.StatusOK, gin.H{
        "pagination": gin.H{
            "total":       total,
            "per_page":    size,
            "current":     page,
            "total_pages": totalPages,
            "next":        next,
            "prev":        prev,
        },
        "applications": resp,
    })
}



type UpdateApplicationStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=interview selected rejected deleted"`
}

func (h *ApplicationTrackerHandler) UpdateApplicationStatus(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	selColl := db.Collection("selected_job_applications")
	userID := c.MustGet("userID").(string)
	jobID := c.Param("job_id")

	var req UpdateApplicationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
		return
	}

	// Update the status
	filter := bson.M{"auth_user_id": userID, "job_id": jobID}
	update := bson.M{"$set": bson.M{"status": req.Status}}

	res, err := selColl.UpdateOne(c, filter, update)
	if err != nil || res.MatchedCount == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated", "status": req.Status})
}


func (h *ApplicationTrackerHandler) GetCVAndCL(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    userID := c.MustGet("userID").(string)

    // ✅ Path parameter: job_id
    jobID := c.Param("job_id")
    if jobID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing job_id in URL"})
        return
    }

    cvColl := db.Collection("cv")
    clColl := db.Collection("cover_letters")

    var cvDoc models.CVData
    var clDoc models.CoverLetterData

    cvErr := cvColl.FindOne(c, bson.M{
        "auth_user_id": userID,
        "job_id":       jobID,
    }).Decode(&cvDoc)
    clErr := clColl.FindOne(c, bson.M{
        "auth_user_id": userID,
        "job_id":       jobID,
    }).Decode(&clDoc)

    if cvErr != nil || clErr != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error":    "CV or Cover Letter not found",
            "cv_error": cvErr.Error(),
            "cl_error": clErr.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "cv_data":   cvDoc.CVData,
        "cv_format": cvDoc.CvFormat,
        "cl_data":   clDoc.CLData,
        "cl_format": clDoc.ClFormat,
    })
}
