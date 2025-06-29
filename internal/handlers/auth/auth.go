package auth

import (

	"RAAS/core/config"
	"RAAS/internal/dto"
	"RAAS/internal/models"
	"RAAS/core/security"
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

	// ✅ Final response
	c.JSON(http.StatusOK, gin.H{
		"issue": "Login successful.",
		"token": token,
		"user": gin.H{
			"email":         user.Email,
			"authUserId":    user.AuthUserID,
			"role":          user.Role,
			"emailVerified": user.EmailVerified,
			"Progress":      progress,
		},
	})
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
