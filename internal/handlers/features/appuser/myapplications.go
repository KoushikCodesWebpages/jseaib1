package appuser


// import (
// 	"RAAS/internal/dto"
// 	"RAAS/internal/models"
// 	"fmt"
// 	"net/http"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )
// // MyApplicationsHandler handles operations related to retrieving user's job applications
// type MyApplicationsHandler struct{}

// // NewMyApplicationsHandler initializes and returns a new instance of MyApplicationsHandler
// func NewMyApplicationsHandler() *MyApplicationsHandler {
// 	return &MyApplicationsHandler{}
// }

// // GetMyApplications returns all job applications for the authenticated user
// func (h *MyApplicationsHandler) GetMyApplications(c *gin.Context) {
// 	// Get the database from the context
// 	db := c.MustGet("db").(*mongo.Database)
// 	selectedJobsCollection := db.Collection("selected_job_applications")

// 	// Retrieve the user ID from the context
// 	userID := c.MustGet("userID").(string)

// 	// Filter for the authenticated user's job applications (no date restriction)
// 	filter := bson.M{
// 		"auth_user_id": userID,
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
// 		fmt.Println("Error fetching job applications:", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching job applications"})
// 		return
// 	}
// 	defer cursor.Close(c)

// 	var applications []dto.SelectedJobResponse

// 	for cursor.Next(c) {
// 		var job models.SelectedJobApplication
// 		if err := cursor.Decode(&job); err != nil {
// 			fmt.Println("Error decoding job application:", err)
// 			continue
// 		}

// 		application := dto.SelectedJobResponse{
// 			AuthUserID:           job.AuthUserID,
// 			Source:               job.Source,
// 			JobID:                job.JobID,
// 			Title:                job.Title,
// 			Company:              job.Company,
// 			Location:             job.Location,
// 			PostedDate:           job.PostedDate,
// 			Processed:            job.Processed,
// 			JobType:              job.JobType,
// 			Skills:               job.Skills,
// 			UserSkills:           job.UserSkills,
// 			ExpectedSalary:       convertSalaryRange(job.ExpectedSalary),
// 			MatchScore:           job.MatchScore,
// 			Description:          job.Description,
// 			Selected:             job.Selected,
// 			CvGenerated:          job.CvGenerated,
// 			CoverLetterGenerated: job.CoverLetterGenerated,
// 			ViewLink:             job.ViewLink,
// 			SelectedDate:         job.SelectedDate.Format(time.RFC3339),
// 		}

// 		applications = append(applications, application)
// 	}

// 	if len(applications) == 0 {
// 		c.JSON(http.StatusNoContent, gin.H{"message": "No job applications found"})
// 		return
// 	}

// 	totalCount, err := selectedJobsCollection.CountDocuments(c, filter)
// 	if err != nil {
// 		fmt.Println("Error counting job applications:", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting job applications"})
// 		return
// 	}

// 	nextPage := ""
// 	if int64(offsetInt+limitInt) < totalCount {
// 		nextPage = fmt.Sprintf("/api/my-applications?offset=%d&limit=%d", offsetInt+limitInt, limitInt)
// 	}

// 	prevPage := ""
// 	if offsetInt > 0 {
// 		prevPage = fmt.Sprintf("/api/my-applications?offset=%d&limit=%d", offsetInt-limitInt, limitInt)
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"applications": applications,
// 		"pagination": gin.H{
// 			"total":    totalCount,
// 			"next":     nextPage,
// 			"prev":     prevPage,
// 			"current":  (offsetInt / limitInt) + 1,
// 			"per_page": limitInt,
// 		},
// 	})
// }


