package models

import (

	"context"
	
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

)


type Job struct {
    JobID          string `bson:"job_id" json:"job_id"`
    Title          string `bson:"title" json:"title"`
    Company        string `bson:"company" json:"company"`
    Location       string `bson:"location" json:"location"`
    PostedDate     string `bson:"posted_date" json:"posted_date"`
    Link           string `bson:"link" json:"link"`
    Processed      bool   `bson:"processed" json:"processed"`
    Source         string `bson:"source" json:"source"`
    JobDescription string `bson:"job_description" json:"job_description"`
    JobType        string `bson:"job_type" json:"job_type"`
    Skills         string `bson:"skills" json:"skills"`
    JobLink        string `bson:"job_link" json:"job_link"`
    SelectedCount  int    `bson:"selected_count" json:"selected_count"` // Added selectedCount field
}

func CreateJobIndexes(collection *mongo.Collection) error {
	// Unique index for job_id
	jobIdIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "job_id", Value: 1}}, // Index on job_id, unique
		Options: options.Index().SetUnique(true),
	}

	// Index for job_type (useful for filtering jobs by type)
	jobTypeIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "job_type", Value: 1}}, // Index on job_type
		Options: options.Index().SetUnique(false),    // Not unique
	}

	// Index for selected_count (useful for filtering or counting by selected count)
	selectedCountIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "selected_count", Value: 1}}, // Index on selected_count
		Options: options.Index().SetUnique(false),         // Not unique
	}

	// Hashed index on title for faster lookups based on job title
	jobTitleIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "title", Value: "hashed"}}, // Hashed index on title
		Options: options.Index().SetUnique(false),       // Not unique
	}

	// Index for posted_date (useful for filtering jobs by date)
	postedDateIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "posted_date", Value: 1}}, // Index on posted_date, ascending
		Options: options.Index().SetUnique(false),      // Not unique
	}

	// Create indexes
	_, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		jobIdIndex, jobTypeIndex, selectedCountIndex, jobTitleIndex, postedDateIndex,
	})
	return err
}


// MatchScore for job seeker match score
type MatchScore struct {
	AuthUserID string `json:"auth_user_id" bson:"auth_user_id"`
	JobID      string    `bson:"jobId" json:"jobId"`          
	MatchScore float64   `bson:"matchScore" json:"matchScore"` 
}


func CreateMatchScoreIndexes(collection *mongo.Collection) error {
	// Compound unique index for authUserId and jobId
	matchScoreIndex := mongo.IndexModel{
		Keys:    bson.D{
			{Key: "authUserId", Value: 1}, // Index on authUserId
			{Key: "jobId", Value: 1},      // Index on jobId
		},
		Options: options.Index().SetUnique(true), // Ensuring the combination is unique
	}

	// Create the compound index
	_, err := collection.Indexes().CreateOne(context.Background(), matchScoreIndex)
	return err
}














