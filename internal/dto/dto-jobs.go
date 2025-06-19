package dto

// LinkResponseDTO represents the response DTO for job application links
type LinkResponseDTO struct {
    JobID   string `json:"job_id" bson:"job_id"`
    JobLink string `json:"job_link" bson:"job_link"`
    Source  string `json:"source" bson:"source"`
}


// Job Retrieval

type SalaryRange struct {
    Min int `json:"min" bson:"min"`
    Max int `json:"max" bson:"max"`
}

type JobDTO struct {
    Source         string       `json:"source" bson:"source"`                   // "linkedin" or "xing"
    ID             uint         `json:"id" bson:"id"`                           // UUID or unique DB ID
    JobID          string       `json:"job_id" bson:"job_id"`                   // Platform-specific Job ID
    Title          string       `json:"title" bson:"title"`
    Company        string       `json:"company" bson:"company"`
    Location       string       `json:"location" bson:"location"`
    PostedDate     string       `json:"posted_date" bson:"posted_date"`
    Processed      bool         `json:"processed" bson:"processed"`
    JobType        string       `json:"job_type" bson:"job_type"`               // e.g., Full-time, Part-time
    Skills         string       `json:"skills" bson:"skills"`                   // Comma-separated required skills
    UserSkills     []string     `json:"user_skills" bson:"user_skills"`         // List of user skills used in matching
    ExpectedSalary SalaryRange  `json:"expected_salary" bson:"expected_salary"` // Expected salary range
    MatchScore     float64      `json:"match_score" bson:"match_score"`         // Match score from 0 to 100
    Description    string       `json:"description" bson:"description"`         // Job description text
}

// JobFilterDTO represents the filter data for job retrieval.
type JobFilterDTO struct {
    Title string `form:"title" bson:"title"` // Query param: /jobs/linkedin?title=developer
}

// MatchScoreResponse represents the response containing the match score for a job.
type MatchScoreResponse struct {
    SeekerID   string  `json:"seeker_id" bson:"seeker_id"` // Changed to string
    JobID      string  `json:"job_id" bson:"job_id"`
    MatchScore float64 `json:"match_score" bson:"match_score"`
}

type SelectedJobResponse struct {
	AuthUserID            string             `json:"auth_user_id"`
	Source                string             `json:"source"`
	JobID                 string             `json:"job_id"`
	Title                 string             `json:"title"`
	Company               string             `json:"company"`
	Location              string             `json:"location"`
	PostedDate            string             `json:"posted_date"`
	Processed             bool               `json:"processed"`
	JobType               string             `json:"job_type"`
	Skills                string             `json:"skills"`
	UserSkills            []string           `json:"user_skills"`
	ExpectedSalary        SalaryRange        `json:"expected_salary" bson:"expected_salary"`
	MatchScore            float64            `json:"match_score"`
	Description           string             `json:"description"`
	Selected              bool               `json:"selected"`
	CvGenerated           bool               `json:"cv_generated"`
	CoverLetterGenerated  bool               `json:"cover_letter_generated"`
	ViewLink              bool               `json:"view_link"`
	SelectedDate          string             `json:"selected_date"`
}

type SelectedJobApplicationInput struct {
	JobID string `json:"job_id" binding:"required"` // Only the job_id is required for input
}
