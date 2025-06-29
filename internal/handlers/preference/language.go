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
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/bson/primitive"
)

type LanguageHandler struct{}

func NewLanguageHandler() *LanguageHandler {
	return &LanguageHandler{}
}

func (h *LanguageHandler) CreateLanguage(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    db := c.MustGet("db").(*mongo.Database)
    seekers := db.Collection("seekers")
    timelines := db.Collection("user_entry_timelines")

    var input dto.LanguageRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        log.Printf("Bind error [CreateLanguage] user=%s: %v", userID, err)
        c.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
            "issue": "Some required fields are missing or invalid.",
        })
        return
    }

    validLevels := map[string]bool{
        "beginner":     true,
        "intermediate": true,
        "fluent":       true,
        "native":       true,
    }
    if !validLevels[input.ProficiencyLevel] {
        msg := fmt.Sprintf("invalid proficiency level: %s", input.ProficiencyLevel)
        log.Printf("Validation error [CreateLanguage] user=%s: %s", userID, msg)
        c.JSON(http.StatusBadRequest, gin.H{
            "error": msg,
            "issue": "Proficiency level must be one of: beginner, intermediate, fluent, native.",
        })
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var seeker models.Seeker
    if err := seekers.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
        log.Printf("DB fetch error [CreateLanguage] user=%s: %v", userID, err)
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{
                "error": err.Error(),
                "issue": "No account found. It might have been removed or reset.",
            })
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": err.Error(),
                "issue": "Could not retrieve your profile. Please try again.",
            })
        }
        return
    }

    if err := repository.AppendToLanguages(&seeker, input, ""); err != nil {
        log.Printf("Append error [CreateLanguage] user=%s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": err.Error(),
            "issue": "Could not save your language details. Try again shortly.",
        })
        return
    }

    updateResult, err := seekers.UpdateOne(ctx, bson.M{"auth_user_id": userID}, bson.M{"$set": bson.M{"languages": seeker.Languages}})
    if err != nil {
        log.Printf("DB update error [CreateLanguage] user=%s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": err.Error(),
            "issue": "Failed to update your language records. Please retry.",
        })
        return
    }
    if updateResult.MatchedCount == 0 {
        log.Printf("No document updated [CreateLanguage] user=%s", userID)
        c.JSON(http.StatusNotFound, gin.H{
            "error": "no seeker updated for " + userID,
            "issue": "Your account couldn't be updated. Please refresh and try again.",
        })
        return
    }

    if _, err := timelines.UpdateOne(ctx, bson.M{"auth_user_id": userID}, bson.M{"$set": bson.M{"languages_completed": true}}); err != nil {
        log.Printf("Timeline update error [CreateLanguage] user=%s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": err.Error(),
            "issue": "Language added, but progress tracking failed.",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{"issue": "Language added successfully"})
}

// GetLanguages handles the retrieval of a user's languages
func (h *LanguageHandler) GetLanguages(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    db := c.MustGet("db").(*mongo.Database)
    seekers := db.Collection("seekers")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var seeker models.Seeker
    if err := seekers.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
        log.Printf("DB fetch error [GetLanguages] user=%s: %v", userID, err)
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{
                "error": err.Error(),
                "issue": "Account not found. Please refresh or contact support.",
            })
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": err.Error(),
                "issue": "Failed to retrieve your profile. Please try again.",
            })
        }
        return
    }

    if len(seeker.Languages) == 0 {
        c.JSON(http.StatusNoContent, gin.H{"message": "No languages found"})
        return
    }

    languages, err := repository.GetLanguages(&seeker)
    if err != nil {
        log.Printf("Processing error [GetLanguages] user=%s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": err.Error(),
            "issue": "We couldn't load your language list. Please try again.",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{"languages": languages})
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
