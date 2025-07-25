package middleware

import (
	"RAAS/core/config"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InjectDB(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		db := client.Database(config.Cfg.Cloud.MongoDBName)
		c.Set("db", db)
		c.Next()
	}
}