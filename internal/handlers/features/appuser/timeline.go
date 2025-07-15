package appuser

import (

	"RAAS/internal/models"
	"RAAS/internal/handlers/features/jobs"
	"RAAS/internal/handlers/repository"

	"fmt"
	"net/http"
	"log"
	"context"
	"time"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetNextEntryStep handles fetching the next incomplete step in the user entry timeline for MongoDB
func GetNextEntryStep() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := c.MustGet("db").(*mongo.Database)
		userID := c.MustGet("userID").(string)
		seekersColl := db.Collection("seekers")
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    	defer cancel()
		if db == nil {
			fmt.Println("Error: MongoDB database is nil")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database unavailable"})
			return
		}
		
		collection := db.Collection("user_entry_timelines")
		var timeline models.UserEntryTimeline
		err := collection.FindOne(c, bson.M{"auth_user_id": userID}).Decode(&timeline)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				fmt.Println("Error fetching timeline: User not found")
				c.JSON(http.StatusNotFound, gin.H{"error": "Timeline not found"})
			} else {
				fmt.Println("Error fetching timeline:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch timeline"})
			}
			return
		}

		steps := []struct {
			Name      string
			Completed bool
			Required  bool
		}{
			{"personal_infos", timeline.PersonalInfoCompleted, timeline.PersonalInfoRequired},
			{"work_experiences", timeline.WorkExperiencesCompleted, timeline.WorkExperiencesRequired},
			{"academics", timeline.AcademicsCompleted, timeline.AcademicsRequired},
			{"past_projects", timeline.PastProjectsCompleted, timeline.PastProjectsRequired},
			{"certificates", timeline.CertificatesCompleted, timeline.CertificatesRequired},
			{"languages", timeline.LanguagesCompleted, timeline.LanguagesRequired},
			{"preferred_job_titles", timeline.JobTitlesCompleted, timeline.JobTitlesRequired},
			{"key_skills", timeline.KeySkillsCompleted, timeline.KeySkillsRequired},
		}
	
		for _, step := range steps {
			if step.Required && !step.Completed {
				c.JSON(http.StatusOK, gin.H{
					"completed": false,
					"next_step": step.Name,
				})
				return
			}
		}

		if !timeline.Completed {
			update := bson.M{
				"$set": bson.M{"completed": true},
			}

			_, err := collection.UpdateOne(c, bson.M{"auth_user_id": userID}, update)
			if err != nil {
				fmt.Println("Error updating timeline:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as completed"})
				return
			}
		}

		    // 7️⃣ Load seeker profile
		var seeker models.Seeker
		if err := seekersColl.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"issue": "Seeker not found"})
				log.Printf("Seeker not found for auth_user_id: %s", userID)
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to retrieve seeker"})
				log.Printf("Failed to retrieve seeker for auth_user_id: %s, Error: %v", userID, err)
			}
			return
		}
		
		completion, missing := repository.CalculateJobProfileCompletion(seeker)
		if completion == 100 || len(missing) == 0 {
			if err = jobs.StartJobMatchScoreCalculation(c, db, userID); err != nil {
				log.Printf("Error starting job match process: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start job match process"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"completed": true,
			"next_step": nil,
		})
	}
}
