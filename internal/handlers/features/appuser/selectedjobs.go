package appuser

// import (
// 	"RAAS/internal/dto"
// 	"RAAS/internal/models"
// 	"RAAS/internal/handlers/repository"

// 	"fmt"
// 	"net/http"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// // SelectedJobsHandler handles job selection operations
// type SelectedJobsHandler struct{}

// // NewSelectedJobsHandler initializes and returns a new instance of SelectedJobsHandler
// func NewSelectedJobsHandler() *SelectedJobsHandler {
// 	return &SelectedJobsHandler{}
// }

// // PostSelectedJob saves the selected job for the authenticated user
// func (h *SelectedJobsHandler) PostSelectedJob(c *gin.Context) {
// 	db := c.MustGet("db").(*mongo.Database)
// 	userID := c.MustGet("userID").(string)

// 	var input struct {
// 		JobID string `json:"job_id"`
// 	}
// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
// 		return
// 	}

// 	selectedJobsCollection := db.Collection("selected_job_applications")

// 		// Reuse helper functions
// 	seeker, skills, err := repository.GetSeekerData(db, userID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching seeker data"})
// 		return
// 	}
// 	if seeker.DailySelectableJobsCount <= 0 {
// 		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot select job. Daily limit over."})
// 		return
// 	}

// 	// Prevent duplicates
// 	var existing models.SelectedJobApplication
// 	err = selectedJobsCollection.FindOne(c, bson.M{"auth_user_id": userID, "job_id": input.JobID}).Decode(&existing)
// 	if err != mongo.ErrNoDocuments {
// 		c.JSON(http.StatusConflict, gin.H{"error": "You have already selected this job"})
// 		return
// 	}

// 	// Reuse helper functions


// 	job, err := repository.GetJobByID(db, input.JobID)
// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
// 		return
// 	}

// 	expectedSalary := repository.GenerateSalaryRange()

// 	selectedJob := models.SelectedJobApplication{
// 		AuthUserID:           userID,
// 		JobID:                job.JobID,
// 		Title:                job.Title,
// 		Company:              job.Company,
// 		Location:             job.Location,
// 		PostedDate:           job.PostedDate,
// 		Processed:            true,
// 		JobType:              job.JobType,
// 		Skills:               job.Skills,
// 		UserSkills:           skills,
// 		ExpectedSalary:       expectedSalary,
// 		MatchScore:           70,
// 		Description:          job.JobDescription,
// 		Selected:             true,
// 		CvGenerated:          false,
// 		CoverLetterGenerated: false,
// 		ViewLink:             false,
// 		SelectedDate:         time.Now(),
// 	}

// 	if _, err := selectedJobsCollection.InsertOne(c, selectedJob); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save selected job"})
// 		return
// 	}

// 	// Update seeker & job stats
// 	_, _ = db.Collection("seekers").UpdateOne(c, bson.M{"auth_user_id": userID}, bson.M{"$inc": bson.M{"daily_selectable_jobs_count": -1}})
// 	_, _ = db.Collection("jobs").UpdateOne(c, bson.M{"job_id": job.JobID}, bson.M{"$inc": bson.M{"selected_count": 1}})

// 	c.JSON(http.StatusCreated, gin.H{"message": "Selected job saved successfully"})
// }


// func (h *SelectedJobsHandler) GetSelectedJobs(c *gin.Context) {
// 	// Get the database from the context
// 	db := c.MustGet("db").(*mongo.Database)
// 	selectedJobsCollection := db.Collection("selected_job_applications")

// 	// Retrieve the user ID from the context
// 	userID := c.MustGet("userID").(string)

// 	// Calculate today's date range in UTC
// 	now := time.Now().UTC()
// 	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
// 	endOfDay := startOfDay.Add(24 * time.Hour)

// 	// Filter for the authenticated user and jobs selected today
// 	filter := bson.M{
// 		"auth_user_id":  userID,
// 		"selected_date": bson.M{"$gte": startOfDay, "$lt": endOfDay},
// 	}

// 	// Access pagination values from context set by middleware
// 	pagination := c.MustGet("pagination").(gin.H)
// 	offsetInt := pagination["offset"].(int)
// 	limitInt := pagination["limit"].(int)

// 	// Define the pagination and sort options
// 	findOptions := options.Find().
// 		SetSkip(int64(offsetInt)).
// 		SetLimit(int64(limitInt)).
// 		SetSort(bson.D{{Key: "selected_date", Value: -1}}) // Sort by most recent

// 	// Query the database
// 	cursor, err := selectedJobsCollection.Find(c, filter, findOptions)
// 	if err != nil {
// 		fmt.Println("Error fetching selected jobs:", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching selected jobs"})
// 		return
// 	}
// 	defer cursor.Close(c)

// 	var selectedJobs []dto.SelectedJobResponse

// 	for cursor.Next(c) {
// 		var selectedJob models.SelectedJobApplication
// 		if err := cursor.Decode(&selectedJob); err != nil {
// 			fmt.Println("Error decoding selected job:", err)
// 			continue
// 		}

// 		selectedJobResponse := dto.SelectedJobResponse{
// 			AuthUserID:           selectedJob.AuthUserID,
// 			Source:               selectedJob.Source,
// 			JobID:                selectedJob.JobID,
// 			Title:                selectedJob.Title,
// 			Company:              selectedJob.Company,
// 			Location:             selectedJob.Location,
// 			PostedDate:           selectedJob.PostedDate,
// 			Processed:            selectedJob.Processed,
// 			JobType:              selectedJob.JobType,
// 			Skills:               selectedJob.Skills,
// 			UserSkills:           selectedJob.UserSkills,
// 			ExpectedSalary:       convertSalaryRange(selectedJob.ExpectedSalary),
// 			MatchScore:           selectedJob.MatchScore,
// 			Description:          selectedJob.Description,
// 			Selected:             selectedJob.Selected,
// 			CvGenerated:          selectedJob.CvGenerated,
// 			CoverLetterGenerated: selectedJob.CoverLetterGenerated,
// 			ViewLink:             selectedJob.ViewLink,
// 			SelectedDate:         selectedJob.SelectedDate.Format(time.RFC3339),
// 		}

// 		selectedJobs = append(selectedJobs, selectedJobResponse)
// 	}

// 	if len(selectedJobs) == 0 {
// 		c.JSON(http.StatusNoContent, gin.H{"message": "No selected jobs found for today"})
// 		return
// 	}

// 	totalCount, err := selectedJobsCollection.CountDocuments(c, filter)
// 	if err != nil {
// 		fmt.Println("Error counting selected jobs:", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting selected jobs"})
// 		return
// 	}

// 	nextPage := ""
// 	if int64(offsetInt+limitInt) < totalCount {
// 		nextPage = fmt.Sprintf("/api/selected-jobs?offset=%d&limit=%d", offsetInt+limitInt, limitInt)
// 	}

// 	prevPage := ""
// 	if offsetInt > 0 {
// 		prevPage = fmt.Sprintf("/api/selected-jobs?offset=%d&limit=%d", offsetInt-limitInt, limitInt)
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"selected_jobs": selectedJobs,
// 		"pagination": gin.H{
// 			"total":    totalCount,
// 			"next":     nextPage,
// 			"prev":     prevPage,
// 			"current":  (offsetInt / limitInt) + 1,
// 			"per_page": limitInt,
// 		},
// 	})
// }

// func convertSalaryRange(modelSalary models.SalaryRange) dto.SalaryRange {
// 	return dto.SalaryRange{
// 		Min: modelSalary.Min,
// 		Max: modelSalary.Max,
// 	}
// }
