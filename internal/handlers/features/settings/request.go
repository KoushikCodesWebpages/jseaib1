package settings

import (
	//"RAAS/internal/dto"
	"RAAS/internal/models"
	"RAAS/internal/handlers/repository"
	"RAAS/utils"
	"RAAS/core/config"

	"context"
	"log"
	"net/http"
	"time"
	"fmt"
	"html"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const adminEmail = "help@arshan.digital"

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
	seeker, err := repository.GetSeekerData(db, userID)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "seeker_not_found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

	firstname :=seeker.PersonalInfo["first_name"]
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
	
	emailBody := fmt.Sprintf(`
	<html>
	<body style="font-family: Arial, sans-serif; background-color: #f9f9f9; padding: 20px;">
		<div style="max-width: 600px; margin:auto; background: #fff; padding: 20px; border-radius: 8px;">
		<h2 style="color: #333; text-align:center;">🛠️ User Feedback / Bug / Feature Request</h2>
		
		<p><strong>Name:</strong> %s</p>
		<p><strong>User Email:</strong> %s</p>
		<p><strong>User ID:</strong> %s</p>
		<hr>

		<h3 style="color: #4CAF50;">Subscription Details</h3>
		<p><strong>Tier:</strong> %s</p>
		<p><strong>Period:</strong> %s</p>
		<p><strong>Interval Start:</strong> %s</p>
		<p><strong>Interval End:</strong> %s</p>
		<p><strong>Stripe Customer ID:</strong> %s</p>
		<hr>

		<h3 style="color: #4CAF50;">Message from User:</h3>
		<p>%s</p>
		<hr>

		<p style="font-size:0.9em; color:#666;">
			📅 <strong>Sent at:</strong> %s
		</p>
		<hr>

		<p style="font-size:0.8em; color:#999;">This notification is auto-generated. Reply formatting preserved from user input.</p>
		</div>
	</body>
	</html>
	`,
		firstname,
		user.Email,
		userID,
		seeker.SubscriptionTier,
		seeker.SubscriptionPeriod,
		seeker.SubscriptionIntervalStart.Format("2006-01-02"),
		seeker.SubscriptionIntervalEnd.Format("2006-01-02"),
		seeker.StripeCustomerID,
		input.Body,
		time.Now().Format("2006-01-02 15:04 MST"),
	)


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
    currentEmail := c.MustGet("email").(string)

    var req struct {
        Body string `json:"body" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"issue": "Invalid input", "error": err.Error()})
        return
    }

    // 1️⃣ Fetch subscription data
    db := c.MustGet("db").(*mongo.Database)
	seeker, err := repository.GetSeekerData(db, userID)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "seeker_not_found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }
	firstname :=seeker.PersonalInfo["first_name"]
    // 2️⃣ Compose email with HTML formatting
    emailBody := fmt.Sprintf(`
    <html>
      <body style="font-family: Arial, sans-serif; background-color: #f9f9f9; padding: 20px;">
        <div style="max-width:600px; margin:auto; background:white; padding:20px; border-radius:8px;">
          <h2 style="color:#333; text-align:center;">🔐 Email Change Request</h2>
          <p><strong>Name:</strong> %s</p>
          <p><strong>Current Email:</strong> %s</p>
          <hr>
          <h3 style="color:#4CAF50;">Subscription Details</h3>
          <p><strong>Tier:</strong> %s</p>
          <p><strong>Period:</strong> %s</p>
          <p><strong>Interval Start:</strong> %s</p>
          <p><strong>Interval End:</strong> %s</p>
          <p><strong>Stripe Customer ID:</strong> %s</p>
          <hr>
          <h3 style="color:#4CAF50;">User Message</h3>
          <p>%s</p>
          <hr>
          <p style="font-size:0.8em; color:#999;">Sent at: %s</p>
        </div>
      </body>
    </html>`,
        firstname,
        currentEmail,
        seeker.SubscriptionTier,
        seeker.SubscriptionPeriod,
        seeker.SubscriptionIntervalStart.Format("2006-01-02"),
        seeker.SubscriptionIntervalEnd.Format("2006-01-02"),
        seeker.StripeCustomerID,
        html.EscapeString(req.Body),
        time.Now().Format("2006-01-02 15:04 MST"),
    )

    // 3️⃣ Send email to admin support
    emailCfg := utils.EmailConfig{
        Host:     config.Cfg.Cloud.EmailHost,
        Port:     config.Cfg.Cloud.EmailPort,
        Username: config.Cfg.Cloud.EmailHostUser,
        Password: config.Cfg.Cloud.EmailHostPassword,
        From:     config.Cfg.Cloud.DefaultFromEmail,
        UseTLS:   config.Cfg.Cloud.EmailUseTLS,
    }

    log.Printf("📩 Sending email-change request for user %s (%s) to admin", userID, currentEmail)
    if err := utils.SendEmail(emailCfg, adminEmail, "User Email Change Request", emailBody); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to send email", "error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"issue": "Email change request sent to support"})
}

func (h *SettingsHandler) SendJobTitleChangeRequest(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    currentEmail := c.MustGet("email").(string)

    var req struct {
        Body string `json:"body" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"issue": "Invalid input", "error": err.Error()})
        return
    }

    // 1️⃣ Fetch subscription data
    db := c.MustGet("db").(*mongo.Database)
	seeker, err := repository.GetSeekerData(db, userID)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "seeker_not_found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }
	firstname :=seeker.PersonalInfo["first_name"]


    // 2️⃣ Compose HTML email
    emailBody := fmt.Sprintf(`
    <html>
      <body style="font-family: Arial, sans-serif; background-color: #f3f3f3; padding: 20px;">
        <div style="max-width:600px; margin:auto; background:white; padding:20px; border-radius:8px;">
          <h2 style="color:#333; text-align:center;">🏢 Job Title Change Request</h2>
          <p><strong>Name:</strong> %s</p>
          <p><strong>Current Email:</strong> %s</p>
          <hr>
          <h3 style="color:#4CAF50;">Subscription Details</h3>
          <p><strong>Tier:</strong> %s</p>
          <p><strong>Period:</strong> %s</p>
          <p><strong>Interval:</strong> %s – %s</p>
          <p><strong>Stripe Customer ID:</strong> %s</p>
          <hr>
          <h3 style="color:#4CAF50;">User Message</h3>
          <p>%s</p>
          <hr>
          <p style="font-size:0.8em; color:#999;">Sent at: %s</p>
        </div>
      </body>
    </html>`,
        firstname,
        currentEmail,
        seeker.SubscriptionTier,
        seeker.SubscriptionPeriod,
        seeker.SubscriptionIntervalStart.Format("2006-01-02"),
        seeker.SubscriptionIntervalEnd.Format("2006-01-02"),
        seeker.StripeCustomerID,
        html.EscapeString(req.Body),
        time.Now().Format("2006-01-02 15:04 MST"),
    )

    // 3️⃣ Send email
    emailCfg := utils.EmailConfig{
        Host:     config.Cfg.Cloud.EmailHost,
        Port:     config.Cfg.Cloud.EmailPort,
        Username: config.Cfg.Cloud.EmailHostUser,
        Password: config.Cfg.Cloud.EmailHostPassword,
        From:     config.Cfg.Cloud.DefaultFromEmail,
        UseTLS:   config.Cfg.Cloud.EmailUseTLS,
    }

    log.Printf("📩 Sending job title change request for user %s (%s) to admin", userID, currentEmail)
    if err := utils.SendEmail(emailCfg, adminEmail, "User Job Title Change Request", emailBody); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to send email", "error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"issue": "Job title change request sent to support"})
}


