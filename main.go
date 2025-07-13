package main

import (
    "RAAS/app/routes"
    // "RAAS/app/workers"
    "RAAS/core/config"
    "RAAS/internal/models"
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
)

func main() {
    gin.SetMode(gin.ReleaseMode)

    // Load config
    if err := config.InitConfig(); err != nil {
        log.Fatalf("Error loading config: %v", err)
    }

    // Init MongoDB
    client, _ := models.InitDB(config.Cfg)

    // Setup Gin router
    r := gin.Default()
    routes.SetupRoutes(r, client, config.Cfg)







	

    // // Start Daily Worker (here: every 5s for testing or switch to 24h)
	// deletionCancel := workers.StartPurgeWorker(client.Database(config.Cfg.Cloud.MongoDBName), 24*time.Second)
	// defer deletionCancel()

    // notifier := workers.StartTestNotifier(client.Database(config.Cfg.Cloud.MongoDBName))
    // defer notifier.Stop()



    // Start HTTP server
    port := os.Getenv("PORT")
    if port == "" {
        port = fmt.Sprintf("%d", config.Cfg.Server.ServerPort)
        log.Printf("🌐 Dev server listening: http://localhost:%s", port)
    }
    srv := &http.Server{
        Addr:    ":" + port,
        Handler: r,
    }
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // Wait for termination signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("🛑 Shutting down...")



    // Shutdown HTTP server with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server shutdown error: %v", err)
    }

    // Close MongoDB connection
    if err := client.Disconnect(context.TODO()); err != nil {
        log.Fatalf("MongoDB disconnect error: %v", err)
    }
    log.Println("✅ Graceful shutdown complete")
}
