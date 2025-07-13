package appuser

import (
	"context"
	"net/http"
    "time"
	"RAAS/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

    selColl := db.Collection("selected_job_applications")
    seekerColl := db.Collection("seekers")
    matchScoreColl := db.Collection("match_scores")
    internalJobColl := db.Collection("jobs")
    externalJobColl := db.Collection("external_jobs")

    // 1️⃣ Fetch selections with all generated flags true
    filter := bson.M{
        "auth_user_id":           userID,
        "cv_generated":           true,
        "cover_letter_generated": true,
        "view_link":              true,
    }
    cursor, err := selColl.Find(context.TODO(), filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
        return
    }
    defer cursor.Close(context.TODO())

    var apps []models.SelectedJobApplication
    if err := cursor.All(context.TODO(), &apps); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode application data"})
        return
    }

    // 2️⃣ Update status to "applied" and reflect change in slice
    for idx := range apps {
        if apps[idx].Status != "applied" {
            _, _ = selColl.UpdateOne(context.TODO(),
                bson.M{"auth_user_id": userID, "job_id": apps[idx].JobID},
                bson.M{"$set": bson.M{"status": "applied"}},
            )
            apps[idx].Status = "applied"
        }
    }

    // 3️⃣ Load seeker skills
    var seeker models.Seeker
    _ = seekerColl.FindOne(context.TODO(), bson.M{"auth_user_id": userID}).Decode(&seeker)

    // 4️⃣ Build response
    resp := make([]ApplicationTrackerResponse, 0, len(apps))
    for _, app := range apps {
        var title, company, location, jobTitle, jobDesc, skills string

        if app.Source == "external" {
            var extJob struct {
                Title       string `bson:"title"`
                Company     string `bson:"company"`
                Description string `bson:"description"`
                Location    string `bson:"location,omitempty"`
                JobTitle    string `bson:"job_title,omitempty"`
                Skills      string `bson:"skills,omitempty"`
            }
            if err := externalJobColl.FindOne(context.TODO(), bson.M{"job_id": app.JobID}).Decode(&extJob); err != nil {
                continue
            }
            title = extJob.Title
            company = extJob.Company
            jobDesc = extJob.Description
            location = extJob.Location
            jobTitle = extJob.JobTitle
            skills = extJob.Skills
        } else {
            var intJob struct {
                Title          string `bson:"title"`
                Company        string `bson:"company"`
                Location       string `bson:"location"`
                JobTitle       string `bson:"job_title"`
                JobDescription string `bson:"job_description"`
                Skills         string `bson:"skills"`
            }
            if err := internalJobColl.FindOne(context.TODO(), bson.M{"job_id": app.JobID}).Decode(&intJob); err != nil {
                continue
            }
            title = intJob.Title
            company = intJob.Company
            jobDesc = intJob.JobDescription
            location = intJob.Location
            jobTitle = intJob.JobTitle
            skills = intJob.Skills
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

    c.JSON(http.StatusOK, resp)
}


type UpdateApplicationStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=interview selected rejected"`
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

    var req struct {
        JobID string `json:"job_id" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing job_id"})
        return
    }

    cvColl := db.Collection("cv")
    clColl := db.Collection("cover_letters")

    // Fetch both documents in parallel
    var (
        cvDoc models.CVData
        clDoc models.CoverLetterData
        cvErr = cvColl.FindOne(c, bson.M{"auth_user_id": userID, "job_id": req.JobID}).Decode(&cvDoc)
        clErr = clColl.FindOne(c, bson.M{"auth_user_id": userID, "job_id": req.JobID}).Decode(&clDoc)
    )

    if cvErr != nil || clErr != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "CV or Cover Letter not found",
            "cv_error": cvErr.Error(),
            "cl_error": clErr.Error(),
        })
        return
    }

    // Return both payload and formats
    c.JSON(http.StatusOK, gin.H{
        "cv_data":   cvDoc.CVData,
        "cv_format": cvDoc.CvFormat,
        "cl_data":   clDoc.CLData,
        "cl_format": clDoc.ClFormat,
    })
}
