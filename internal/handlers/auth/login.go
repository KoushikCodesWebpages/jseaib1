package auth

import (
	"RAAS/internal/dto"
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
        c.JSON(http.StatusForbidden, gin.H{"issue": "This account was deleted. Please contact help@arshan.digital for further help", "error": "user_deleted"})
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

    // 6️⃣ Generate JWT
    token, err := security.GenerateJWT(user.AuthUserID, user.Email, "seeker")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"issue": "Login successful but token generation failed.", "error": "jwt_token_error", "details": err.Error()})
        return
    }

    completed, next_step, err := repository.UpdateTimelineStepAndCheckCompletion(ctx, db, user.AuthUserID, "")
    if err != nil {
        log.Printf("Timeline update error [SetKeySkills] user=%s: %v", user.AuthUserID, err)
    }

    if completed{
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
            "progress_completed":     completed,
            "next_step": next_step,
            
        },
    })
}
