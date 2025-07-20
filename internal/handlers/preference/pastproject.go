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
// CreatePastProject handles creation or update of a single past project entry
func (h *PastProjectHandler) CreatePastProject(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekers := db.Collection("seekers")

	var input dto.PastProjectRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("❌ Bind error [CreatePastProject] user=%s: %v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"issue": "Some required fields are missing or invalid.",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1️⃣ Fetch seeker
	var seeker models.Seeker
	if err := seekers.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		status := http.StatusInternalServerError
		issue := "Could not retrieve your profile. Please try again."

		if err == mongo.ErrNoDocuments {
			status = http.StatusNotFound
			issue = "No account found. It might have been removed or reset."
		}

		log.Printf("❌ DB fetch error [CreatePastProject] user=%s: %v", userID, err)
		c.JSON(status, gin.H{
			"error": err.Error(),
			"issue": issue,
		})
		return
	}

	// 2️⃣ Append to past projects
	if err := repository.AppendToPastProjects(&seeker, input); err != nil {
		log.Printf("❌ Processing error [CreatePastProject] user=%s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Could not save your project details. Try again shortly.",
		})
		return
	}

	// 3️⃣ Update seeker document
	update := bson.M{"$set": bson.M{"past_projects": seeker.PastProjects}}
	updateResult, err := seekers.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update)
	if err != nil || updateResult.MatchedCount == 0 {
		status := http.StatusInternalServerError
		issue := "Failed to update your project records. Please retry."

		if updateResult.MatchedCount == 0 {
			status = http.StatusNotFound
			issue = "Your account couldn't be updated. Please refresh and try again."
		}

		log.Printf("❌ DB update error [CreatePastProject] user=%s: %v", userID, err)
		c.JSON(status, gin.H{
			"error": err.Error(),
			"issue": issue,
		})
		return
	}

	// 4️⃣ Update entry timeline
	if err := func() error {
		_, _, err := repository.UpdateTimelineStepAndCheckCompletion(ctx, db, userID, "past_projects_completed")
		return err
	}(); err != nil {
		log.Printf("⚠️ Timeline update error [CreatePastProject] user=%s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Project saved, but progress tracking failed.",
		})
		return
	}

	// ✅ Success
	c.JSON(http.StatusOK, gin.H{
		"issue": "Project added successfully",
	})
}

// GetPastProjects handles retrieval of a user's past project records
func (h *PastProjectHandler) GetPastProjects(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    db := c.MustGet("db").(*mongo.Database)
    seekersCollection := db.Collection("seekers")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var seeker models.Seeker
    if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
        log.Printf("DB fetch error [GetPastProjects] user=%s: %v", userID, err)
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{
                "error": err.Error(),
                "issue": "Account not found. Please refresh or contact support.",
            })
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": err.Error(),
                "issue": "Failed to retrieve your profile. Please try again.",
            })
        }
        return
    }

    if len(seeker.PastProjects) == 0 {
        c.JSON(http.StatusNoContent, gin.H{"message": "No past projects found"})
        return
    }

    projects, err := repository.GetPastProjects(&seeker)
    if err != nil {
        log.Printf("Processing error [GetPastProjects] user=%s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": err.Error(),
            "issue": "We couldn't load your project history. Please try again.",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{"past_projects": projects})
}
func (h *PastProjectHandler) UpdatePastProject(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	indexStr := c.Param("id")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_index",
			"issue": "Invalid project index",
		})
		return
	}

	var input dto.PastProjectRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"issue": "Invalid input format",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker struct {
		PastProjects []bson.M `bson:"past_projects"`
	}
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		status := http.StatusInternalServerError
		issue := "Failed to retrieve seeker"
		if err == mongo.ErrNoDocuments {
			status = http.StatusNotFound
			issue = "Seeker not found"
		}
		c.JSON(status, gin.H{
			"error": err.Error(),
			"issue": issue,
		})
		return
	}

	if index > len(seeker.PastProjects) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "index_out_of_range",
			"issue": "Project index is out of range",
		})
		return
	}

	seeker.PastProjects[index-1] = bson.M{
		"project_name":        input.ProjectName,
		"institution":         input.Institution,
		"start_date":          input.StartDate,
		"end_date":            input.EndDate,
		"project_description": input.ProjectDescription,
		"updated_at":          time.Now(),
	}

	update := bson.M{
		"$set": bson.M{
			"past_projects": seeker.PastProjects,
		},
		"$currentDate": bson.M{"updated_at": true},
	}

	if _, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Failed to update past project",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issue": "Past project updated successfully",
	})
}

func (h *PastProjectHandler) DeletePastProject(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	indexStr := c.Param("id")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_index",
			"issue": "Invalid project index",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker struct {
		PastProjects []bson.M `bson:"past_projects"`
	}
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		status := http.StatusInternalServerError
		issue := "Failed to retrieve seeker"
		if err == mongo.ErrNoDocuments {
			status = http.StatusNotFound
			issue = "Seeker not found"
		}
		c.JSON(status, gin.H{
			"error": err.Error(),
			"issue": issue,
		})
		return
	}

	if index > len(seeker.PastProjects) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "index_out_of_range",
			"issue": "Project index is out of range",
		})
		return
	}

	seeker.PastProjects = append(seeker.PastProjects[:index-1], seeker.PastProjects[index:]...)

	update := bson.M{
		"$set": bson.M{
			"past_projects": seeker.PastProjects,
		},
		"$currentDate": bson.M{"updated_at": true},
	}

	if _, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Failed to delete past project",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issue": "Past project deleted successfully",
	})
}
