package repository

import (
	"RAAS/internal/models"

	"context"
	"fmt"
	"time"
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


func FetchAppliedJobIDs(ctx context.Context, collection *mongo.Collection, userID string) ([]string, error) {
	filter := bson.M{
		"auth_user_id": userID,
		"status": bson.M{
			"$in": []string{"applied", "interview", "rejected", "selected"},
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var appliedIDs []string
	for cursor.Next(ctx) {
		var result struct {
			JobID string `bson:"job_id"`
		}
		if err := cursor.Decode(&result); err == nil {
			appliedIDs = append(appliedIDs, result.JobID)
		}
	}

	return appliedIDs, nil
}


func BuildJobFilter(preferredTitles, appliedJobIDs []string, jobLang string) bson.M {
	var andConditions []bson.M

	// Step 0: Filter jobs within last 14 days
	twoWeeksAgo := time.Now().AddDate(0, 0, -14).Format("2006-01-02")
	andConditions = append(andConditions, bson.M{"posted_date": bson.M{"$gte": twoWeeksAgo}})

	// Step 1: Preferred job titles from user
	if len(preferredTitles) > 0 {
		var titleConditions []bson.M
		for _, title := range preferredTitles {
			titleConditions = append(titleConditions, bson.M{"title": bson.M{"$regex": title, "$options": "i"}})
		}
		andConditions = append(andConditions, bson.M{"$or": titleConditions})
	}

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

