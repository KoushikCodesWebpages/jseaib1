package repository

import (
	"RAAS/internal/models"
)
func CanAccessAdvancedFeatures(seeker *models.Seeker) bool {
	return seeker.SubscriptionTier == "advanced" || seeker.ProficiencyTest > 0
}
