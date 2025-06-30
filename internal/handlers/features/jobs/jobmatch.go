package jobs

import (
	"RAAS/internal/models"
    "RAAS/internal/handlers/repository"
	"net/http"
    "fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson/primitive"

    "math"
	"strings"

	"github.com/agnivade/levenshtein"
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


func CalculateMatchScore(seeker models.Seeker, job models.Job) (float64, error) {
	experienceSummaryObjs, _ := repository.GetWorkExperience(&seeker)
	certificateObjs, _ := repository.GetCertificates(&seeker)
	languageObjs, _ := repository.GetLanguages(&seeker)
	educationObjs, _ := repository.GetAcademics(&seeker)
	projectObjs, _ := repository.GetPastProjects(&seeker)

	var tokens []string

	// fmt.Println("▶ Key Skills:")
	for _, skill := range seeker.KeySkills {
		// fmt.Println(" -", skill)
		tokens = append(tokens, strings.ToLower(skill))
	}

	// fmt.Println("▶ Work Experience:")
	for _, e := range experienceSummaryObjs {
		fields := extractFields(e, "job_title", "company_name", "key_responsibilities")
		// fmt.Println(" -", fields)
		tokens = append(tokens, fields...)
	}

	// fmt.Println("▶ Education:")
	for _, e := range educationObjs {
		fields := extractFields(e, "degree", "field_of_study", "institution")
		// fmt.Println(" -", fields)
		tokens = append(tokens, fields...)
	}

	// fmt.Println("▶ Certificates:")
	for _, cert := range certificateObjs {
		fields := extractFields(cert, "certificate_name", "provider")
		// fmt.Println(" -", fields)
		tokens = append(tokens, fields...)
	}

	// fmt.Println("▶ Languages:")
	for _, lang := range languageObjs {
		fields := extractFields(lang, "language", "proficiency")
		// fmt.Println(" -", fields)
		tokens = append(tokens, fields...)
	}

	// fmt.Println("▶ Past Projects:")
	for _, p := range projectObjs {
		fields := extractFields(p, "project_name", "project_description", "institution")
		// fmt.Println(" -", fields)
		tokens = append(tokens, fields...)
	}

	seekerText := normalizeText(strings.Join(tokens, " "))
	jobText := normalizeText(job.Title + " " + job.JobDescription + " " + job.Skills)

	// fmt.Println("▶ Normalized Seeker Text:\n", seekerText)
	// fmt.Println("▶ Normalized Job Text:\n", jobText)

	cosineScore := cosineSimilarity(seekerText, jobText)
	// fmt.Printf("✅ Cosine Similarity: %.4f\n", cosineScore)

	levScore := levenshteinSimilarity(seekerText, jobText)
	// fmt.Printf("✅ Levenshtein Similarity: %.4f\n", levScore)

	finalScore := (0.6 * cosineScore) + (0.4 * levScore)
    // Map it to 50–100 range
    scaledScore := 55 + (finalScore * 45) // 0.0 → 50, 1.0 → 100

    // Round to 2 decimal places
    finalScoreRounded := math.Round(scaledScore*100) / 100
	// fmt.Printf("🎯 Final Match Score: %.2f\n", finalScoreRounded)

	return finalScoreRounded, nil
}


// Extracts fields from bson.M and lowercases them
func extractFields(m map[string]interface{}, keys ...string) []string {
	var out []string
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch val := v.(type) {
			case string:
				out = append(out, strings.ToLower(val))
			case primitive.DateTime:
				if !val.Time().IsZero() {
					out = append(out, val.Time().Format("Jan 2006"))
				}
			}
		}
	}
	return out
}

// Normalizes text for similarity comparisons
func normalizeText(text string) string {
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.Join(strings.Fields(text), " ") // remove extra spaces
	return text
}

