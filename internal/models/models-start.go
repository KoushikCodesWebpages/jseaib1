package models

import (
	"RAAS/core/config"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global MongoDB instance
var MongoDB *mongo.Database

// Collection name constants
const (
	CollectionAuthUsers            	= "auth_users"
	CollectionSeekers              	= "seekers"
	CollectionAdmins               	= "admins"
	CollectionSavedJobs            	= "saved_jobs"
	CollectionUserEntryTimelines   	= "user_entry_timelines"
	CollectionSelectedJobApps      	= "selected_job_applications"
	CollectionCoverLetters         	= "cover_letters"
	CollectionCV                   	= "cv"
	CollectionMatchScores          	= "match_scores"
	CollectionExtJobs			   	= "external_jobs"
	CollectionJobResearch			= "job_research"
	CollectionJobs                 	= "jobs"
	CollectionCounter			   	= "counters"
	CollectionProfilePic		   	= "profile_pic"
	CollectionPreferences		   	= "preferences"
	CollectionNotifications		   	= "notifications"
	CollectionQuestions				= "questions"
	CollectionResults				= "exam_results"
	CollectionAnnouncements		   	= "announcements"
	
)

// InitDB connects to MongoDB, initializes indexes, and optionally creates collections
func InitDB(cfg *config.Config) (*mongo.Client, *mongo.Database) {
	clientOptions := options.Client().ApplyURI(cfg.Cloud.MongoDBUri)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("❌ Error connecting to MongoDB: %v", err)
	}
	
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("❌ Error pinging MongoDB: %v", err)
	}

	MongoDB = client.Database(cfg.Cloud.MongoDBName)
	log.Println("✅ MongoDB connection established")

	// resetCollections()

	// // // Explicit collection creation (optional)
	CreateCollectionsExplicitly([]string{
	// // 	CollectionAuthUsers,
	// // 	CollectionSeekers,
	// // 	CollectionAdmins,
	// // 	CollectionSavedJobs,
	// // 	CollectionUserEntryTimelines,
	// // 	CollectionSelectedJobApps,
	// // 	CollectionCoverLetters,
	// // 	CollectionCV,
	// 	CollectionMatchScores,
	// // 	CollectionProfilePic,
	// 	CollectionJobs,
	CollectionJobResearch,
	CollectionAnnouncements,
	// // 	CollectionExtJobs,
	// // 	CollectionPreferences,
	// // 	CollectionNotifications,
	// // 	CollectionQuestions,
	// // 	CollectionResults,
	
	})
	
	// // // // Create indexes
	// CreateAllIndexes()

	return client, MongoDB
}

// Explicitly create collections if not present
func CreateCollectionsExplicitly(collectionNames []string) {
	for _, col := range collectionNames {
		err := MongoDB.CreateCollection(context.TODO(), col)
		if err != nil && !mongo.IsDuplicateKeyError(err) {
			log.Printf("⚠️ Failed to explicitly create collection %s: %v", col, err)
		} else {
			log.Printf("📁 Collection %s ensured", col)
		}
	}
}

// Optional: Reset/Drop collections (dev/test use only)
func resetCollections() {
	collections := []string{
		// CollectionAuthUsers,
		// CollectionSeekers,
		// CollectionAdmins,
		// CollectionSavedJobs,
		// CollectionUserEntryTimelines,
		// CollectionSelectedJobApps,
		// CollectionCoverLetters,
		// CollectionCV,
		// CollectionMatchScores,
		// CollectionJobs,
		// CollectionExtJobs,
		// CollectionJobResearch
		// CollectionCounter,
		// CollectionProfilePic,
		// CollectionNotifications,
		// CollectionPreferences,
		// CollectionQuestions,
		// CollectionResults,
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

// Index creation task
type IndexCreationTask struct {
	CollectionName    string
	CreateIndexesFunc func(collection *mongo.Collection) error
}

// Register all index tasks
func CreateAllIndexes() {
	tasks := []IndexCreationTask{
		// {CollectionAuthUsers, CreateAuthUserIndexes},
		// {CollectionSeekers, CreateSeekerIndexes},
		// {CollectionAdmins, CreateAdminIndexes},
		// {CollectionSavedJobs, CreateSavedJobApplicationIndexes},
		// {CollectionUserEntryTimelines, CreateUserEntryTimelineIndexes},
		// {CollectionSelectedJobApps, CreateSelectedJobApplicationIndexes},
		// {CollectionCoverLetters, CreateCoverLetterIndexes},
		// {CollectionCV, CreateCVIndexes},
		{CollectionMatchScores, CreateMatchScoreIndexes},
		{CollectionJobs, CreateJobIndexes},
		{CollectionJobResearch,CreateUserJobResearchIndexes},
		// {CollectionProfilePic,CreateProfilePicIndexes},
		// {CollectionNotifications, CreateUserNotificationsIndexes},
		// {CollectionPreferences, CreateUserPreferencesIndexes},
		// {CollectionQuestions,CreateQuestionIndexes},
		// {CollectionResults,CreateResultsIndexes},
	}

	for _, task := range tasks {
		collection := MongoDB.Collection(task.CollectionName)
		if err := task.CreateIndexesFunc(collection); err != nil {
			log.Fatalf("❌ Failed to create indexes for %s: %v", task.CollectionName, err)
		} else {
			log.Printf("✅ Indexes for %s created", task.CollectionName)
		}
	}
}
