package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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



type SelectedJobApplication struct {
	ID                     	primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	AuthUserID             	string             	`bson:"auth_user_id" json:"auth_user_id"`
	JobID                  	string             	`bson:"job_id" json:"job_id"`
	CoverLetterGenerated  	bool               	`bson:"cover_letter_generated" json:"cover_letter_generated"`
	CvGenerated           	bool               	`bson:"cv_generated" json:"cv_generated"`
	SelectedDate          	time.Time          	`bson:"selected_date" json:"selected_date"`
	ViewLink              	bool               	`bson:"view_link" json:"view_link"`
	Status					string				`bson:"status" json:"status"`
	Source 					string				`bson:"source" json:"source"`
	Company					string				`bson:"company" json:"company"`
}

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
		Keys:    bson.D{{Key: "selected_date", Value: -1}},
		Options: options.Index().SetUnique(false),
	}
	
	indexModel4 := mongo.IndexModel{
		Keys:    bson.D{{Key: "status", Value: "hashed"}},
		Options: options.Index().SetUnique(false),
	}

	indexModel5 := mongo.IndexModel{
		Keys:    bson.D{{Key: "source", Value: "hashed"}},
		Options: options.Index().SetUnique(false),
	}



	_, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{indexModel1, indexModel2, indexModel3,indexModel4,indexModel5})
	return err
}


type SavedJob struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AuthUserID            string             `bson:"auth_user_id" json:"auth_user_id"`
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

//Preferences 

type UserPreferences struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AuthUserID          string             `bson:"auth_user_id" json:"auth_user_id"`
	Language            string             `bson:"language" json:"language"` // "english", "german"
	Timezone            string             `bson:"timezone" json:"timezone"` // e.g. "Europe/Berlin"
	CookiePolicy        bool               `bson:"cookie_policy" json:"cookie_policy"`
	CreatedAt           time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt           time.Time          `bson:"updated_at" json:"updated_at"`
}
func CreateUserPreferencesIndexes(collection *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "auth_user_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("unique_auth_user_id_prefs"),
	}
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	return err
}


type NotificationSettings struct {
	AuthUserID         string    `json:"auth_user_id" bson:"auth_user_id"`
	Subscription       bool      `json:"subscription" bson:"subscription"`
	RecommendedJobs    bool      `json:"recommended_jobs" bson:"recommended_jobs"`
	GermanTest         bool      `json:"german_test" bson:"german_test"`
	Announcements      bool      `json:"announcements" bson:"announcements"`
	CreatedAt          time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" bson:"updated_at"`
}

func CreateUserNotificationsIndexes(collection *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "auth_user_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("unique_auth_user_id_notify"),
	}
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	return err
}


type Option struct {
	ID        			string 		`bson:"id" json:"id"`                         // Unique within question
	Text      			string 		`bson:"text" json:"text"`                     // Answer text
	Media     			string 		`bson:"media,omitempty" json:"media,omitempty"` // Optional image/audio/video URL
	IsCorrect 			bool   		`bson:"is_correct,omitempty" json:"is_correct,omitempty"` // Only stored in backend
}

type Question struct {               	// Question ID
	QuestionID        	string   	`bson:"question_id" json:"question_id"`				// Unique question code (e.g., QB-123)
	
	Type             	string    	`bson:"type" json:"type"` 							// "mcq", "checkbox", "short", "long", "coding"
	Question			string 		`bson:"question" json:"question"`
	Options          	[]Option  	`bson:"options,omitempty" json:"options,omitempty"` 	// For MCQ/checkbox types
	Difficulty       	string    	`bson:"difficulty,omitempty" json:"difficulty,omitempty"` // easy, medium, hard
	Language         	string    	`bson:"language,omitempty" json:"language,omitempty"`
	RandomizeOptions 	bool      	`bson:"randomize_options,omitempty" json:"randomize_options,omitempty"` // For frontend shuffle
	
	Title            	string    	`bson:"title" json:"title"`                         	// Short title/summary
	Description      	string    	`bson:"description" json:"description"`             	// Full question text                          
	CorrectOptionIDs 	[]string  	`bson:"correct_option_ids,omitempty" json:"correct_option_ids,omitempty"` // For validation/evaluation
	AnswerKey        	string    	`bson:"answer_key,omitempty" json:"answer_key,omitempty"` // For subjective/coding
	Marks            	float64   	`bson:"marks" json:"marks"`                         // Default marks
	NegativeMark     	float64   	`bson:"negative_mark,omitempty" json:"negative_mark,omitempty"`
	Tags             	[]string  	`bson:"tags,omitempty" json:"tags,omitempty"`       // e.g., ["algebra", "chapter-2"]
	Category         	string    	`bson:"category,omitempty" json:"category,omitempty"` // e.g., "Math", "Programming"
	SubCategory      	string   	`bson:"sub_category,omitempty" json:"sub_category,omitempty"`
    // For bilingual support
	Attachments      	[]string  	`bson:"attachments,omitempty" json:"attachments,omitempty"` // File URLs
	Explanation      	string    	`bson:"explanation,omitempty" json:"explanation,omitempty"` // For post-submission feedback
	IsActive         	bool      	`bson:"is_active" json:"is_active"`                 // For enabling/disabling in pool
	CreatedAt        	time.Time 	`bson:"created_at" json:"created_at"`
	UpdatedAt        	time.Time 	`bson:"updated_at" json:"updated_at"`
}

