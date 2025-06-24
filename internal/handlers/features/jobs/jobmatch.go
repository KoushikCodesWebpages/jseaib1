package jobs

import (
	"RAAS/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	//"github.com/google/uuid"
)

// MatchScoreHandler handles match score retrieval
type MatchScoreHandler struct{}

func NewMatchScoreHandler() *MatchScoreHandler {
    return &MatchScoreHandler{}
}

// GET /b1/match-scores?job_id=...
func (h *MatchScoreHandler) GetMatchScores(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    coll := db.Collection("match_scores")
    userID := c.MustGet("userID").(string)

    filter := bson.M{"auth_user_id": userID}
    if jobID := c.Query("job_id"); jobID != "" {
        filter["job_id"] = jobID
    }

    cursor, err := coll.Find(c, filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer cursor.Close(c)

    var results []models.MatchScore
    if err := cursor.All(c, &results); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing results"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": results})
}
