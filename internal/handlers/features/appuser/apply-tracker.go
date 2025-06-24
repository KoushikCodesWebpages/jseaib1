package appuser

import (
	"context"
	"net/http"

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
}

func (h *ApplicationTrackerHandler) GetApplicationTracker(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    userID := c.MustGet("userID").(string)

    selColl := db.Collection("selected_job_applications")
    jobColl := db.Collection("jobs")
    seekerColl := db.Collection("seekers")
    matchScoreColl := db.Collection("match_scores")

    // 1️⃣ Find applications where all three generation flags are true
    filter := bson.M{
        "auth_user_id":            userID,
        "cv_generated":            true,
        "cover_letter_generated":  true,
        "view_link":               true,
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

    // 2️⃣ Update status to "applied" if not already
    for _, app := range apps {
        if app.Status != "applied" {
            _, _ = selColl.UpdateOne(context.TODO(), bson.M{
                "auth_user_id": userID,
                "job_id":       app.JobID,
            }, bson.M{"$set": bson.M{"status": "applied"}})
            app.Status = "applied" // Keep the struct in sync
        }
    }

    // 3️⃣ Load seeker key skills once
    var seeker models.Seeker
    _ = seekerColl.FindOne(context.TODO(), bson.M{"auth_user_id": userID}).Decode(&seeker)

    // 4️⃣ Build response
    var resp []ApplicationTrackerResponse
    for _, app := range apps {
        var job models.Job
        if err := jobColl.FindOne(context.TODO(), bson.M{"job_id": app.JobID}).Decode(&job); err != nil {
            continue
        }

        var match struct{ MatchScore float64 `bson:"match_score"` }
        _ = matchScoreColl.FindOne(context.TODO(), bson.M{
            "auth_user_id": userID,
            "job_id":       app.JobID,
        }).Decode(&match)

		if app.Status=="pending"{
			app.Status="applied"
		}
        resp = append(resp, ApplicationTrackerResponse{
            JobID:      app.JobID,
            Title:      job.Title,
            Company:    job.Company,
            Location:   job.Location,
            JobTitle:   job.JobTitle,
            JobDesc:    job.JobDescription,
            Skills:     job.Skills,
            KeySkills:  seeker.KeySkills,
            MatchScore: match.MatchScore,
            Status:     app.Status,  // ← using dynamic status from DB
            Source:     app.Source,
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
