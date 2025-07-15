package settings

import (
	"RAAS/internal/dto"
	"RAAS/internal/models"

	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SettingsHandler struct{}

func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

func (h *SettingsHandler) GetGeneralSettings(c *gin.Context) {
	// Declaration
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	authUsersCollection := db.Collection("auth_users")

	// Derivation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Operational
	var user models.AuthUser
	err := authUsersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"issue": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to fetch user", "error": err.Error()})
		return
	}

	// Output
	c.JSON(http.StatusOK, gin.H{"email": user.Email})
}

// GetPreferences retrieves user preferences
func (h *SettingsHandler) GetPreferences(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	preferencesCollection := db.Collection("preferences")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var prefs models.UserPreferences
	if err := preferencesCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&prefs); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"issue": "Preferences not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve preferences"})
		}
		return
	}

	response := dto.PreferencesResponse{
		AuthUserID:   prefs.AuthUserID,
		Language:     prefs.Language,
		Timezone:     prefs.Timezone,
		CookiePolicy: prefs.CookiePolicy,
		CreatedAt:    prefs.CreatedAt,
		UpdatedAt:    prefs.UpdatedAt,
	}
	c.JSON(http.StatusOK, response)
}

// UpdatePreferences sets or updates preferences
func (h *SettingsHandler) UpdatePreferences(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	preferencesCollection := db.Collection("preferences")

	var req dto.PreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"issue": "Invalid input", "error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"language":      req.Language,
			"timezone":      req.Timezone,
			"cookie_policy": req.CookiePolicy,
			"updated_at":    time.Now(),
		},
		"$setOnInsert": bson.M{
			"auth_user_id": userID,
			"created_at":   time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	res, err := preferencesCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update, opts)
if err != nil {
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preferences"})
	return
}

	c.JSON(http.StatusOK, gin.H{
		"issue":     "Preferences updated successfully",
		"matched":     res.MatchedCount,
		"modified":    res.ModifiedCount,
		"upserted_id": res.UpsertedID,
	})

	
}

// GetNotificationSettings retrieves notification toggles
func (h *SettingsHandler) GetNotificationSettings(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	collection := db.Collection("notifications")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var ns models.NotificationSettings
	err := collection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&ns)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"issue": "Notification settings not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notification settings"})
		return
	}

	response := dto.NotificationSettingsResponse{
		AuthUserID:      ns.AuthUserID,
		Subscription:    ns.Subscription,
		RecommendedJobs: ns.RecommendedJobs,
		GermanTest:      ns.GermanTest,
		Announcements:   ns.Announcements,
		CreatedAt:       ns.CreatedAt,
		UpdatedAt:       ns.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateNotificationSettings updates or creates notification toggles
func (h *SettingsHandler) UpdateNotificationSettings(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	collection := db.Collection("notifications")

	var req dto.NotificationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"subscription":     req.Subscription,
			"recommended_jobs": req.RecommendedJobs,
			"german_test":      req.GermanTest,
			"announcements":    req.Announcements,
			"updated_at":       time.Now(),
		},
		"$setOnInsert": bson.M{
			"auth_user_id": userID,
			"created_at":   time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification settings updated successfully"})
}

