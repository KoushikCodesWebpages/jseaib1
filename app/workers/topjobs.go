// workers/tasks.go

package workers

// import (
//     "bytes"
//     "context"
//     "html/template"
//     "log"
//     "time"

//     "github.com/go-co-op/gocron"
//     "go.mongodb.org/mongo-driver/bson"
//     "go.mongodb.org/mongo-driver/mongo"
//     "go.mongodb.org/mongo-driver/mongo/options"

//     "RAAS/internal/models"
//     "RAAS/utils"
// )

// // StartMatchNotifier runs worker every 2 days
// func StartMatchNotifier(db *mongo.Database) *gocron.Scheduler {
//     s := gocron.NewScheduler(time.UTC)
//     s.Every(2).Days().Do(func() { notifyUsers(db) })
//     s.StartAsync()
//     log.Println("[MatchNotifier] started: every 2 days")
//     return s
// }

// func notifyUsers(db *mongo.Database) {
//     ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
//     defer cancel()

//     cur, err := db.Collection("auth_users").Find(ctx, bson.M{"email": bson.M{"$ne": ""}})
//     if err != nil {
//         log.Println("[MatchNotifier] failed to list users:", err)
//         return
//     }
//     defer cur.Close(ctx)

//     for cur.Next(ctx) {
//         var u struct {
//             AuthUserID string `bson:"auth_user_id"`
//             Email      string `bson:"email"`
//         }
//         if err := cur.Decode(&u); err != nil {
//             log.Println("[MatchNotifier] decode user:", err)
//             continue
//         }

//         matches := fetchTopMatches(ctx, db, u.AuthUserID, 3)
//         if len(matches) == 0 {
//             continue
//         }

//         jobs := fetchJobsByIDs(ctx, db, matches)
//         if len(jobs) == 0 {
//             continue
//         }

//         if err := sendJobMatchesEmail(u.Email, jobs); err != nil {
//             log.Printf("[MatchNotifier] send to %s failed: %v", u.Email, err)
//         }
//     }
// }

// func fetchTopMatches(ctx context.Context, db *mongo.Database, userID string, limit int) []models.MatchScore {
//     var results []models.MatchScore
//     cursor, err := db.Collection("match_scores").Find(ctx,
//         bson.M{"auth_user_id": userID},
//         options.Find().SetSort(bson.D{{"match_score", -1}}).SetLimit(int64(limit)),
//     )
//     if err != nil {
//         log.Println("[MatchNotifier] fetch matches error:", err)
//         return nil
//     }
//     defer cursor.Close(ctx)
//     cursor.All(ctx, &results)
//     return results
// }

// // fetchJobsByIDs joins match entries to actual job objects
// func fetchJobsByIDs(ctx context.Context, db *mongo.Database, matches []models.MatchScore) []models.Job {
//     ids := make([]string, len(matches))
//     for i, m := range matches {
//         ids[i] = m.JobID
//     }

//     cursor, err := db.Collection("jobs").Find(ctx, bson.M{"job_id": bson.M{"$in": ids}})
//     if err != nil {
//         log.Println("[MatchNotifier] fetch jobs error:", err)
//         return nil
//     }
//     defer cursor.Close(ctx)

//     var jobs []models.Job
//     cursor.All(ctx, &jobs)
//     return jobs
// }

// func sendJobMatchesEmail(to string, jobs []models.Job) error {
//     cfg := utils.GetEmailConfig()

//     const tmplStr = `
//     <h2>Top {{len .}} Job Matches</h2>
//     <ul>
//     {{range .}}
//       <li style="margin-bottom:12px;">
//         <strong>{{.JobTitle}}</strong> at <em>{{.Company}}</em><br/>
//         📍 {{.Location}} | 🛠 Skills: {{.Skills}} | 📌 Type: {{.JobType}}<br/>
//         <a href="{{.JobLink}}">View Job</a>
//       </li>
//     {{end}}
//     </ul>
//     `

//     t := template.Must(template.New("jobsEmail").Parse(tmplStr))
//     var buf bytes.Buffer
//     if err := t.Execute(&buf, jobs); err != nil {
//         return err
//     }

//     return utils.SendEmail(cfg, to, "Your Top Job Matches 🚀", buf.String())
// }



// workers/tasks.go
import (
    "bytes"
    "context"
    "html/template"
    "log"
    "time"

    "github.com/go-co-op/gocron"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"

    "RAAS/utils"
)

// StartTestNotifier schedules test emails every 2 days
func StartTestNotifier(db *mongo.Database) *gocron.Scheduler {
    s := gocron.NewScheduler(time.UTC)
    s.Every(2).Days().Do(func() { notifyUsersTest(db) })
    s.StartAsync()
    log.Println("[TestNotifier] started: test email every 2 days")
    return s
}

// // StartTestNotifier schedules test emails every 10 seconds
// func StartTestNotifier(db *mongo.Database) *gocron.Scheduler {
//     s := gocron.NewScheduler(time.UTC)
//     // Schedule every 10 seconds for testing
//     _, err := s.Every(10).Seconds().Do(func() { notifyUsersTest(db) })
//     if err != nil {
//         log.Fatalf("[TestNotifier] failed to schedule: %v", err)
//     }
//     s.StartAsync()
//     log.Println("[TestNotifier] started: sending test email every 10 seconds")
//     return s
// }
func notifyUsersTest(db *mongo.Database) {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
    defer cancel()

    cur, err := db.Collection("auth_users").Find(ctx, bson.M{"email": bson.M{"$ne": ""}})
    if err != nil {
        log.Println("[TestNotifier] failed to list users:", err)
        return
    }
    defer cur.Close(ctx)

    for cur.Next(ctx) {
        var u struct {
            Email string `bson:"email"`
        }
        if err := cur.Decode(&u); err != nil {
            log.Println("[TestNotifier] decode user:", err)
            continue
        }

        // 📧 Send a simple test email
        if err := sendTestEmail(u.Email); err != nil {
            log.Printf("[TestNotifier] sending to %s failed: %v", u.Email, err)
        }
    }
}

func sendTestEmail(to string) error {
    cfg := utils.GetEmailConfig()

    const tmplStr = `
    <h2>This is a test email</h2>
    <p>Just verifying your email configuration is working correctly.</p>
    `

    t := template.Must(template.New("testEmail").Parse(tmplStr))
    var buf bytes.Buffer
    if err := t.Execute(&buf, nil); err != nil {
        return err
    }

    return utils.SendEmail(cfg, to, "🔧 Test Email from Our Service", buf.String())
}
