package appuser

import (
    "context"
    "net/http"
    "time"

    "RAAS/internal/models"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type SelectedJobHandler struct{}

func NewSelectedJobHandler() *SelectedJobHandler {
    return &SelectedJobHandler{}
}

func (h *SelectedJobHandler) GetSelectedJobApplications(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    coll := db.Collection("selected_job_applications")
    userID := c.MustGet("userID").(string)

    // Build filter
    filter := bson.M{"auth_user_id": userID}
    if jobID := c.Query("job_id"); jobID != "" {
        filter["job_id"] = jobID
    }
    if status := c.Query("status"); status != "" {
        filter["status"] = status
    }
    if source := c.Query("source"); source != "" {
        filter["source"] = source // matches equality :contentReference[oaicite:0]{index=0}
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cursor, err := coll.Find(ctx, filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "DB query error"})
        return
    }
    defer cursor.Close(ctx)

    var results []models.SelectedJobApplication
    if err := cursor.All(ctx, &results); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse data"})
        return
    }

    c.JSON(http.StatusOK, results)
}
