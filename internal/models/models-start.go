package models

import (

	"RAAS/core/config"

	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoDB *mongo.Database
func InitDB(cfg *config.Config) (*mongo.Client, *mongo.Database) {
	// Create MongoDB client options using the URI from the config
	clientOptions := options.Client().ApplyURI(cfg.Cloud.MongoDBUri)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("❌ Error connecting to MongoDB: %v", err)
	}

	// Ping MongoDB to check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("❌ Error pinging MongoDB: %v", err)
	}

	// Set the global MongoDB variable to the database from the config
	MongoDB = client.Database(cfg.Cloud.MongoDBName)
	log.Println("✅ MongoDB connection established")

	// Print all collections (optional)
	// air


	// Optionally reset collections (this function could be defined elsewhere if needed)
	// resetCollections()

	// Call the CreateAllIndexes function to create the necessary indexes for all models
	CreateAllIndexes()

	// Now, select the "jobs" collection
	// collection := MongoDB.Collection("jobs") // Replace with your actual collection name
	// SeedJobs(collection)


	// Return the client and MongoDB database instances
	return client, MongoDB
}

// Reset collections if necessary
func resetCollections() {
	collections := []string{
		"admins",
		"auth_users",
		"seekers", 
		"match_scores",
		"user_entry_timelines",
		"selected_job_applications",
		"cover_letters", 
		"cv", 
		"saved_jobs", 
		
	}

	for _, col := range collections {
		err := MongoDB.Collection(col).Drop(context.TODO())
		if err != nil {
			log.Printf("⚠️ Error resetting collection %s: %v", col, err)
		} else {
			log.Printf("✅ Collection %s reset", col)
		}
	}
}

// Print all collections
func PrintAllCollections() {
	collections, err := MongoDB.ListCollectionNames(context.TODO(), bson.M{})
	if err != nil {
		log.Fatalf("❌ Error fetching collection names: %v", err)
	}

	log.Println("📦 Collections in the database:")
	for _, col := range collections {
		log.Println(" -", col)
	}
}






type IndexCreationTask struct {
	CollectionName     string
	CreateIndexesFunc  func(collection *mongo.Collection) error
}

func CreateAllIndexes() {
	// Define the index creation tasks for all collections
	indexTasks := []IndexCreationTask{
		{
			CollectionName:    "auth_users",
			CreateIndexesFunc: CreateAuthUserIndexes,
		},
		{
			CollectionName:    "seekers",
			CreateIndexesFunc: CreateSeekerIndexes,
		},
		{
			CollectionName:    "admins",
			CreateIndexesFunc: CreateAdminIndexes,
		},
		{
			CollectionName: "saved_jobs",
			CreateIndexesFunc: CreateSavedJobApplicationIndexes,
		},
		{
			CollectionName:    "user_entry_timelines",
			CreateIndexesFunc: CreateUserEntryTimelineIndexes,
		},
		{
			CollectionName:    "selected_job_applications",
			CreateIndexesFunc: CreateSelectedJobApplicationIndexes,
		},
		
		{
			CollectionName:    "cover_letters",
			CreateIndexesFunc: CreateCoverLetterIndexes, // Add CoverLetter index creation
		},
		{
			CollectionName:    "cv",
			CreateIndexesFunc: CreateCVIndexes, // Add CV index creation
		},
		{
			CollectionName:    "match_scores", // Add MatchScore index creation
			CreateIndexesFunc: CreateMatchScoreIndexes, // Add MatchScore compound index for authUserId and jobId
		},
		{
			CollectionName:    "jobs", // Add Job index creation
			CreateIndexesFunc: CreateJobIndexes, // Add Job index creation (hash for selected count and unique for jobId/jobLink)
		},
	}
	
	// Iterate over each task and execute the index creation
	for _, task := range indexTasks {
		collection := MongoDB.Collection(task.CollectionName)
		if err := task.CreateIndexesFunc(collection); err != nil {
			log.Fatalf("Failed to create indexes for %s: %v", task.CollectionName, err)
		} else {
			// log.Printf("Indexes for %s created successfully!", task.CollectionName)
		}
	}
}