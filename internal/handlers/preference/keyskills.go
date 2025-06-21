package preference

import (
	"RAAS/internal/models"
	"RAAS/internal/dto"

	"context"
	"log"
	"net/http"
	"time"
	"strings"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type KeySkillsHandler struct{}

func NewKeySkillsHandler() *KeySkillsHandler {
	return &KeySkillsHandler{}
}
// SetKeySkills sets or updates the key skills of the authenticated user
func (h *KeySkillsHandler) SetKeySkills(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")
	timelineCollection := db.Collection("user_entry_timelines")

	var input dto.KeySkillsRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		log.Printf("Error binding key skills input: %v", err)
		return
	}

	// Clean up skills
	var cleanedSkills []string
	for _, skill := range input.Skills {
		cleaned := strings.ReplaceAll(skill, "\n", "")
		cleaned = strings.TrimSpace(cleaned)
		if cleaned != "" {
			cleanedSkills = append(cleanedSkills, cleaned)
		}
	}

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

	operation := "updated"
	if len(seeker.KeySkills) == 0 {
		operation = "created"
	}

	// Update seeker's key skills
	update := bson.M{
		"$set": bson.M{
			"key_skills": cleanedSkills,
			"updated_at": time.Now(),
		},
	}

	result, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set key skills"})
		log.Printf("Failed to set key skills for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching seeker found"})
		log.Printf("No seeker found to set key skills for auth_user_id: %s", userID)
		return
	}

	// ✅ Update UserEntryTimeline to mark key_skills_completed = true
	timelineUpdate := bson.M{
		"$set": bson.M{
			"key_skills_completed": true,
		},
	}
	_, err = timelineCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, timelineUpdate)
	if err != nil {
		log.Printf("Warning: Failed to update key_skills_completed for user timeline. auth_user_id: %s, error: %v", userID, err)
		// Don't block success response due to timeline update failure
	}

	c.JSON(http.StatusOK, gin.H{"message": "Key skills " + operation + " successfully"})
}


// GetKeySkills retrieves the key skills of the authenticated user
func (h *KeySkillsHandler) GetKeySkills(c *gin.Context) {
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

	if len(seeker.KeySkills) == 0 {
		c.JSON(http.StatusNoContent, gin.H{})
		return
	}

	c.JSON(http.StatusOK, dto.KeySkillsResponse{
		AuthUserID: userID,
		Skills:     seeker.KeySkills,
		CreatedAt:  seeker.CreatedAt,
		UpdatedAt:  seeker.UpdatedAt,
	})
}
