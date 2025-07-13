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

const resetPasskey = "reset@arshan.de" // You can use os.Getenv("RESET_PASSKEY") in production to make this more secure

// ResetRequest defines the structure of the reset request payload
type ResetRequest struct {
	Passkey string `json:"passkey"`
	Email   string `json:"email"`
}

// ResetDBHandler handles the logic for resetting the DB (deleting a user and associated data)
func ResetDBHandler(c *gin.Context) {
	var req ResetRequest

	// Validate request
	if err := c.ShouldBindJSON(&req); err != nil || req.Passkey != resetPasskey {
		log.Println("❌ Invalid passkey or bad request")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid passkey or bad request"})
		return
	}

	// Fetch user by email from MongoDB using the database object
	db := c.MustGet("db").(*mongo.Database) // Changed to *mongo.Database
	var authUser models.AuthUser
	log.Printf("🔄 Fetching user by email: %s", req.Email)
	err := db.Collection("auth_users").FindOne(c, bson.M{"email": req.Email}).Decode(&authUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("❌ User not found: %s", req.Email)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			log.Printf("❌ DB error retrieving user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		}
		return
	}

	// Ensure UUID is converted to string
	userID := authUser.AuthUserID // Convert UUID to string
	log.Printf("🔄 Reset triggered for user: %s (ID: %s)", req.Email, userID)

	// Attempt to delete user by string ID
	log.Printf("🔄 Attempting to delete user from auth_users with ID: %s", userID)
	_, err = db.Collection("auth_users").DeleteOne(c, bson.M{"_id": authUser.AuthUserID})
	if err != nil {
		log.Printf("❌ Failed to delete user from auth_users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	// Log the result
	log.Printf("✅ Deleted user with ID: %s", userID)

	// Clean each collection where user data might exist
	collections := []string{
		// "seekers", 
		// "user_entry_timelines", 
		// "cover_letters", 
		// "cv", 
		"selected_job_applications",
		// "jobs", 
			"admins",
			"match_scores", 
			"auth_users",
			"saved_jobs",
			"preferences",
			"notifications",
	}

	// Delete user data from each collection
	for _, collectionName := range collections {
		deleteUserDataFromCollection(c, db, collectionName, userID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "User and associated data deleted successfully."})
}

// Helper function for deleting data from each collection
func deleteUserDataFromCollection(c context.Context, db *mongo.Database, collectionName string, userID string) {
	count, err := db.Collection(collectionName).CountDocuments(c, bson.M{"auth_user_id": userID})
	if err != nil {
		log.Printf("❌ Error checking collection '%s': %v", collectionName, err)
		return // Skip this collection if there's an error
	}

	if count > 0 {
		// Perform deletion
		_, err := db.Collection(collectionName).DeleteMany(c, bson.M{"auth_user_id": userID})
		if err != nil {
			log.Printf("❌ Error deleting from %s: %v", collectionName, err)
		} else {
			log.Printf("✅ Deleted data from %s", collectionName)
		}
	} else {
		log.Printf("🔄 No documents found in collection '%s', skipping deletion.", collectionName)
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
