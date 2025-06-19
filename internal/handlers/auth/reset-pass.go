package auth

// import (
// 	"RAAS/core/config"
// 	"RAAS/internal/models"
// 	"RAAS/utils"

// 	"fmt"
// 	"github.com/gin-gonic/gin"
// 	"golang.org/x/crypto/bcrypt"
// 	"gorm.io/gorm"
// 	"net/http"

// 	"time"
// )


// // ForgotPasswordHandler sends a reset password link to the user's email
// func ForgotPasswordHandler(c *gin.Context) {
// 	var req struct {
// 		Email string `json:"email" binding:"required,email"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	db := c.MustGet("db").(*gorm.DB)
// 	var user models.AuthUser
// 	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
// 		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent."})
// 		return
// 	}

// 	// Generate token and send email
// 	token, err := generateResetTokenForUser(db, &user)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
// 		return
// 	}

// 	resetLink := fmt.Sprintf("%s/reset-password?token=%s", config.Cfg.Project.FrontendBaseUrl, token)
// 	body := fmt.Sprintf(`<p>Click to reset your password:</p><a href="%s">%s</a>`, resetLink, resetLink)

// 	emailCfg := utils.EmailConfig{
// 		Host:     config.Cfg.Cloud.EmailHost,
// 		Port:     config.Cfg.Cloud.EmailPort,
// 		Username: config.Cfg.Cloud.EmailHostUser,
// 		Password: config.Cfg.Cloud.EmailHostPassword,
// 		From:     config.Cfg.Cloud.DefaultFromEmail,
// 		UseTLS:   config.Cfg.Cloud.EmailUseTLS,
// 	}

// 	if err := utils.SendEmail(emailCfg, user.Email, "Reset your password", body); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset email"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent."})
// }

// // SystemInitiatedResetTokenHandler generates a reset token for system use
// func SystemInitiatedResetTokenHandler(c *gin.Context) {
// 	var req struct {
// 		Email string `json:"email" binding:"required,email"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	db := c.MustGet("db").(*gorm.DB)
// 	var user models.AuthUser
// 	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 		return
// 	}

// 	token, err := generateResetTokenForUser(db, &user)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Reset token generated",
// 		"token":   token,
// 	})
// }

// // ResetPasswordHandler handles the password reset process
// func ResetPasswordHandler(c *gin.Context) {
// 	var req struct {
// 		Token           string `json:"token" binding:"required"`
// 		NewPassword     string `json:"new_password" binding:"required,min=8"`
// 		ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	db := c.MustGet("db").(*gorm.DB)
// 	var user models.AuthUser
// 	if err := db.Where("verification_token = ?", req.Token).First(&user).Error; err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
// 		return
// 	}

// 	// Check if token has expired
// 	if user.ResetTokenExpiry.Before(time.Now()) {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Token has expired"})
// 		return
// 	}

// 	// Hash the new password and save it
// 	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
// 		return
// 	}

// 	// Save new password, invalidate token, and clear expiry
// 	user.Password = string(hashed)
// 	user.VerificationToken = ""
// 	user.ResetTokenExpiry = nil

// 	if err := db.Save(&user).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save new password"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
// }
