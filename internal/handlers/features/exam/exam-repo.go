package exam

import (
	"RAAS/internal/dto"
	"RAAS/internal/models"
)

// func compareOptions(selected, correct []string) bool {
// 	if len(selected) != len(correct) {
// 		return false
// 	}

// 	selectedMap := make(map[string]bool)
// 	for _, s := range selected {
// 		selectedMap[s] = true
// 	}

// 	for _, c := range correct {
// 		if !selectedMap[c] {
// 			return false
// 		}
// 	}

// 	return true
// }


func deref(ptr *string, fallback string) string {
	if ptr != nil {
		return *ptr
	}
	return fallback
}

func derefBool(ptr *bool, fallback bool) bool {
	if ptr != nil {
		return *ptr
	}
	return fallback
}

func ToQuestionDTO(q models.Question) dto.ExamQuestionDTO {
	var optionDTOs []dto.OptionDTO
	for _, option := range q.Options {
		// Correctly map OptionID from the original model
		optionDTOs = append(optionDTOs, dto.OptionDTO{
			OptionID: option.OptionID, // This line is crucial
			Text:     option.Text,
		})
	}

	return dto.ExamQuestionDTO{
		QuestionID:  q.QuestionID,
		Type:        q.Type,
		Question:    q.Question,
		Options:     optionDTOs, // Pass the new, sanitized options
		Language:    q.Language,
		Difficulty:  q.Difficulty,
		Title:       q.Title,
		Description: q.Description,
		Marks:       q.Marks,
		Tags:        q.Tags,
		Category:    q.Category,
		SubCategory: q.SubCategory,
	}
}

func mapOptionDTOs(input []dto.OptionDTO) []models.Option {
	opts := make([]models.Option, len(input))
	for i, o := range input {
		opts[i] = models.Option{
			OptionID:  o.OptionID,
			Text:      o.Text,
			Media:     o.Media,
			IsCorrect: o.IsCorrect,
		}
	}
	return opts
}