package auth

import (

	"RAAS/core/config"
	"RAAS/internal/dto"
	"RAAS/internal/models"
	
	"RAAS/utils"



	"context"
	"strings"
	"fmt"
	"net/http"
	
	"time"
	// "errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)


func SeekerSignUp(c *gin.Context) {
	var input dto.SeekerSignUpInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"issue":   "Invalid input data format.",
			"error":   "invalid_input",
			"details": err.Error(),
		})
		return
	}

	db := c.MustGet("db").(*mongo.Database)
	userRepo := NewUserRepo(db)

	if err := userRepo.ValidateSeekerSignUpInput(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"issue":   "Please fill all required fields.",
			"error":   "validation_error",
			"details": err.Error(),
		})
		return
	}

	// Check if user already exists
	var existingUser models.AuthUser
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.Collection("auth_users").FindOne(ctx, bson.M{"email": input.Email}).Decode(&existingUser)
	if err == nil {
		if existingUser.EmailVerified {
			c.JSON(http.StatusConflict, gin.H{
				"issue": "Account already exists and is verified. Please login.",
				"error": "email_already_registered",
			})
			return
		}

		// Resend verification email
		verificationLink := fmt.Sprintf("%s/b1/auth/verify-email?token=%s", config.Cfg.Project.FrontendBaseUrl, existingUser.VerificationToken)
		emailBody := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; background-color: #f9f9f9; margin: 0; padding: 0;">
			<div style="max-width: 600px; margin: 40px auto; background: #ffffff; padding: 30px; border-radius: 10px; box-shadow: 0 2px 8px rgba(0,0,0,0.05);">
				<h2 style="color: #4CAF50; text-align: center;">Welcome back to JSE AI!</h2>
				<p>Hi %s,</p>
				<p>You already signed up but didn't verify your email. Please click the button below to complete verification:</p>
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"issue":   "Verification email could not be resent. Please try again later.",
				"error":   "email_resent_failed",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusConflict, gin.H{
			"issue": "Email already registered but not verified. Verification email resent.",
			"error": "email_unverified_resent",
		})
		return
	} else if err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{
			"issue":   "Error checking existing account.",
			"error":   "check_existing_user_failed",
			"details": err.Error(),
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"issue": "Could not secure your password. Please try again.",
			"error": "password_hash_error",
		})
		return
	}

	// Create user
	if err := userRepo.CreateSeeker(input, string(hashedPassword)); err != nil {
		msg := err.Error()

		switch {
		case strings.Contains(msg, "email is already taken"):
			c.JSON(http.StatusConflict, gin.H{
				"issue": "Email is already taken.",
				"error": "email_duplicate",
			})
		case strings.Contains(msg, "phone number is already taken"):
			c.JSON(http.StatusConflict, gin.H{
				"issue": "Phone number is already taken.",
				"error": "phone_duplicate",
			})
		case strings.Contains(msg, "failed to create auth user"):
			c.JSON(http.StatusInternalServerError, gin.H{
				"issue": "Account creation failed. Please try again.",
				"error": "create_auth_failed",
			})
		case strings.Contains(msg, "failed to create seeker profile"):
			c.JSON(http.StatusInternalServerError, gin.H{
				"issue": "User created but profile setup failed.",
				"error": "create_seeker_profile_failed",
			})
		case strings.Contains(msg, "failed to create entry timeline"):
			c.JSON(http.StatusInternalServerError, gin.H{
				"issue": "User created but onboarding timeline failed.",
				"error": "create_timeline_failed",
			})
		case strings.Contains(msg, "failed to send verification email"):
			c.JSON(http.StatusInternalServerError, gin.H{
				"issue": "Verification email could not be sent. Please try again later.",
				"error": "email_send_failed",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"issue":   "Something went wrong. Please try again.",
				"error":   "create_seeker_failed",
				"details": msg,
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"issue":   "User registered and verification mail sent.",
		"message": "seeker_registered_successfully",
	})
}