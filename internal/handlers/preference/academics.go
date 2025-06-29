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
// CreateAcademics creates or updates an academic entry
func (h *AcademicsHandler) CreateAcademics(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekers := db.Collection("seekers")
	timelines := db.Collection("user_entry_timelines")

	var input dto.AcademicsRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Bind error [CreateAcademics] user=%s: %v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"issue": "Some fields are missing or contain invalid values.",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := seekers.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Seeker not found [CreateAcademics] user=%s", userID)
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
				"issue": "We couldn't find your account. Please contact support.",
			})
		} else {
			log.Printf("DB fetch error [CreateAcademics] user=%s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"issue": "Something went wrong while retrieving your profile. Please try again.",
			})
		}
		return
	}

	if err := repository.AppendToAcademics(&seeker, input); err != nil {
		log.Printf("Processing error [CreateAcademics] user=%s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Could not save your educational record. Try again shortly.",
		})
		return
	}

	updateResult, err := seekers.UpdateOne(ctx,
		bson.M{"auth_user_id": userID},
		bson.M{"$set": bson.M{"academics": seeker.Academics}},
	)
	if err != nil {
		log.Printf("DB update error [CreateAcademics] user=%s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Failed to update your education records. Please retry.",
		})
		return
	}
	if updateResult.MatchedCount == 0 {
		log.Printf("No document matched [CreateAcademics] user=%s", userID)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "no seeker updated for " + userID,
			"issue": "Your account couldn't be updated. Please refresh and try again.",
		})
		return
	}

	if _, err := timelines.UpdateOne(ctx, bson.M{"auth_user_id": userID},
		bson.M{"$set": bson.M{"academics_completed": true}},
	); err != nil {
		log.Printf("Timeline update error [CreateAcademics] user=%s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Education saved, but progress tracking failed.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"issue": "Academics added successfully"})
}

// GetAcademics retrieves a user's academics records
func (h *AcademicsHandler) GetAcademics(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekers := db.Collection("seekers")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := seekers.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Seeker not found [GetAcademics] user=%s", userID)
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
				"issue": "Account not found. Please refresh or contact support.",
			})
		} else {
			log.Printf("DB fetch error [GetAcademics] user=%s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"issue": "Failed to retrieve your profile. Please try again.",
			})
		}
		return
	}

	if len(seeker.Academics) == 0 {
		c.JSON(http.StatusNoContent, gin.H{"message": "No academic records found"})
		return
	}

	academics, err := repository.GetAcademics(&seeker)
	if err != nil {
		log.Printf("Processing error [GetAcademics] user=%s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "We couldn't load your education history. Please try again.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"academics": academics})
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
