package jobs

import (
	"RAAS/internal/dto"
	"RAAS/internal/handlers/repository"
	"RAAS/internal/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func JobRetrievalHandler(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	userID := c.MustGet("userID").(string)

	// Get seeker for skills, etc.
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

	// Get applied job IDs to exclude
	appliedJobIDs, err := repository.FetchAppliedJobIDs(c, db.Collection("selected_job_applications"), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching applied job data"})
		return
	}
	appliedSet := make(map[string]bool)
	for _, id := range appliedJobIDs {
		appliedSet[id] = true
	}

	// Pagination
	pagination := c.MustGet("pagination").(gin.H)
	offset := pagination["offset"].(int)
	limit := pagination["limit"].(int)

	// Step 1: Get all match scores for user, sorted by match_score DESC
	matchCursor, err := db.Collection("match_scores").
		Find(c,
			bson.M{"auth_user_id": userID},
			options.Find().SetSort(bson.D{{Key: "match_score", Value: -1}}),
		)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching match scores"})
		return
	}
	defer matchCursor.Close(c)

	var scores []models.MatchScore
	if err := matchCursor.All(c, &scores); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding match scores"})
		return
	}

	// Step 2: Filter out applied jobs and collect job IDs in order
	var matchedJobIDs []string
	for _, ms := range scores {
		if !appliedSet[ms.JobID] {
			matchedJobIDs = append(matchedJobIDs, ms.JobID)
		}
	}

	total := len(matchedJobIDs)
	if total == 0 {
		c.JSON(http.StatusOK, gin.H{
			"pagination": gin.H{
				"total":    0,
				"next":     "",
				"prev":     "",
				"current":  1,
				"per_page": limit,
			},
			"jobs": []dto.JobDTO{},
		})
		return
	}

	// Step 3: Paginate the job IDs
	end := offset + limit
	if end > total {
		end = total
	}
	pagedJobIDs := matchedJobIDs[offset:end]

	// Step 4: Build filter for jobs
	filter := bson.M{
		"job_id": bson.M{"$in": pagedJobIDs},
		"posted_date": bson.M{
			"$gte": time.Now().AddDate(0, 0, -14).Format("2006-01-02"),
		},
	}
	if lang := c.Query("job_language"); lang != "" {
		filter["job_language"] = bson.M{"$regex": lang, "$options": "i"}
	}
	if title := c.Query("title"); title != "" {
		filter["title"] = bson.M{"$regex": title, "$options": "i"}
	}
	if company := c.Query("company"); company != "" {
		filter["company"] = bson.M{"$regex": company, "$options": "i"}
	}

	jobCursor, err := db.Collection("jobs").Find(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching job data"})
		return
	}
	defer jobCursor.Close(c)

	// Build job map
	jobMap := make(map[string]models.Job)
	for jobCursor.Next(c) {
		var job models.Job
		if err := jobCursor.Decode(&job); err != nil {
			continue
		}
		jobMap[job.JobID] = job
	}

	// Fetch selected applications for current page jobs
	selectedFilter := bson.M{
		"auth_user_id": userID,
		"job_id":       bson.M{"$in": pagedJobIDs},
	}
	selectedCursor, err := db.Collection("selected_job_applications").Find(c, selectedFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching selected applications"})
		return
	}
	defer selectedCursor.Close(c)

	selectedMap := make(map[string]models.SelectedJobApplication)
	for selectedCursor.Next(c) {
		var s models.SelectedJobApplication
		if err := selectedCursor.Decode(&s); err != nil {
			continue
		}
		selectedMap[s.JobID] = s
	}

	// Step 5: Build DTOs
	var jobs []dto.JobDTO
	var index uint = 1
	for _, jobID := range pagedJobIDs {
		job, ok := jobMap[jobID]
		if !ok {
			continue
		}

		score := repository.GetMatchScoreForJob(c, db, userID, jobID)
		isSelected := repository.IsJobSelected(c, db, userID, jobID)
		selected := selectedMap[jobID]

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
			MatchScore:  score,
			Description: job.JobDescription,
			JobLang:     job.JobLang,
			JobTitle:    job.JobTitle,
			Selected:    isSelected,

			LinkViewed:   selected.ViewLink,
			CvGenerated:  selected.CvGenerated,
			ClGenerated:  selected.CoverLetterGenerated,
		})
		index++
	}

	// Step 6: Build pagination
	nextPage := ""
	if end < total {
		nextPage = fmt.Sprintf("/b1/api/jobs?offset=%d&limit=%d", end, limit)
	}
	prevPage := ""
	if offset > 0 {
		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		prevPage = fmt.Sprintf("/b1/api/jobs?offset=%d&limit=%d", prevOffset, limit)
	}

	c.JSON(http.StatusOK, gin.H{
		"pagination": gin.H{
			"total":    total,
			"next":     nextPage,
			"prev":     prevPage,
			"current":  (offset / limit) + 1,
			"per_page": limit,
		},
		"jobs": jobs,
	})
}
