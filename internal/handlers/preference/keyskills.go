package preference

import (
    "context"
    "log"
    "net/http"
    "strings"
    "time"

    "RAAS/internal/dto"
    "RAAS/internal/models"
    "RAAS/internal/handlers/repository"
    "RAAS/internal/handlers/features/jobs"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type KeySkillsHandler struct{}

func NewKeySkillsHandler() *KeySkillsHandler {
    return &KeySkillsHandler{}
}

// SetKeySkills inserts or updates key skills for the authenticated user
func (h *KeySkillsHandler) SetKeySkills(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var input dto.KeySkillsRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"issue": "Please provide your skills in the correct format.",
		})
		return
	}

	// Clean and deduplicate skills
	skillSet := make(map[string]struct{})
	for _, s := range input.Skills {
		s = strings.TrimSpace(strings.ReplaceAll(s, "\n", ""))
		if s != "" {
			skillSet[s] = struct{}{}
		}
	}
	if len(skillSet) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no valid skills provided",
			"issue": "Please include at least one valid skill.",
		})
		return
	}

	cleaned := make([]string, 0, len(skillSet))
	for s := range skillSet {
		cleaned = append(cleaned, s)
	}

	seekers := db.Collection("seekers")
	var existing struct {
		KeySkills []string `bson:"key_skills"`
	}
	err := seekers.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&existing)
	if err != nil {
		code := http.StatusInternalServerError
		issue := "Unable to find your profile. Try again later."
		if err == mongo.ErrNoDocuments {
			code = http.StatusNotFound
			issue = "User not found. Please log in again."
		}
		c.JSON(code, gin.H{"error": err.Error(), "issue": issue})
		return
	}

	op := "updated"
	if len(existing.KeySkills) == 0 {
		op = "added"
	}

	_, err = seekers.UpdateOne(ctx,
		bson.M{"auth_user_id": userID},
		bson.M{
			"$set": bson.M{
				"key_skills": cleaned,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Couldn't save your skills. Please try again.",
		})
		return
	}

	// Update timeline step
	completed, _, err := repository.UpdateTimelineStepAndCheckCompletion(ctx, db, userID, "key_skills_completed")
	if err != nil {
		log.Printf("❌ Timeline update error [SetKeySkills] user=%s: %v", userID, err)
	}

	// Trigger job matching if completed
	if completed {
		if err := jobs.StartJobMatchScoreCalculation(c, db, userID); err != nil {
			log.Printf("❌ Job match process error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to start job match process",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"issue": "Key skills " + op + " successfully",
	})
}

// GetKeySkills returns existing key skills
func (h *KeySkillsHandler) GetKeySkills(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    db := c.MustGet("db").(*mongo.Database)
    seekers := db.Collection("seekers")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var seeker models.Seeker
    if err := seekers.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
        log.Printf("DB fetch error [GetKeySkills] user=%s: %v", userID, err)
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{
                "error": err.Error(),
                "issue": "User not found. Please log in again.",
            })
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": err.Error(),
                "issue": "Unable to fetch your skills. Please try again later.",
            })
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
