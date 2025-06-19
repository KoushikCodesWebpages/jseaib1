package jobs

// import (
// 	"net/http"
// 	"github.com/gin-gonic/gin"
// 	"gorm.io/gorm"
// 	"RAAS/dto"
// 	"RAAS/models"
// 	//"github.com/google/uuid"
// )

// // MatchScoreHandler handles match score related requests
// type MatchScoreHandler struct {
// 	DB *gorm.DB
// }

// func (h *MatchScoreHandler) GetAllMatchScores(c *gin.Context) {
// 	var matchScores []models.MatchScore

// 	// Fetch all match scores from the database
// 	if err := h.DB.Find(&matchScores).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch match scores"})
// 		return
// 	}

// 	// Map match scores to response DTO
// 	var matchScoreResponses []dto.MatchScoreResponse
// 	for _, matchScore := range matchScores {
// 		matchScoreResponses = append(matchScoreResponses, dto.MatchScoreResponse{
// 			SeekerID:   matchScore.AuthUserID,
// 			JobID:      matchScore.JobID,
// 			MatchScore: matchScore.MatchScore,
// 		})
// 	}

// 	// Return all match scores as a response
// 	c.JSON(http.StatusOK, gin.H{
// 		"match_scores": matchScoreResponses,
// 	})
// }