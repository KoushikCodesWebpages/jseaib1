package jobs

import (
	"RAAS/internal/dto"
	"RAAS/internal/handlers/repository"
	"RAAS/internal/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func JobRetrievalHandler(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	userID := c.MustGet("userID").(string)

	// Get seeker and skills
	seeker, err := repository.GetSeekerData(db, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching seeker data"})
		return
	}
	if seeker.PrimaryTitle == "" {
		c.JSON(http.StatusNoContent, gin.H{"error": "No preferred job title set for user."})
		return
	}
	skills := seeker.KeySkills

	preferredTitles := repository.CollectPreferredTitles(seeker)
	if len(preferredTitles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No preferred job titles set for user."})
		return
	}

	appliedJobIDs, err := repository.FetchAppliedJobIDs(c, db.Collection("selected_job_applications"), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching applied job data"})
		return
	}
	if appliedJobIDs == nil {
		appliedJobIDs = []string{}
	}

	// Optional job_language query param
	jobLanguage := c.Query("job_language")

	// Build job filter (includes last 2 weeks filtering)
	filter := repository.BuildJobFilter(preferredTitles, appliedJobIDs, jobLanguage)

	// Pagination values
	pagination := c.MustGet("pagination").(gin.H)
	offset := pagination["offset"].(int)
	limit := pagination["limit"].(int)

	// Set Mongo FindOptions with sorting by posted_date (newest first)
	findOpts := options.Find().
		SetSort(bson.D{{Key: "posted_date", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := db.Collection("jobs").Find(c, filter, findOpts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching job data"})
		return
	}
	defer cursor.Close(c)

	var jobs []dto.JobDTO
	var index uint = 1
	for cursor.Next(c) {
		var job models.Job
		if err := cursor.Decode(&job); err != nil {
			fmt.Println("Error decoding job:", err)
			continue
		}

		matchScore := repository.GetMatchScoreForJob(c, db, userID, job.JobID)
		isSelected := repository.IsJobSelected(c, db, userID, job.JobID)

		jobs = append(jobs, dto.JobDTO{
			Source:      "seeker",
			ID:          index,
			JobID:       job.JobID,
			Title:       job.Title,
			Company:     job.Company,
			Location:    job.Location,
			PostedDate:  job.PostedDate,
			Processed:   job.Processed,
			JobType:     job.JobType,
			Skills:      job.Skills,
			UserSkills:  skills,
			MatchScore:  matchScore,
			Description: job.JobDescription,
			JobLang:     job.JobLang,
			JobTitle:    job.JobTitle,
			Selected:    isSelected,
		})
		index++
	}

	// Count total jobs matching filter
	totalCount, err := db.Collection("jobs").CountDocuments(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting job data"})
		return
	}

	// Build pagination links
	nextPage := ""
	if int64(offset+limit) < totalCount {
		nextPage = fmt.Sprintf("/api/jobs?offset=%d&limit=%d", offset+limit, limit)
	}
	prevPage := ""
	if offset > 0 {
		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		prevPage = fmt.Sprintf("/api/jobs?offset=%d&limit=%d", prevOffset, limit)
	}

	// Final JSON response
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
