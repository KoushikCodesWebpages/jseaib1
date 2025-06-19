package preference

import (
	"RAAS/internal/dto"
	"RAAS/internal/handlers/repository"
	"RAAS/internal/models"


	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LanguageHandler struct{}

func NewLanguageHandler() *LanguageHandler {
	return &LanguageHandler{}
}

// CreateLanguage handles the creation or update of a single language entry
// CreateLanguage handles the creation or update of a single language entry (no file upload)
func (h *LanguageHandler) CreateLanguage(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")
	entryTimelineCollection := db.Collection("user_entry_timelines")

	var input dto.LanguageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		log.Printf("Error binding input: %v", err)
		return
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
	validLevels := map[string]bool{
	"beginner":    true,
	"intermediate": true,
	"fluent":      true,
	}

	if !validLevels[input.ProficiencyLevel] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proficiency level. Must be one of: beginner, intermediate, fluent, native"})
		return
	}
	// Append the new language (no file URL passed)
	if err := repository.AppendToLanguages(&seeker, input, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process language"})
		log.Printf("Failed to process language for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	// Update seeker document
	update := bson.M{
		"$set": bson.M{
			"languages": seeker.Languages,
		},
	}

	updateResult, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save language"})
		log.Printf("Failed to update language for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	if updateResult.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching seeker found to update"})
		log.Printf("No matching seeker found for auth_user_id: %s", userID)
		return
	}

	// Update user entry timeline to mark languages completed
	timelineUpdate := bson.M{
		"$set": bson.M{
			"languages_completed": true,
		},
	}
	if _, err := entryTimelineCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, timelineUpdate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user entry timeline"})
		log.Printf("Failed to update user entry timeline for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Language added successfully",
	})
}


// GetLanguages handles the retrieval of a user's languages
func (h *LanguageHandler) GetLanguages(c *gin.Context) {
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

	if len(seeker.Languages) == 0 {
		c.JSON(http.StatusNoContent, gin.H{"message": "No languages found"})
		return
	}

	languages, err := repository.GetLanguages(&seeker)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing languages"})
		log.Printf("Error processing languages for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	// Map to LanguageResponse
	var response []dto.LanguageResponse
	for _, lang := range languages {
		response = append(response, dto.LanguageResponse{
			AuthUserID:       userID,
			LanguageName:     lang["language"].(string),
			ProficiencyLevel: lang["proficiency"].(string),
			CreatedAt:        lang["created_at"].(primitive.DateTime).Time(),
			UpdatedAt:        lang["updated_at"].(primitive.DateTime).Time(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"languages": response,
	})
}

// UpdateLanguage handles the update of a language entry (without file upload)
func (h *LanguageHandler) UpdateLanguage(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	id := c.Param("id")
	index, err := strconv.Atoi(id)
	if err != nil || index <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid language index. Must be a positive integer."})
		return
	}

	var input dto.LanguageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		log.Printf("Error binding input: %v", err)
		return
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

	if index > len(seeker.Languages) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Language index out of range"})
		return
	}

	updatedLanguage := bson.M{
		"language":        input.LanguageName,
		"proficiency":     input.ProficiencyLevel,
	}

	seeker.Languages[index-1] = updatedLanguage

	update := bson.M{
		"$set": bson.M{
			"languages": seeker.Languages,
		},
	}

	updateResult, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save updated language"})
		log.Printf("Failed to update language for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	if updateResult.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching seeker found to update"})
		log.Printf("No matching seeker found for auth_user_id: %s", userID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Language updated successfully",
	})
}


// DeleteLanguage handles deleting an existing language entry
func (h *LanguageHandler) DeleteLanguage(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	id := c.Param("id")

	index, err := strconv.Atoi(id)
	if err != nil || index <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid language index"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve seeker"})
		}
		return
	}

	if index > len(seeker.Languages) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Language index out of range"})
		return
	}

	// Remove the language entry at index-1
	seeker.Languages = append(seeker.Languages[:index-1], seeker.Languages[index:]...)

	update := bson.M{
		"$set": bson.M{
			"languages": seeker.Languages,
		},
	}

	if _, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete language entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Language deleted successfully"})
}
