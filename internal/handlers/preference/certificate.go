package preference

import (
    "context"
    "log"
    "net/http"
    "time"
    "strconv"
    
    "RAAS/internal/dto"
    "RAAS/internal/handlers/repository"
    "RAAS/internal/models"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

type CertificateHandler struct{}

func NewCertificateHandler() *CertificateHandler {
    return &CertificateHandler{}
}

// CreateCertificate handles creating or updating a certificate
func (h *CertificateHandler) CreateCertificate(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    db := c.MustGet("db").(*mongo.Database)
    seekers := db.Collection("seekers")
    timelines := db.Collection("user_entry_timelines")

    var input dto.CertificateRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        log.Printf("Bind error [CreateCertificate] user=%s: %v", userID, err)
        c.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
            "issue": "Please check your input and try again.",
        })
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var seeker models.Seeker
    if err := seekers.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
        log.Printf("DB fetch error [CreateCertificate] user=%s: %v", userID, err)
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{
                "error": err.Error(),
                "issue": "Could not find your account. Please log in again.",
            })
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": err.Error(),
                "issue": "Failed to retrieve your profile. Try again later.",
            })
        }
        return
    }

    if err := repository.AppendToCertificates(&seeker, input); err != nil {
        log.Printf("Append error [CreateCertificate] user=%s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": err.Error(),
            "issue": "Could not save your certificate details. Try again shortly.",
        })
        return
    }

    updateResult, err := seekers.UpdateOne(ctx, bson.M{"auth_user_id": userID}, bson.M{"$set": bson.M{"certificates": seeker.Certificates}})
    if err != nil {
        log.Printf("DB update error [CreateCertificate] user=%s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": err.Error(),
            "issue": "Failed to save your certificate. Please try again.",
        })
        return
    }
    if updateResult.MatchedCount == 0 {
        log.Printf("No document matched [CreateCertificate] user=%s", userID)
        c.JSON(http.StatusNotFound, gin.H{
            "error": "no seeker matched",
            "issue": "Your account was not found to update.",
        })
        return
    }

    if _, err := timelines.UpdateOne(ctx, bson.M{"auth_user_id": userID}, bson.M{"$set": bson.M{"certificates_completed": true}}); err != nil {
        log.Printf("Timeline update error [CreateCertificate] user=%s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": err.Error(),
            "issue": "Certificate saved, but progress tracking failed.",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{"issue": "Certificate added successfully"})
}

// GetCertificates retrieves a user's certificate records
func (h *CertificateHandler) GetCertificates(c *gin.Context) {
    userID := c.MustGet("userID").(string)
    db := c.MustGet("db").(*mongo.Database)
    seekers := db.Collection("seekers")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var seeker models.Seeker
    if err := seekers.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
        log.Printf("DB fetch error [GetCertificates] user=%s: %v", userID, err)
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{
                "error": err.Error(),
                "issue": "Account not found. Please log in.",
            })
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": err.Error(),
                "issue": "Failed to retrieve your certificates. Try again later.",
            })
        }
        return
    }

    if len(seeker.Certificates) == 0 {
        c.JSON(http.StatusNoContent, gin.H{"message": "No certificates found"})
        return
    }

    certificatesRaw, err := repository.GetCertificates(&seeker)
    if err != nil {
        log.Printf("Parse error [GetCertificates] user=%s: %v", userID, err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": err.Error(),
            "issue": "Failed to load your certificates. Data Has Been Corrupted",
        })
        return
    }

    var response []dto.CertificateResponse
    for _, cert := range certificatesRaw {
        name, _ := cert["certificate_name"].(string)
        ctype, _ := cert["certificate_type"].(string)
        providerPtr, _ := cert["provider"].(string)
        completionRaw, _ := cert["completion_date"].(primitive.DateTime)
        createdRaw, _ := cert["created_at"].(primitive.DateTime)
        updatedRaw, _ := cert["updated_at"].(primitive.DateTime)

        response = append(response, dto.CertificateResponse{
            AuthUserID:       userID,
            CertificateName:  name,
            CertificateType:  ctype,
            Provider:         &providerPtr,
            CompletionDate:   completionRaw.Time(),
            CreatedAt:        createdRaw.Time(),
            UpdatedAt:        updatedRaw.Time(),
        })
    }

    c.JSON(http.StatusOK, gin.H{"certificates": response})
}

func (h *CertificateHandler) UpdateCertificate(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	indexStr := c.Param("id")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_index",
			"issue": "Invalid certificate index",
		})
		return
	}

	var input dto.CertificateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"issue": "Invalid input format",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker struct {
		Certificates []bson.M `bson:"certificates"`
	}
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		status := http.StatusInternalServerError
		issue := "Failed to retrieve seeker"
		if err == mongo.ErrNoDocuments {
			status = http.StatusNotFound
			issue = "Seeker not found"
		}
		c.JSON(status, gin.H{
			"error": err.Error(),
			"issue": issue,
		})
		return
	}

	if index > len(seeker.Certificates) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "index_out_of_range",
			"issue": "Certificate index is out of range",
		})
		return
	}

	// Update certificate at index - 1
	seeker.Certificates[index-1] = bson.M{
		"certificate_name": input.CertificateName,
		"certificate_type": input.CertificateType,
		"provider":         input.Provider,
		"completion_date":  input.CompletionDate,
		"updated_at":       time.Now(),
	}

	update := bson.M{
		"$set": bson.M{"certificates": seeker.Certificates},
		"$currentDate": bson.M{"updated_at": true},
	}

	if _, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Failed to update certificate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"issue": "Certificate updated successfully"})
}

func (h *CertificateHandler) DeleteCertificate(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	db := c.MustGet("db").(*mongo.Database)
	seekersCollection := db.Collection("seekers")

	indexStr := c.Param("id")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_index",
			"issue": "Invalid certificate index",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var seeker struct {
		Certificates []bson.M `bson:"certificates"`
	}
	if err := seekersCollection.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
		status := http.StatusInternalServerError
		issue := "Failed to retrieve seeker"
		if err == mongo.ErrNoDocuments {
			status = http.StatusNotFound
			issue = "Seeker not found"
		}
		c.JSON(status, gin.H{
			"error": err.Error(),
			"issue": issue,
		})
		return
	}

	if index > len(seeker.Certificates) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "index_out_of_range",
			"issue": "Certificate index is out of range",
		})
		return
	}

	seeker.Certificates = append(seeker.Certificates[:index-1], seeker.Certificates[index:]...)

	update := bson.M{
		"$set": bson.M{"certificates": seeker.Certificates},
		"$currentDate": bson.M{"updated_at": true},
	}

	if _, err := seekersCollection.UpdateOne(ctx, bson.M{"auth_user_id": userID}, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"issue": "Failed to delete certificate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"issue": "Certificate deleted successfully"})
}
