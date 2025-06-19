package workers

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserEntryTimeline struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	AuthUserID string             `bson:"auth_user_id"`
	CreatedAt  time.Time          `bson:"created_at"`
}

type Seeker struct {
	ID                          primitive.ObjectID `bson:"_id,omitempty"`
	AuthUserID                  string             `bson:"auth_user_id"`
	SubscriptionTier            string             `bson:"subscription_tier"`
	DailySelectableJobsCount    int                `bson:"daily_selectable_jobs_count"`
	DailyGeneratableCV          int                `bson:"daily_generatable_cv"`
	DailyGeneratableCoverletter int                `bson:"daily_generatable_coverletter"`
}

// TierLimits defines limits for each subscription tier
var TierLimits = map[string]struct {
	SelectableJobsLimit         int
	GeneratableCVLimit          int
	GeneratableCoverletterLimit int
}{
	"free": {
		SelectableJobsLimit:         10,
		GeneratableCVLimit:          100,
		GeneratableCoverletterLimit: 100,
	},
	"basic": {
		SelectableJobsLimit:         0,
		GeneratableCVLimit:          0,
		GeneratableCoverletterLimit: 0,
	},
	"student": {
		SelectableJobsLimit:         0,
		GeneratableCVLimit:          0,
		GeneratableCoverletterLimit: 0,
	},
	"premium": {
		SelectableJobsLimit:         0,
		GeneratableCVLimit:          0,
		GeneratableCoverletterLimit: 0,
	},
	"advanced": {
		SelectableJobsLimit:         0,
		GeneratableCVLimit:          0,
		GeneratableCoverletterLimit: 0,
	},
}

func RunDailyWorker(ctx context.Context, seekersColl, timelinesColl *mongo.Collection) error {
	now := time.Now()

	cursor, err := timelinesColl.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to fetch timelines: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var timeline UserEntryTimeline
		if err := cursor.Decode(&timeline); err != nil {
			return fmt.Errorf("failed to decode timeline: %w", err)
		}

		var seeker Seeker
		err = seekersColl.FindOne(ctx, bson.M{"auth_user_id": timeline.AuthUserID}).Decode(&seeker)
		if err != nil {
			fmt.Printf("seeker not found for auth_user_id: %s\n", timeline.AuthUserID)
			continue
		}

		limits, ok := TierLimits[seeker.SubscriptionTier]
		if !ok || limits.SelectableJobsLimit == 0 {
			// Skip tiers not configured yet
			continue
		}

		updateFields := bson.M{}

		// Daily selectable jobs refill
		if seeker.DailySelectableJobsCount < limits.SelectableJobsLimit {
			updateFields["daily_selectable_jobs_count"] = limits.SelectableJobsLimit
		}

		// Weekly CV & Coverletter refill
		weeksSinceCreation := int(now.Sub(timeline.CreatedAt).Hours() / (24 * 7))
		lastWeeklyReset := timeline.CreatedAt.Add(time.Duration(weeksSinceCreation*7*24) * time.Hour)
		if now.After(lastWeeklyReset) && now.Sub(lastWeeklyReset) < 24*time.Hour {
			if seeker.DailyGeneratableCV < limits.GeneratableCVLimit {
				updateFields["daily_generatable_cv"] = limits.GeneratableCVLimit
				
			}
			if seeker.DailyGeneratableCoverletter < limits.GeneratableCoverletterLimit {
				updateFields["daily_generatable_coverletter"] = limits.GeneratableCoverletterLimit
	
			}
		}

		if len(updateFields) > 0 {
			_, err := seekersColl.UpdateOne(ctx,
				bson.M{"_id": seeker.ID},
				bson.M{"$set": updateFields},
			)
			if err != nil {
				fmt.Printf("failed to update seeker %s: %v\n", seeker.AuthUserID, err)
			}
		}
	}

	if err := cursor.Err(); err != nil {
		return fmt.Errorf("cursor error: %w", err)
	}

	return nil
}

// ✅ Run DailyWorker every 10 seconds for testing (switch back later)
func StartDailyWorker(db *mongo.Database) {
	seekersColl := db.Collection("seekers")
	timelinesColl := db.Collection("user_entry_timelines")

	// // 🟡 For testing: every 10 seconds
	// ticker := time.NewTicker(10 * time.Second)

	//✅ Production: every 24h (commented now)
	ticker := time.NewTicker(24 * time.Hour)

	defer ticker.Stop()

	// Run immediately at startup
	RunDailyWorkerOnce(seekersColl, timelinesColl)

	for range ticker.C {
	RunDailyWorkerOnce(seekersColl, timelinesColl)
	}
}


func RunDailyWorkerOnce(seekersColl, timelinesColl *mongo.Collection) {

	err := RunDailyWorker(context.Background(), seekersColl, timelinesColl)
	if err != nil {
		log.Printf("[DailyWorker] Error: %v", err)
	} else {
		// log.Println("[DailyWorker] Completed successfully")
	}
}
