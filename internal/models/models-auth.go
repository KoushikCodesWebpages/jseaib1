package models

import (
	"time"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuthUser struct {
	AuthUserID           	string     `json:"auth_user_id" bson:"auth_user_id"` // Changed to string for MongoDB UUID storage
	Email                	string     `json:"email" bson:"email"`
	Phone                	string     `json:"phone" bson:"phone"`
	Password             	string     `json:"password" bson:"password"`
	Role                 	string     `json:"role" bson:"role"`
	EmailVerified        	bool       `json:"email_verified" bson:"email_verified"`
	Provider             	string     `json:"provider" bson:"provider,omitempty"`

	VerificationToken    	string     `json:"verification_token" bson:"verification_token"`
	ResetTokenExpiry     	*time.Time `json:"reset_token_expiry" bson:"reset_token_expiry"`

	IsActive             	bool       `json:"is_active" bson:"is_active"`
	
	CreatedBy            	string     `json:"created_by" bson:"created_by"` // Changed to string for MongoDB UUID storage
	UpdatedBy            	string     `json:"updated_by" bson:"updated_by"`
	CreatedAt 				*time.Time `json:"created_at" bson:"created_at"` // Changed to string for MongoDB UUID storage
	UpdatedAt 				*time.Time `json:"updated_at" bson:"updated_at"`
	LastLoginAt          	*time.Time `json:"last_login_at,omitempty" bson:"last_login_at,omitempty"`
	PasswordLastUpdated  	*time.Time `json:"password_last_updated,omitempty" bson:"password_last_updated,omitempty"`
	TwoFactorEnabled     	bool       `json:"two_factor_enabled" bson:"two_factor_enabled"`
	TwoFactorSecret      	*string    `json:"two_factor_secret,omitempty" bson:"two_factor_secret,omitempty"`
}


func CreateAuthUserIndexes(collection *mongo.Collection) error {
	indexModelEmail := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	indexModelPhone := mongo.IndexModel{
		Keys:    bson.D{{Key: "phone", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	indexModelCompound := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}, {Key: "phone", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		indexModelEmail,
		indexModelPhone,
		indexModelCompound,
	})
	return err
}


type Seeker struct {
	ID                          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	AuthUserID                  string             `json:"auth_user_id" bson:"auth_user_id"`

	PhotoUrl					string 				`json:"photo_url,omitempty" bson:"photo_url,omitempty"`


	TotalApplications           int                `json:"total_applications" bson:"total_applications"`
	WeeklyAppliedJobs           int                `json:"weekly_applications_count" bson:"weekly_applications_count"`
	TopJobs                     int                `json:"top_jobs_count" bson:"top_jobs_count"`
	
	StripeCustomerID			string				`json:"stripe_customer_id" bson:"stripe_customer_id"`
	SubscriptionTier          	string    			`json:"subscription_tier" bson:"subscription_tier"`
	SubscriptionPeriod        	string    			`json:"subscription_period" bson:"subscription_period"` // e.g., "monthly", "quarterly"
	SubscriptionIntervalStart 	time.Time 			`json:"subscription_interval_start" bson:"subscription_interval_start"`
	SubscriptionIntervalEnd   	time.Time 			`json:"subscription_interval_end" bson:"subscription_interval_end"`


	ExternalApplications         int                `json:"external_application_count" bson:"external_application_count"`
	InternalApplications         int                `json:"internal_application_count" bson:"internal_application_count"`
	ProficiencyTest            	int                	`json:"proficiency_test" bson:"proficiency_test"`

	PersonalInfo                bson.M             `json:"personal_info" bson:"personal_info"`
	WorkExperiences             []bson.M           `json:"work_experiences" bson:"work_experiences"`
	Academics                   []bson.M           `json:"academics" bson:"academics"`
	PastProjects                []bson.M           `json:"past_projects" bson:"past_projects"`
	Certificates                []bson.M           `json:"certificates" bson:"certificates"`
	Languages                   []bson.M           `json:"languages" bson:"languages"`
	KeySkills                   []string           `json:"key_skills" bson:"key_skills"`

	PrimaryTitle                string             `json:"primary_title" bson:"primary_title"`
	SecondaryTitle              *string            `json:"secondary_title,omitempty" bson:"secondary_title,omitempty"`
	TertiaryTitle               *string            `json:"tertiary_title,omitempty" bson:"tertiary_title,omitempty"`

	CvFormat					string			   `json:"cv_format" bson:"cv_format"`
	ClFormat					string			   `json:"cl_format" bson:"cl_format"`
	
	CreatedAt                   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt                   time.Time          `json:"updated_at" bson:"updated_at"`
}

//modern_deedy

func CreateSeekerIndexes(collection *mongo.Collection) error {
	// Create index for AuthUserID to be unique
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "auth_user_id", Value: 1}}, 
		Options: options.Index().SetUnique(true),        
	}
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return err
	}

	// Create hashed index for primary_title
	primaryTitleIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "primary_title", Value: "hashed"}},
		Options: options.Index().SetName("primary_title_hashed"),
	}
	_, err = collection.Indexes().CreateOne(context.Background(), primaryTitleIndex)
	if err != nil {
		return err
	}

	// Create hashed index for secondary_title
	secondaryTitleIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "secondary_title", Value: "hashed"}},
		Options: options.Index().SetName("secondary_title_hashed"),
	}
	_, err = collection.Indexes().CreateOne(context.Background(), secondaryTitleIndex)
	if err != nil {
		return err
	}

	// Create hashed index for tertiary_title
	tertiaryTitleIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "tertiary_title", Value: "hashed"}},
		Options: options.Index().SetName("tertiary_title_hashed"),
	}
	_, err = collection.Indexes().CreateOne(context.Background(), tertiaryTitleIndex)
	if err != nil {
		return err
	}

	return nil
}

type Admin struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	AuthUserID string             `json:"auth_user_id" bson:"auth_user_id"` // Change uuid.UUID to string
}

func CreateAdminIndexes(collection *mongo.Collection) error {
	// Create index for AuthUserID to be unique
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "auth_user_id", Value: 1}}, 
		Options: options.Index().SetUnique(true),      
	}
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	return err
}

type ProfilePic struct {
	AuthUserID    			string			`bson:"auth_user_id" json:"auth_user_id"`           // Reference to the user
	Image     				[]byte  					`bson:"image" json:"-"`                   // Binary data (image file)
	MimeType  				string             			`bson:"mime_type" json:"mime_type"`       // e.g. image/png, image/jpeg
	CreatedAt 				time.Time          			`bson:"created_at" json:"created_at"`
	UpdatedAt 				time.Time          			`bson:"updated_at" json:"updated_at"`
}

func CreateProfilePicIndexes(collection *mongo.Collection) error {
	// Create index for AuthUserID to be unique
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "auth_user_id", Value: 1}}, 
		Options: options.Index().SetUnique(true),      
	}
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	return err
}