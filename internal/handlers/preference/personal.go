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
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")
	entryTimelineCollection := db.Collection("user_entry_timelines")

	var input dto.PersonalInfoRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Error binding input for user %s: %v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"issue": "Some fields are missing or invalid.",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Seeker not found for user %s", userID)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "no seeker found",
				"issue": "We couldn't find your account. It may have been reset or removed.",
			})
		} else {
			log.Printf("Error retrieving seeker for user %s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"issue": "Couldn't fetch your account details. Try again.",
			})
		}
		return
	}

	if err := repository.SetPersonalInfo(&seeker, &input); err != nil {
		log.Printf("Failed to process personal info for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "We couldn't process your personal info. Please retry.",
		})
		return
	}

	updateResult, err := seekersCollection.UpdateOne(ctx,
		bson.M{"auth_user_id": userID},
		bson.M{"$set": bson.M{"personal_info": seeker.PersonalInfo}},
	)
	if err != nil {
		log.Printf("Failed to update seeker for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Couldn't save your details due to a system issue.",
		})
		return
	}

	if updateResult.MatchedCount == 0 {
		log.Printf("No matching seeker found to update for user %s", userID)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "no document matched for auth_user_id " + userID,
			"issue": "Your account couldn't be found to update.",
		})
		return
	}

	message := "Personal info saved successfully"
	if repository.IsFieldFilled(seeker.PersonalInfo) {
		message = "Personal info updated successfully"
	}

	_, err = entryTimelineCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, bson.M{
		"$set": bson.M{"personal_info_completed": true},
	})
	if err != nil {
		log.Printf("Failed to update entry timeline for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Details saved, but progress tracking failed.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issue": message,
	})
}

// GetPersonalInfo retrieves the personal information
func (h *PersonalInfoHandler) GetPersonalInfo(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)

	seekerCollection := db.Collection("seekers")
	authUserCollection := db.Collection("auth_users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	err := seekerCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker)
	if err != nil {
		log.Printf("Seeker fetch failed for user %s: %v", userID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"issue": "We couldn't find your profile. Your Account is unavailable.",
		})
		return
	}

	var authUser struct {
		Email string `bson:"email"`
		Phone string `bson:"phone"`
	}
	err = authUserCollection.FindOne(ctx, bson.M{"auth_user_id": seeker.AuthUserID}).Decode(&authUser)
	if err != nil {
		log.Printf("Failed to fetch auth user for %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Could not load your contact details. Your Account is unavailable.",
		})
		return
	}

	if seeker.PersonalInfo == nil || !repository.IsFieldFilled(seeker.PersonalInfo) {
		c.JSON(http.StatusOK, gin.H{
			"email": authUser.Email,
			"phone": authUser.Phone,
		})
		return
	}

	personalInfo, err := repository.GetPersonalInfo(&seeker)
	if err != nil {
		log.Printf("Failed to unmarshal personal info for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "We couldn't load your personal details. System data has been corrupted.",
		})
		return
	}

	personalInfo.AuthUserID = userID
	personalInfo.Email = authUser.Email
	personalInfo.Phone = authUser.Phone

	c.JSON(http.StatusOK, gin.H{
		"personal_info": personalInfo,
	})
}
