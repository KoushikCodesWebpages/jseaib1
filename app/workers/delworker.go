// workers/tasks.go

package workers

import (
    "context"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

var purgeCollections = []string{
    "seekers", "user_entry_timelines", "cover_letters", "cv",
    "selected_job_applications", "admins", "match_scores",
    "auth_users", "saved_jobs", "preferences", "notifications",
}

// PurgeOlddeletedUsers finds and purges users deleted over 30 days ago.
func PurgeOldDeletedUsers(ctx context.Context, db *mongo.Database) error {








    cutoff := time.Now().Add(-30 * 24 * time.Hour)
    // cutoff := time.Now().Add( 5 * time.Second)






    



    
    cursor, err := db.Collection("auth_users").
        Find(ctx, bson.M{"is_deleted": true, "deleted_at": bson.M{"$lte": cutoff}})
    if err != nil {
        return err
    }
    defer cursor.Close(ctx)

    var u struct{ AuthUserID string `bson:"auth_user_id"` }
    for cursor.Next(ctx) {
        if err := cursor.Decode(&u); err != nil {
            log.Printf("[Purge] decode error: %v", err)
            continue
        }
        purgeAllUserData(ctx, db, u.AuthUserID)
    }
    return cursor.Err()
}

// Deletes data for a user across related collections.
func purgeAllUserData(ctx context.Context, db *mongo.Database, userID string) {
    for _, coll := range purgeCollections {
        _, err := db.Collection(coll).DeleteMany(ctx, bson.M{"auth_user_id": userID})
        if err != nil {
            log.Printf("[Purge] error purging %s for %s: %v", coll, userID, err)
        } else {
            // log.Printf("[Purge] purged %d docs from %s for %s", res.DeletedCount, coll, userID)
        }
    }
}

// RunDailyPurgeTasks wraps your purge logic.
func RunDailyPurgeTasks(ctx context.Context, db *mongo.Database) {
	log.Println("[PurgeWorker] Starting purge task...")
	if err := PurgeOldDeletedUsers(ctx, db); err != nil {
		log.Printf("[PurgeWorker] purge error: %v", err)
	}
}

// Scheduler loop with cancellation support.
func workerLoop(ctx context.Context, interval time.Duration, db *mongo.Database) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    RunDailyPurgeTasks(ctx, db)
    for {
        select {
        case <-ctx.Done():
            log.Println("[PurgeWorker] stopped")
            return
        case <-ticker.C:
            RunDailyPurgeTasks(ctx, db)
        }
    }
}

// StartPurgeWorker starts the task scheduler.
func StartPurgeWorker(db *mongo.Database, interval time.Duration) context.CancelFunc {
    ctx, cancel := context.WithCancel(context.Background())
    go workerLoop(ctx, interval, db)
    return cancel
}
