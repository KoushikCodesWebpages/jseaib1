package preference

import (
	"RAAS/internal/models"
	"RAAS/internal/dto"

	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type JobTitleHandler struct{}

// NewJobTitleHandler creates a new JobTitleHandler
func NewJobTitleHandler() *JobTitleHandler {
	return &JobTitleHandler{}
}

// CreateJobTitleOnce allows setting the job titles only once for the authenticated user
func (h *JobTitleHandler) CreateJobTitleOnce(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")
	entryTimelineCollection := db.Collection("user_entry_timelines")

	var input dto.JobTitleInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		log.Printf("Error binding input: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the seeker
	var seeker models.Seeker
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
			log.Printf("Seeker not found for auth_user_id: %s", userID)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving seeker"})
			log.Printf("Error retrieving seeker for auth_user_id: %s, Error: %v", userID, err)
		}
		return
	}

	// Check if job titles are already set
	if seeker.PrimaryTitle != "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Job titles are already set and cannot be changed.",
		})
		log.Printf("Attempt to modify existing job titles by auth_user_id: %s", userID)
		return
	}

	// Titles not set yet: proceed with update
	update := bson.M{
		"$set": bson.M{
			"primary_title":   input.PrimaryTitle,
			"secondary_title": input.SecondaryTitle,
			"tertiary_title":  input.TertiaryTitle,
		},
	}

	updateResult, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job titles", "details": err.Error()})
		log.Printf("Failed to update job titles for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	if updateResult.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching seeker found to update"})
		log.Printf("No matching seeker found for auth_user_id: %s", userID)
		return
	}

	// Update timeline: PreferredJobTitlesCompleted = true
	timelineUpdate := bson.M{
		"$set": bson.M{
			"preferred_job_titles_completed": true,
		},
	}
	if _, err := entryTimelineCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, timelineUpdate); err != nil {
		log.Printf("Failed to update timeline for auth_user_id: %s, Error: %v", userID, err)
		// Not fatal for the main request, so no return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job titles set successfully"})
}

// GetJobTitle retrieves the preferred job titles for the authenticated user
func (h *JobTitleHandler) GetJobTitle(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
			log.Printf("Seeker not found for auth_user_id: %s", userID)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving seeker"})
			log.Printf("Error retrieving seeker for auth_user_id: %s, Error: %v", userID, err)
		}
		return
	}

	// If the primary title is empty, return a 204 No Content
	if seeker.PrimaryTitle == "" {
		c.JSON(http.StatusNoContent, gin.H{})
		return
	}

	jobTitleResponse := dto.JobTitleResponse{
		AuthUserID:     seeker.AuthUserID,
		PrimaryTitle:   seeker.PrimaryTitle,
		SecondaryTitle: seeker.SecondaryTitle,
		TertiaryTitle:  seeker.TertiaryTitle,
	}

	c.JSON(http.StatusOK, jobTitleResponse)
}



// // PatchJobTitle allows partial update of job titles for the authenticated user
// func (h *JobTitleHandler) PatchJobTitle(c *gin.Context) {
// 	userID := c.MustGet("userID").(uuid.UUID)
// 	var input dto.JobTitleInput

// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
// 		return
// 	}

// 	// Find the Seeker model by AuthUserID
// 	var seeker models.Seeker
// 	if err := h.DB.Where("auth_user_id = ?", userID).First(&seeker).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found", "details": err.Error()})
// 		return
// 	}

// 	// Update only the fields that are provided
// 	if input.PrimaryTitle != "" {
// 		seeker.PrimaryTitle = input.PrimaryTitle
// 	}
// 	if input.SecondaryTitle != nil {
// 		seeker.SecondaryTitle = input.SecondaryTitle
// 	}
// 	if input.TertiaryTitle != nil {
// 		seeker.TertiaryTitle = input.TertiaryTitle
// 	}

// 	if err := h.DB.Save(&seeker).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch job title", "details": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Job title patched successfully", "jobTitle": seeker})
// }
