package jobs

import (

	"RAAS/internal/models"

	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

)

// LinkResponseDTO represents the response DTO for job application links
type LinkResponseDTO struct {
	JobID   string `json:"job_id"`
	JobLink string `json:"job_link"`
	Source  string `json:"source"`
}

// LinkProviderHandler handles requests for job application links
type LinkProviderHandler struct{}

// NewLinkProviderHandler returns a new instance of LinkProviderHandler
func NewLinkProviderHandler() *LinkProviderHandler {
	return &LinkProviderHandler{}
}

// PostAndGetLink handles POST requests to retrieve job application links
func (h *LinkProviderHandler) PostAndGetLink(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	authUserID, ok := c.MustGet("userID").(string)

	var req struct {
		JobID string `json:"job_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println("❌ Failed to bind request:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid job_id in request body"})
		return
	}
	jobID := req.JobID


	if !ok {
		fmt.Println("❌ Failed to get userID from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract user ID from JWT claims"})
		return
	}

	// Check if the job was selected by the user
	selectedJobCollection := db.Collection("selected_job_applications")
	var selectedJob models.SelectedJobApplication
	err := selectedJobCollection.FindOne(c, bson.M{
		"auth_user_id": authUserID,
		"job_id":       jobID,
	}).Decode(&selectedJob)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("🚫 Job not selected by the user")
			c.JSON(http.StatusForbidden, gin.H{"error": "Job not selected by the user"})
			return
		}
		fmt.Println("❌ DB error while checking selected job:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify job selection"})
		return
	}

	// Retrieve JobLink from the unified Job model
	jobCollection := db.Collection("jobs")
	var job models.Job
	err = jobCollection.FindOne(c, bson.M{"job_id": jobID}).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("🚫 Job not found in jobs table")
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			return
		}
		fmt.Println("❌ DB error while fetching job:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve job info"})
		return
	}

	// Build and return the response
	response := LinkResponseDTO{
		JobID:   job.JobID,
		JobLink: job.JobLink,
		Source:  job.Source,
	}

	// Update view_link = true
	_, err = selectedJobCollection.UpdateOne(c,
		bson.M{
			"auth_user_id": authUserID,
			"job_id":       jobID,
		},
		bson.M{
			"$set": bson.M{"view_link": true},
		},
	)
	if err != nil {
		fmt.Println("⚠️ Failed to update view_link field:", err)
	}

	c.JSON(http.StatusOK, response)
}
