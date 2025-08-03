package jobs

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "RAAS/internal/models"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

// JobsHandler handles job retrievals.
type JobsHandler struct{}

func NewJobsHandler() *JobsHandler {
    return &JobsHandler{}
}

// GET /b1/all-jobs
func (h *JobsHandler) GetAllJobs(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    cursor, err := db.Collection("jobs").Find(ctx, bson.M{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query jobs"})
        return
    }
    defer cursor.Close(ctx)

    var jobs []models.Job
    for cursor.Next(ctx) {
        var job models.Job
        if err := cursor.Decode(&job); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode job"})
            return
        }

        // Log the first document only
        if len(jobs) == 0 {
            fmt.Println("🧪 Sample raw job struct:")
            fmt.Printf("%+v\n", job)
        }

        jobs = append(jobs, job)
    }

    if err := cursor.Err(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "cursor error"})
        return
    }

    if len(jobs) == 0 {
        c.JSON(http.StatusOK, gin.H{"message": "no jobs found"})
        return
    }

    c.JSON(http.StatusOK, jobs)
}

// DELETE /b1/all-jobs
func (h *JobsHandler) DeleteAllJobs(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    res, err := db.Collection("jobs").DeleteMany(ctx, bson.M{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete jobs"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message":       "All jobs deleted",
        "deleted_count": res.DeletedCount,
    })
}
