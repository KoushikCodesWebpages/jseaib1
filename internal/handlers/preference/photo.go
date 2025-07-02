package preference

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"time"

	"github.com/disintegration/imaging"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	_ "image/gif"
	_ "image/png"
)

type PhotoHandler struct{}

func NewPhotoHandler() *PhotoHandler {
	return &PhotoHandler{}
}

func (h *PhotoHandler) UploadProfilePhoto(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	userIDStr := c.MustGet("userID").(string)

	authUserID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"issue": "Invalid user ID", "error": err.Error()})
		return
	}

	file, _, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"issue": "No photo file found in request", "error": err.Error()})
		return
	}
	defer file.Close()

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

	resized := imaging.Resize(img, 640, 640, imaging.Lanczos)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 75}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to encode image", "error": err.Error()})
		return
	}

	finalBytes := buf.Bytes()
	now := time.Now()

	// Save to profile_pics collection
	profilePics := db.Collection("profile_pics")
	_, err = profilePics.UpdateOne(c,
		bson.M{"auth_user_id": authUserID},
		bson.M{
			"$set": bson.M{
				"image":      finalBytes,
				"mime_type":  "image/jpeg",
				"updated_at": now,
			},
			"$setOnInsert": bson.M{
				"auth_user_id": authUserID,
				"created_at":   now,
			},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to store profile photo", "error": err.Error()})
		return
	}

	// Update photo URL in seekers collection
	photoURL := "/b1/photo/view/" + userIDStr
	seekers := db.Collection("seekers")
	_, err = seekers.UpdateOne(c,
		bson.M{"auth_user_id": userIDStr},
		bson.M{"$set": bson.M{"photo": photoURL, "updated_at": now}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to update seeker photo URL", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issue":     "Profile photo uploaded successfully",
		"photo_url": photoURL,
	})
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

// shared logic to send profile photo from profile_pics collection
func sendProfilePhotoByID(c *gin.Context, db *mongo.Database, userID string) {
	profilePics := db.Collection("profile_pics")

	authUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"issue": "Invalid user ID", "error": err.Error()})
		return
	}

	var result struct {
		Image    []byte `bson:"image"`
		MimeType string `bson:"mime_type"`
	}

	err = profilePics.FindOne(c, bson.M{"auth_user_id": authUserID}).Decode(&result)
	if err != nil || len(result.Image) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"issue": "Profile photo not found"})
		return
	}

	contentType := result.MimeType
	if contentType == "" {
		contentType = "image/jpeg" // fallback default
	}

	c.Data(http.StatusOK, contentType, result.Image)
}
