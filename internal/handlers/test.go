package handlers

import (
	"net/http"
	// "time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAcademicDatesHandler(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	authUserID := c.Query("auth_user_id")

	if authUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "auth_user_id is required"})
		return
	}

	seekColl := db.Collection("seekers")

	var seeker bson.M
	err := seekColl.FindOne(c, bson.M{"auth_user_id": authUserID}).Decode(&seeker)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "seeker not found"})
		return
	}

	academicsRaw, ok := seeker["academics"].(primitive.A)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"academics": []string{}})
		return
	}

	type Response struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}

	var results []Response

	for _, raw := range academicsRaw {
		entry, ok := raw.(bson.M)
		if !ok {
			continue
		}

		var startStr, endStr string

		if start, ok := entry["start_date"].(primitive.DateTime); ok {
			startStr = start.Time().Format("2006-01-02")
		}

		if end, ok := entry["end_date"].(primitive.DateTime); ok && !end.Time().IsZero() {
			endStr = end.Time().Format("2006-01-02")
		} else {
			endStr = "Present"
		}

		results = append(results, Response{
			StartDate: startStr,
			EndDate:   endStr,
		})
	}

	c.JSON(http.StatusOK, results)
}