func (h *SettingsHandler) SignUpBenefitEmail(c *gin.Context) {


    var req struct {
        Name                string          `json:"name" binding:"required"`
        Email               string          `json:"email" binding:"required"`
        ExpectedProfession  string          `json:"expected_profession" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"issue": "Invalid input", "error": err.Error()})
        return
    }



    // 2️⃣ Compose HTML email
    emailBody := fmt.Sprintf(`
    <html>
    <head>
        <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;600&display=swap');
        </style>
    </head>
    <body style="margin:0; padding:0; background-color:#f4f4f4; font-family:'Inter', Arial, sans-serif;">
        <div style="max-width:600px; margin:40px auto; background-color:#ffffff; border-radius:12px; box-shadow:0 4px 12px rgba(0,0,0,0.1); overflow:hidden;">
        <div style="background:#4A90E2; color:#ffffff; padding:24px 32px; text-align:center;">
            <h1 style="margin:0; font-size:24px;">🚀 Welcome to SignUpBenefits</h1>
            <p style="margin-top:8px; font-size:16px;">Thanks for signing up!</p>
        </div>
        <div style="padding:32px;">
            <h2 style="color:#333333; font-size:20px; margin-top:0;">👤 User Details</h2>
            <p><strong>Name:</strong> %s</p>
            <p><strong>Email:</strong> %s</p>
            <hr style="border:none; border-top:1px solid #eee; margin:24px 0;">
            <h2 style="color:#333333; font-size:20px;">🎁 Sign Up Bonus</h2>
            <p><strong>Profession Interested In:</strong> %s</p>
            <p><strong>Date:</strong> %s</p>
            <p><strong>Time Zone:</strong> %s</p>
            <hr style="border:none; border-top:1px solid #eee; margin:24px 0;">
            <p style="font-size:12px; color:#888888; text-align:center;">This email was generated automatically. Please do not reply.</p>
        </div>
        </div>
    </body>
    </html>`,
        req.Name,
        req.Email,
        req.ExpectedProfession,
        time.Now().Format("January 2, 2006"),
        time.Now().Format("MST"),
    )


    // 3️⃣ Send email
    emailCfg := utils.EmailConfig{
        Host:     config.Cfg.Cloud.EmailHost,
        Port:     config.Cfg.Cloud.EmailPort,
        Username: config.Cfg.Cloud.EmailHostUser,
        Password: config.Cfg.Cloud.EmailHostPassword,
        From:     config.Cfg.Cloud.DefaultFromEmail,
        UseTLS:   config.Cfg.Cloud.EmailUseTLS,
    }

    // log.Printf("📩 Sending job title change request for user %s (%s) to admin", userID, currentEmail)
    if err := utils.SendEmail(emailCfg, adminEmail, "Signup bonus", emailBody); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to send signup bonus", "error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"issue": "Signup bonus sent"})
}