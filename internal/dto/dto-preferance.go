package dto

import (


	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"

)


// =======================
// PERSONAL INFO
// =======================

type PersonalInfoRequest struct {
	FirstName       string  `json:"first_name" binding:"required" bson:"first_name"`
	SecondName      *string `json:"second_name,omitempty" bson:"second_name,omitempty"`
	Country         *string `json:"country,omitempty" bson:"country,omitempty"`
	State           *string `json:"state,omitempty" bson:"state,omitempty"`
	City            *string `json:"city,omitempty" bson:"city,omitempty"`
	LinkedInProfile *string `json:"linkedin_profile,omitempty" bson:"linkedin_profile,omitempty"`
	Portfolio       *string `json:"portfolio,omitempty" bson:"portfolio,omitempty"`
	Resume          *string `json:"resume,omitempty" bson:"resume,omitempty"`
	Blog            *string `json:"blog,omitempty" bson:"blog,omitempty"`
}

type PersonalInfoResponse struct {
	AuthUserID      string            	`json:"auth_user_id" bson:"auth_user_id"`
	FirstName       string             	`json:"first_name" bson:"first_name"`
	SecondName      *string            	`json:"second_name,omitempty" bson:"second_name,omitempty"`
	Email       	string     			`json:"email" bson:"email"`
	Phone           string     			`json:"phone" bson:"phone"`			
	Country         *string            	`json:"country,omitempty" bson:"country,omitempty"`
	State           *string            	`json:"state,omitempty" bson:"state,omitempty"`
	City            *string            	`json:"city,omitempty" bson:"city,omitempty"`
	LinkedInProfile *string            	`json:"linkedin_profile,omitempty" bson:"linkedin_profile,omitempty"`
	Portfolio       *string            	`json:"portfolio,omitempty" bson:"portfolio,omitempty"`
	Resume          *string            	`json:"resume,omitempty" bson:"resume,omitempty"`
	Blog            *string            	`json:"blog,omitempty" bson:"blog,omitempty"`
	CreatedAt       time.Time          	`json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time         	`json:"updated_at" bson:"updated_at"`
}

// =======================
// WORK EXPERIENCE
// =======================

// Request payload (for creating or updating)
type WorkExperienceRequest struct {
	JobTitle            string     `json:"job_title" binding:"required" bson:"job_title"`
	CompanyName         string     `json:"company_name" binding:"required" bson:"company_name"`
	Location            string     `json:"location" bson:"location"`
	StartDate           time.Time  `json:"start_date" binding:"required" bson:"start_date"`
	EndDate             *time.Time `json:"end_date,omitempty" bson:"end_date,omitempty"`
	KeyResponsibilities *string     `json:"key_responsibilities,omitempty" bson:"key_responsibilities,omitempty"`
}

// Response payload
type WorkExperienceResponse struct {
	ID                  primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AuthUserID          string             `json:"auth_user_id" bson:"auth_user_id"`
	JobTitle            string             `json:"job_title" bson:"job_title"`
	CompanyName         string             `json:"company_name" bson:"company_name"`
	Location            string            `json:"location" bson:"location"`
	StartDate           time.Time          `json:"start_date" bson:"start_date"`
	EndDate             *time.Time         `json:"end_date,omitempty" bson:"end_date,omitempty"`
	KeyResponsibilities *string             `json:"key_responsibilities,omitempty" bson:"key_responsibilities,omitempty"`
	CreatedAt           time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at" bson:"updated_at"`
}

// =======================
// EDUCATION
// =======================

type AcademicsRequest struct {
	Institution     string     `json:"institution" binding:"required" bson:"institution"`
	City            *string    `json:"city,omitempty" bson:"city,omitempty"`
	Degree          string     `json:"degree" binding:"required" bson:"degree"`
	FieldOfStudy    string     `json:"field_of_study" binding:"required" bson:"field_of_study"`
	StartDate       time.Time  `json:"start_date" binding:"required" bson:"start_date"`
	EndDate         *time.Time `json:"end_date,omitempty" bson:"end_date,omitempty"`
	Description     *string    `json:"description,omitempty" bson:"description,omitempty"` // Additional info like achievements
}


