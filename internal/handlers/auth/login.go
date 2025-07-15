package auth

import (
	"RAAS/internal/dto"
	"RAAS/internal/models"
	"RAAS/core/security"
	"RAAS/internal/handlers/features/jobs"
	"RAAS/internal/handlers/repository"

	"context"
	"strings"
	"net/http"
	"log"
	"time"
	// "errors"
	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)



func SeekerLogin(c *gin.Context) {
    // 1️⃣ Bind input
    var input dto.LoginInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "issue":   "Invalid input format.",
            "error":   "invalid_input",
            "details": err.Error(),
        })
        return
    }

    // 2️⃣ Setup DB & repos
    db := c.MustGet("db").(*mongo.Database)
    seekersColl := db.Collection("seekers")
    authColl := db.Collection("auth_users")
    userRepo := NewUserRepo(db)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // 3️⃣ Authenticate user
    user, err := userRepo.AuthenticateUser(ctx, input.Email, input.Password)
if err != nil {
    msg := err.Error()
    switch {
    case strings.Contains(msg, "user_not_found"):
        c.JSON(http.StatusUnauthorized, gin.H{"issue": "Account with this email doesn't exist.", "error": "user_not_found"})
    case strings.Contains(msg, "user_deleted"):
        c.JSON(http.StatusForbidden, gin.H{"issue": "This account was deleted. Contact support if you think this is a mistake.", "error": "user_deleted"})
    case strings.Contains(msg, "email_not_verified"):
        c.JSON(http.StatusUnauthorized, gin.H{"issue": "Please verify your email before logging in.", "error": "email_unverified"})
    case strings.Contains(msg, "invalid_password"):
        c.JSON(http.StatusUnauthorized, gin.H{"issue": "Incorrect password.", "error": "wrong_password"})
    case strings.Contains(msg, "db_error"):
        c.JSON(http.StatusInternalServerError, gin.H{"issue": "Database error. Please try again.", "error": "db_error", "details": msg})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"issue": "Unexpected error during login.", "error": "login_failed", "details": msg})
    }
    return
}


    // 4️⃣ Capture login metadata
    // clientIP := c.ClientIP()
    nowLogin := time.Now()
    if _, err := authColl.UpdateOne(ctx,
        bson.M{"auth_user_id": user.AuthUserID},
        bson.M{"$set": bson.M{"last_login_at": nowLogin}},
    ); err != nil {
        log.Printf("⚠️ Failed to update last login metadata for %s: %v", user.AuthUserID, err)
    }

    // 5️⃣ Check entry timeline
    var timeline models.UserEntryTimeline
    progress := false
    if err := db.Collection("entry_progress_timelines").FindOne(ctx, bson.M{"auth_user_id": user.AuthUserID}).Decode(&timeline); err == nil && timeline.Completed {
        progress = true
    }

    // 6️⃣ Generate JWT
    token, err := security.GenerateJWT(user.AuthUserID, user.Email, "seeker")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"issue": "Login successful but token generation failed.", "error": "jwt_token_error", "details": err.Error()})
        return
    }

    // 7️⃣ Load seeker profile
    var seeker models.Seeker
    if err := seekersColl.FindOne(ctx, bson.M{"auth_user_id": user.AuthUserID}).Decode(&seeker); err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"issue": "Seeker not found"})
            log.Printf("Seeker not found for auth_user_id: %s", user.AuthUserID)
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to retrieve seeker"})
            log.Printf("Failed to retrieve seeker for auth_user_id: %s, Error: %v", user.AuthUserID, err)
        }
        return
    }

    // 8️⃣ Trigger job match score if profile complete
    matchscore := false
    completion, missing := repository.CalculateJobProfileCompletion(seeker)
    if completion == 100 || len(missing) == 0 {
        matchscore = true
        if err = jobs.StartJobMatchScoreCalculation(c, db, user.AuthUserID); err != nil {
            log.Printf("Error starting job match process: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start job match process"})
            return
        }
    }

    // 📤 Final response
    c.JSON(http.StatusOK, gin.H{
        "issue": "Login successful.",
        "token": token,
        "user": gin.H{
            "email":          user.Email,
            "auth_user_id":   user.AuthUserID,
            "role":           user.Role,
            "email_verified": user.EmailVerified,
            "progress":       progress,
            "match_score":    matchscore,
        },
    })
}
