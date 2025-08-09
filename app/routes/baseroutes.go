// routes/auth_routes.go
package routes

import (
	"RAAS/core/config"
	"RAAS/core/middlewares"
	"RAAS/internal/handlers/features/base"
	// "RAAS/internal/handlers/oauth"


	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

)

func SetupBaseRoutes(r *gin.Engine, client *mongo.Client, cfg *config.Config) {
	// Rate limiter configurations
	r.Use(middleware.InjectDB(client))
	baseLimiter := middleware.RateLimiterMiddleware(5, time.Minute)
	pagem := middleware.PaginationMiddleware
	authm := middleware.AuthMiddleware()
	baseRoutes := r.Group("/b1/base", authm,baseLimiter,pagem)
	{
		baseRoutes.GET("/announcements",base.AnnouncementHandler)
	}
	
}