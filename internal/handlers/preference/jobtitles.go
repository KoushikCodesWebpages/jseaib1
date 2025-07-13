package preference

import (
    "context"
    "log"
    "net/http"
    "time"

    "RAAS/internal/dto"
    "RAAS/internal/models"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type JobTitleHandler struct{}

func NewJobTitleHandler() *JobTitleHandler {
    return &JobTitleHandler{}
}

// CreateJobTitleOnce allows setting the job titles only once per user
func (h *JobTitleHandler) CreateJobTitleOnce(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    db := c.MustGet("db").(*mongo.Database)
    seekers := db.Collection("seekers")
    timelines := db.Collection("user_entry_timelines")

    var input dto.JobTitleInput
    if err := c.ShouldBind(&input); err != nil {
        log.Printf("Bind error [CreateJobTitleOnce] user=%s: %v", userID, err)
        c.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
            "issue": "Please provide valid job title information.",
        })
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var seeker models.Seeker
    if err := seekers.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
        log.Printf("DB fetch error [CreateJobTitleOnce] user=%s: %v", userID, err)
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{
                "error": err.Error(),
                "issue": "User account not found. Please login again.",
            })
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": err.Error(),
                "issue": "Failed to retrieve your profile. Try again later.",
            })
        }
        return
    }

    if seeker.PrimaryTitle != "" {
        log.Printf("Attempt to reset titles [CreateJobTitleOnce] user=%s", userID)
        c.JSON(http.StatusForbidden, gin.H{
            "error": "titles_already_set",
            "issue": "Job titles cannot be changed once set.",
        })
        return
    }

    update := bson.M{"$set": bson.M{
        "primary_title":    input.PrimaryTitle,
        "secondary_title":  input.SecondaryTitle,
        "tertiary_title":   input.TertiaryTitle,
    }}

    res, err := seekers.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update)
    if err != nil {
        log.Printf("DB update error [CreateJobTitleOnce] user=%s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": err.Error(),
            "issue": "Failed to save your job titles. Please try again.",
        })
        return
    }
    if res.MatchedCount == 0 {
        log.Printf("No seeker updated [CreateJobTitleOnce] user=%s", userID)
        c.JSON(http.StatusNotFound, gin.H{
            "error": "no seeker matched",
            "issue": "Your account was not found. Please refresh and try again.",
        })
        return
    }

    if _, err := timelines.UpdateOne(ctx, bson.M{"auth_user_id": userID}, bson.M{"$set": bson.M{"job_titles_completed": true}}); err != nil {
        log.Printf("Timeline update error [CreateJobTitleOnce] user=%s: %v", userID, err)
    }

    c.JSON(http.StatusOK, gin.H{"issue": "Job titles set successfully"})
}

// GetJobTitle retrieves the user's job titles
func (h *JobTitleHandler) GetJobTitle(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    db := c.MustGet("db").(*mongo.Database)
    seekers := db.Collection("seekers")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var seeker models.Seeker
    if err := seekers.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
        log.Printf("DB fetch error [GetJobTitle] user=%s: %v", userID, err)
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{
                "error": err.Error(),
                "issue": "User account not found. Please login.",
            })
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": err.Error(),
                "issue": "Failed to retrieve your job titles. Try again later.",
            })
        }
        return
    }

    if seeker.PrimaryTitle == "" {
        c.JSON(http.StatusNoContent, gin.H{})
        return
    }

    resp := dto.JobTitleResponse{
        AuthUserID:     seeker.AuthUserID,
        PrimaryTitle:   seeker.PrimaryTitle,
        SecondaryTitle: seeker.SecondaryTitle,
        TertiaryTitle:  seeker.TertiaryTitle,
    }
    c.JSON(http.StatusOK, resp)
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