func CreateQuestionIndexes(collection *mongo.Collection) error {


	indexModel1 := mongo.IndexModel{
		Keys: bson.D{{Key: "question_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	indexModel2 := mongo.IndexModel{
		Keys:    bson.D{{Key: "created_at", Value: -1}},
		Options: options.Index().SetUnique(false),
	}
	
	indexModel3 := mongo.IndexModel{
		Keys:    bson.D{{Key: "difficulty", Value: "hashed"}},
		Options: options.Index().SetUnique(false),
	}

	indexModel4 := mongo.IndexModel{
		Keys:    bson.D{{Key: "language", Value: "hashed"}},
		Options: options.Index().SetUnique(false),
	}
	indexModel5 := mongo.IndexModel{
		Keys:    bson.D{{Key: "type", Value: "hashed"}},
		Options: options.Index().SetUnique(false),
	}

	_, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{indexModel1, indexModel2, indexModel3,indexModel4,indexModel5})
	return err
}

// AnswerResult captures a user’s answer and how it was scored
type AnswerResult struct {
	QuestionID        string   `bson:"question_id" json:"question_id"`                 // Which question
	SelectedOptionIDs []string `bson:"selected_option_ids,omitempty" json:"selected_option_ids,omitempty"` // For MCQ/checkbox
	WrittenAnswer     string   `bson:"written_answer,omitempty" json:"written_answer,omitempty"`           // For subjective/coding
	AutoScore         float64  `bson:"auto_score,omitempty" json:"auto_score,omitempty"`                   // Based on key/check
	ManualScore       float64  `bson:"manual_score,omitempty" json:"manual_score,omitempty"`               // If reviewed
	MaxMarks          float64  `bson:"max_marks" json:"max_marks"`                     // Total marks for the question
	NegativeMark      float64  `bson:"negative_mark,omitempty" json:"negative_mark,omitempty"` // Applied penalty
	TimeSpent         int      `bson:"time_spent,omitempty" json:"time_spent,omitempty"` // Seconds
	IsFlagged         bool     `bson:"is_flagged,omitempty" json:"is_flagged,omitempty"` // "Mark for review"
	Remarks           string   `bson:"remarks,omitempty" json:"remarks,omitempty"`       // Optional reviewer notes
}



type ExamResult struct {
	AuthUserID      string    	`json:"auth_user_id" bson:"auth_user_id"`
	ExamID          string    	`json:"exam_id" bson:"exam_id"`
	Title           string    	`json:"title" bson:"title"`
	Language        string    	`json:"language" bson:"language"`
	Level           string    	`json:"level,omitempty" bson:"level,omitempty"` // difficulty
	Score           float64   	`json:"score" bson:"score"`
	Grade           string    	`json:"grade" bson:"grade"`
	TotalMarks      float64   	`json:"total_marks" bson:"total_marks"`
	Percentage      float64   	`json:"percentage" bson:"percentage"`
	Attempted       int       	`json:"attempted" bson:"attempted"`
	Correct         int       	`json:"correct" bson:"correct"`
	Wrong           int       	`json:"wrong" bson:"wrong"`
	Skipped         int       	`json:"skipped" bson:"skipped"`
	NegativeMarks   float64   	`json:"negative_marks,omitempty" bson:"negative_marks,omitempty"`
	Rank            *int      	`json:"rank,omitempty" bson:"rank,omitempty"` // optional leaderboard integration
	Remarks         string    	`json:"remarks,omitempty" bson:"remarks,omitempty"`
	ReviewedBy      *string   	`json:"reviewed_by,omitempty" bson:"reviewed_by,omitempty"`
	SubmittedAt     time.Time 	`json:"submitted_at" bson:"submitted_at"`
	DurationMinutes int       	`json:"duration_minutes" bson:"duration_minutes"`
	IsPass          bool      	`json:"is_pass" bson:"is_pass"`
}

func CreateResultsIndexes(collection *mongo.Collection) error {


	indexModel1 := mongo.IndexModel{
		Keys: bson.D{{Key: "exam_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	indexModel2 := mongo.IndexModel{
		Keys:    bson.D{{Key: "created_at", Value: -1}},
		Options: options.Index().SetUnique(false),
	}
	
	indexModel3 := mongo.IndexModel{
		Keys:    bson.D{{Key: "exam_id", Value: "hashed"}},
		Options: options.Index().SetUnique(false),
	}

	indexModel4 := mongo.IndexModel{
		Keys:    bson.D{{Key: "auth_user_id", Value: "hashed"}},
		Options: options.Index().SetUnique(false),
	}

	_, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{indexModel1, indexModel2, indexModel3,indexModel4})
	return err
}

type Announcement struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Message   string             `bson:"message" json:"message"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}