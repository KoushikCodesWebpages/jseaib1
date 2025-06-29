package appuser

import (
	"context"
	"log"
	"net/http"
	"time"

	"RAAS/internal/handlers/repository"
	"RAAS/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SeekerHandler struct{}

func NewSeekerHandler() *SeekerHandler {
	return &SeekerHandler{}
}

// GetSeekerProfile fetches the full seeker document by userID
func (h *SeekerHandler) GetSeekerProfile(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
			log.Printf("Seeker not found for auth_user_id: %s", userID)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve seeker"})
			log.Printf("Failed to retrieve seeker for auth_user_id: %s, Error: %v", userID, err)
		}
		return
	}

	profileCompletion, missing := repository.CalculateProfileCompletion(seeker)
	// Return the entire seeker document as JSON
	c.JSON(http.StatusOK, gin.H{
		"seeker": seeker,
		"profile_completion": profileCompletion,
		"not_completed":missing,
	})
}
