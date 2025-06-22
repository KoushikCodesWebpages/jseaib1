package models

import (
	"time"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserEntryTimeline struct {
	ID                     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AuthUserID             string             `bson:"auth_user_id" json:"auth_user_id"`

	// Compulsory steps
	PersonalInfoCompleted   bool              `bson:"personal_info_completed" json:"personal_info_completed"`
	PersonalInfoRequired    bool              `bson:"personal_info_required" json:"personal_info_required"`

	AcademicsCompleted      bool              `bson:"academics_completed" json:"academics_completed"`
	AcademicsRequired       bool              `bson:"academics_required" json:"academics_required"`

	LanguagesCompleted      bool              `bson:"languages_completed" json:"languages_completed"`
	LanguagesRequired       bool              `bson:"languages_required" json:"languages_required"`

	JobTitlesCompleted      bool              `bson:"job_titles_completed" json:"job_titles_completed"`
	JobTitlesRequired       bool              `bson:"job_titles_required" json:"job_titles_required"`

	KeySkillsCompleted      bool              `bson:"key_skills_completed" json:"key_skills_completed"`
	KeySkillsRequired       bool              `bson:"key_skills_required" json:"key_skills_required"`

	// Optional steps
	WorkExperiencesCompleted bool             `bson:"work_experiences_completed" json:"work_experiences_completed"`
	WorkExperiencesRequired  bool             `bson:"work_experiences_required" json:"work_experiences_required"`

	PastProjectsCompleted    bool             `bson:"past_projects_completed" json:"past_projects_completed"`
	PastProjectsRequired     bool             `bson:"past_projects_required" json:"past_projects_required"`

	CertificatesCompleted    bool             `bson:"certificates_completed" json:"certificates_completed"`
	CertificatesRequired     bool             `bson:"certificates_required" json:"certificates_required"`

	// Overall
	Completed                bool             `bson:"completed" json:"completed"`

	CreatedAt                time.Time        `bson:"created_at" json:"created_at"`
	UpdatedAt                time.Time        `bson:"updated_at" json:"updated_at"`
}


func CreateUserEntryTimelineIndexes(collection *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "auth_user_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("unique_auth_user_id"),
	}
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	return err
}


type SalaryRange struct {
	Min int `bson:"min" json:"min"`
	Max int `bson:"max" json:"max"`
}

type SelectedJobApplication struct {
	ID                     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AuthUserID             string             `bson:"auth_user_id" json:"auth_user_id"`
	JobID                  string             `bson:"job_id" json:"job_id"`
	SelectedDate          time.Time          `bson:"selected_date" json:"selected_date"` // Changed to time.Time
}

// type SelectedJobApplication struct {
// 	ID                     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
// 	AuthUserID             string             `bson:"auth_user_id" json:"auth_user_id"`
// 	Source                 string             `bson:"source" json:"source"`
// 	JobID                  string             `bson:"job_id" json:"job_id"`
// 	Title                 string             `bson:"title" json:"title"`
// 	Company               string             `bson:"company" json:"company"`
// 	Location              string             `bson:"location" json:"location"`
// 	PostedDate            string             `bson:"posted_date" json:"posted_date"`
// 	Processed             bool               `bson:"processed" json:"processed"`
// 	JobType               string             `bson:"job_type" json:"job_type"`
// 	Skills                string             `bson:"skills" json:"skills"`
// 	UserSkills            []string           `bson:"user_skills" json:"user_skills"` 
// 	ExpectedSalary        SalaryRange        `json:"expected_salary" bson:"expected_salary"`
// 	MatchScore            float64            `bson:"match_score" json:"match_score"`
// 	Description           string             `bson:"description" json:"description"`
// 	Selected              bool               `bson:"selected" json:"selected"`
// 	CvGenerated           bool               `bson:"cv_generated" json:"cv_generated"`
// 	CoverLetterGenerated  bool               `bson:"cover_letter_generated" json:"cover_letter_generated"`
// 	ViewLink              bool               `bson:"view_link" json:"view_link"`
// 	SelectedDate          time.Time          `bson:"selected_date" json:"selected_date"` // Changed to time.Time
// }

func CreateSelectedJobApplicationIndexes(collection *mongo.Collection) error {
	indexModel1 := mongo.IndexModel{
		Keys:    bson.D{{Key: "auth_user_id", Value: 1}, {Key: "job_id", Value: 1}}, 
		Options: options.Index().SetUnique(true),
	}


	indexModel2 := mongo.IndexModel{
		Keys:    bson.D{{Key: "job_id", Value: "hashed"}},
		Options: options.Index().SetUnique(false),
	}

	indexModel3 := mongo.IndexModel{
		Keys:    bson.D{{Key: "match_score", Value: "hashed"}}, 
		Options: options.Index().SetUnique(false),
	}
	indexModel4 := mongo.IndexModel{
		Keys:    bson.D{{Key: "selected_date", Value: -1}},
		Options: options.Index().SetUnique(false),
	}


	_, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{indexModel1, indexModel2, indexModel3, indexModel4})
	return err
}
type CV struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AuthUserID string             `bson:"auth_user_id" json:"auth_user_id"`
	JobID      string             `bson:"job_id" json:"job_id"`
	CVUrl      string             `bson:"cv_url" json:"cv_url"` 
}

func CreateCVIndexes(collection *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "auth_user_id", Value: 1}, {Key: "job_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	return err
}

type CoverLetter struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AuthUserID    string             `bson:"auth_user_id" json:"auth_user_id"` // Changed to string
	JobID         string             `bson:"job_id" json:"job_id"` // Changed to string
	CoverLetterURL string            `bson:"cover_letter_url" json:"cover_letter_url"` // Changed to string
}

func CreateCoverLetterIndexes(collection *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "auth_user_id", Value: 1}, {Key: "job_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	return err
}


type SavedJob struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AuthUserID            string             `bson:"auth_user_id" json:"auth_user_id"`
	Source                string             `bson:"source" json:"source"`
	JobID                 string             `bson:"job_id" json:"job_id"`
	// SavedDate          time.Time          `bson:"savedDate" json:"selected_date"`
}

func CreateSavedJobApplicationIndexes(collection *mongo.Collection) error {
	indexModel1 := mongo.IndexModel{
		Keys:    bson.D{{Key: "auth_user_id", Value: 1}, {Key: "job_id", Value: 1}}, 
		Options: options.Index().SetUnique(true),
	}


	indexModel2 := mongo.IndexModel{
		Keys:    bson.D{{Key: "job_id", Value: "hashed"}},
		Options: options.Index().SetUnique(false),
	}

	indexModel3 := mongo.IndexModel{
		Keys: bson.D{{Key: "auth_user_id", Value: 1}},
	}
	

	_, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{indexModel1, indexModel2,indexModel3})
	return err
}