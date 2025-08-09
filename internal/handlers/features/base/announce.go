package base

import (
	
	"net/http"


	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"RAAS/internal/models"
)
func AnnouncementHandler(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)

	// Always filter active
	filter := bson.M{"is_active": true}

	// If "recent" param is true → return single latest announcement
	if c.Query("recent") == "true" {
		var announcement models.Announcement
		err := db.Collection("announcements").
			FindOne(c,
				filter,
				options.FindOne().
					SetSort(bson.D{{Key: "created_at", Value: -1}}),
			).Decode(&announcement)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusOK, gin.H{"announcement": nil})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching recent announcement"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"announcement": announcement})
		return
	}

	// Otherwise → paginated active announcements list
	pagination := c.MustGet("pagination").(gin.H)
	offset := pagination["offset"].(int)
	limit := pagination["limit"].(int)

	total, err := db.Collection("announcements").CountDocuments(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting announcements"})
		return
	}

	cursor, err := db.Collection("announcements").
		Find(c,
			filter,
			options.Find().
				SetSort(bson.D{{Key: "created_at", Value: -1}}).
				SetSkip(int64(offset)).
				SetLimit(int64(limit)),
		)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching announcements"})
		return
	}
	defer cursor.Close(c)

	var announcements []models.Announcement
	if err := cursor.All(c, &announcements); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding announcements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pagination": gin.H{
			"total":    total,
			"current":  (offset / limit) + 1,
			"per_page": limit,
		},
		"announcements": announcements,
	})
}
