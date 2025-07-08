package auth

import (
	"RAAS/internal/models"
	"RAAS/core/security"


	"context"
	"log"

	"net/http"
	"time"
	// "errors"
	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)


func VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.String(http.StatusBadRequest, "Missing token")
		return
	}

	db := c.MustGet("db").(*mongo.Database)

	var user models.AuthUser
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.Collection("auth_users").FindOne(ctx, bson.M{"verification_token": token}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		c.String(http.StatusNotFound, "Invalid or expired token")
		return
	} else if err != nil {
		c.String(http.StatusInternalServerError, "Database error")
		return
	}

	if user.EmailVerified {
		c.String(http.StatusOK, "Email already verified.")
		return
	}

	_, err = db.Collection("auth_users").UpdateOne(
		ctx,
		bson.M{"auth_user_id": user.AuthUserID},
		bson.M{"$set": bson.M{"email_verified": true, "verification_token": ""}},
	)

	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to verify email")
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>Email Verified</title>
			<style>
				body { font-family: Arial, sans-serif; background-color: #f2f4f8; color: #333; text-align: center; padding-top: 100px; }
				.card { background: white; padding: 40px; margin: auto; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); width: 90%; max-width: 500px; }
				h1 { color: #28a745; }
				p { margin-top: 10px; font-size: 18px; }
				a { display: inline-block; margin-top: 20px; text-decoration: none; color: white; background-color: #007bff; padding: 10px 20px; border-radius: 5px; }
			</style>
		</head>
		<body>
			<div class="card">
				<h1>✅ Email Verified</h1>
				<p>Your email has been successfully verified.</p>
				<a href="https://dev.arshan.digital" target="_blank" rel="noopener noreferrer">Go to Login</a>

			</div>
		</body>
		</html>
	`))
}

func AdminRefreshToken(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"issue":   "Invalid input data. Email and password are required.",
			"error":   "invalid_input",
			"details": err.Error(),
		})
		return
	}

	if input.Password != "admin@123" {
		c.JSON(http.StatusForbidden, gin.H{
			"issue": "Admin password incorrect.",
			"error": "access_denied",
		})
		return
	}

	db := c.MustGet("db").(*mongo.Database)
	userRepo := NewUserRepo(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := userRepo.FindUserByEmail(ctx, input.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"issue": "User with the provided email not found.",
			"error": "user_not_found",
		})
		return
	}

	if !user.EmailVerified {
		c.JSON(http.StatusForbidden, gin.H{
			"issue": "Email not verified. Cannot generate token.",
			"error": "email_not_verified",
		})
		return
	}

	token, err := security.GenerateJWT(user.AuthUserID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"issue": "Failed to generate token. Please try again.",
			"error": "token_generation_failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}



func RequestPasswordResetHandler(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"issue":   "Invalid input. Please provide a valid email address.",
			"error":   "invalid_input",
			"details": err.Error(),
		})
		return
	}

	db := c.MustGet("db").(*mongo.Database)
	userRepo := NewUserRepo(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := userRepo.RequestPasswordReset(ctx, input.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"issue": "Failed to send reset email. Please try again.",
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issue": "Password reset link has been sent to your email.",
	})
}

func ResetPasswordHandler(c *gin.Context) {
	var input struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password"`
	}

	log.Println("🔐 [ResetPassword] Incoming reset request")

	// Parse JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Println("❌ [ResetPassword] Input bind error:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"issue":   "Invalid input. Please fill all fields correctly.",
			"error":   "invalid_input",
			"details": err.Error(),
		})
		return
	}

	log.Println("✅ [ResetPassword] Token received:", input.Token)
	log.Println("✅ [ResetPassword] NewPassword length:", len(input.NewPassword))

	// MongoDB connection
	db := c.MustGet("db").(*mongo.Database)
	userRepo := NewUserRepo(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call core logic
	err := userRepo.ResetPassword(ctx, input.Token, input.NewPassword)
	if err != nil {
		log.Println("❌ [ResetPassword] Reset failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"issue": "Invalid or expired token. Please request a new reset link.",
			"error": err.Error(),
		})
		return
	}

	log.Println("✅ [ResetPassword] Password reset success")

	c.JSON(http.StatusOK, gin.H{
		"issue": "Password reset successfully. You can now log in with your new password.",
	})
}
