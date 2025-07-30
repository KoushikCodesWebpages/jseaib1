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
	var opts []dto.OptionDTO
	for _, o := range q.Options {
		opts = append(opts, dto.OptionDTO{
			OptionID:    o.ID,
			Text:  o.Text,
			Media: o.Media,
		})
	}

	return dto.ExamQuestionDTO{
		QuestionID:  q.QuestionID,
		Type:        q.Type,
		Question:    q.Question,
		Options:     opts, // Now correct type
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