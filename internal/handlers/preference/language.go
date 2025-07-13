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

func (h *LanguageHandler) UpdateLanguage(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	indexStr := c.Param("id")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_index",
			"issue": "Invalid language index",
		})
		return
	}

	var input dto.LanguageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"issue": "Invalid input format",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker struct {
		Languages []bson.M `bson:"languages"`
	}
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		status := http.StatusInternalServerError
		issue := "Failed to retrieve seeker"
		if err == mongo.ErrNoDocuments {
			status = http.StatusNotFound
			issue = "Seeker not found"
		}
		c.JSON(status, gin.H{
			"error": err.Error(),
			"issue": issue,
		})
		return
	}

	if index > len(seeker.Languages) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "index_out_of_range",
			"issue": "Language index is out of range",
		})
		return
	}

	seeker.Languages[index-1] = bson.M{
		"language":    input.LanguageName,
		"proficiency": input.ProficiencyLevel,
		"updated_at":  time.Now(),
	}

	update := bson.M{
		"$set": bson.M{
			"languages": seeker.Languages,
		},
		"$currentDate": bson.M{"updated_at": true},
	}

	if _, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Failed to update language",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issue": "Language updated successfully",
	})
}


func (h *LanguageHandler) DeleteLanguage(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	indexStr := c.Param("id")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_index",
			"issue": "Invalid language index",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker struct {
		Languages []bson.M `bson:"languages"`
	}
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		status := http.StatusInternalServerError
		issue := "Failed to retrieve seeker"
		if err == mongo.ErrNoDocuments {
			status = http.StatusNotFound
			issue = "Seeker not found"
		}
		c.JSON(status, gin.H{
			"error": err.Error(),
			"issue": issue,
		})
		return
	}

	if index > len(seeker.Languages) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "index_out_of_range",
			"issue": "Language index is out of range",
		})
		return
	}

	seeker.Languages = append(seeker.Languages[:index-1], seeker.Languages[index:]...)

	update := bson.M{
		"$set": bson.M{
			"languages": seeker.Languages,
		},
		"$currentDate": bson.M{"updated_at": true},
	}

	if _, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Failed to delete language",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issue": "Language deleted successfully",
	})
}
