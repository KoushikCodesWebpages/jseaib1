package appuser

import (
	"RAAS/internal/dto"
	"RAAS/internal/handlers/repository"
	"RAAS/internal/models"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SeekerProfileHandler struct{}

func NewSeekerProfileHandler() *SeekerProfileHandler {
	return &SeekerProfileHandler{}
}

// GetSeekerProfile retrieves the profile for the authenticated user
func (h *SeekerProfileHandler) GetSeekerProfile(c *gin.Context) {
	// Get authenticated user ID and db from context
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)

	seekersCollection := db.Collection("seekers")

	// Set a timeout for the MongoDB operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the seeker by auth_user_id
	var seeker models.Seeker
	err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker profile not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving seeker profile"})
		}
		log.Printf("Error retrieving seeker profile for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	// Extract language names
	var languageNames []string
	for _, language := range seeker.Languages {
		// Ensure 'language' is a map (bson.M), and access the "language" key
		if lang, ok := language["language"].(string); ok {
			languageNames = append(languageNames, lang) // Collect the language name
		} else {
			log.Printf("[WARN] Invalid or missing 'language' field in languages array")
		}
	}



	workExperiencesBson, err := repository.GetWorkExperience(&seeker)
	if err != nil {
		log.Fatalf("Error retrieving work experiences: %v", err)
	}

	workExperiences, err := repository.ConvertBsonMToWorkExperienceRequest(workExperiencesBson)
	if err != nil {
		log.Fatalf("Error converting work experiences: %v", err)
	}

	totalMonths, err := repository.GetExperienceInMonths(workExperiences)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error calculating total work experience"})
		log.Printf("Error calculating total work experience for userID %s: %v", userID, err)
		return
	}
	// Map seeker to SeekerProfileDTO
	profile := dto.SeekerProfileDTO{
		AuthUserID:                  seeker.AuthUserID,
		FirstName:                   repository.DereferenceString(repository.GetOptionalField(seeker.PersonalInfo, "first_name")),
		SecondName:                  repository.GetOptionalField(seeker.PersonalInfo, "second_name"),
		Skills:                      seeker.KeySkills,
		TotalExperienceInMonths:     totalMonths, // Set the correct total experience here
		Certificates:                repository.ExtractCertificates(seeker.Certificates),
		PreferredJobTitle:           seeker.PrimaryTitle,
		SubscriptionTier:            seeker.SubscriptionTier,
		DailySelectableJobsCount:    seeker.DailySelectableJobsCount,
		DailyGeneratableCV:          seeker.DailyGeneratableCV,
		DailyGeneratableCoverletter: seeker.DailyGeneratableCoverletter,
		TotalApplications:           seeker.TotalApplications,
		TotalJobsAvailable:          0, // For now, as you said
		ProfileCompletion:           repository.CalculateProfileCompletion(seeker),
		Languages:                   languageNames,
	}

	// Send the profile as a response
	c.JSON(http.StatusOK, profile)
}
