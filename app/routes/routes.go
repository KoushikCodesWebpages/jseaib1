package routes

import (
	"RAAS/core/config"
	"RAAS/core/middlewares"
	"RAAS/internal/handlers/features/appuser"


	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(r *gin.Engine, client *mongo.Client, cfg *config.Config) {
	// CORS
	origins := strings.Split(cfg.Project.CORSAllowedOrigins, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}

	corsConfig := cors.Config{
		AllowOrigins:  origins,
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Accept", "Origin", "Cache-Control", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge: 12 * time.Hour,
	}

	//INJECT
	r.Use(cors.New(corsConfig))
	r.Use(middleware.InjectDB(client))


	//STATIC
	r.Static("/assets", "./public/dist/assets")
	r.GET("/", func(c *gin.Context) {
		c.File("./public/dist/index.html")
	})
	r.NoRoute(func(c *gin.Context) {
		c.File("./app/templates/noroutes.html")
	})
	
	

	// SETUP
	SetupAuthRoutes(r, cfg)
	SetupDataEntryRoutes(r, client, cfg)
	SetupFeatureRoutes(r, client, cfg)

	// EXPOSED
	r.POST("/b1/api/reset-db", appuser.ResetDBHandler)
	r.POST("/b1/api/print-all-collections", appuser.PrintAllCollectionsHandler)
}
