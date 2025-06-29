package preference

import (

	"RAAS/internal/dto"
	"RAAS/internal/models"
	"RAAS/internal/handlers/repository"

	"context"
	"log"
	"net/http"
	"time"
	// "strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
)
type WorkExperienceHandler struct{}

func NewWorkExperienceHandler() *WorkExperienceHandler {
	return &WorkExperienceHandler{}
}

// CreateWorkExperience handles the creation or update of a single work experience
func (h *WorkExperienceHandler) CreateWorkExperience(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")
	entryTimelineCollection := db.Collection("user_entry_timelines")

	var input dto.WorkExperienceRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Bind error for user %s: %v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"issue": "Some required fields are missing or invalid.",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Seeker not found for user %s", userID)
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
				"issue": "No account found. It might have been removed or reset.",
			})
		} else {
			log.Printf("DB read error for user %s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"issue": "Could not retrieve your profile. Please try again.",
			})
		}
		return
	}

	if err := repository.AppendToWorkExperience(&seeker, input); err != nil {
		log.Printf("Processing error for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Something went wrong adding your work experience. Try again.",
		})
		return
	}

	update := bson.M{"$set": bson.M{"work_experiences": seeker.WorkExperiences}}
	updateResult, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update)
	if err != nil {
		log.Printf("DB update error for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Unable to save your work experience right now.",
		})
		return
	}
	if updateResult.MatchedCount == 0 {
		log.Printf("Update matched zero docs for user %s", userID)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "no seeker updated for " + userID,
			"issue": "Your account wasn't found to update. Please refresh and try again.",
		})
		return
	}

	if _, err := entryTimelineCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, bson.M{"$set": bson.M{"work_experiences_completed": true}}); err != nil {
		log.Printf("Timeline update error for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Work experience saved, but progress tracking failed.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Work experience added successfully"})
}

// GetWorkExperience retrieves all work experiences for the user
func (h *WorkExperienceHandler) GetWorkExperience(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Seeker not found for user %s", userID)
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
				"issue": "No account found. It might have been removed or reset.",
			})
		} else {
			log.Printf("DB read error for user %s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"issue": "Could not retrieve your profile. Please try again.",
			})
		}
		return
	}

	if len(seeker.WorkExperiences) == 0 {
		c.JSON(http.StatusNoContent, gin.H{"message": "No work experiences found"})
		return
	}

	workExperiences, err := repository.GetWorkExperience(&seeker)
	if err != nil {
		log.Printf("GetWorkExperience error for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Failed to load your work experiences. Please try again.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"work_experiences": workExperiences})
}


// func (h *WorkExperienceHandler) UpdateWorkExperience(c *gin.Context) {
//     userID := c.MustGet("userID").(string)
//     db := c.MustGet("db").(*mongo.Database)
//     seekersCollection := db.Collection("seekers")

//     id := c.Param("id") // Assume this is the work experience index or unique identifier (better to use index in your case)

//     var input dto.WorkExperienceRequest
//     if err := c.ShouldBindJSON(&input); err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
//         return
//     }

//     ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//     defer cancel()

//     // Fetch seeker with projection to minimize data
//     var seeker models.Seeker
//     projection := bson.M{"work_experiences": 1}
//     if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}, options.FindOne().SetProjection(projection)).Decode(&seeker); err != nil {
//         if err == mongo.ErrNoDocuments {
//             c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
//         } else {
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving seeker"})
//         }
//         return
//     }

//     // Convert id param to int (index based like your example)
//     index, err := strconv.Atoi(id)
//     if err != nil || index <= 0 || index > len(seeker.WorkExperiences) {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid work experience index"})
//         return
//     }

//     // Replace entry at index-1
// 	seeker.WorkExperiences[index-1] = bson.M{
// 		"auth_user_id":			userID,
// 		"job_title":            input.JobTitle,
// 		"company_name":         input.CompanyName,
// 		"start_date":           input.StartDate,
// 		"end_date":             input.EndDate,
// 		"key_responsibilities": input.KeyResponsibilities,
// 		"location" : 			input.Location,
// 		"created_at":			input.StartDate,
// 		"updated_at":           time.Now(),
// 	}


//     // Save updated seeker document
// 	update := bson.M{
// 		"$set": bson.M{
// 			"work_experiences": seeker.WorkExperiences,
// 		},
// 	}

// 	if _, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update work experience"})
// 		return
// 	}

//     // Return updated entry as response
//     c.JSON(http.StatusOK, gin.H{
//         "message": "Work experience updated successfully",
//     })
// }

// func (h *WorkExperienceHandler) DeleteWorkExperience(c *gin.Context) {
// 	userID := c.MustGet("userID").(string)
// 	db := c.MustGet("db").(*mongo.Database)
// 	seekersCollection := db.Collection("seekers")

// 	id := c.Param("id")

// 	// Parse id as integer index
// 	index, err := strconv.Atoi(id)
// 	if err != nil || index <= 0 {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid work experience index"})
// 		return
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	var seeker struct {
// 		WorkExperiences []bson.M `bson:"work_experiences"`
// 	}

// 	// Fetch current work experiences
// 	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
// 		} else {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve seeker"})
// 		}
// 		return
// 	}

// 	// Check if index is valid
// 	if index > len(seeker.WorkExperiences) {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Work experience index out of range"})
// 		return
// 	}

// 	// Remove the item at index-1
// 	seeker.WorkExperiences = append(seeker.WorkExperiences[:index-1], seeker.WorkExperiences[index:]...)

// 	// Update the seeker document with the new array
// 	update := bson.M{
// 		"$set": bson.M{
// 			"work_experiences": seeker.WorkExperiences,
// 		},
// 	}

// 	if _, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update work experiences"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Work experience deleted successfully"})
// }

