package repository

import (
	"RAAS/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

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
func BuildJobFilter(preferredTitles, appliedJobIDs []string) bson.M {
	var titleConditions []bson.M
	for _, title := range preferredTitles {
		titleConditions = append(titleConditions, bson.M{"title": bson.M{"$regex": title, "$options": "i"}})
	}

	filter := bson.M{
		"$and": []bson.M{
			{"$or": titleConditions},
			{"job_id": bson.M{"$nin": appliedJobIDs}}, // safe now
		},
	}
	return filter
}