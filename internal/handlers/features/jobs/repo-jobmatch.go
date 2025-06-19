package jobs
// import (
// 	"RAAS/models"
// 	"gorm.io/gorm"
// 	"strings"
// 	"math"
// 	"encoding/json"
// 	"github.com/google/uuid"
// )

// // CalculateMatchScore calculates the match score based on the job and user data
// func CalculateMatchScore(db *gorm.DB, jobID string, userID uuid.UUID) float64 {
// 	// Fetch user skills from their professional summary
// 	var userSummary models.ProfessionalSummary
// 	if err := db.Where("auth_user_id = ?", userID).First(&userSummary).Error; err != nil {
// 		return 0 // No skills data available
// 	}

// 	// Assuming Skills is stored as a datatypes.JSON field, let's unmarshal it into a slice of strings
// 	var userSkills []string
// 	if err := json.Unmarshal(userSummary.Skills, &userSkills); err != nil {
// 		return 0 // Error unmarshalling skills
// 	}

// 	// Fetch the job description, skills, and type (from LinkedIn or Xing)
// 	var jobDesc models.LinkedInJobDescription // Assuming LinkedIn for example, adjust for Xing
// 	if err := db.Where("job_id = ?", jobID).First(&jobDesc).Error; err != nil {
// 		return 0
// 	}

// 	// Split the job description's skills into a slice
// 	jobSkills := strings.Split(jobDesc.Skills, ",")

// 	// Count matching skills
// 	matchingSkills := 0
// 	for _, userSkill := range userSkills {
// 		for _, jobSkill := range jobSkills {
// 			if strings.Contains(strings.ToLower(jobSkill), strings.ToLower(userSkill)) {
// 				matchingSkills++
// 			}
// 		}
// 	}

// 	// For simplicity, let's assume the score is directly proportional to the number of matched skills
// 	matchScore := float64(matchingSkills) / float64(len(jobSkills)) * 100

// 	// Return the match score
// 	return math.Round(matchScore*100) / 100
// }
