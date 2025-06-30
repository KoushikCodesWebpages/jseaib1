package preference

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	_ "image/gif"
	_ "image/png"
)

type PhotoHandler struct{}

func NewPhotoHandler() *PhotoHandler {
	return &PhotoHandler{}
}

// UploadProfilePhoto handles photo upload with resizing to 640x640
func (h *PhotoHandler) UploadProfilePhoto(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	userID := c.MustGet("userID").(string)
	seekers := db.Collection("seekers")

	file, _, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"issue": "No photo file found in request", "error": err.Error()})
		return
	}
	defer file.Close()

	// Decode the uploaded image
	imgData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to read photo data", "error": err.Error()})
		return
	}
	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"issue": "Invalid image format", "error": err.Error()})
		return
	}

	// Resize to 640x640
	resized := imaging.Resize(img, 640, 640, imaging.Lanczos)

	// Encode as JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 75}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to encode image", "error": err.Error()})
		return
	}
	finalBytes := buf.Bytes()

	// Update seeker doc
	filter := bson.M{"auth_user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"photo":      finalBytes,
			"updated_at": time.Now(),
		},
	}

	res, err := seekers.UpdateOne(c, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to store profile photo", "error": err.Error()})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"issue": "Seeker not found"})
		return
	}

	photoURL := "/b1/photo/view/" + userID
	c.JSON(http.StatusOK, gin.H{"issue": "Profile photo uploaded successfully", "photo_url": photoURL})
}

// GetProfilePhoto returns authenticated user's profile photo
func (h *PhotoHandler) GetProfilePhoto(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	userID := c.MustGet("userID").(string)
	sendProfilePhotoByID(c, db, userID)
}

// PublicGetProfilePhoto allows viewing a profile photo via URL
func (h *PhotoHandler) PublicGetProfilePhoto(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	userID := c.Param("user_id")
	sendProfilePhotoByID(c, db, userID)
}

// shared logic to send photo
func sendProfilePhotoByID(c *gin.Context, db *mongo.Database, userID string) {
	seekers := db.Collection("seekers")

	var result struct {
		Photo []byte `bson:"photo"`
	}
	err := seekers.FindOne(c, bson.M{"auth_user_id": userID}).Decode(&result)
	if err != nil || len(result.Photo) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"issue": "Photo not found"})
		return
	}
	c.Data(http.StatusOK, "image/jpeg", result.Photo)
}
