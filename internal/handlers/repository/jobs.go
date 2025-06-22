package repository

import (
	"RAAS/internal/models"

	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func IsJobSelected(c context.Context, db *mongo.Database, userID, jobID string) bool {
	collection := db.Collection("selected_job_applications")

	filter := bson.M{
		"auth_user_id": userID,
		"job_id":       jobID,
	}

	err := collection.FindOne(c, filter).Err()
	if err == mongo.ErrNoDocuments {
		return false
	}
	if err != nil {
		fmt.Println("Error checking selected job:", err)
		return false
	}
	return true
}


func GetMatchScoreForJob(c context.Context, db *mongo.Database, userID, jobID string) float64 {
	collection := db.Collection("match_scores")

	var result models.MatchScore
	err := collection.FindOne(c, bson.M{
		"auth_user_id": userID,
		"job_id":       jobID,
	}).Decode(&result)

	if err == mongo.ErrNoDocuments {
		// Not found, return default
		return 50.0
	}
	if err != nil {
		fmt.Println("Error fetching match score:", err)
		return 50.0
	}
	return result.MatchScore
}


// Fetch applied job IDs to exclude
func FetchAppliedJobIDs(c *gin.Context, col *mongo.Collection, userID string) ([]string, error) {
	jobIDs := []string{} // Always return a non-nil slice

	cursor, err := col.Find(c, bson.M{"auth_user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(c)

	for cursor.Next(c) {
		var application models.SelectedJobApplication
		if err := cursor.Decode(&application); err == nil {
			jobIDs = append(jobIDs, application.JobID)
		}
	}
	return jobIDs, nil
}

// Construct the job query filter
func BuildJobFilter(preferredTitles, appliedJobIDs []string, jobLang string) bson.M {
	var andConditions []bson.M

	// Step 1: Preferred job titles from user
	var titleConditions []bson.M
	for _, title := range preferredTitles {
		titleConditions = append(titleConditions, bson.M{"title": bson.M{"$regex": title, "$options": "i"}})
	}
	andConditions = append(andConditions, bson.M{"$or": titleConditions})

	// Step 2: Exclude already applied jobs
	if len(appliedJobIDs) > 0 {
		andConditions = append(andConditions, bson.M{"job_id": bson.M{"$nin": appliedJobIDs}})
	}

	// Step 3: Optional job language filter
	if jobLang != "" {
		andConditions = append(andConditions, bson.M{"job_language": bson.M{"$regex": jobLang, "$options": "i"}})
	}

	return bson.M{"$and": andConditions}
}
