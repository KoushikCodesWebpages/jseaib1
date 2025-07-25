package models

import (

	"context"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"

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

	// New Fields
	JobLang		   string `bson:"job_language" json:"job_language"`
	JobTitle	   string `bson:"job_title" json:"job_title"`
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
		// Hashed index on title for faster lookups based on job title
	jobLangIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "job_language", Value: "hashed"}}, // Hashed index on title
		Options: options.Index().SetUnique(false),       // Not unique
	}

	// Index for posted_date (useful for filtering jobs by date)
	postedDateIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "posted_date", Value: 1}}, // Index on posted_date, ascending
		Options: options.Index().SetUnique(false),      // Not unique
	}

	// Create indexes
	_, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		jobIdIndex, jobTypeIndex, selectedCountIndex, jobTitleIndex, jobLangIndex ,postedDateIndex,
	})
	return err
}

type ExternalJob struct {
    JobID         string    `bson:"job_id" json:"job_id"`
    Title         string    `bson:"title" json:"title"`
    Company       string    `bson:"company" json:"company"`
    Description   string    `bson:"description" json:"description"`
    JobLanguage   string    `bson:"job_language" json:"job_language"`
    PostedDate    time.Time `bson:"posted_date" json:"posted_date"`
}


type MatchScore struct {
    AuthUserID string             `json:"auth_user_id" bson:"auth_user_id"`
    JobID      string             `json:"job_id" bson:"job_id"`
    MatchScore float64            `json:"match_score" bson:"match_score"`
    CreatedAt  time.Time          `json:"created_at" bson:"created_at"`  
}

func CreateMatchScoreIndexes(collection *mongo.Collection) error {
    // Create compound unique index on auth_user_id and job_id
    model := mongo.IndexModel{
        Keys: bson.D{
            {Key: "auth_user_id", Value: 1},
            {Key: "job_id", Value: 1},
			
        },
        Options: options.Index().SetUnique(true),
    }

    _, err := collection.Indexes().CreateOne(context.Background(), model)
    return err
}



type CoverLetterData struct {
    ID          primitive.ObjectID     	`bson:"_id,omitempty" json:"id"`
    AuthUserID  string                 	`bson:"auth_user_id" json:"auth_user_id"`
    JobID       string                 	`bson:"job_id" json:"job_id"`
    CLData      map[string]interface{} 	`bson:"cl_data" json:"cl_data"`
	ClFormat	string					`bson:"cl_format" json:"cl_format"`
}

func CreateCoverLetterIndexes(collection *mongo.Collection) error {
	// Compound unique index for authUserId and jobId
	coverLetterIndexes := mongo.IndexModel{
		Keys:    bson.D{
			{Key: "auth_user_id", Value: 1}, // Index on authUserId
			{Key: "job_id", Value: 1},      // Index on jobId
		},
		Options: options.Index().SetUnique(true), // Ensuring the combination is unique
	}

	// Create the compound index
	_, err := collection.Indexes().CreateOne(context.Background(), coverLetterIndexes)
	return err
}

type CVData struct {
    ID         		primitive.ObjectID     		`bson:"_id,omitempty" json:"id"`
    AuthUserID 		string                 		`bson:"auth_user_id" json:"auth_user_id"`
    JobID      		string                 		`bson:"job_id" json:"job_id"`
    CVData     		map[string]interface{} 	  	`bson:"cv_data" json:"cv_data"`
	CvFormat		string						`bson:"cv_format" json:"cv_format"`
}

func CreateCVIndexes(collection *mongo.Collection) error {
	// Compound unique index for authUserId and jobId
	cvLetterIndexes := mongo.IndexModel{
		Keys:    bson.D{
			{Key: "auth_user_id", Value: 1}, // Index on authUserId
			{Key: "job_id", Value: 1},      // Index on jobId
		},
		Options: options.Index().SetUnique(true), // Ensuring the combination is unique
	}

	// Create the compound index
	_, err := collection.Indexes().CreateOne(context.Background(), cvLetterIndexes)
	return err
}









