package preference

import (

	"RAAS/internal/dto"
	"RAAS/internal/models"
	"RAAS/internal/handlers/repository"

	"context"
	"log"
	"net/http"
	"time"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		log.Printf("Error binding input: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
			log.Printf("Seeker not found for auth_user_id: %s", userID)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving seeker"})
			log.Printf("Error retrieving seeker for auth_user_id: %s, Error: %v", userID, err)
		}
		return
	}

	// Create a dto.WorkExperienceRequest from the input
	workExperience := dto.WorkExperienceRequest{
		JobTitle:           	input.JobTitle,
		CompanyName:        	input.CompanyName,
		StartDate:          	input.StartDate,
		EndDate:            	input.EndDate,
		Location: 				input.Location,
		KeyResponsibilities: 	input.KeyResponsibilities,
	}

	// Use AppendToWorkExperience to add the new experience
	if err := repository.AppendToWorkExperience(&seeker, workExperience); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process work experience"})
		log.Printf("Failed to process work experience for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	// Update seeker document with the new work experiences
	update := bson.M{
		"$set": bson.M{
			"work_experiences": seeker.WorkExperiences, // Save updated work experiences
		},
	}

	updateResult, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save work experience"})
		log.Printf("Failed to update work experience for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	if updateResult.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching seeker found to update"})
		log.Printf("No matching seeker found for auth_user_id: %s", userID)
		return
	}

	// Update user entry timeline to mark work experiences completed
	timelineUpdate := bson.M{
		"$set": bson.M{
			"work_experiences_completed": true,
		},
	}

	if _, err := entryTimelineCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, timelineUpdate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user entry timeline"})
		log.Printf("Failed to update user entry timeline for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Work experience added successfully",
	})
}

// GetWorkExperienceHandler handles the retrieval of a user's work experiences
func (h *WorkExperienceHandler) GetWorkExperience(c *gin.Context) {
    // Extract user ID from the context
    userID := c.MustGet("userID").(string)
    db := c.MustGet("db").(*mongo.Database)
    seekersCollection := db.Collection("seekers")

    // Set up a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Find the seeker by their auth_user_id
    var seeker models.Seeker
    if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
            log.Printf("Seeker not found for auth_user_id: %s", userID)
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving seeker"})
            log.Printf("Error retrieving seeker for auth_user_id: %s, Error: %v", userID, err)
        }
        return
    }

    // Check if the user has work experiences
    if len(seeker.WorkExperiences) == 0 {
        c.JSON(http.StatusNoContent, gin.H{"message": "No work experiences found"})
        return
    }

    // Convert bson.M to the expected dto.WorkExperienceRequest type
    workExperiences, err := repository.GetWorkExperience(&seeker)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing work experiences"})
        log.Printf("Error processing work experiences for auth_user_id: %s, Error: %v", userID, err)
        return
    }

    // Return the work experiences in the response
    c.JSON(http.StatusOK, gin.H{
        "work_experiences": workExperiences,
    })
}


func (h *WorkExperienceHandler) UpdateWorkExperience(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    db := c.MustGet("db").(*mongo.Database)
    seekersCollection := db.Collection("seekers")

    id := c.Param("id") // Assume this is the work experience index or unique identifier (better to use index in your case)

    var input dto.WorkExperienceRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Fetch seeker with projection to minimize data
    var seeker models.Seeker
    projection := bson.M{"work_experiences": 1}
    if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}, options.FindOne().SetProjection(projection)).Decode(&seeker); err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving seeker"})
        }
        return
    }

    // Convert id param to int (index based like your example)
    index, err := strconv.Atoi(id)
    if err != nil || index <= 0 || index > len(seeker.WorkExperiences) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid work experience index"})
        return
    }

    // Replace entry at index-1
	seeker.WorkExperiences[index-1] = bson.M{
		"auth_user_id":			userID,
		"job_title":            input.JobTitle,
		"company_name":         input.CompanyName,
		"start_date":           input.StartDate,
		"end_date":             input.EndDate,
		"key_responsibilities": input.KeyResponsibilities,
		"location" : 			input.Location,
		"created_at":			input.StartDate,
		"updated_at":           time.Now(),
	}


    // Save updated seeker document
	update := bson.M{
		"$set": bson.M{
			"work_experiences": seeker.WorkExperiences,
		},
	}

	if _, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update work experience"})
		return
	}

    // Return updated entry as response
    c.JSON(http.StatusOK, gin.H{
        "message": "Work experience updated successfully",
    })
}

func (h *WorkExperienceHandler) DeleteWorkExperience(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	id := c.Param("id")

	// Parse id as integer index
	index, err := strconv.Atoi(id)
	if err != nil || index <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid work experience index"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker struct {
		WorkExperiences []bson.M `bson:"work_experiences"`
	}

	// Fetch current work experiences
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve seeker"})
		}
		return
	}

	// Check if index is valid
	if index > len(seeker.WorkExperiences) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Work experience index out of range"})
		return
	}

	// Remove the item at index-1
	seeker.WorkExperiences = append(seeker.WorkExperiences[:index-1], seeker.WorkExperiences[index:]...)

	// Update the seeker document with the new array
	update := bson.M{
		"$set": bson.M{
			"work_experiences": seeker.WorkExperiences,
		},
	}

	if _, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update work experiences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Work experience deleted successfully"})
}