// Basic cosine similarity on word sets
func cosineSimilarity(a, b string) float64 {
	vecA := toFreqMap(strings.Fields(a))
	vecB := toFreqMap(strings.Fields(b))

	var dotProduct, normA, normB float64
	for word, freqA := range vecA {
		freqB := vecB[word]
		dotProduct += float64(freqA * freqB)
	}
	for _, freq := range vecA {
		normA += float64(freq * freq)
	}
	for _, freq := range vecB {
		normB += float64(freq * freq)
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Converts word slice to frequency map
func toFreqMap(words []string) map[string]int {
	freqMap := make(map[string]int)
	for _, word := range words {
		freqMap[word]++
	}
	return freqMap
}

// Normalized Levenshtein score
func levenshteinSimilarity(a, b string) float64 {
	distance := levenshtein.ComputeDistance(a, b)
	maxLen := math.Max(float64(len(a)), float64(len(b)))
	if maxLen == 0 {
		return 1
	}
	return 1 - float64(distance)/maxLen
}






// // CalculateMatchScore computes a weighted match score based on field-by-field cosine similarity.
// func CalculateMatchScore(seeker models.Seeker, job models.Job) (float64, error) {
// 	// --- Get seeker components from raw bson.M slices ---
// 	experienceObjs, _ := repository.GetWorkExperience(&seeker)
// 	certificateObjs, _ := repository.GetCertificates(&seeker)
// 	languageObjs, _ := repository.GetLanguages(&seeker)
// 	educationObjs, _ := repository.GetAcademics(&seeker)
// 	projectObjs, _ := repository.GetPastProjects(&seeker)

// 	// --- Normalize Job Fields ---
// 	jobText := normalizeText(job.Title + " " + job.JobDescription + " " + job.Skills)

// 	// --- Flatten Seeker Fields ---
// 	skillsText := normalizeText(strings.Join(seeker.KeySkills, " "))
// 	experienceText := flattenAndNormalize(experienceObjs)
// 	educationText := flattenAndNormalize(educationObjs)
// 	certificatesText := flattenAndNormalize(certificateObjs)
// 	languagesText := flattenAndNormalize(languageObjs)
// 	projectsText := flattenAndNormalize(projectObjs)

// 	// --- Cosine Similarities (0–1 per field) ---
// 	skillsScore := cosineSimilarity(skillsText, jobText)
// 	experienceScore := cosineSimilarity(experienceText, jobText)
// 	educationScore := cosineSimilarity(educationText, jobText)
// 	certificateScore := cosineSimilarity(certificatesText, jobText)
// 	languageScore := cosineSimilarity(languagesText, jobText)
// 	projectScore := cosineSimilarity(projectsText, jobText)

// 	// --- Weights (adjust based on importance) ---
// 	weights := map[string]float64{
// 		"skills":       0.25,
// 		"experience":   0.25,
// 		"education":    0.20,
// 		"certificates": 0.10,
// 		"languages":    0.10,
// 		"projects":     0.10,
// 	}

// 	// --- Final weighted score ---
// 	finalScore := (skillsScore * weights["skills"]) +
// 		(experienceScore * weights["experience"]) +
// 		(educationScore * weights["education"]) +
// 		(certificateScore * weights["certificates"]) +
// 		(languageScore * weights["languages"]) +
// 		(projectScore * weights["projects"])

// 	// --- Debug print ---
// 	fmt.Println("---- Field-wise Similarity Scores ----")
// 	fmt.Printf("✅ Skills:        %.4f\n", skillsScore)
// 	fmt.Printf("✅ Experience:    %.4f\n", experienceScore)
// 	fmt.Printf("✅ Education:     %.4f\n", educationScore)
// 	fmt.Printf("✅ Certificates:  %.4f\n", certificateScore)
// 	fmt.Printf("✅ Languages:     %.4f\n", languageScore)
// 	fmt.Printf("✅ Projects:      %.4f\n", projectScore)
// 	fmt.Printf("🎯 Final Match Score: %.2f\n", finalScore*100)

// 	return math.Round(finalScore*10000) / 100, nil // Return 2-decimal percentage score
// }

// // Converts bson.M slice to flat string and normalizes
// func flattenAndNormalize(data []bson.M) string {
// 	var result []string
// 	for _, item := range data {
// 		for _, val := range item {
// 			if str, ok := val.(string); ok {
// 				result = append(result, str)
// 			}
// 		}
// 	}
// 	return normalizeText(strings.Join(result, " "))
// }

// // Normalize and clean string
// func normalizeText(text string) string {
// 	text = strings.ToLower(text)
// 	text = strings.ReplaceAll(text, "\n", " ")
// 	text = strings.ReplaceAll(text, "\t", " ")
// 	text = strings.ReplaceAll(text, ",", " ")
// 	text = strings.ReplaceAll(text, ".", " ")
// 	return strings.Join(strings.Fields(text), " ") // remove extra spaces
// }

// // Converts string slice to frequency map
// func toFreqMap(words []string) map[string]int {
// 	freqMap := make(map[string]int)
// 	for _, word := range words {
// 		freqMap[word]++
// 	}
// 	return freqMap
// }

// // Computes cosine similarity between two text blobs
// func cosineSimilarity(a, b string) float64 {
// 	vecA := toFreqMap(strings.Fields(a))
// 	vecB := toFreqMap(strings.Fields(b))

// 	var dotProduct, normA, normB float64
// 	for word, freqA := range vecA {
// 		dotProduct += float64(freqA * vecB[word])
// 	}
// 	for _, freq := range vecA {
// 		normA += float64(freq * freq)
// 	}
// 	for _, freq := range vecB {
// 		normB += float64(freq * freq)
// 	}
// 	if normA == 0 || normB == 0 {
// 		return 0
// 	}
// 	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
// }
