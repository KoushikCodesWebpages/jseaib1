package preference

import (

	"RAAS/internal/dto"
	"RAAS/internal/models"
	"RAAS/internal/handlers/repository"

	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

)

type PersonalInfoHandler struct{}

func NewPersonalInfoHandler() *PersonalInfoHandler {
	return &PersonalInfoHandler{}
}

// CreatePersonalInfo handles the creation of personal information
func (h *PersonalInfoHandler) CreatePersonalInfo(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	
	seekersCollection := c.MustGet("db").(*mongo.Database).Collection("seekers")
	entryTimelineCollection := c.MustGet("db").(*mongo.Database).Collection("user_entry_timelines")

	var input dto.PersonalInfoRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		log.Printf("Error binding input: %s", err.Error())
		return
	}

	// Set context and timeout for MongoDB operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	// Fetch seeker by "auth_user_id"
	err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
			log.Printf("Seeker not found for auth_user_id: %s", userID)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving seeker"})
			log.Printf("Error retrieving seeker for auth_user_id: %s, Error: %s", userID, err.Error())
		}
		return
	}

	// Process and set personal info using the new reusable function
	if err := repository.SetPersonalInfo(&seeker, &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process personal info"})
		log.Printf("Failed to process personal info for auth_user_id: %s, Error: %s", userID, err.Error())
		return
	}

	// Update personal info in MongoDB (seekers collection)
	updateResult, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, bson.M{"$set": bson.M{"personal_info": seeker.PersonalInfo}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save personal info"})
		log.Printf("Failed to update personal info for auth_user_id: %s, Error: %s", userID, err.Error())
		return
	}

	// Check if any document was matched
	if updateResult.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching seeker found to update"})
		log.Printf("No matching seeker found for auth_user_id: %s", userID)
		return
	}

	// Determine the message based on whether we were creating or updating
	message := "Personal info created"
	if repository.IsFieldFilled(seeker.PersonalInfo) {
		message = "Personal info updated"
	}

	// Now, update the user entry progress in user_entry_timelines collection
	// Set `personal_infos_completed` to true
	timelineUpdate := bson.M{
		"$set": bson.M{
			"personal_info_completed": true, // Mark personal info as completed
		},
	}

	// Update the user entry timeline for the specific user
	_, err = entryTimelineCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, timelineUpdate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user entry timeline"})
		log.Printf("Failed to update user entry timeline for auth_user_id: %s, Error: %s", userID, err.Error())
		return
	}


	// Respond with appropriate message
	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

// GetPersonalInfo retrieves the personal information
func (h *PersonalInfoHandler) GetPersonalInfo(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	collection := c.MustGet("db").(*mongo.Database).Collection("seekers")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	err := collection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker)
	if err != nil {
		// Handle case where seeker is not found
		c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
		return
	}

	// Check if personal info is empty
	if seeker.PersonalInfo == nil || !repository.IsFieldFilled(seeker.PersonalInfo) {
		// Respond with 204 No Content and a custom message
		c.JSON(http.StatusNoContent, gin.H{"message": "Personal information not filled"})
		return
	}

	personalInfo, err := repository.GetPersonalInfo(&seeker)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal personal info"})
		return
	}
	// Replace or inject values
	personalInfo.AuthUserID = userID

	// Add email/phone if needed (use pointer to string)
	// Fetch email & phone from auth_user collection
	authUserCollection := c.MustGet("db").(*mongo.Database).Collection("auth_users")

	var authUser struct {
		Email string `bson:"email"`
		Phone string `bson:"phone"`
	}

	err = authUserCollection.FindOne(c, bson.M{"auth_user_id": seeker.AuthUserID}).Decode(&authUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch auth user"})
		return
	}


	personalInfo.Email = authUser.Email
	personalInfo.Phone = authUser.Phone
	c.JSON(http.StatusOK, gin.H{
		"personal_info": personalInfo,
	})
	
}