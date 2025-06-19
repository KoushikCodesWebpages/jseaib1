// routes/auth_routes.go
package routes

import (
	"RAAS/core/config"
	"RAAS/core/middlewares"
	"RAAS/internal/handlers/auth"
	"RAAS/internal/handlers/oauth"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions"

	"time"

	"github.com/gin-gonic/gin"

)

func SetupAuthRoutes(r *gin.Engine, cfg *config.Config) {
	// Rate limiter configurations
	signupLimiter := middleware.RateLimiterMiddleware(5, time.Minute)
	loginLimiter := middleware.RateLimiterMiddleware(10, time.Minute)
	// forgotPassLimiter := middleware.RateLimiterMiddleware(3, time.Minute)
	// resetPassLimiter := middleware.RateLimiterMiddleware(3, time.Minute)
	verifyEmailLimiter := middleware.RateLimiterMiddleware(10, time.Minute)
	googleLoginLimiter := middleware.RateLimiterMiddleware(10, time.Minute)
	googleCallbackLimiter := middleware.RateLimiterMiddleware(20, time.Minute)
	store := cookie.NewStore([]byte("your-secret-key"))
    r.Use(sessions.Sessions("session", store))

	authGroup := r.Group("/b1/auth")
	{
		// Standard auth routes (rate-limited where necessary)

 		authGroup.GET("/google/login", googleLoginLimiter, oauth.GoogleLoginHandler)
        authGroup.GET("/google/callback", googleCallbackLimiter, oauth.GoogleCallbackHandler)

        authGroup.POST("/signup", signupLimiter, auth.SeekerSignUp)
        authGroup.GET("/verify-email", verifyEmailLimiter, auth.VerifyEmail)
        authGroup.POST("/login", loginLimiter, auth.Login)
        authGroup.POST("/admin/refresh-token", auth.AdminRefreshToken)

		
		// authGroup.POST("/forgot-password", forgotPassLimiter, auth.ForgotPasswordHandler)
		// authGroup.POST("/admin-reset-token", auth.SystemInitiatedResetTokenHandler) // No limiter
		// authGroup.GET("/reset-password", auth.ResetPasswordPage)                     // Optional
		// authGroup.POST("/reset-password", resetPassLimiter, auth.ResetPasswordHandler)
	}
}