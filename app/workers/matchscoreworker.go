package workers

// import (
// 	"fmt"
// 	"log"
// 	"RAAS/models"
// 	"gorm.io/gorm"
// 	"time"
// 	"strings"
// 	"github.com/google/uuid"
// )

// // MatchScoreWorker continuously calculates match scores and stores them
// type MatchScoreWorker struct {
// 	DB *gorm.DB
// }

// func (w *MatchScoreWorker) calculateAndStoreMatchScore(seekerAuthUserID uuid.UUID, jobID string) error {
// 	log.Printf("📊 Starting match score calculation for Seeker: %s, Job: %s", seekerAuthUserID, jobID)

// 	// Fetch seeker details
// 	var seeker models.Seeker
// 	if err := w.DB.Where("auth_user_id = ?", seekerAuthUserID).First(&seeker).Error; err != nil {
// 		log.Printf("❌ Failed to fetch seeker: %v", err)
// 		return fmt.Errorf("failed to find seeker: %v", err)
// 	}
// 	log.Printf("✅ Fetched seeker: %s", seeker.AuthUserID)

// 	// Fetch job details
// 	var job models.Job
// 	if err := w.DB.Where("job_id = ?", jobID).First(&job).Error; err != nil {
// 		log.Printf("❌ Failed to fetch job: %v", err)
// 		return fmt.Errorf("failed to find job: %v", err)
// 	}
// 	log.Printf("✅ Fetched job: %s - %s", job.JobID, job.Title)

// 	// Calculate the match score
// 	matchScore, err := CalculateMatchScore(seeker, job)
// 	if err != nil {
// 		log.Printf("❌ Failed to calculate match score: %v", err)
// 		return fmt.Errorf("failed to calculate match score: %v", err)
// 	}
// 	log.Printf("✅ Calculated match score: %.2f for Seeker: %s and Job: %s", matchScore, seeker.AuthUserID, job.JobID)

// 	// Create and save match score
// 	matchScoreEntry := models.MatchScore{
// 		AuthUserID:   seeker.AuthUserID,
// 		JobID:      jobID,
// 		MatchScore: matchScore,
// 	}
// 	log.Printf("💾 Saving match score entry for Seeker: %s, Job: %s", seeker.AuthUserID, jobID)

// 	if err := w.DB.Save(&matchScoreEntry).Error; err != nil {
// 		log.Printf("❌ Failed to save match score to DB: %v", err)
// 		return fmt.Errorf("failed to save match score: %v", err)
// 	}
// 	log.Printf("✅ Match score saved successfully for Seeker: %s, Job: %s", seeker.AuthUserID, jobID)

// 	return nil
// }


// func (w *MatchScoreWorker) Run() {
// 	log.Println("✅ MatchScoreWorker started running...")

// 	for {
// 		log.Println("🔄 Starting new cycle...")

// 		// 1. Fetch all seekers
// 		var seekers []models.Seeker
// 		if err := w.DB.Find(&seekers).Error; err != nil {
// 			log.Printf("❌ Error fetching seekers: %v", err)
// 			time.Sleep(time.Minute)
// 			continue
// 		}
// 		log.Printf("🔍 Found %d seekers", len(seekers))

// 		// 2. Loop through all seekers
// 		for _, seeker := range seekers {
// 			log.Printf("👤 Processing seeker: %s", seeker.AuthUserID)

// 			// Collect preferred titles
// 			var preferredTitles []string
// 			if seeker.PrimaryTitle != "" {
// 				preferredTitles = append(preferredTitles, seeker.PrimaryTitle)
// 			}
// 			if seeker.SecondaryTitle != nil && *seeker.SecondaryTitle != "" {
// 				preferredTitles = append(preferredTitles, *seeker.SecondaryTitle)
// 			}
// 			if seeker.TertiaryTitle != nil && *seeker.TertiaryTitle != "" {
// 				preferredTitles = append(preferredTitles, *seeker.TertiaryTitle)
// 			}

// 			log.Printf("🎯 Preferred titles for seeker %s: %v", seeker.AuthUserID, preferredTitles)

// 			if len(preferredTitles) == 0 {
// 				log.Printf("⚠️ No preferred job titles for seeker %s, skipping", seeker.AuthUserID)
// 				continue
// 			}

// 			// Build job title filtering query
// 			var conditions []string
// 			var values []interface{}
// 			for _, title := range preferredTitles {
// 				conditions = append(conditions, "LOWER(title) LIKE ?")
// 				values = append(values, "%"+strings.ToLower(title)+"%")
// 			}
// 			whereClause := strings.Join(conditions, " OR ")

// 			// 3. Fetch matching jobs
// 			var jobs []models.Job
// 			if err := w.DB.Where(whereClause, values...).Find(&jobs).Error; err != nil {
// 				log.Printf("❌ Error fetching jobs for seeker %s: %v", seeker.AuthUserID, err)
// 				continue
// 			}
// 			log.Printf("📄 Found %d matching jobs for seeker %s", len(jobs), seeker.AuthUserID)

// 			// 4. Process each job
// 			for _, job := range jobs {
// 				log.Printf("⚙️ Checking job: %s - %s", job.JobID, job.Title)

// 				// Check if match score already exists
// 				var existingMatchScore models.MatchScore
// 				if err := w.DB.Where("seeker_id = ? AND job_id = ?", seeker.AuthUserID, job.JobID).First(&existingMatchScore).Error; err == nil {
// 					log.Printf("⏭️ Match score already exists for seeker %s and job %s, skipping", seeker.AuthUserID, job.JobID)
// 					continue
// 				} else {
// 					log.Printf("➕ No existing match score, calculating new one...")
// 				}

// 				// Calculate and store match score
// 				if err := w.calculateAndStoreMatchScore(seeker.AuthUserID, job.JobID); err != nil {
// 					log.Printf("❌ Error calculating match score for seeker %s and job %s: %v", seeker.AuthUserID, job.JobID, err)
// 				} else {
// 					log.Printf("✅ Match score calculated and saved for seeker %s and job %s", seeker.AuthUserID, job.JobID)
// 				}
// 			}
// 		}

// 		log.Println("🛌 MatchScoreWorker completed cycle. Sleeping 1 minute...")
// 		time.Sleep(time.Minute)
// 	}
// }
