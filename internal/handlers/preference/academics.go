package preference

import (
	"context"
	"log"
	"net/http"
	"time"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"RAAS/internal/dto"
	"RAAS/internal/models"
	"RAAS/internal/handlers/repository"
)

type AcademicsHandler struct{}

func NewAcademicsHandler() *AcademicsHandler {
	return &AcademicsHandler{}
}

// CreateEducation handles the creation or update of a single education entry
func (h *AcademicsHandler) CreateAcademics(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")
	entryTimelineCollection := db.Collection("user_entry_timelines")

	var input dto.AcademicsRequest
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

	// Create an EducationRequest from the input
	academics := dto.AcademicsRequest{
		Degree:       	input.Degree,
		Institution:  	input.Institution,
		FieldOfStudy: 	input.FieldOfStudy,
		City: 			input.City,
		StartDate:    	input.StartDate,
		EndDate:      	input.EndDate,
		Description: 	input.Description,
	}

	// Use AppendToEducation to add the new education
	if err := repository.AppendToAcademics(&seeker, academics); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process education"})
		log.Printf("Failed to process education for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	// Update seeker document with the new education
	update := bson.M{
		"$set": bson.M{
			"academics": seeker.Academics, // Save updated education records
		},
	}

	updateResult, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save education"})
		log.Printf("Failed to update education for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	if updateResult.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching seeker found to update"})
		log.Printf("No matching seeker found for auth_user_id: %s", userID)
		return
	}

	// Update user entry timeline to mark education completed
	timelineUpdate := bson.M{
		"$set": bson.M{
			"academics_completed": true,
		},
	}

	if _, err := entryTimelineCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, timelineUpdate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user entry timeline"})
		log.Printf("Failed to update user entry timeline for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Education added successfully",
	})
}

// GetEducationHandler handles the retrieval of a user's education records
func (h *AcademicsHandler) GetAcademics(c *gin.Context) {
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
    // Check if the user has any education records
    if len(seeker.Academics) == 0 {
        c.JSON(http.StatusNoContent, gin.H{"message": "No academics records found"})
        return
    }

    // Fetch the education data (could be a function similar to GetWorkExperience)
    academics, err := repository.GetAcademics(&seeker)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing academics records"})
        log.Printf("Error processing academics records for auth_user_id: %s, Error: %v", userID, err)
        return
    }

    // Return the education data in the response
    c.JSON(http.StatusOK, gin.H{
        "academics": academics,
    })
}

func (h *AcademicsHandler) UpdateAcademics(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	id := c.Param("id")

	var input dto.AcademicsRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	index, err := strconv.Atoi(id)
	if err != nil || index <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid education index"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve seeker"})
		}
		return
	}

	if index > len(seeker.Academics) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Education index out of range"})
		return
	}

	// Replace the education at index-1
	seeker.Academics[index-1] = bson.M{
		"degree":        	input.Degree,
		"institution":   	input.Institution,
		"city":				input.City,
		"field_of_study": 	input.FieldOfStudy,
		"start_date":    	input.StartDate,
		"end_date":      	input.EndDate,
		"achievements":  	input.Description,
		"updated_at":		time.Now(),
	}

	update := bson.M{
		"$set": bson.M{
			"academics": seeker.Academics,
		},
	}

	if _, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update academics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Academics updated successfully"})
}

func (h *AcademicsHandler) DeleteAcademics(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	id := c.Param("id")

	index, err := strconv.Atoi(id)
	if err != nil || index <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid academics index"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve seeker"})
		}
		return
	}

	if index > len(seeker.Academics) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Academics index out of range"})
		return
	}

	// Remove the education entry at index-1
	seeker.Academics = append(seeker.Academics[:index-1], seeker.Academics[index:]...)

	update := bson.M{
		"$set": bson.M{
			"academics": seeker.Academics,
		},
	}

	if _, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete academics entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Academics deleted successfully"})
}
