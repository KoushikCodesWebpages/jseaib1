package jobs

import (
	"RAAS/internal/dto"
	"RAAS/internal/handlers/repository"
	"RAAS/internal/models"

	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func JobRetrievalHandler(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	userID := c.MustGet("userID").(string)

	// Fetch seeker and skills using helper
	seeker, err := repository.GetSeekerData(db, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching seeker data"})
		return
	}
	skills := seeker.KeySkills
	if seeker.PrimaryTitle == "" {
		c.JSON(http.StatusNoContent, gin.H{"error": "No preferred job title set for user."})
		return
	}
	
	// Collect preferred titles using helper
	preferredTitles := repository.CollectPreferredTitles(seeker)
	if len(preferredTitles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No preferred job titles set for user."})
		return
	}

	// Get applied jobs using helper
	appliedJobIDs, err := repository.FetchAppliedJobIDs(c, db.Collection("selected_job_applications"), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching applied job data"})
		return
	}

	// Ensure appliedJobIDs is not nil
	if appliedJobIDs == nil {
		appliedJobIDs = []string{}
	}

	// Build MongoDB query
	filter := repository.BuildJobFilter(preferredTitles, appliedJobIDs)

	// Pagination
	pagination := c.MustGet("pagination").(gin.H)
	offset := pagination["offset"].(int)
	limit := pagination["limit"].(int)

	// Query jobs
	cursor, err := db.Collection("jobs").Find(c, filter, options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching job data"})
		return
	}
	defer cursor.Close(c)

	
	var jobs []dto.JobDTO
	for cursor.Next(c) {
		var job models.Job
		if err := cursor.Decode(&job); err != nil {
			fmt.Println("Error decoding job:", err)
			continue
		}

		// Get the next auto-increment ID
		jobID, err := repository.GetNextSequence(db, "job_id")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating job ID"})
			return
		}


		// Append job with auto-incremented ID
		jobs = append(jobs, dto.JobDTO{
			Source:         "seeker",
			ID:             jobID, // Set the auto-increment ID
			JobID:          job.JobID,
			Title:          job.Title,
			Company:        job.Company,
			Location:       job.Location,
			PostedDate:     job.PostedDate,
			Processed:      job.Processed,
			JobType:        job.JobType,
			Skills:         job.Skills,
			UserSkills:     skills,
			
			MatchScore:     70,
			Description:    job.JobDescription,
		})
	}

	// Count total jobs
	totalCount, err := db.Collection("jobs").CountDocuments(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting job data"})
		return
	}

	// Build pagination response
	nextPage := ""
	if int64(offset+limit) < totalCount {
		nextPage = fmt.Sprintf("/api/jobs?offset=%d&limit=%d", offset+limit, limit)
	}
	prevPage := ""
	if offset > 0 {
		prevPage = fmt.Sprintf("/api/jobs?offset=%d&limit=%d", offset-limit, limit)
	}

	// Send response
	c.JSON(http.StatusOK, gin.H{
		"pagination": gin.H{
			"total":    totalCount,
			"next":     nextPage,
			"prev":     prevPage,
			"current":  (offset / limit) + 1,
			"per_page": limit,
		},
		"jobs": jobs,
	})
}



