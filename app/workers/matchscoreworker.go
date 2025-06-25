package workers

import (
    "context"
    "log"
    "strings"
    "sync"
    "time"

    "github.com/cenkalti/backoff/v4"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "RAAS/core/config"
    "RAAS/internal/models"
)

type Metrics struct {
    TotalCalculations int
    Errors            int
    Mutex             sync.Mutex
}

func (m *Metrics) IncCalc() {
    m.Mutex.Lock()
    m.TotalCalculations++
    m.Mutex.Unlock()
}
func (m *Metrics) IncError() {
    m.Mutex.Lock()
    m.Errors++
    m.Mutex.Unlock()
}

type MatchScoreWorker struct {
    Client  *mongo.Client
    Metrics *Metrics
}

func NewMatchScoreWorker(client *mongo.Client) *MatchScoreWorker {
    return &MatchScoreWorker{
        Client:  client,
        Metrics: &Metrics{},
    }
}

func (w *MatchScoreWorker) calculateAndStore(ctx context.Context, userID, jobID string) {
    op := func() error {
        return w.calculateAndStoreOnce(ctx, userID, jobID)
    }
    b := backoff.NewExponentialBackOff()
    b.MaxElapsedTime = 2 * time.Minute
    b.RandomizationFactor = 0.3

    err := backoff.Retry(op, backoff.WithContext(b, ctx))
    if err != nil {
        log.Printf("❌ Persistent failure for %s/%s: %v", userID, jobID, err)
        w.Metrics.IncError()
    }
}
func (w *MatchScoreWorker) calculateAndStoreOnce(ctx context.Context, userID, jobID string) error {
    db := w.Client.Database(config.Cfg.Cloud.MongoDBName)
    seekerColl := db.Collection("seekers")
    jobColl := db.Collection("jobs")
    scoreColl := db.Collection("match_scores")

    // Check if score already exists
    var existing bson.M
    err := scoreColl.FindOne(ctx, bson.M{"auth_user_id": userID, "job_id": jobID}).Decode(&existing)
    if err == nil {
        // Score already exists, skip calculation
        return nil
    }
    if err != mongo.ErrNoDocuments {
        // An actual error occurred
        return err
    }

    // Load seeker and job data
    var seeker models.Seeker
    if err := seekerColl.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
        return err
    }

    var job models.Job
    if err := jobColl.FindOne(ctx, bson.M{"job_id": jobID}).Decode(&job); err != nil {
        return err
    }

    // Calculate match score
    score, err := CalculateMatchScore(seeker, job)
    if err != nil {
        return err
    }

    // Store the score
    _, err = scoreColl.UpdateOne(ctx,
        bson.M{"auth_user_id": userID, "job_id": jobID},
        bson.M{"$set": bson.M{"match_score": score}},
        options.Update().SetUpsert(true),
    )
    if err == nil {
        w.Metrics.IncCalc()
    }
    return err
}


func (w *MatchScoreWorker) Run(ctx context.Context) {
    log.Println("✅ MatchScoreWorker started")
    dbName := config.Cfg.Cloud.MongoDBName

    ticker := time.NewTicker(5*time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            log.Println("🛑 Worker shutdown")
            return
        case <-ticker.C:
            w.processCycle(ctx, dbName)
        }
    }
}

func (w *MatchScoreWorker) processCycle(ctx context.Context, dbName string) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db := w.Client.Database(dbName)
	timelineColl := db.Collection("user_entry_timelines")
	authUsersColl := db.Collection("auth_users")
	seekersColl := db.Collection("seekers")
	jobsColl := db.Collection("jobs")
	scoresColl := db.Collection("match_scores")

	timelineCursor, err := timelineColl.Find(ctx, bson.M{"completed": true}, options.Find().SetBatchSize(200))
	if err != nil {
		log.Printf("❌ Error fetching timelines: %v", err)
		return
	}
	defer timelineCursor.Close(ctx)

	var wg sync.WaitGroup

	for timelineCursor.Next(ctx) {
		var timeline models.UserEntryTimeline
		if err := timelineCursor.Decode(&timeline); err != nil {
			continue
		}

		var authUser models.AuthUser
		err := authUsersColl.FindOne(ctx, bson.M{"auth_user_id": timeline.AuthUserID}).Decode(&authUser)
		if err != nil {
			log.Printf("⚠️ Could not find auth user for ID: %s", timeline.AuthUserID)
			continue
		}

		var seeker models.Seeker
		err = seekersColl.FindOne(ctx, bson.M{"auth_user_id": timeline.AuthUserID}).Decode(&seeker)
		if err != nil {
			log.Printf("⚠️ No seeker data for user: %s", timeline.AuthUserID)
			continue
		}

		titles := []string{seeker.PrimaryTitle}
		if seeker.SecondaryTitle != nil && *seeker.SecondaryTitle != "" {
			titles = append(titles, *seeker.SecondaryTitle)
		}
		if seeker.TertiaryTitle != nil && *seeker.TertiaryTitle != "" {
			titles = append(titles, *seeker.TertiaryTitle)
		}

		log.Printf("👤 User: %s | Titles: %v", authUser.Email, titles)

		// Build OR query for job title match
		var orClauses bson.A
		for _, t := range titles {
			orClauses = append(orClauses, bson.M{
				"title": bson.M{"$regex": "(?i)" + strings.TrimSpace(t)},
			})
		}

		jobCursor, err := jobsColl.Find(ctx, bson.M{"$or": orClauses}, options.Find().SetBatchSize(100))
		if err != nil {
			log.Printf("⚠️ Error fetching jobs for user %s: %v", authUser.Email, err)
			continue
		}

		for jobCursor.Next(ctx) {
			var job models.Job
			if err := jobCursor.Decode(&job); err != nil {
				continue
			}

			log.Printf("   📌 Related Job: %s | Company: %s | Link: %s", job.Title, job.Company, job.JobLink)

			// Check if match score already exists
			existsErr := scoresColl.FindOne(ctx,
				bson.M{"auth_user_id": seeker.AuthUserID, "job_id": job.JobID},
				options.FindOne().SetProjection(bson.M{"_id": 1}),
			).Err()
			if existsErr == nil {
				continue // score exists
			}

			wg.Add(1)
			go func(uid, jid string) {
				defer wg.Done()
				w.calculateAndStore(ctx, uid, jid)
			}(seeker.AuthUserID, job.JobID)
		}
		jobCursor.Close(ctx)
	}

	wg.Wait()
	log.Printf("✅ Completed cycle: Total Calculations = %d, Errors = %d", w.Metrics.TotalCalculations, w.Metrics.Errors)
}