type AcademicsResponse struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AuthUserID      string             `json:"auth_user_id" bson:"auth_user_id"`
	Institution     string             `json:"institution" bson:"institution"`
	City            string            `json:"city,omitempty" bson:"city,omitempty"`
	Degree          string             `json:"degree" bson:"degree"`
	FieldOfStudy    string             `json:"field_of_study" bson:"field_of_study"`
	StartDate       time.Time          `json:"start_date" bson:"start_date"`
	EndDate         *time.Time         `json:"end_date,omitempty" bson:"end_date,omitempty"`
	Description     *string            `json:"description,omitempty" bson:"description,omitempty"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}

// =======================
// PAST PROJECT
// =======================

type PastProjectRequest struct {
	ProjectName        string     `json:"project_name" binding:"required" bson:"project_name"`
	Institution        string     `json:"institution" binding:"required" bson:"institution"` // University or Company
	StartDate          time.Time  `json:"start_date" binding:"required" bson:"start_date"`
	EndDate            *time.Time `json:"end_date,omitempty" bson:"end_date,omitempty"`
	ProjectDescription *string     `json:"project_description,omitempty" bson:"project_description,omitempty"`
}

type PastProjectResponse struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AuthUserID         string             `json:"auth_user_id" bson:"auth_user_id"`
	ProjectName        string             `json:"project_name" bson:"project_name"`
	Institution        string             `json:"institution" bson:"institution"`
	StartDate          time.Time          `json:"start_date" bson:"start_date"`
	EndDate            *time.Time         `json:"end_date,omitempty" bson:"end_date,omitempty"`
	ProjectDescription *string             `json:"project_description,omitempty" bson:"project_description,omitempty"`
	CreatedAt          time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at" bson:"updated_at"`
}

// =======================
// LANGUAGES
// =======================

type LanguageRequest struct {
	LanguageName     string `json:"language" binding:"required" bson:"language"`
	ProficiencyLevel string `json:"proficiency" binding:"required" bson:"proficiency"`
}



type LanguageResponse struct {
	AuthUserID       string             `json:"auth_user_id" bson:"auth_user_id"`
	LanguageName     string             `json:"language" bson:"language"`
	ProficiencyLevel string             `json:"proficiency" bson:"proficiency"`
	CreatedAt        time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" bson:"updated_at"`
}


// =======================
// CERTIFICATE
// =======================

type CertificateRequest struct {
	CertificateName string     `json:"certificate_name" binding:"required" bson:"certificate_name"`
	Platform        *string     `json:"platform,omitempty" bson:"platform,omitempty"`
	StartDate       time.Time  `json:"start_date" binding:"required" bson:"start_date"`
	EndDate         *time.Time `json:"end_date,omitempty" bson:"end_date,omitempty"`
}


type CertificateResponse struct {
	AuthUserID      string             `json:"auth_user_id" bson:"auth_user_id"`
	CertificateName string             `json:"certificate_name" bson:"certificate_name"`
	Platform        string             `json:"platform" bson:"platform"`
	StartDate       time.Time          `json:"start_date" bson:"start_date"`
	EndDate         *time.Time         `json:"end_date,omitempty" bson:"end_date,omitempty"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}




// =======================
// JOB TITLE
// =======================

type JobTitleInput struct {
	PrimaryTitle   string  `json:"primary_title" bson:"primary_title"`
	SecondaryTitle *string `json:"secondary_title,omitempty" bson:"secondary_title,omitempty"`
	TertiaryTitle  *string `json:"tertiary_title,omitempty" bson:"tertiary_title,omitempty"`
}

type JobTitleResponse struct {
	AuthUserID     string  `json:"auth_user_id" bson:"auth_user_id"`
	PrimaryTitle   string  `json:"primary_title" bson:"primary_title"`
	SecondaryTitle *string `json:"secondary_title,omitempty" bson:"secondary_title,omitempty"`
	TertiaryTitle  *string `json:"tertiary_title,omitempty" bson:"tertiary_title,omitempty"`
}

// =======================
// KEY SKILLS
// =======================

type KeySkillsRequest struct {
	Skills []string `json:"skills" binding:"required" bson:"skills"`
}

type KeySkillsResponse struct {
	AuthUserID string             `json:"auth_user_id" bson:"auth_user_id"`
	Skills     []string           `json:"skills" bson:"skills"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}
