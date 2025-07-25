package settings

import (
	"RAAS/internal/handlers/auth"


	"context"
	"log"

	"net/http"
	"time"
	// "errors"
	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/mongo"
)


func RequestPasswordChangeHandler(c *gin.Context) {
    // 1️⃣ Get email from context (user must be authenticated)
    userEmail := c.MustGet("email").(string)
    if userEmail == "" {
        c.JSON(http.StatusUnauthorized, gin.H{
            "issue": "User email not found in context; please authenticate first.",
            "error": "unauthorized",
        })
        return
    }

    // 2️⃣ Perform password reset using repo
    db := c.MustGet("db").(*mongo.Database)
    userRepo := auth.NewUserRepo(db)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := userRepo.RequestPasswordReset(ctx, userEmail); err != nil {
        if err.Error() == "limit_exceeded" {
            log.Printf("⚠️ [RateLimiter] Too many reset attempts for: %s", userEmail)
            c.JSON(http.StatusTooManyRequests, gin.H{
                "issue": "Too many reset requests. Try again later.",
                "error": "limit_exceeded",
            })
            return
        }
        log.Printf("❌ [PasswordReset] Failed to send reset link for %s: %v", userEmail, err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "issue": "Failed to send reset email. Please try again.",
            "error": err.Error(),
        })
        return
    }

    log.Printf("✅ [PasswordReset] Reset link sent to: %s", userEmail)
    c.JSON(http.StatusOK, gin.H{
        "issue": "Password reset link has been sent to your email.",
    })
}
