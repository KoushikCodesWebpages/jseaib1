package auth

import (

	"RAAS/internal/models"
	"RAAS/utils"
    "RAAS/core/config"

	"context"
	"errors"
	"fmt"
    "log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func (r *UserRepo) RequestPasswordReset(ctx context.Context, email string) error {
	coll := r.DB.Collection("auth_users")

	var user models.AuthUser
	err := coll.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		log.Printf("❌ [PasswordReset] No user found with email: %s", email)
		return fmt.Errorf("user not found with given email")
	}

	token := utils.GenerateVerificationToken()
	expiry := time.Now().Add(30 * time.Minute)

	update := bson.M{
		"$set": bson.M{
			"verification_token": token,
			"reset_token_expiry": &expiry,
		},
	}

	res, err := coll.UpdateOne(ctx, bson.M{"auth_user_id": user.AuthUserID}, update)
	if err != nil || res.MatchedCount == 0 {
		log.Printf("❌ [PasswordReset] Update failed for user: %s", user.AuthUserID)
		return fmt.Errorf("failed to save reset token")
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", config.Cfg.Project.FrontendBaseUrl, token)

	emailCfg := utils.GetEmailConfig()
	emailBody := fmt.Sprintf(`
    <html>
    <body style="font-family: Arial, sans-serif; background-color: #f9f9f9; margin: 0; padding: 0;">
        <div style="max-width: 600px; margin: 40px auto; background: #ffffff; padding: 30px; border-radius: 10px; box-shadow: 0 2px 8px rgba(0,0,0,0.05);">
            <h2 style="color: #2196F3; text-align: center;">Reset Your Password</h2>
            <p>Hi %s,</p>
            <p>We received a request to reset your password. Click the button below to continue. This link is valid for 30 minutes:</p>
            <div style="text-align: center; margin: 30px 0;">
                <a href="%s" style="background-color: #2196F3; color: #ffffff; padding: 14px 24px; text-decoration: none; border-radius: 6px; font-weight: bold;">
                    Reset Password
                </a>
            </div>
            <p>If you didn’t request a password reset, you can safely ignore this email.</p>
            <p>Cheers,<br><strong>The JSE AI Team</strong></p>
        </div>
    </body>
    </html>`, user.Email, resetLink)

	err = utils.SendEmail(emailCfg, user.Email, "Reset Your Password", emailBody)
	if err != nil {
		log.Printf("❌ [PasswordReset] Email send failed: %v", err)
		return err
	}

	log.Printf("✅ [PasswordReset] Email sent to: %s", user.Email)
	return nil
}


func (r *UserRepo) ResetPassword(ctx context.Context, token, newPassword string) error {
	log.Println("🔐 [ResetPassword] Starting process")

	coll := r.DB.Collection("auth_users")
	now := time.Now()

	filter := bson.M{
		"verification_token": token,
		"reset_token_expiry": bson.M{"$gt": now},
	}

	var user models.AuthUser
	err := coll.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		log.Printf("❌ [ResetPassword] Token invalid or expired")
		return errors.New("invalid or expired token")
	}

	// Ensure new password is different
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(newPassword)); err == nil {
		log.Println("⚠️ [ResetPassword] New password is same as old")
		return errors.New("new password must be different from the old password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("❌ [ResetPassword] Hashing failed: %v", err)
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"password":               string(hashedPassword),
			"password_last_updated": time.Now(),
			"updated_at":            time.Now(),
		},
		"$unset": bson.M{
			"verification_token": "",
			"reset_token_expiry": "",
		},
	}

	res, err := coll.UpdateOne(ctx, bson.M{"auth_user_id": user.AuthUserID}, update)
	if err != nil {
		log.Printf("❌ [ResetPassword] UpdateOne failed: %v", err)
		return err
	}

	log.Printf("✅ [ResetPassword] Updated %d document(s)", res.ModifiedCount)
	return nil
}
