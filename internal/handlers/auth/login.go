package auth

import (
	"RAAS/internal/dto"
	"RAAS/internal/models"
	"RAAS/core/security"
	"RAAS/internal/handlers/features/jobs"
	"RAAS/internal/handlers/repository"

	"context"
	"strings"
	"fmt"
	"net/http"
	"log"
	"time"
	// "errors"
	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)



func SeekerLogin(c *gin.Context) {
	var input dto.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"issue":   "Invalid input format.",
			"error":   "invalid_input",
			"details": err.Error(),
		})
		return
	}

	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")
	userRepo := NewUserRepo(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := userRepo.AuthenticateUser(ctx, input.Email, input.Password)
	if err != nil {
		msg := err.Error()

		switch {
		case strings.Contains(msg, "user_not_found"):
			c.JSON(http.StatusUnauthorized, gin.H{
				"issue": "Account with this email doesn't exist.",
				"error": "user_not_found",
			})
		case strings.Contains(msg, "email_not_verified"):
			c.JSON(http.StatusUnauthorized, gin.H{
				"issue": "Please verify your email before logging in.",
				"error": "email_unverified",
			})
		case strings.Contains(msg, "invalid_password"):
			c.JSON(http.StatusUnauthorized, gin.H{
				"issue": "Incorrect password.",
				"error": "wrong_password",
			})
		case strings.Contains(msg, "db_error"):
			c.JSON(http.StatusInternalServerError, gin.H{
				"issue":   "Database error. Please try again.",
				"error":   "db_error",
				"details": msg,
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"issue":   "Unexpected error during login.",
				"error":   "login_failed",
				"details": msg,
			})
		}
		return
	}

	// ✅ Check entry timeline for completion
	var timeline models.UserEntryTimeline
	err = db.Collection("entry_progress_timelines").FindOne(ctx, bson.M{
		"auth_user_id": user.AuthUserID,
	}).Decode(&timeline)

	progress := false
	if err == nil && timeline.Completed {
		progress = true
	}

	// Generate JWT
	token, err := security.GenerateJWT(user.AuthUserID, user.Email, "seeker")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"issue":   "Login successful but token generation failed.",
			"error":   "jwt_token_error",
			"details": err.Error(),
		})
		return
	}
		var seeker models.Seeker
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": user.AuthUserID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"issue": "Seeker not found"})
			log.Printf("Seeker not found for auth_user_id: %s", user.AuthUserID)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to retrieve seeker"})
			log.Printf("Failed to retrieve seeker for auth_user_id: %s, Error: %v", user.AuthUserID, err)
		}
		return
	}
	matchscore := false
	completion, missing := repository.CalculateJobProfileCompletion(seeker)
		if completion == 100 || len(missing) == 0 {
			matchscore=true

			fmt.Println("starting match score calculation")
			// ✅ Trigger job match score calculation
			err = jobs.StartJobMatchScoreCalculation(c, db, user.AuthUserID)
			if err != nil {
				fmt.Println("Error starting job match score calculation:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start job match process"})
				return
		}
	}



	// ✅ Final response
	c.JSON(http.StatusOK, gin.H{
		"issue": "Login successful.",
		"token": token,
		"user": gin.H{
			"email":         	user.Email,
			"auth_user_id":    	user.AuthUserID,
			"role":          	user.Role,
			"email_verified": 	user.EmailVerified,
			"progress":      	progress,
			"match_score": 		matchscore,
		},
	})
}