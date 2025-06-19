package auth

import (
	"crypto/rand"
	"encoding/hex"
	// "html/template"
)

func GenerateResetToken() (string, error) {
	bytes := make([]byte, 32) // 256-bit token
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// // Helper function to generate reset token and update the user model
// func generateResetTokenForUser(db *gorm.DB, user *models.AuthUser) (string, error) {
// 	token, err := utils.GenerateResetToken()
// 	if err != nil {
// 		return "", err
// 	}
// 	user.VerificationToken = token
// 	expiryTime := time.Now().Add(time.Hour)
// 	user.ResetTokenExpiry = &expiryTime

// 	if err := db.Save(user).Error; err != nil {
// 		return "", err
// 	}
// 	return token, nil
// }





// // ResetPasswordPage renders the password reset form with token validation
// func ResetPasswordPage(c *gin.Context) {
// 	token := c.Query("token")
// 	if token == "" {
// 		c.String(http.StatusBadRequest, "Missing token")
// 		return
// 	}

// 	db := c.MustGet("db").(*gorm.DB)
// 	var user models.AuthUser
// 	if err := db.Where("verification_token = ?", token).First(&user).Error; err != nil {
// 		c.String(http.StatusNotFound, "Invalid or expired token")
// 		return
// 	}

// 	// Check if token has expired
// 	if user.ResetTokenExpiry.Before(time.Now()) {
// 		c.String(http.StatusBadRequest, "Token has expired")
// 		return
// 	}

// 	// Use Go's html/template package to render the reset password page
// 	tmpl, err := template.New("resetPasswordPage").Parse(`
// 		<!DOCTYPE html>
// 		<html lang="en">
// 		<head>
// 			<meta charset="UTF-8">
// 			<title>Reset Password</title>
// 			<style>
// 				body { font-family: Arial, sans-serif; background-color: #f2f4f8; color: #333; text-align: center; padding-top: 100px; }
// 				.card { background: white; padding: 40px; margin: auto; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); width: 90%; max-width: 500px; }
// 				h1 { color: #007bff; }
// 				input[type="password"] { padding: 10px; margin: 10px 0; width: 100%; border-radius: 5px; border: 1px solid #ccc; }
// 				button { padding: 10px 20px; background-color: #007bff; color: white; border: none; border-radius: 5px; cursor: pointer; }
// 				button:hover { background-color: #0056b3; }
// 				.error { color: red; font-size: 12px; }
// 			</style>
// 		</head>
// 		<body>
// 			<div class="card">
// 				<h1>Reset Your Password</h1>
// 				<form id="resetPasswordForm" action="{{.FrontendBaseUrl}}/reset-password" method="POST">
// 					<input type="hidden" name="token" value="{{.Token}}" />
// 					<input type="password" id="newPassword" name="new_password" placeholder="New Password" required />
// 					<input type="password" id="confirmPassword" name="confirm_password" placeholder="Confirm Password" required />
// 					<div id="errorMessage" class="error"></div>
// 					<button type="submit">Submit</button>
// 				</form>
// 			</div>
// 			<script>
// 				document.getElementById('resetPasswordForm').onsubmit = function(event) {
// 					var newPassword = document.getElementById('newPassword').value;
// 					var confirmPassword = document.getElementById('confirmPassword').value;
// 					var errorMessage = document.getElementById('errorMessage');

// 					errorMessage.textContent = '';

// 					if (newPassword !== confirmPassword) {
// 						errorMessage.textContent = 'Passwords do not match!';
// 						event.preventDefault();
// 					}
// 				};
// 			</script>
// 		</body>
// 		</html>
// 	`)

// 	if err != nil {
// 		c.String(http.StatusInternalServerError, fmt.Sprintf("Error loading template: %s", err.Error()))
// 		return
// 	}

// 	data := struct {
// 		FrontendBaseUrl string
// 		Token            string
// 	}{
// 		FrontendBaseUrl: config.Cfg.Project.FrontendBaseUrl,
// 		Token:            token,
// 	}

// 	c.Header("Content-Type", "text/html; charset=utf-8")
// 	if err := tmpl.Execute(c.Writer, data); err != nil {
// 		c.String(http.StatusInternalServerError, fmt.Sprintf("Error rendering template: %s", err.Error()))
// 	}
// }
