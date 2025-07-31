// routes/auth_routes.go
package routes

import (
	"RAAS/core/config"
	"RAAS/core/middlewares"
	"RAAS/internal/handlers/auth"
	// "RAAS/internal/handlers/oauth"


	"time"

	"github.com/gin-gonic/gin"

)

func SetupAuthRoutes(r *gin.Engine, cfg *config.Config) {
	// Rate limiter configurations
	signupLimiter := middleware.RateLimiterMiddleware(5, time.Minute)
	loginLimiter := middleware.RateLimiterMiddleware(10, time.Minute)
	// forgotPassLimiter := middleware.RateLimiterMiddleware(3, time.Minute)
	resetPassLimiter := middleware.RateLimiterMiddleware(100, time.Hour* 24)
	verifyEmailLimiter := middleware.RateLimiterMiddleware(10, time.Minute)


	authRoutes := r.Group("/b1/auth")
	{
		// Standard auth routes (rate-limited where necessary)

        authRoutes.POST("/signup", signupLimiter, auth.SeekerSignUp)
        authRoutes.GET("/verify-email", verifyEmailLimiter, auth.VerifyEmail)
        authRoutes.POST("/login", loginLimiter, auth.SeekerLogin)
        authRoutes.POST("/admin/refresh-token", auth.AdminRefreshToken)
		authRoutes.POST("/request-password-reset",resetPassLimiter, auth.RequestPasswordResetHandler )
		authRoutes.POST("/reset-password", resetPassLimiter, auth.ResetPasswordHandler)
		
	}
}