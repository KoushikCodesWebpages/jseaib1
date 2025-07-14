package jobs

import (
	"RAAS/internal/handlers/repository"
	"RAAS/internal/models"
	"fmt"
	// "log"
	"net/http"
    "math/rand"
    "math"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	// "math"
	"strings"
	// "github.com/ugurkorkmaz/multiversal/cosine_similarity"
	// "github.com/texttheater/golang-levenshtein/levenshtein"
	// "RAAS/internal/models"
	//"github.com/google/uuid"
)

// MatchScoreHandler handles match score retrieval
type MatchScoreHandler struct{}

func NewMatchScoreHandler() *MatchScoreHandler {
    return &MatchScoreHandler{}
}

// GET /b1/match-scores?job_id=...
func (h *MatchScoreHandler) GetMatchScores(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    coll := db.Collection("match_scores")
    userID := c.MustGet("userID").(string)

    filter := bson.M{"auth_user_id": userID}
    if jobID := c.Query("job_id"); jobID != "" {
        filter["job_id"] = jobID
    }

    cursor, err := coll.Find(c, filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer cursor.Close(c)

    var results []models.MatchScore
    if err := cursor.All(c, &results); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing results"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": results})
}

func StartJobMatchScoreCalculation(c *gin.Context, db *mongo.Database, userID string) error {
	fmt.Println("Starting job match scoring for user:", userID)

	// 1. Fetch seeker data
	seeker, err := repository.GetSeekerData(db, userID)
	if err != nil {
		return fmt.Errorf("failed to fetch seeker data: %v", err)
	}

	// 2. Collect preferred titles
	preferredTitles := repository.CollectPreferredTitles(seeker)
	if len(preferredTitles) == 0 {
		return fmt.Errorf("no preferred job titles found for seeker")
	}

	// 3. Build job filter
	filter := repository.BuildJobFilter(preferredTitles, nil, "") // nil = no applied filter, "" = all languages

	// 4. Query all matching jobs
	cursor, err := db.Collection("jobs").Find(c, filter)
	if err != nil {
		return fmt.Errorf("failed to query jobs: %v", err)
	}
	defer cursor.Close(c)

	// 5. Iterate through jobs
	matchCollection := db.Collection("match_scores")
	for cursor.Next(c) {
		var job models.Job
		if err := cursor.Decode(&job); err != nil {
			fmt.Println("error decoding job:", err)
			continue
		}

		// 6. Check if match score already exists
		exists, _ := matchCollection.CountDocuments(c, bson.M{
			"auth_user_id": userID,
			"job_id":       job.JobID,
		})
		if exists > 0 {
			continue // skip if already calculated
		}

		// 7. Calculate match score
		score, err := CalculateMatchScore(seeker, job)
		if err != nil {
			fmt.Println("error calculating score for job:", job.JobID, err)
			continue
		}

		// 8. Insert new match score
		_, err = matchCollection.InsertOne(c, models.MatchScore{
			AuthUserID: userID,
			JobID:      job.JobID,
			MatchScore: score,
		})
		if err != nil {
			fmt.Println("error inserting match score:", err)
			continue
		}
	}

	return nil
}

// Configuration: section weights sum to 1.0
var (
    skillsWeight = 1.0
    // certsWeight  = 0.3
    // langsWeight  = 0.2
)

// CalculateMatchScore returns a match score (0.0–1.0)
// using keyword-based matching per section.
func CalculateMatchScore(seeker models.Seeker, job models.Job) (float64, error) {
    // 1. Token extraction (unchanged)
    // certificateObjs, _ := repository.GetCertificates(&seeker)
    // languageObjs, _ := repository.GetLanguages(&seeker)

    var skillsTokens []string
	// var certTokens []string
	// var langTokens []string

    // 1️⃣ Key skills
    for _, skill := range seeker.KeySkills {
        skillsTokens = append(skillsTokens, strings.ToLower(skill))
    }

    // 2️⃣ Extract cert titles
    // for _, cert := range certificateObjs {
    //     name, _ := cert["certificate_name"].(string)
    //     if name != "" {
    //         certTokens  = append(certTokens , strings.ToLower(name))
    //     }
    // }

    // // 3️⃣ Format language proficiency
    // for _, lang := range languageObjs {
    //     langName, _ := lang["language"].(string)
    //     proficiency, _ := lang["proficiency"].(string)
    //     if langName != "" {
    //         formatted := fmt.Sprintf("%s: %s", langName, proficiency)
    //         langTokens = append(langTokens, strings.ToLower(formatted))
    //     }
    // }

    // 2. Prepare job text once
    jobText := strings.ToLower(job.Title + " " + job.JobDescription + " " + job.Skills)
	fmt.Println("🔍 Skills Tokens:", skillsTokens)
    // fmt.Println("🔍 Certificate Tokens:", certTokens)
    // fmt.Println("🔍 Language Tokens:", langTokens)
    fmt.Println("🔍 Job Text snippet:", jobText[:min(len(jobText), 20)])

    // 3. Compute per-section scores
    skillScore := keywordMatch(skillsTokens, jobText)
    // certScore := keywordMatch(certTokens, jobText)
    // langScore := keywordMatch(langTokens, jobText)
	// fmt.Printf("skill: %.2f, cert: %.2f, lang: %.2f\n", skillScore, certScore, langScore)
    // 4. Weighted aggregation
    raw := skillScore*skillsWeight 
    scaled := 60 + raw*40
    final := math.Round(scaled*100) / 100

    // Apply directional deviation
    deviation := (rand.Float64()*1 + 1) / 100 * final // ±1–2%
    if final > 60 {
        final = math.Round((final) * 100) / 100
        final = final - deviation
    } else {
        final = math.Round((final) * 100) / 100
        final = final + deviation
    }

    return final, nil
}

// Helper: lowercase a slice
func toLower(input []string) []string {
    out := make([]string, len(input))
    for i, s := range input {
        out[i] = strings.ToLower(s)
    }
    return out
}

// keywordMatch computes how many unique tokens appear in jobText.
// Returns match rate in [0,1].
func keywordMatch(tokens []string, jobText string) float64 {
    if len(tokens) == 0 {
        return 0.0
    }
    seen := make(map[string]struct{})
    matches := 0

    for _, t := range tokens {
        t = strings.ToLower(strings.TrimSpace(t))
        if t == "" {
            continue
        }
        if _, ok := seen[t]; ok {
            continue
        }
        seen[t] = struct{}{}

        if strings.Contains(jobText, t) {
            matches++
        }
    }
    return float64(matches) / float64(len(seen))
}


// Stub: your existing implementation
// func extractFields(obj interface{}, fields ...string) []string {
//     // your existing logic here
//     return nil
// }
