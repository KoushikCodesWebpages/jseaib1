package auth

import (

	"RAAS/core/config"
	"RAAS/internal/dto"
	"RAAS/internal/models"
	"RAAS/core/security"
	"RAAS/utils"


	"context"
	"fmt"
	"net/http"
	"time"
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func SeekerSignUp(c *gin.Context) {
	var input dto.SeekerSignUpInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "details": err.Error()})
		return
	}

	db := c.MustGet("db").(*mongo.Database)
	userRepo := NewUserRepo(db)

	if err := userRepo.ValidateSeekerSignUpInput(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": err.Error()})
		return
	}

	// Check for duplicate email or phone
	emailTaken, phoneTaken, err := userRepo.CheckDuplicateEmailOrPhone(input.Email, input.Number)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "check_duplicate_failed", "details": err.Error()})
		return
	}

	if emailTaken || phoneTaken {
		// Try to fetch the existing user to decide how to respond
		var user models.AuthUser
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := db.Collection("auth_users").FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
		if err == nil {
			if user.EmailVerified {
				c.JSON(http.StatusBadRequest, gin.H{"error": "email_already_verified"})
				return
			}

			// Resend verification email
			verificationLink := fmt.Sprintf("%s/b1/auth/verify-email?token=%s", config.Cfg.Project.FrontendBaseUrl, user.VerificationToken)
			emailBody := fmt.Sprintf(`
			<html>
			<body style="font-family: Arial, sans-serif; background-color: #f9f9f9; margin: 0; padding: 0;">
				<div style="max-width: 600px; margin: 40px auto; background: #ffffff; padding: 30px; border-radius: 10px; box-shadow: 0 2px 8px rgba(0,0,0,0.05);">
				<h2 style="color: #4CAF50; text-align: center;">Welcome to JSE AI!</h2>
				<p>Hi %s,</p>
				<p>Thanks for signing up! To get started, please confirm your email address by clicking the button below:</p>
				<div style="text-align: center; margin: 30px 0;">
					<a href="%s" style="background-color: #4CAF50; color: #ffffff; padding: 14px 24px; text-decoration: none; border-radius: 6px; font-weight: bold;">
					Verify Email
					</a>
				</div>
				<p>If you didn’t create this account, you can safely ignore this email.</p>
				<p>Cheers,<br><strong>The Team</strong></p>
				</div>
			</body>
			</html>
			`, input.Email, verificationLink)

			emailCfg := utils.EmailConfig{
				Host:     config.Cfg.Cloud.EmailHost,
				Port:     config.Cfg.Cloud.EmailPort,
				Username: config.Cfg.Cloud.EmailHostUser,
				Password: config.Cfg.Cloud.EmailHostPassword,
				From:     config.Cfg.Cloud.DefaultFromEmail,
				UseTLS:   config.Cfg.Cloud.EmailUseTLS,
			}

			if err := utils.SendEmail(emailCfg, input.Email, "Verify your email", emailBody); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_send_verification_email", "details": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "verification_email_resent"})
			return
		}

		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusConflict, gin.H{"error": "email_taken_but_user_not_found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "details": err.Error()})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password_hash_error"})
		return
	}

	// Create the seeker account
	if err := userRepo.CreateSeeker(input, string(hashedPassword)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create_seeker_failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "seeker_registered_successfully"})
}


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
				<a href="https://arshan.digital" target="_blank" rel="noopener noreferrer">Go to Login</a>

			</div>
		</body>
		</html>
	`))
}

func Login(c *gin.Context) {
	var input dto.LoginInput

	db := c.MustGet("db").(*mongo.Database)
	userRepo := NewUserRepo(db)

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "details": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := userRepo.AuthenticateUser(ctx, input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := security.GenerateJWT(user.AuthUserID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_generation_failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func AdminRefreshToken(c *gin.Context) {
	clientIP := c.ClientIP()
	if !isIPAllowed(clientIP) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access_denied"})
		return
	}

	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "details": err.Error()})
		return
	}

	db := c.MustGet("db").(*mongo.Database)
	userRepo := NewUserRepo(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := userRepo.FindUserByEmail(ctx, input.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if !user.EmailVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "email not verified"})
		return
	}

	token, err := security.GenerateJWT(user.AuthUserID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_generation_failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func isIPAllowed(ip string) bool {
	allowedIPs := []string{
		"106.222.223.57",
		"106.222.220.147",
		"168.231.109.75",
		"::1",
		"2401:4900:1f2d:5646:e544:6315:f174:47e",
		"2401:4900:1f2d:166e:6430:1a60:de37:48bd",
		"2a02:4780:41:4a1b::1",
	}

	fmt.Printf("Client IP: %s\n", ip)
	fmt.Printf("Allowed IPs: %v\n", allowedIPs)

	for _, allowedIP := range allowedIPs {
		if ip == allowedIP {
			fmt.Println("IP matched: access granted")
			return true
		}
	}

	fmt.Println("IP not allowed: access denied")
	return false
}

