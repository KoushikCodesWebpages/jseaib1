package settings

import (

	"RAAS/internal/models"

	"log"
	"net/http"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"context"
)

const resetPasskey = "reset@arshan.de" 
// You can use os.Getenv("RESET_PASSKEY") in production to make this more secure
var allowedCollections = map[string]struct{}{
	"auth_users":				{},
    "seekers":                  {},
    "preferences":              {},
	"profile_pic":              {},
    "notifications":            {},


    "cover_letters":            {},
    "cv":                       {},
    "selected_job_applications": {},
    "external_jobs":            {},
    "match_scores":             {},

	"saved_jobs":               {},
	"exam_results":             {},
    "job_research_results":     {},
    // add more if needed
}

type ResetRequest struct {
    Passkey     string   `json:"passkey"`
    Email       string   `json:"email"`
    Collections []string `json:"collections,omitempty"`
}

// ResetDBHandler handles the logic for resetting the DB (deleting a user and associated data)
func ResetDBHandler(c *gin.Context) {
    var req ResetRequest
    if err := c.ShouldBindJSON(&req); err != nil || req.Passkey != resetPasskey {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid passkey or bad request"})
        return
    }

    db := c.MustGet("db").(*mongo.Database)
    var authUser models.AuthUser
    if err := db.Collection("auth_users").FindOne(c, bson.M{"email": req.Email}).Decode(&authUser); err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
        }
        return
    }
    userID := authUser.AuthUserID

    if _, err := db.Collection("auth_users").DeleteOne(c, bson.M{"_id": authUser.AuthUserID}); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
        return
    }

    collectionsToDelete := req.Collections
    if len(collectionsToDelete) == 0 {
        for col := range allowedCollections {
            collectionsToDelete = append(collectionsToDelete, col)
        }
    }

    for _, col := range collectionsToDelete {
        if _, ok := allowedCollections[col]; !ok {
            log.Printf("❌ Attempt to delete unauthorized collection: %s", col)
            continue
        }
        deleteUserDataFromCollection(c, db, col, userID)
    }

    c.JSON(http.StatusOK, gin.H{"message": "User and associated data deleted successfully."})
}


func deleteUserDataFromCollection(ctx context.Context, db *mongo.Database, collectionName, userID string) {
    count, err := db.Collection(collectionName).CountDocuments(ctx, bson.M{"auth_user_id": userID})
    if err != nil {
        log.Printf("❌ Error checking '%s': %v", collectionName, err)
        return
    }
    if count == 0 {
        log.Printf("🔄 No documents found in '%s', skipping.", collectionName)
        return
    }
    if _, err := db.Collection(collectionName).DeleteMany(ctx, bson.M{"auth_user_id": userID}); err != nil {
        log.Printf("❌ Error deleting from %s: %v", collectionName, err)
    } else {
        log.Printf("✅ Deleted data from %s", collectionName)
    }
}


// PrintAllCollectionsHandler handles the logic for listing all collections in the DB
func PrintAllCollectionsHandler(c *gin.Context) {
	// Get MongoDB database object
	db := c.MustGet("db").(*mongo.Database) // Changed to *mongo.Database
	log.Println("🔄 Fetching all collections from the database")

	// List all collections in the database
	collections, err := db.ListCollectionNames(c, bson.M{})
	if err != nil {
		log.Printf("❌ Error listing collections: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list collections"})
		return
	}

	// Iterate over all collections and print their contents
	for _, collectionName := range collections {
		// Fetch all documents from the collection
		cursor, err := db.Collection(collectionName).Find(c, bson.M{})
		if err != nil {
			log.Printf("❌ Error fetching documents from collection %s: %v", collectionName, err)
			continue
		}

		var documents []bson.M
		if err := cursor.All(c, &documents); err != nil {
			log.Printf("❌ Error reading documents from collection %s: %v", collectionName, err)
			continue
		}

		// You can also print this directly to the response if needed
		c.JSON(http.StatusOK, gin.H{
			"collection": collectionName,
			"documents":  documents,
		})
	}
}
