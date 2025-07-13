package preference

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// FormatRequest is the expected body for updating format
type FormatRequest struct {
	Format string `json:"format" binding:"required"`
}

// Allowed formats for CV and CL
var allowedCvFormats = map[string]bool{
	"minimalist": true,
	"classic":    true,
	"modern":     true,
	"elegant":    true,
	"creative":   true,
}

var allowedClFormats = map[string]bool{
	"formal":     true,
	"casual":     true,
	"concise":    true,
	"narrative":  true,
	"technical":  true,
}

// FormatHandler groups the format-related handlers
type FormatHandler struct{}

func NewFormatHandler() *FormatHandler {
	return &FormatHandler{}
}

// UpdateCvFormat updates the CV format
func (h *FormatHandler) UpdateCvFormat(c *gin.Context) {
	updateFormatField(c, "cv_format", allowedCvFormats)
}

// UpdateClFormat updates the Cover Letter format
func (h *FormatHandler) UpdateClFormat(c *gin.Context) {
	updateFormatField(c, "cl_format", allowedClFormats)
}

// updateFormatField is the shared logic for format updates
func updateFormatField(c *gin.Context, field string, allowed map[string]bool) {
	db := c.MustGet("db").(*mongo.Database)
	userID := c.MustGet("userID").(string)
	seekers := db.Collection("seekers")

	var req FormatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"issue": "Invalid request body",
			"error": err.Error(),
		})
		return
	}

	if !allowed[req.Format] {
		c.JSON(http.StatusBadRequest, gin.H{
			"issue":           "Invalid format",
			"allowed_formats": keys(allowed),
		})
		return
	}

	filter := bson.M{"auth_user_id": userID}
	update := bson.M{
		"$set": bson.M{
			field:       req.Format,
			"updated_at": time.Now(),
		},
	}

	res, err := seekers.UpdateOne(c, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"issue": "Failed to update format",
			"error": err.Error(),
		})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"issue": "Seeker not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issue": field + " updated successfully",
	})
}

// keys returns a list of allowed formats
func keys(m map[string]bool) []string {
	k := make([]string, 0, len(m))
	for key := range m {
		k = append(k, key)
	}
	return k
}
