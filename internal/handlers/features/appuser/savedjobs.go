package appuser

import (
    "fmt"
    "net/http"
    "strconv"

    "RAAS/internal/dto"
    "RAAS/internal/handlers/repository"
    "RAAS/internal/models"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// SavedJobsHandler handles saving and retrieving saved jobs
type SavedJobsHandler struct{}

// NewSavedJobsHandler initializes and returns a new SavedJobsHandler instance
func NewSavedJobsHandler() *SavedJobsHandler {
    return &SavedJobsHandler{}
}

// POST /saved-jobs
func (h *SavedJobsHandler) SaveJob(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    userID := c.MustGet("userID").(string)

    var payload struct {
        JobID  string `json:"job_id" binding:"required"`
    }
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
        return
    }

    coll := db.Collection("saved_jobs")
    _, err := coll.InsertOne(c, models.SavedJob{
        AuthUserID: userID,

        JobID:      payload.JobID,
    })
    if err != nil {
        if mongo.IsDuplicateKeyError(err) {
            c.JSON(http.StatusConflict, gin.H{"error": "Job already saved"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save job"})
        }
        return
    }

    c.JSON(http.StatusCreated, gin.H{"issue": "Job saved successfully"})
}

// DELETE /saved-jobs/:job_id
func (h *SavedJobsHandler) DeleteSavedJob(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    userID := c.MustGet("userID").(string)
    jobID := c.Param("job_id") // captured from URL

    res, err := db.Collection("saved_jobs").DeleteOne(c, bson.M{
        "auth_user_id": userID,
        "job_id":       jobID,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete saved job"})
        return
    }
    if res.DeletedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Saved job not found"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"issue": "Saved Job Deleted"})
}
func (h *SavedJobsHandler) GetSavedJobs(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    userID := c.MustGet("userID").(string)

    // 1️⃣ Fetch saved job IDs
    savedIDs, err := repository.FetchSavedJobIDs(c, db.Collection("saved_jobs"), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching saved job IDs"})
        return
    }
    if len(savedIDs) == 0 {
        c.JSON(http.StatusNoContent, gin.H{"jobs": []dto.JobDTO{}})
        return
    }

    // 2️⃣ Pagination
    page := 1
    size := 10
    if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
        page = p
    }
    if s, err := strconv.Atoi(c.DefaultQuery("size", "10")); err == nil && s > 0 {
        size = s
    }
    offset := (page - 1) * size

    // 3️⃣ Fetch job documents
    cursor, err := db.Collection("jobs").Find(c,
        bson.M{"job_id": bson.M{"$in": savedIDs}},
        options.Find().SetSkip(int64(offset)).SetLimit(int64(size)),
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching jobs"})
        return
    }
    defer cursor.Close(c)

    // 4️⃣ Load seeker data (for skills)
    seeker, err := repository.GetSeekerData(db, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching seeker data"})
        return
    }

    // 5️⃣ Build JobDTO list
    var jobs []dto.JobDTO
    index := uint(offset + 1)
    for cursor.Next(c) {
        var job models.Job
        if err := cursor.Decode(&job); err != nil {
            continue
        }

        score := repository.GetMatchScoreForJob(c, db, userID, job.JobID)
        isSelected := repository.IsJobSelected(c, db, userID, job.JobID)
        selected := models.SelectedJobApplication{} // optional lookup if needed

        // If you want the link/cv/cl flags:
        if isSelected {
            _ = db.Collection("selected_job_applications").
                FindOne(c, bson.M{"auth_user_id": userID, "job_id": job.JobID}).
                Decode(&selected)
        }

        jobs = append(jobs, dto.JobDTO{
            Source:        "saved",
            ID:            index,
            JobID:         job.JobID,
            Title:         job.Title,
            Company:       job.Company,
            Location:      job.Location,
            PostedDate:    job.PostedDate,
            Processed:     job.Processed,
            JobType:       job.JobType,
            Skills:        job.Skills,
            UserSkills:    seeker.KeySkills,
            MatchScore:    score,
            Description:   job.JobDescription,
            JobLang:       job.JobLang,
            JobTitle:      job.JobTitle,
            Selected:      isSelected,
            LinkViewed:    selected.ViewLink,
            CvGenerated:   selected.CvGenerated,
            ClGenerated:   selected.CoverLetterGenerated,
        })
        index++
    }

    // 6️⃣ Build pagination metadata
    total := len(savedIDs)
    totalPages := (total + size - 1) / size
    next, prev := "", ""
    if page < totalPages {
        next = fmt.Sprintf("/saved-jobs?page=%d&size=%d", page+1, size)
    }
    if page > 1 {
        prev = fmt.Sprintf("/saved-jobs?page=%d&size=%d", page-1, size)
    }

    // 7️⃣ Return response
    c.JSON(http.StatusOK, gin.H{
        "pagination": gin.H{
            "total":       total,
            "per_page":    size,
            "current":     page,
            "total_pages": totalPages,
            "next":        next,
            "prev":        prev,
        },
        "jobs": jobs,
    })
}
