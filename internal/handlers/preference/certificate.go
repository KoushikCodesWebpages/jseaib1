package preference

import (
	"RAAS/internal/dto"
	"RAAS/internal/models"
	"RAAS/internal/handlers/repository"

	"context"
	"log"
	"net/http"
	"time"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CertificateHandler struct{}

func NewCertificateHandler() *CertificateHandler {
	return &CertificateHandler{}
}
func (h *CertificateHandler) CreateCertificate(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")
	entryTimelineCollection := db.Collection("user_entry_timelines")

	var input dto.CertificateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		log.Printf("Error binding input: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
			log.Printf("Seeker not found for auth_user_id: %s", userID)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving seeker"})
			log.Printf("Error retrieving seeker for auth_user_id: %s, Error: %v", userID, err)
		}
		return
	}

	// Append the new certificate
	if err := repository.AppendToCertificates(&seeker, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process certificate"})
		log.Printf("Failed to process certificate for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	// Update seeker document
	update := bson.M{
		"$set": bson.M{
			"certificates": seeker.Certificates,
		},
	}

	updateResult, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save certificate"})
		log.Printf("Failed to update certificate for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	if updateResult.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching seeker found to update"})
		log.Printf("No matching seeker found for auth_user_id: %s", userID)
		return
	}

	// Update user entry timeline to mark certificates completed
	timelineUpdate := bson.M{
		"$set": bson.M{
			"certificates_completed": true,
		},
	}
	if _, err := entryTimelineCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, timelineUpdate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user entry timeline"})
		log.Printf("Failed to update user entry timeline for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Certificate added successfully",
	})
}

// GetCertificates handles the retrieval of a user's certificates
func (h *CertificateHandler) GetCertificates(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
			log.Printf("Seeker not found for auth_user_id: %s", userID)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving seeker"})
			log.Printf("Error retrieving seeker for auth_user_id: %s, Error: %v", userID, err)
		}
		return
	}

	if len(seeker.Certificates) == 0 {
		c.JSON(http.StatusNoContent, gin.H{"message": "No certificates found"})
		return
	}

	certificatesRaw, err := repository.GetCertificates(&seeker)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing certificates"})
		log.Printf("Error processing certificates for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	var response []dto.CertificateResponse

	for _, cert := range certificatesRaw {
		certificateName, _ := cert["certificate_name"].(string)
		platform, _ := cert["platform"].(string)
		startDateRaw, _ := cert["start_date"].(primitive.DateTime)
		var endDate *time.Time
		if rawEndDate, ok := cert["end_date"].(primitive.DateTime); ok {
			t := rawEndDate.Time()
			endDate = &t
		}
		createdAtRaw, _ := cert["created_at"].(primitive.DateTime)
		updatedAtRaw, _ := cert["updated_at"].(primitive.DateTime)

		response = append(response, dto.CertificateResponse{
			AuthUserID:      userID,
			CertificateName: certificateName,
			Platform:        platform,
			StartDate:       startDateRaw.Time(),
			EndDate:         endDate,
			CreatedAt:       createdAtRaw.Time(),
			UpdatedAt:       updatedAtRaw.Time(),
		})
		}

		c.JSON(http.StatusOK, gin.H{
			"certificates": response,
		})
	}

func (h *CertificateHandler) UpdateCertificate(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	id := c.Param("id")
	index, err := strconv.Atoi(id)
	if err != nil || index <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid certificate index. Must be a positive integer."})
		return
	}

	var input dto.CertificateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		log.Printf("Error binding input: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
			log.Printf("Seeker not found for auth_user_id: %s", userID)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving seeker"})
			log.Printf("Error retrieving seeker for auth_user_id: %s, Error: %v", userID, err)
		}
		return
	}

	if index > len(seeker.Certificates) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Certificate index out of range"})
		return
	}

	// Update the certificate fields

	updatedCertificate := bson.M{
		"auth_user_id":     userID,
		"certificate_name": input.CertificateName,
		"platform":         input.Platform,
		"start_date":       input.StartDate,
		"end_date":         input.EndDate,
	}

	// Replace the certificate at the given index (1-based)
	seeker.Certificates[index-1] = updatedCertificate

	update := bson.M{
		"$set": bson.M{
			"certificates": seeker.Certificates,
		},
	}

	updateResult, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update certificate"})
		log.Printf("Failed to update certificate for auth_user_id: %s, Error: %v", userID, err)
		return
	}

	if updateResult.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching seeker found to update"})
		log.Printf("No matching seeker found for auth_user_id: %s", userID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Certificate updated successfully",
	})
}



// DeleteCertificate handles deleting a certificate entry
func (h *CertificateHandler) DeleteCertificate(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	id := c.Param("id")
	index, err := strconv.Atoi(id)
	if err != nil || index <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid certificate index"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker models.Seeker
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seeker not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving seeker"})
		return
	}

	if index > len(seeker.Certificates) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Certificate index out of range"})
		return
	}

	// Remove the certificate entry at index-1
	seeker.Certificates = append(seeker.Certificates[:index-1], seeker.Certificates[index:]...)

	// Save updated certificates to DB
	update := bson.M{
		"$set": bson.M{
			"certificates": seeker.Certificates,
		},
	}

	_, err = seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete certificate"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Certificate deleted successfully"})
}
