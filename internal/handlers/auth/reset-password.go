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

	// Find the user
	var user models.AuthUser
	err := coll.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		log.Printf("❌ [PasswordReset] No user found with email: %s", email)
		return fmt.Errorf("user not found with given email")
	}

	// log.Printf("🔍 [PasswordReset] Found user: AUTH_USER_ID=%v, Email=%s", user.AuthUserID, user.Email)

	// Generate token & expiry
	token := utils.GenerateVerificationToken()
	expiry := time.Now().Add(30 * time.Minute)

	// log.Printf("🔐 [PasswordReset] Generated token: %s (expires at %s)", token, expiry.Format(time.RFC3339))

	// Save token and expiry to DB
	update := bson.M{
		"$set": bson.M{
			"verification_token": token,
			"reset_token_expiry": &expiry,
		},
	}

	log.Printf("✍️ [PasswordReset] Attempting to update user by ID: %v", user.AuthUserID)
    res, err := coll.UpdateOne(ctx, bson.M{"auth_user_id": user.AuthUserID}, update)
    if err != nil {
        log.Printf("❌ [PasswordReset] UpdateOne failed: %v", err)
        return err
    }
    if res.MatchedCount == 0 {
        log.Printf("❌ [PasswordReset] No document matched for auth_user_id: %s", user.AuthUserID)
        return fmt.Errorf("user not found for update")
    }
    log.Printf("✅ [PasswordReset] Token saved for user ID %s", user.AuthUserID)

	// Construct reset link and send email
	emailCfg := utils.GetEmailConfig()
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", config.Cfg.Project.FrontendBaseUrl, token)
	log.Printf("📩 [PasswordReset] Reset link generated: %s", resetLink)

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
    </html>
    `, user.Email, resetLink)

	log.Printf("📤 [PasswordReset] Sending reset email to %s...", user.Email)
	err = utils.SendEmail(emailCfg, user.Email, "Reset Your Password", emailBody)
	if err != nil {
		log.Printf("❌ [PasswordReset] Failed to send email: %v", err)
		return err
	}
	log.Printf("✅ [PasswordReset] Reset email sent to %s", user.Email)

	return nil
}


func (r *UserRepo) ResetPassword(ctx context.Context, token, newPassword string) error {
	log.Println("🔐 [ResetPassword] Incoming reset request")
	log.Printf("✅ [ResetPassword] Token received: %s", token)
	log.Printf("✅ [ResetPassword] NewPassword length: %d", len(newPassword))

	coll := r.DB.Collection("auth_users")

	// Construct filter
	now := time.Now()
	filter := bson.M{
		"verification_token": token,
		"reset_token_expiry": bson.M{"$gt": now},
	}
	log.Printf("🔎 [UserRepo] Query filter: %+v", filter)

	// Try to find matching user
	var user models.AuthUser
	err := coll.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		log.Printf("❌ [UserRepo] Token lookup failed: %v", err)

		// Optional: check if token alone exists (helps identify expired case)
		tokenCheck := bson.M{"verification_token": token}
		var debugUser bson.M
		err2 := coll.FindOne(ctx, tokenCheck).Decode(&debugUser)
		if err2 != nil {
			log.Printf("❌ [Debug] Token doesn't exist in DB: %v", err2)
		} else {
			log.Printf("⚠️ [Debug] Token found without valid expiry: %+v", debugUser)
		}

		return errors.New("invalid or expired token")
	}

	log.Printf("✅ [UserRepo] User found: %s", user.Email)
	log.Printf("🔐 [UserRepo] Updating password for user: %s", user.AuthUserID)

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("❌ [UserRepo] Failed to hash password: %v", err)
		return err
	}

	// Update document
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

	res, err := coll.UpdateByID(ctx, user.AuthUserID, update)
	if err != nil {
		log.Printf("❌ [UserRepo] Failed to update password: %v", err)
		return err
	}

	log.Printf("✅ [UserRepo] Password updated successfully. Matched count: %d", res.MatchedCount)
	return nil
}
