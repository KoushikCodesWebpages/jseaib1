package routes

import (
	"RAAS/core/config"
	"RAAS/core/middlewares"
	//"RAAS/internal/handlers/features/appuser"
	"RAAS/internal/handlers/features/settings"

	"strings"
	"time"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(r *gin.Engine, client *mongo.Client, cfg *config.Config) {
    // --- Debug Logging Middleware ---
    r.Use(func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
        method := c.Request.Method
        rawPath := c.Request.URL.Path
        matched := c.FullPath() // empty if route not matched
        log.Printf("➡️ Incoming %s %s │ Origin: %s │ FullPath: %q", method, rawPath, origin, matched)
        c.Next()
        acao := c.Writer.Header().Get("Access-Control-Allow-Origin")
        log.Printf("⬅️ After handlers │ ACAO header: %q", acao)
    })
	
    // --- CORS middleware ---
    origins := strings.Split(cfg.Project.CORSAllowedOrigins, ",")
    for i := range origins {
        origins[i] = strings.TrimSpace(origins[i])
    }
    corsConfig := cors.Config{
        AllowOrigins:     origins,
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Accept", "Origin", "Cache-Control", "X-Requested-With"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }
    r.Use(cors.New(corsConfig))

    // --- Optional: Wrap DB middleware afterward ---
    r.Use(middleware.InjectDB(client))

    // --- Static Routes & Fallback ---
    r.Static("/assets", "./public/dist/assets")
    r.GET("/", func(c *gin.Context) {
    c.String(200, "This website’s backend is completely built and hosted by Koushik.")
})
    r.GET("/reset-password", func(c *gin.Context) { c.File("./app/templates/resetpassword.html") })


    // --- API Routes ---
    SetupAuthRoutes(r, cfg)
    SetupDataEntryRoutes(r, client, cfg)
    SetupFeatureRoutes(r, client, cfg)
    SetupBaseRoutes(r,client,cfg)

    // --- Exposed admin routes ---
    r.POST("/b1/api/reset-db", settings.ResetDBHandler)
    r.POST("/b1/api/print-all-collections", settings.PrintAllCollectionsHandler)
    r.POST("/b1/api/sign-up-bonus", settings.NewSettingsHandler().SignUpBenefitEmail)
}
