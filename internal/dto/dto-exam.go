package dto

import (
	"RAAS/internal/models"
)

type ExamDTORequest struct {
	Title         *string   `json:"title,omitempty"`         // e.g., "Final Test"
	Description   *string   `json:"description,omitempty"`   // e.g., "Covers full syllabus"
	Category      *string   `json:"category,omitempty"`      // e.g., "Math", "Programming"
	SubCategory   *string   `json:"sub_category,omitempty"`  // e.g., "Algebra", "Loops"
	TotalMarks    *float64  `json:"total_marks,omitempty"`   // Total exam marks (can be inferred from questions)
	DurationMins  *int      `json:"duration_mins,omitempty"` // e.g., 60 mins
	QuestionIDs   *[]string `json:"question_ids,omitempty"`  // List of question IDs
	Tags          *[]string `json:"tags,omitempty"`          // For filtering/search
	Difficulty    *string   `json:"difficulty,omitempty"`    // average difficulty: Easy/Medium/Hard
	IsPublic      *bool     `json:"is_public,omitempty"`     // whether exam is discoverable
	StartsAt      *string   `json:"starts_at,omitempty"`     // optional scheduled start time (ISO8601)
	EndsAt        *string   `json:"ends_at,omitempty"`       // optional scheduled end time (ISO8601)
	Language      *string   `json:"language,omitempty"`      // e.g., "English", "Hindi"
	Level         *string   `json:"level,omitempty"`         // e.g., "Beginner", "Advanced"
	TotalQuestions *int     `json:"total_questions,omitempty"` // optional if inferred from QuestionIDs
}

type ExamResponseDTO struct {
	ExamID               string           `json:"exam_id"`
	Title            string           `json:"title"`
	Description      string           `json:"description,omitempty"`
	Questions        []ExamQuestionDTO    `json:"questions"`         // Full data of each question
	DurationMinutes  int              `json:"duration_minutes"`
	TotalMarks       float64          `json:"total_marks"`
	AllowNegativeMark bool            `json:"allow_negative_mark"`
	AttemptsAllowed  int              `json:"attempts_allowed"`
	IsPublic         bool             `json:"is_public"`
	StartsAt         string           `json:"starts_at,omitempty"`
	EndsAt           string           `json:"ends_at,omitempty"`
	CreatedAt        string           `json:"created_at"`
	UpdatedAt        string           `json:"updated_at"`
}




type QuestionDTO struct {
	QuestionID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Options     models.Option`json:"options,omitempty"`
	Marks       float64  `json:"marks"`
	Attachments []string `json:"attachments,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Difficulty  string   `json:"difficulty,omitempty"`
}

type OptionDTO struct {
	OptionID        string `json:"option_id"`
	Text      string `json:"text"`
	Media     string `json:"media,omitempty"`
	IsCorrect bool   `json:"is_correct,omitempty"` // Only respected in backend
}

type CreateQuestionDTO struct {
	QuestionID        string      `json:"question_id" binding:"required"`
	Type              string      `json:"type" binding:"required"`
	Question          string      `json:"question" binding:"required"`
	Options           []OptionDTO `json:"options,omitempty"`
	Difficulty        string      `json:"difficulty,omitempty"`
	Language          string      `json:"language,omitempty"`
	RandomizeOptions  bool        `json:"randomize_options,omitempty"`
	Title             string      `json:"title" binding:"required"`
	Description       string      `json:"description" binding:"required"`
	CorrectOptionIDs  []string    `json:"correct_option_ids,omitempty"`
	AnswerKey         string      `json:"answer_key,omitempty"`
	Marks             float64     `json:"marks" binding:"required"`
	NegativeMark      float64     `json:"negative_mark,omitempty"`
	Tags              []string    `json:"tags,omitempty"`
	Category          string      `json:"category,omitempty"`
	SubCategory       string      `json:"sub_category,omitempty"`
	Attachments       []string    `json:"attachments,omitempty"`
	Explanation       string      `json:"explanation,omitempty"`
	IsActive          bool        `json:"is_active"`
}

type UpdateQuestionDTO struct {
	Title            *string      `json:"title,omitempty"`
	Description      *string      `json:"description,omitempty"`
	Type             *string      `json:"type,omitempty"`
	Question         *string      `json:"question,omitempty"`
	Options          *[]OptionDTO `json:"options,omitempty"`
	CorrectOptionIDs *[]string    `json:"correct_option_ids,omitempty"`
	AnswerKey        *string      `json:"answer_key,omitempty"`
	Marks            *float64     `json:"marks,omitempty"`
	NegativeMark     *float64     `json:"negative_mark,omitempty"`
	Difficulty       *string      `json:"difficulty,omitempty"`
	Language         *string      `json:"language,omitempty"`
	Tags             *[]string    `json:"tags,omitempty"`
	Category         *string      `json:"category,omitempty"`
	SubCategory      *string      `json:"sub_category,omitempty"`
	RandomizeOptions *bool        `json:"randomize_options,omitempty"`
	Attachments      *[]string    `json:"attachments,omitempty"`
	Explanation      *string      `json:"explanation,omitempty"`
	IsActive         *bool        `json:"is_active,omitempty"`
}





type ExamOptionDTO struct {
	OptionID    string `json:"option_id"`
	Text  string `json:"text"`
	Media string `json:"media,omitempty"`
}

type ExamQuestionDTO struct {
	QuestionID string      `json:"question_id"`
	Type       string      `json:"type"`
	Question   string      `json:"question"`
	Options    []OptionDTO `json:"options,omitempty"`
	Language   string      `json:"language,omitempty"`
	Difficulty string      `json:"difficulty,omitempty"`
	Title      string      `json:"title"`
	Description string     `json:"description"`
	Marks      float64     `json:"marks"`
	Tags       []string    `json:"tags,omitempty"`
	Category   string      `json:"category,omitempty"`
	SubCategory string     `json:"sub_category,omitempty"`
}


type AnswerSubmissionDTO struct {
	QuestionID 		string `json:"question_id" binding:"required"`
	Answer     		string `json:"answer" binding:"required"`       // For text-based or typed answers
	OptionID   			string `json:"option_id,omitempty"`    // For MCQ, this is the chosen option
}

type ExamResultRequestDTO struct {
	ExamID    		string                 	`json:"exam_id" binding:"required"`
	Title 			string 					`json:"title" binding:"required"`
	Answers   		[]AnswerSubmissionDTO 	`json:"answers" binding:"required,dive"`
	StartsAt  		string                 	`json:"starts_at" binding:"required"` // ISO 8601
	EndsAt    		string                 	`json:"ends_at" binding:"required"`   // ISO 8601
}
