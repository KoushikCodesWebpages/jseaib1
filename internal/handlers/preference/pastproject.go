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

type PastProjectHandler struct{}

func NewPastProjectHandler() *PastProjectHandler {
	return &PastProjectHandler{}
}

// CreatePastProject handles the creation or update of a single past project entry
func (h *PastProjectHandler) CreatePastProject(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")
	entryTimelineCollection := db.Collection("user_entry_timelines")

	var input dto.PastProjectRequest
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

	// Create a PastProjectRequest from the input
	project := dto.PastProjectRequest{
		ProjectName:        input.ProjectName,
		Institution:        input.Institution,
		StartDate:          input.StartDate,
		EndDate:            input.EndDate,
		ProjectDescription: input.ProjectDescription,
	}

	// Use AppendToPastProjects to add the new project (you need to define this in repository)
	if err := repository.AppendToPastProjects(&seeker, project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process past project"})
		log.Printf("Failed to process past project for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	// Update seeker document with the new past projects
	update := bson.M{
		"$set": bson.M{
			"past_projects": seeker.PastProjects, // Save updated past projects
		},
	}

	updateResult, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save past project"})
		log.Printf("Failed to update past project for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	if updateResult.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching seeker found to update"})
		log.Printf("No matching seeker found for auth_user_id: %s", userID)
		return
	}

	// Update user entry timeline to mark past projects completed
	timelineUpdate := bson.M{
		"$set": bson.M{
			"past_projects_completed": true,
		},
	}

	if _, err := entryTimelineCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, timelineUpdate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user entry timeline"})
		log.Printf("Failed to update user entry timeline for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Past project added successfully",
	})
}

// GetPastProjects handles the retrieval of a user's past project records
func (h *PastProjectHandler) GetPastProjects(c *gin.Context) {
    // Extract user ID from context
    userID := c.MustGet("userID").(string)
    db := c.MustGet("db").(*mongo.Database)
    seekersCollection := db.Collection("seekers")

    // Set timeout context
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Find seeker by auth_user_id
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

    // Check if there are any past projects
    if len(seeker.PastProjects) == 0 {
        c.JSON(http.StatusNoContent, gin.H{"message": "No past projects found"})
        return
    }

    // Use repository layer to format and return the data
    projects, err := repository.GetPastProjects(&seeker)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing past project records"})
        log.Printf("Error processing past projects for auth_user_id: %s, Error: %v", userID, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "past_projects": projects,
    })
}

func (h *PastProjectHandler) UpdatePastProject(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	id := c.Param("id") // expects index (e.g. /update/1)

	var input dto.PastProjectRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	index, err := strconv.Atoi(id)
	if err != nil || index <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project index"})
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

	if index > len(seeker.PastProjects) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project index out of range"})
		return
	}

	// Replace the project at index-1
	seeker.PastProjects[index-1] = bson.M{
		"project_name":        input.ProjectName,
		"institution":         input.Institution,
		"start_date":          input.StartDate,
		"end_date":            input.EndDate,
		"project_description": input.ProjectDescription,
		"created_at":          time.Now(),   // optional: keep original if stored
		"updated_at":          time.Now(),
	}

	update := bson.M{
		"$set": bson.M{
			"past_projects": seeker.PastProjects,
		},
	}

	if _, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update past project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Past project updated successfully"})
}


func (h *PastProjectHandler) DeletePastProject(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	id := c.Param("id")

	index, err := strconv.Atoi(id)
	if err != nil || index <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project index"})
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

	if index > len(seeker.PastProjects) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project index out of range"})
		return
	}

	// Remove the project at index-1
	seeker.PastProjects = append(seeker.PastProjects[:index-1], seeker.PastProjects[index:]...)

	update := bson.M{
		"$set": bson.M{
			"past_projects": seeker.PastProjects,
		},
	}

	if _, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Past project deleted successfully"})
}
