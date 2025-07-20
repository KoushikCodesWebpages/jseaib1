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

func (h *PersonalInfoHandler) CreatePersonalInfo(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var input dto.PersonalInfoRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("❌ Invalid input for user %s: %v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"issue": "Some fields are missing or invalid.",
		})
		return
	}

	seekersColl := db.Collection("seekers")
	var seeker models.Seeker
	if err := seekersColl.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("❌ Seeker not found for user %s", userID)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Seeker not found",
				"issue": "We couldn't find your account. It may have been reset or removed.",
			})
		} else {
			log.Printf("❌ DB error retrieving seeker for user %s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"issue": "Couldn't fetch your account details. Try again.",
			})
		}
		return
	}

	if err := repository.SetPersonalInfo(&seeker, &input); err != nil {
		log.Printf("❌ Failed to process personal info for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "We couldn't process your personal info. Please retry.",
		})
		return
	}

	update := bson.M{"$set": bson.M{"personal_info": seeker.PersonalInfo}}
	res, err := seekersColl.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update)
	if err != nil || res.MatchedCount == 0 {
		log.Printf("❌ Failed to update seeker for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save personal info",
			"issue": "Your information couldn't be saved. Please try again.",
		})
		return
	}

	// ✅ Update timeline and get next step
	_ , _, err = repository.UpdateTimelineStepAndCheckCompletion(
		ctx, db, userID, "personal_info_completed",
	)
	if err != nil {
		log.Printf("⚠️ Timeline update failed for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Details saved, but progress tracking failed.",
		})
		return
	}

	response := gin.H{"issue": "Personal info saved successfully"}
	if repository.IsFieldFilled(seeker.PersonalInfo) {
		response["issue"] = "Personal info updated successfully"
	}

	c.JSON(http.StatusOK, response)
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
