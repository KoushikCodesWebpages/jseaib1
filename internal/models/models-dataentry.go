package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)


// =======================
// PERSONAL INFO
// =======================

type PersonalInfo struct {
	ID              primitive.ObjectID 	`bson:"_id,omitempty"`
	AuthUserID      string            	`bson:"auth_user_id"`
	FirstName       string             	`bson:"first_name"`
	SecondName      *string            	`bson:"second_name,omitempty"`
	Country         *string            	`bson:"country,omitempty"`
	State           *string            	`bson:"state,omitempty"`
	City            *string            	`bson:"city,omitempty"`
	LinkedInProfile *string            	`bson:"linkedin_profile,omitempty"`
	Portfolio       *string            	`bson:"portfolio,omitempty"`
	Resume          *string            	`bson:"resume,omitempty"`
	Blog            *string            	`bson:"blog,omitempty"`
	CreatedAt       time.Time          	`bson:"created_at"`
	UpdatedAt       time.Time          	`bson:"updated_at"`
}

// =======================
// WORK EXPERIENCE
// =======================

type WorkExperience struct {
	ID                  	primitive.ObjectID `bson:"_id,omitempty"`
	AuthUserID          	string             `bson:"auth_user_id"`

	JobTitle           		string     			`json:"job_title"`
	CompanyName        		string     			`json:"company_name"`
	StartDate          		time.Time  			`json:"start_date"`
	EndDate            		*time.Time 			`json:"end_date,omitempty"`
	KeyResponsibilities 	*string    			`json:"key_responsibilities,omitempty"`
	Location           		string     			`json:"location"` // not a pointer   // Description of work
	CreatedAt       		time.Time          	`bson:"created_at"`
	UpdatedAt       		time.Time          	`bson:"updated_at"`
}

// =======================
// Academics
// =======================

type Academics struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	AuthUserID      string             `bson:"auth_user_id"`
	Institution     string             `bson:"institution"`
	City            string            `bson:"city,omitempty"`
	Degree          string             `bson:"degree"`
	FieldOfStudy    string             `bson:"field_of_study"`
	StartDate       time.Time          `bson:"start_date"`
	EndDate         *time.Time         `bson:"end_date,omitempty"`
	Description     *string            `bson:"description,omitempty"` // e.g., Achievements, activities, awards
	CreatedAt       time.Time          `bson:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at"`
}

// =======================
// PAST PROJECT
// =======================

type PastProject struct {
	ID                	primitive.ObjectID `bson:"_id,omitempty"`
	AuthUserID        	string             `bson:"auth_user_id"`                    // Link to the user
	ProjectName       	string             `bson:"project_name"`                    // Name of the project
	Institution       	string             `bson:"institution"`                     // University or Company
	StartDate         	time.Time          `bson:"start_date"`                      // Required
	EndDate           	*time.Time         `bson:"end_date,omitempty"`              // Optional
	ProjectDescription 	*string            `bson:"project_description"`             // Summary or detail of the project
	CreatedAt         	time.Time          `bson:"created_at"`
	UpdatedAt         	time.Time          `bson:"updated_at"`
}

// =======================
// LANGUAGES
// =======================

type Language struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	AuthUserID       string             `bson:"auth_user_id"`
	LanguageName     string             `bson:"language"`
	ProficiencyLevel string             `bson:"proficiency"`
	CreatedAt        time.Time          `bson:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at"`
}

// =======================
// CERTIFICATE
// =======================

type Certificate struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	AuthUserID     string             `bson:"auth_user_id"`
	CertificateName string            `bson:"certificate_name"`  // Name of the certificate
	Platform        *string            `bson:"platform"`          // Platform or issuer (e.g., Coursera, Udemy)
	StartDate       time.Time         `bson:"start_date"`        // Required
	EndDate         *time.Time        `bson:"end_date,omitempty"`// Optional
	CreatedAt       time.Time         `bson:"created_at"`
	UpdatedAt       time.Time         `bson:"updated_at"`
}