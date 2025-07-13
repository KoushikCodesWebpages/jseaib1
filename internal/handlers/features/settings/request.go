package settings

import (
	//"RAAS/internal/dto"
	"RAAS/internal/models"
	"RAAS/utils"
	"RAAS/core/config"

	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const adminEmail = "koushik@arshan.de"

func (h *SettingsHandler) RequestFeedbackEmail(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	authUsersCollection := db.Collection("auth_users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get user email
	var user models.AuthUser
	if err := authUsersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Could not retrieve user", "details": err.Error()})
		return
	}

	// Parse feedback input inline (no DTO)
	var input struct {
		Subject    string `json:"subject" binding:"required"`
		Body string `json:"body" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"issue": "Invalid input", "details": err.Error()})
		return
	}

	emailCfg := utils.EmailConfig{
		Host:     config.Cfg.Cloud.EmailHost,
		Port:     config.Cfg.Cloud.EmailPort,
		Username: config.Cfg.Cloud.EmailHostUser,
		Password: config.Cfg.Cloud.EmailHostPassword,
		From:     config.Cfg.Cloud.DefaultFromEmail,
		UseTLS:   config.Cfg.Cloud.EmailUseTLS,
	}


	// Validate feedback type
	if input.Subject != "bug" && input.Subject != "request" && input.Subject != "feedback" {
		c.JSON(http.StatusBadRequest, gin.H{"issue": "Invalid feedback type. Must be 'bug', 'request', or 'feedback'"})
		return
	}
	//reciever := "koushik@arshan.de"
	
	emailBody := "feedback or bug or request from user:\n\n" +
		"User ID: " + userID + "\n" +
		"Current Email: " + user.Email + "\n\n" +
		"Message:\n" + input.Body
	// Log it (or send an email, save to DB, etc.)

	log.Printf("New %s from %s (%s): %s", input.Subject, adminEmail, userID, input.Body)

	if err := utils.SendEmail(emailCfg, adminEmail, input.Subject, emailBody); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to send email", "error": err.Error()})
		return
	}

	// TODO: optionally send this to admin via email or notification

	c.JSON(http.StatusOK, gin.H{"issue": "Feedback received, thank you!"})
}

func (h *SettingsHandler) SendEmailChangeRequest(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	userEmail := c.MustGet("email").(string)

	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"issue": "Invalid input", "error": err.Error()})
		return
	}

	emailCfg := utils.EmailConfig{
		Host:     config.Cfg.Cloud.EmailHost,
		Port:     config.Cfg.Cloud.EmailPort,
		Username: config.Cfg.Cloud.EmailHostUser,
		Password: config.Cfg.Cloud.EmailHostPassword,
		From:     config.Cfg.Cloud.DefaultFromEmail,
		UseTLS:   config.Cfg.Cloud.EmailUseTLS,
	}

	subject := "Request for Email Change"
	//adminEmail := "koushik@arshan.de" // 📬 send to admin!

	emailBody := "Email change request from user:\n\n" +
		"User ID: " + userID + "\n" +
		"Current Email: " + userEmail + "\n\n" +
		"Message:\n" + req.Body

	log.Printf("📬 Forwarding email change request for user %s (%s) to admin", userID, userEmail)

	if err := utils.SendEmail(emailCfg, adminEmail, subject, emailBody); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to send email", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"issue": "Email change request sent to support"})
}


func (h *SettingsHandler) SendJobTitleChangeRequest(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	userEmail := c.MustGet("email").(string)

	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"issue": "Invalid input", "error": err.Error()})
		return
	}

	emailCfg := utils.EmailConfig{
		Host:     config.Cfg.Cloud.EmailHost,
		Port:     config.Cfg.Cloud.EmailPort,
		Username: config.Cfg.Cloud.EmailHostUser,
		Password: config.Cfg.Cloud.EmailHostPassword,
		From:     config.Cfg.Cloud.DefaultFromEmail,
		UseTLS:   config.Cfg.Cloud.EmailUseTLS,
	}

	subject := "Request for Job Title Change"
	//adminEmail := "koushik@arshan.de" // 📬 send to admin!

	emailBody := "Job Title change request from user:\n\n" +
		"User ID: " + userID + "\n" +
		"Current Email: " + userEmail + "\n\n" +
		"Message:\n" + req.Body

	log.Printf("📬 Forwarding job title change request for user %s (%s) to admin", userID, userEmail)

	if err := utils.SendEmail(emailCfg, adminEmail, subject, emailBody); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to send email", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"issue": "Job title change request sent to support"})
}


