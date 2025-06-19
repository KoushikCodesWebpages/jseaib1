package repository

import (
	"context"
	"RAAS/internal/models"
	"RAAS/internal/dto"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

)

// Fetch seeker only (no skill extraction)
func GetSeekerData(db *mongo.Database, userID string) (models.Seeker, error) {
	var seeker models.Seeker
	err := db.Collection("seekers").FindOne(context.TODO(), bson.M{"auth_user_id": userID}).Decode(&seeker)
	if err != nil {
		return models.Seeker{}, err
	}
	return seeker, nil
}


// Extract preferred titles from seeker
func CollectPreferredTitles(seeker models.Seeker) []string {
	var titles []string
	if seeker.PrimaryTitle != "" {
		titles = append(titles, seeker.PrimaryTitle)
	}
	if seeker.SecondaryTitle != nil && *seeker.SecondaryTitle != "" {
		titles = append(titles, *seeker.SecondaryTitle)
	}
	if seeker.TertiaryTitle != nil && *seeker.TertiaryTitle != "" {
		titles = append(titles, *seeker.TertiaryTitle)
	}
	return titles
}

// Fetch job by job ID
func GetJobByID(db *mongo.Database, jobID string) (models.Job, error) {
	var job models.Job
	err := db.Collection("jobs").FindOne(context.TODO(), bson.M{"job_id": jobID}).Decode(&job)
	if err != nil {
		return models.Job{}, err
	}
	return job, nil
}

func CountJobsByTitles(db *mongo.Database, titles []string) (int64, error) {
	if len(titles) == 0 {
		return 0, fmt.Errorf("no titles provided for counting")
	}

	// Build the $or conditions for matching titles in primary, secondary, tertiary title fields (case-insensitive)
	var orConditions []bson.M
	for _, title := range titles {
		regexFilter := bson.M{
			"$regex":   title,
			"$options": "i", // Case-insensitive match
		}
		orConditions = append(orConditions, bson.M{"primary_title": regexFilter})
		orConditions = append(orConditions, bson.M{"secondary_title": regexFilter})
		orConditions = append(orConditions, bson.M{"tertiary_title": regexFilter})
	}

	// Final filter: $or condition
	filter := bson.M{
		"$or": orConditions,
	}

	log.Printf("[DEBUG] Counting jobs with filter: %+v", filter)

	// Perform the count
	count, err := db.Collection("jobs").CountDocuments(context.TODO(), filter)
	if err != nil {
		log.Printf("[ERROR] Failed to count jobs by titles: %v", err)
		return 0, err
	}

	log.Printf("[DEBUG] Found %d jobs matching given titles", count)
	return count, nil
}



func ExtractKeySkills(seeker bson.M) []string {
	val, ok := seeker["key_skills"].(primitive.A)
	if !ok {
		return []string{}
	}

	skills := make([]string, 0, len(val))
	for _, skill := range val {
		if str, ok := skill.(string); ok {
			skills = append(skills, str)
		}
	}
	return skills
}


// Helper function to fetch saved job IDs
func FetchSavedJobIDs(c *gin.Context, col *mongo.Collection, userID string) ([]string, error) {
	var jobIDs []string
	cursor, err := col.Find(c, bson.M{"auth_user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(c)

	for cursor.Next(c) {
		var saved models.SavedJob
		if err := cursor.Decode(&saved); err == nil {
			jobIDs = append(jobIDs, saved.JobID)
		}
	}
	return jobIDs, nil
}

// Convert BSON to []WorkExperienceRequest
func ConvertBsonMToWorkExperienceRequest(workExperiencesBson []bson.M) ([]dto.WorkExperienceRequest, error) {
	var workExperiences []dto.WorkExperienceRequest

	for _, weBson := range workExperiencesBson {
		var we dto.WorkExperienceRequest

		// Start Date (required)
		startDateRaw, ok := weBson["start_date"].(primitive.DateTime)
		if !ok {
			return nil, fmt.Errorf("missing or invalid start_date")
		}
		we.StartDate = startDateRaw.Time()

		// End Date (optional)
		if endDateRaw, exists := weBson["end_date"]; exists && endDateRaw != nil {
			if endDateDT, ok := endDateRaw.(primitive.DateTime); ok {
				t := endDateDT.Time()
				we.EndDate = &t
			}
		}

		// Other fields
		we.CompanyName, _ = weBson["company_name"].(string)
		we.JobTitle, _ = weBson["job_title"].(string)
		we.Location, _ = weBson["location"].(string)
		we.KeyResponsibilities, _ = weBson["key_responsibilities"].(string)

		workExperiences = append(workExperiences, we)
	}

	return workExperiences, nil
}

// Get total work experience in months
func GetExperienceInMonths(workExperiences []dto.WorkExperienceRequest) (int, error) {
	totalMonths := 0

	for _, we := range workExperiences {
		startDate := we.StartDate
		endDate := time.Now()

		if we.EndDate != nil && !we.EndDate.IsZero() {
			endDate = *we.EndDate
		}

		years, months, _ := CalculateWorkExperience(startDate, endDate)
		totalMonths += years*12 + months
	}

	return totalMonths, nil
}

// Calculate difference in years, months, days
func CalculateWorkExperience(startDate, endDate time.Time) (years, months, days int) {
	if endDate.Before(startDate) {
		return 0, 0, 0
	}

	years = endDate.Year() - startDate.Year()
	months = int(endDate.Month() - startDate.Month())
	days = endDate.Day() - startDate.Day()

	if days < 0 {
		endDate = endDate.AddDate(0, -1, 0)
		days += daysIn(endDate)
		months--
	}

	if months < 0 {
		months += 12
		years--
	}

	return
}

// Get days in a given month
func daysIn(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
}

// Format experience nicely
func FormatExperience(years, months, days int) string {
	return fmt.Sprintf("%d years, %d months, %d days", years, months, days)
}



// Helper function to extract certificates
func ExtractCertificates(certificates []bson.M) []string {
	var result []string
	for _, cert := range certificates {
		if certName, ok := cert["certificate_name"].(string); ok {
			result = append(result, certName)
		}
	}
	return result
}

