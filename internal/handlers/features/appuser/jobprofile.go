package appuser

import (
	"context"
	"log"
	"net/http"
	"time"
	"fmt"

	"RAAS/internal/handlers/repository"
	"RAAS/internal/models"
	"RAAS/internal/dto"
	"RAAS/internal/handlers/features/jobs"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SeekerHandler struct{}

func NewSeekerHandler() *SeekerHandler {
	return &SeekerHandler{}
}
// GetSeekerProfile fetches the seeker profile and returns a DTO
func (h *SeekerHandler) GetSeekerProfile(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"issue": "Seeker not found"})
			log.Printf("Seeker not found for auth_user_id: %s", userID)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to retrieve seeker"})
			log.Printf("Failed to retrieve seeker for auth_user_id: %s, Error: %v", userID, err)
		}
		return
	}

	// Calculate profile completion
	completion, missing := repository.CalculateJobProfileCompletion(seeker)

	// Build DTO
	dto := dto.SeekerDTO{
		ID:                    		seeker.ID,
		AuthUserID:            		seeker.AuthUserID,
		StripeCustomerID: 			seeker.StripeCustomerID,	
		PhotoUrl:                 	seeker.PhotoUrl,
		TotalApplications:     		seeker.TotalApplications,
		WeeklyAppliedJobs:     		seeker.WeeklyAppliedJobs,
		TopJobs:               		seeker.TopJobs,
		SubscriptionTier:      		seeker.SubscriptionTier,
		SubscriptionPeriod:    		seeker.SubscriptionPeriod,
		SubscriptionIntervalStart: 	seeker.SubscriptionIntervalStart,
		SubscriptionIntervalEnd:   	seeker.SubscriptionIntervalEnd,
		ExternalApplications:  		seeker.ExternalApplications,
		InternalApplications:  		seeker.InternalApplications,
		ProficiencyTest:      		seeker.ProficiencyTest,
		PersonalInfo:          		seeker.PersonalInfo,
		WorkExperiences:       		seeker.WorkExperiences,
		Academics:             		seeker.Academics,
		PastProjects:          		seeker.PastProjects,
		Certificates:          		seeker.Certificates,
		Languages:             		seeker.Languages,
		KeySkills:             		seeker.KeySkills,
		PrimaryTitle:          		seeker.PrimaryTitle,
		SecondaryTitle:        		seeker.SecondaryTitle,
		TertiaryTitle:         		seeker.TertiaryTitle,
		CvFormat:              		seeker.CvFormat,
		ClFormat:              		seeker.ClFormat,
		CreatedAt:             		seeker.CreatedAt,
		UpdatedAt:             		seeker.UpdatedAt,
		FirstName:             		repository.DereferenceString(repository.GetOptionalField(seeker.PersonalInfo, "first_name")),
		ProfileCompletion:     		completion,
	}

	matchscore := false
		if completion == 100 || len(missing) == 0 {
			matchscore=true

			fmt.Println("starting match score calculation")
			// ✅ Trigger job match score calculation
			err := jobs.StartJobMatchScoreCalculation(c, db, userID)
			if err != nil {
				fmt.Println("Error starting job match score calculation:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start job match process"})
				return
		}
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"seeker":            dto,
		"profile_completion": completion,
		"not_completed":     missing,
		"match_score":matchscore,
	})
}
