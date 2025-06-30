package dto

import (
    "RAAS/internal/models"
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/bson"
)

type SeekerSignUpInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    Number   string `json:"number" binding:"required,min=10,max=15"`
}

type LoginInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// AuthUserMinimal represents minimal user details for response
type AuthUserMinimal struct {
    Email         string `json:"email"`
    EmailVerified bool   `json:"emailVerified"`
    Provider      string `json:"provider"`
    Number        string `json:"number" binding:"required,len=10"`
}

// SeekerResponse represents the response structure for Seeker details
type SeekerResponse struct {
    ID               primitive.ObjectID `json:"id"`        // Use primitive.ObjectID for Seeker ID in MongoDB
    AuthUserID       string              `json:"authUserId"` // UUID as string for AuthUserID
    AuthUser         AuthUserMinimal     `json:"authUser"`
    SubscriptionTier string              `json:"subscriptionTier"`
}

func SeekerProfileResponse(seeker models.Seeker) SeekerResponse {
    return SeekerResponse{
        ID:               seeker.ID,                 // ID as primitive.ObjectID
        AuthUserID:       seeker.AuthUserID, // UUID as string for AuthUserID
        SubscriptionTier: seeker.SubscriptionTier,
    }
}



type SeekerDTO struct {
	ID                          primitive.ObjectID `json:"_id,omitempty"`
	AuthUserID                  string             `json:"auth_user_id"`
    FirstName                   string              `json:"first_name"`
    ProfileCompletion           int                 `json:profile_completion`
	Photo                       string             `json:"photo,omitempty"` // URL string instead of binary

	TotalApplications           int                `json:"total_applications"`
	WeeklyAppliedJobs           int                `json:"weekly_applications_count"`
	TopJobs                     int                `json:"top_jobs_count"`

	SubscriptionTier            string             `json:"subscription_tier"`
	SubscriptionPeriod          string             `json:"subscription_period"` // e.g., "monthly", "quarterly"
	SubscriptionIntervalStart   time.Time          `json:"subscription_interval_start"`
	SubscriptionIntervalEnd     time.Time          `json:"subscription_interval_end"`

	ExternalApplications        int                `json:"external_application_count"`
	InternalApplications        int                `json:"internal_application_count"`
	ProficicencyTest            int                `json:"proficicency_test"`

	PersonalInfo                bson.M             `json:"personal_info"`
	WorkExperiences             []bson.M           `json:"work_experiences"`
	Academics                   []bson.M           `json:"academics"`
	PastProjects                []bson.M           `json:"past_projects"`
	Certificates                []bson.M           `json:"certificates"`
	Languages                   []bson.M           `json:"languages"`
	KeySkills                   []string           `json:"key_skills"`

	PrimaryTitle                string             `json:"primary_title"`
	SecondaryTitle              *string            `json:"secondary_title,omitempty"`
	TertiaryTitle               *string            `json:"tertiary_title,omitempty"`

	CvFormat                    string             `json:"cv_format"`
	ClFormat                    string             `json:"cl_format"`

	CreatedAt                   time.Time          `json:"created_at"`
	UpdatedAt                   time.Time          `json:"updated_at"`
}

