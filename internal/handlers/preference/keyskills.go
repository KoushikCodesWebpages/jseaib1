package preference

import (
    "context"
    "log"
    "net/http"
    "strings"
    "time"

    "RAAS/internal/dto"
    "RAAS/internal/models"

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
    seekers := db.Collection("seekers")
    timelines := db.Collection("user_entry_timelines")

    var input dto.KeySkillsRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        log.Printf("Bind error [SetKeySkills] user=%s: %v", userID, err)
        c.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
            "issue": "Please provide your skills in the correct format.",
        })
        return
    }

    // Clean skills list
    cleaned := make([]string, 0, len(input.Skills))
    for _, s := range input.Skills {
        t := strings.TrimSpace(strings.ReplaceAll(s, "\n", ""))
        if t != "" {
            cleaned = append(cleaned, t)
        }
    }
    if len(cleaned) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "no valid skills provided",
            "issue": "Please include at least one valid skill.",
        })
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Fetch seeker
    var seeker models.Seeker
    if err := seekers.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
        log.Printf("DB fetch error [SetKeySkills] user=%s: %v", userID, err)
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{
                "error": err.Error(),
                "issue": "User not found. Please log in again.",
            })
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": err.Error(),
                "issue": "Unable to find your profile. Try again later.",
            })
        }
        return
    }

    op := "updated"
    if len(seeker.KeySkills) == 0 {
        op = "added"
    }

    // Update DB
    update := bson.M{"$set": bson.M{"key_skills": cleaned, "updated_at": time.Now()}}
    res, err := seekers.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update)
    if err != nil {
        log.Printf("DB update error [SetKeySkills] user=%s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": err.Error(),
            "issue": "Couldn't save your skills. Please try again.",
        })
        return
    }
    if res.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "no seeker matched",
            "issue": "Your account was not found. Please refresh and try again.",
        })
        return
    }

    // Mark timeline
    if _, err := timelines.UpdateOne(ctx, bson.M{"auth_user_id": userID}, bson.M{"$set": bson.M{"key_skills_completed": true}}); err != nil {
        log.Printf("Timeline update warning [SetKeySkills] user=%s: %v", userID, err)
    }

    c.JSON(http.StatusOK, gin.H{"issue": "Key skills " + op + " successfully"})
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
