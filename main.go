package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"RAAS/app/routes"
	"RAAS/app/workers"
	"RAAS/core/config"
	"RAAS/internal/models"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	// Initialize configuration
	if err := config.InitConfig(); err != nil {
		log.Fatalf("Error initializing config: %v", err)
	}

	// Initialize MongoDB
	client, _ := models.InitDB(config.Cfg)

	// Set up HTTP router
	r := gin.Default()
	routes.SetupRoutes(r, client, config.Cfg)

	// Launch worker
	worker := workers.NewMatchScoreWorker(client)
	ctx, cancel := context.WithCancel(context.Background())
	go worker.Run(ctx)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = fmt.Sprintf("%d", config.Cfg.Server.ServerPort)
		log.Printf("🌐 Starting server on dev port: http://localhost:%s", port)
	} else {
		log.Printf("🌐 Starting server on prod port: %s", port)
	}

	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for termination
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownSignal

	log.Println("🛑 Shutting down server and worker...")
	cancel()

	// Cleanup MongoDB
	if err := client.Disconnect(context.TODO()); err != nil {
		log.Fatalf("Error disconnecting MongoDB: %v", err)
	}
	log.Println("✅ MongoDB connection closed gracefully")
}
