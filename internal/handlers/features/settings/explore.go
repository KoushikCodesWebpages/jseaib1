package settings

import (
	"context"
	"net/http"
	"time"

	"RAAS/internal/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PlanDTO struct {
	Plan   string  `json:"plan"`
	Price  float64 `json:"price"`
	Period string  `json:"period"`
	Status string  `json:"status"`
}

var explorePlanMatrix = map[string]map[string][]PlanDTO{
	"free_monthly": {
		"monthly": {
			{Plan: "free", Price: 0, Period: "1 month", Status: "active"},
			{Plan: "basic", Price: 25, Period: "1 month", Status: "upgrade"},
			{Plan: "advanced", Price: 35, Period: "1 month", Status: "coming soon"},
			{Plan: "premium", Price: 55, Period: "1 month", Status: "coming soon"},
		},
		"quarterly": {
			{Plan: "free", Price: 0, Period: "3 months", Status: "active"},
			{Plan: "basic", Price: 68, Period: "3 months", Status: "upgrade"},
			{Plan: "advanced", Price: 95, Period: "3 months", Status: "coming soon"},
			{Plan: "premium", Price: 149, Period: "3 months", Status: "coming soon"},
		},
	},
	"basic_monthly": {
		"monthly": {
			{Plan: "basic", Price: 25, Period: "1 month", Status: "cancel"},
			{Plan: "advanced", Price: 35, Period: "1 month", Status: "coming soon"},
			{Plan: "premium", Price: 55, Period: "1 month", Status: "coming soon"},
		},
		"quarterly": {
			{Plan: "basic", Price: 68, Period: "3 months", Status: "upgrade"},
			{Plan: "advanced", Price: 95, Period: "3 months", Status: "coming soon"},
			{Plan: "premium", Price: 149, Period: "3 months", Status: "coming soon"},
		},
	},

    	"basic_quarterly": {
		"monthly": {
			{Plan: "basic", Price: 25, Period: "1 month", Status: "downgrade"},
			{Plan: "advanced", Price: 35, Period: "1 month", Status: "coming soon"},
			{Plan: "premium", Price: 55, Period: "1 month", Status: "coming soon"},
		},
		"quarterly": {
			{Plan: "basic", Price: 68, Period: "3 months", Status: "cancel"},
			{Plan: "advanced", Price: 95, Period: "3 months", Status: "coming soon"},
			{Plan: "premium", Price: 149, Period: "3 months", Status: "coming soon"},
		},
	},

	"advanced_monthly": {
        "monthly": {
			{Plan: "basic", Price: 25, Period: "1 month", Status: "downgrade"},
			{Plan: "advanced", Price: 35, Period: "1 month", Status: "cancel"},
			{Plan: "premium", Price: 55, Period: "1 month", Status: "coming soon"},
		},
		"quarterly": {
			{Plan: "basic", Price: 68, Period: "3 months", Status: "downgrade"},
			{Plan: "advanced", Price: 95, Period: "3 months", Status: "upgrade"},
			{Plan: "premium", Price: 149, Period: "3 months", Status: "coming soon"},
		},
	},

    	"advanced_quarterly": {
		"monthly": {
			{Plan: "basic", Price: 25, Period: "1 month", Status: "downgrade"},
			{Plan: "advanced", Price: 35, Period: "1 month", Status: "downgrade"},
			{Plan: "premium", Price: 55, Period: "1 month", Status: "coming soon"},
		},
		"quarterly": {
			{Plan: "basic", Price: 68, Period: "3 months", Status: "downgrade"},
			{Plan: "advanced", Price: 95, Period: "3 months", Status: "cancel"},
			{Plan: "premium", Price: 149, Period: "3 months", Status: "coming soon"},
		},
	},
    
	"premium_monthly": {
        "monthly": {
			{Plan: "basic", Price: 25, Period: "1 month", Status: "downgrade"},
			{Plan: "advanced", Price: 35, Period: "1 month", Status: "downgrade"},
			{Plan: "premium", Price: 55, Period: "1 month", Status: "cancel"},
		},
		"quarterly": {
			{Plan: "basic", Price: 68, Period: "3 months", Status: "downgrade"},
			{Plan: "advanced", Price: 95, Period: "3 months", Status: "downgrade"},
			{Plan: "premium", Price: 149, Period: "3 months", Status: "coming soon"},
		},
	},


	"premium_quarterly": {
		"monthly": {
			{Plan: "basic", Price: 25, Period: "1 month", Status: "downgrade"},
			{Plan: "advanced", Price: 35, Period: "1 month", Status: "downgrade"},
			{Plan: "premium", Price: 55, Period: "1 month", Status: "downgrade"},
		},
		"quarterly": {
			{Plan: "basic", Price: 68, Period: "3 months", Status: "downgrade"},
			{Plan: "advanced", Price: 95, Period: "3 months", Status: "cancel"},
			{Plan: "premium", Price: 149, Period: "3 months", Status: "cancel"},
		},
	},
}

// GetExplorePlans returns the correct set of plans for a user based on their subscription tier and period
func (h *SettingsHandler) GetExplorePlans(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := db.Collection("seekers").
		FindOne(ctx, bson.M{"auth_user_id": userID}).
		Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		}
		return
	}

	// Determine matrix key
	key := seeker.SubscriptionTier + "_" + seeker.SubscriptionPeriod
	plans, ok := explorePlanMatrix[key]
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid subscription data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plans": plans})
}
