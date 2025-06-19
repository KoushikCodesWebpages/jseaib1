package workers

// import (
	
// 	"encoding/json"
// 	"fmt"
// 	// "io"
// 	// "bytes"
// 	// "net/http"
// 	"log"
// 	"github.com/joho/godotenv"
// 	"RAAS/models"
// 	//"log"
// 	"math"
// 	"errors"
// 	"strings"
// 	"github.com/spf13/viper"
// )

// // RoundRobinModelIndex keeps track of the current model index for round-robin
// var (
// 	RoundRobinModelIndex int = 0
// )

// // LoadHFModels loads the Hugging Face models from Viper configuration
// func LoadHFModels(prefix string) ([]string, error) {
// 	var models []string
// 	for i := 1; i <= 10; i++ {
// 		model := viper.GetString(fmt.Sprintf("%s_%d", prefix, i))
// 		if model == "" {
// 			return nil, fmt.Errorf("model %s_%d is not defined", prefix, i)
// 		}
// 		models = append(models, model)
// 	}
// 	//log.Printf("Loaded Hugging Face models with prefix %s: %+v", prefix, models)  // Debugging log
// 	return models, nil
// }

// // CalculateMatchScore calculates the match score between the seeker and the job
// func CalculateMatchScore(seeker models.Seeker, job models.Job) (float64, error) {
// 	log.Println("📥 Starting match score calculation...")

// 	if err := godotenv.Load(); err != nil {
// 		log.Printf("⚠️ Error loading .env file: %v", err)
// 		// Not returning here—.env might already be loaded
// 	}

// 	// Load API key
// 	hfAPIKey := viper.GetString("HF_API_KEY")
// 	if hfAPIKey == "" {
// 		log.Println("❌ Hugging Face API key not found in config")
// 		return 0, fmt.Errorf("hugging Face API key not found")
// 	}
// 	log.Println("✅ Hugging Face API key loaded.")

// 	// Load models
// 	modelsList, err := LoadHFModels("HF_MODEL_FOR_MS")
// 	if err != nil {
// 		log.Printf("❌ Failed to load models: %v", err)
// 		return 0, fmt.Errorf("error loading models: %v", err)
// 	}
// 	currentModel := modelsList[RoundRobinModelIndex]
// 	RoundRobinModelIndex = (RoundRobinModelIndex + 1) % len(modelsList)
// 	log.Printf("🔁 Using model: %s", currentModel)

// 	// === Professional Summary ===
// 	var summary struct {
// 		About        string   `json:"about"`
// 		Skills       []string `json:"skills"`
// 		AnnualIncome float64  `json:"annualIncome"`
// 	}
// 	if err := json.Unmarshal(seeker.ProfessionalSummary, &summary); err != nil {
// 		log.Printf("❌ Failed to parse professional summary: %v", err)
// 		return 0, fmt.Errorf("failed to parse professional summary: %v", err)
// 	}
// 	skillsStr := joinStrings(summary.Skills, ", ")
// 	log.Printf("✅ Parsed skills: %s", skillsStr)

// 	// === Work Experience ===
// 	var workExperiences []map[string]interface{}
// 	if err := json.Unmarshal(seeker.WorkExperiences, &workExperiences); err != nil {
// 		log.Printf("❌ Failed to parse work experiences: %v", err)
// 		return 0, fmt.Errorf("failed to parse work experiences: %v", err)
// 	}
// 	log.Printf("✅ Parsed %d work experience entries", len(workExperiences))

// 	var workExpText string
// 	for _, we := range workExperiences {
// 		workExpText += fmt.Sprintf("Job Title: %s. Responsibilities: %s. ",
// 			getStr(we["jobTitle"]), getStr(we["keyResponsibilities"]))
// 	}

// 	// === Education ===
// 	var educations []map[string]interface{}
// 	if err := json.Unmarshal(seeker.Educations, &educations); err != nil {
// 		log.Printf("❌ Failed to parse education: %v", err)
// 		return 0, fmt.Errorf("failed to parse education: %v", err)
// 	}
// 	log.Printf("✅ Parsed %d education entries", len(educations))

// 	var eduText string
// 	for _, edu := range educations {
// 		eduText += fmt.Sprintf("Degree: %s in %s. Achievements: %s. ",
// 			getStr(edu["degree"]), getStr(edu["fieldOfStudy"]), getStr(edu["achievements"]))
// 	}

// 	// === Certificates ===
// 	var certificates []map[string]interface{}
// 	if err := json.Unmarshal(seeker.Certificates, &certificates); err != nil {
// 		log.Printf("❌ Failed to parse certificates: %v", err)
// 		return 0, fmt.Errorf("failed to parse certificates: %v", err)
// 	}
// 	log.Printf("✅ Parsed %d certificates", len(certificates))

// 	var certText string
// 	for _, cert := range certificates {
// 		certText += fmt.Sprintf("Certificate: %s. ", getStr(cert["certificateName"]))
// 	}

// 	// === Languages ===
// 	var languages []map[string]interface{}
// 	if err := json.Unmarshal(seeker.Languages, &languages); err != nil {
// 		log.Printf("❌ Failed to parse languages: %v", err)
// 		return 0, fmt.Errorf("failed to parse languages: %v", err)
// 	}
// 	log.Printf("✅ Parsed %d languages", len(languages))

// 	var langText string
// 	for _, lang := range languages {
// 		langText += fmt.Sprintf("Language: %s (%s). ",
// 			getStr(lang["language"]), getStr(lang["proficiency"]))
// 	}

// 	// === Preferred Titles ===
// 	titles := []string{}
// 	if seeker.PrimaryTitle != "" {
// 		titles = append(titles, seeker.PrimaryTitle)
// 	}
// 	if seeker.SecondaryTitle != nil && *seeker.SecondaryTitle != "" {
// 		titles = append(titles, *seeker.SecondaryTitle)
// 	}
// 	if seeker.TertiaryTitle != nil && *seeker.TertiaryTitle != "" {
// 		titles = append(titles, *seeker.TertiaryTitle)
// 	}
// 	jobTitles := "Preferred Job Titles: " + strings.Join(titles, ", ") + "."

// 	seekerText := fmt.Sprintf(
// 		"Skills: %s. About: %s. Work Experience: %s Education: %s Certificates: %s Languages: %s %s",
// 		skillsStr, summary.About, workExpText, eduText, certText, langText, jobTitles,
// 	)

// 	jobText := fmt.Sprintf("Title: %s. Description: %s. Skills: %s. Type: %s.",
// 		job.Title, job.JobDescription, job.Skills, job.JobType)

// 	log.Println("🧠 Computing cosine similarity...")
// 	matchScore, err := CosineSimilarity(seekerText, jobText)
// 	if err != nil {
// 		log.Printf("❌ Error from CosineSimilarity: %v", err)
// 		return 0, fmt.Errorf("error getting match score from model %s: %v", currentModel, err)
// 	}
// 	log.Printf("✅ Match score computed: %.2f", matchScore)

// 	return matchScore, nil
// }



// func getStr(val interface{}) string {
// 	if str, ok := val.(string); ok {
// 		return str
// 	}
// 	return ""
// }





















// // Apply sigmoid function for scaling
// func sigmoid(x float64) float64 {
// 	return 100 / (1 + math.Exp(-x)) // Converts into a range of 0 to 100
// }

// // CosineSimilarity calculates the cosine similarity between two text strings and applies sigmoid scaling
// func CosineSimilarity(text1, text2 string) (float64, error) {
// 	log.Println("📐 Starting CosineSimilarity calculation...")

// 	tokens1 := tokenize(text1)
// 	tokens2 := tokenize(text2)

// 	log.Printf("🔤 Tokens1: %d tokens | Tokens2: %d tokens", len(tokens1), len(tokens2))

// 	if len(tokens1) == 0 || len(tokens2) == 0 {
// 		return 0, errors.New("one of the input texts is empty after tokenization")
// 	}

// 	tf1 := termFrequency(tokens1)
// 	tf2 := termFrequency(tokens2)

// 	dotProduct := 0.0
// 	magnitude1 := 0.0
// 	magnitude2 := 0.0

// 	for word, freq1 := range tf1 {
// 		freq2 := tf2[word]
// 		dotProduct += freq1 * freq2
// 		magnitude1 += freq1 * freq1
// 	}
// 	for _, freq2 := range tf2 {
// 		magnitude2 += freq2 * freq2
// 	}

// 	if magnitude1 == 0 || magnitude2 == 0 {
// 		log.Println("❌ One of the vectors has zero magnitude")
// 		return 0, errors.New("one of the vectors has zero magnitude")
// 	}

// 	cosineSim := dotProduct / (math.Sqrt(magnitude1) * math.Sqrt(magnitude2))
// 	log.Printf("📏 Cosine similarity (raw): %.4f", cosineSim)

// 	matchScore := sigmoid(cosineSim * 10) // Amplify
// 	log.Printf("🎯 Match score (scaled with sigmoid): %.2f", matchScore)

// 	return matchScore, nil
// }


// // tokenize splits a string into words (tokens)
// func tokenize(text string) []string {
// 	// Convert the text to lowercase and split by non-alphanumeric characters
// 	// This basic tokenizer can be improved by using a more sophisticated library.
// 	text = strings.ToLower(text)
// 	words := strings.Fields(text)
// 	return words
// }

// // termFrequency calculates the term frequency of each word in a list of tokens
// func termFrequency(tokens []string) map[string]float64 {
// 	tf := make(map[string]float64)
// 	for _, token := range tokens {
// 		tf[token]++
// 	}
// 	// Normalize by the total number of words
// 	for word := range tf {
// 		tf[word] /= float64(len(tokens))
// 	}
// 	return tf
// }



// // func CosineSimilarity(text1, text2 string) (float64, error) {
// // 	// Tokenize the texts by splitting into words
// // 	tokens1 := tokenize(text1)
// // 	tokens2 := tokenize(text2)

// // 	// If either of the texts results in no tokens, return an error
// // 	if len(tokens1) == 0 || len(tokens2) == 0 {
// // 		return 0, errors.New("one of the input texts is empty after tokenization")
// // 	}

// // 	// Calculate term frequencies (TF)
// // 	tf1 := termFrequency(tokens1)
// // 	tf2 := termFrequency(tokens2)

// // 	// Calculate dot product and magnitudes
// // 	dotProduct := 0.0
// // 	magnitude1 := 0.0
// // 	magnitude2 := 0.0

// // 	// Calculate the dot product and magnitudes
// // 	for word, freq1 := range tf1 {
// // 		freq2 := tf2[word]
// // 		dotProduct += freq1 * freq2
// // 		magnitude1 += freq1 * freq1
// // 	}

// // 	for _, freq2 := range tf2 {
// // 		magnitude2 += freq2 * freq2
// // 	}

// // 	// Compute the cosine similarity
// // 	if magnitude1 == 0 || magnitude2 == 0 {
// // 		return 0, errors.New("one of the vectors has zero magnitude")
// // 	}

// // 	// Calculate cosine similarity in the range of 0 to 1
// // 	cosineSim := dotProduct / (math.Sqrt(magnitude1) * math.Sqrt(magnitude2))

// // 	// Convert the cosine similarity to the range of 1 to 100
// // 	matchScore := cosineSim * 100
// // 	if matchScore < 0 {
// // 		matchScore = 0 // Ensure score is non-negative
// // 	} else if matchScore > 100 {
// // 		matchScore = 100 // Cap the score at 100
// // 	}

// // 	return matchScore, nil
// // }




// // getDirectMatchScoreFromHuggingFace sends a request to Hugging Face API to get the match score
// // func getDirectMatchScoreFromHuggingFace(apiKey, seekerText, jobText, model string) (float64, error) {
// // 	url := fmt.Sprintf("https://api-inference.huggingface.co/models/%s", model)
// // 	payload := map[string]interface{}{
// // 		"inputs": fmt.Sprintf("%s %s", seekerText, jobText), // Combine both texts for matching
// // 	}

// // 	payloadBytes, err := json.Marshal(payload)
// // 	if err != nil {
// // 		return 0, fmt.Errorf("error marshaling payload: %v", err)
// // 	}

// // 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
// // 	if err != nil {
// // 		return 0, fmt.Errorf("error creating request: %v", err)
// // 	}

// // 	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))
// // 	req.Header.Add("Content-Type", "application/json")

// // 	client := &http.Client{}
// // 	resp, err := client.Do(req)
// // 	if err != nil {
// // 		return 0, fmt.Errorf("error making request to Hugging Face: %v", err)
// // 	}
// // 	defer resp.Body.Close()

// // 	// Check for successful response
// // 	if resp.StatusCode != http.StatusOK {
// // 		return 0, fmt.Errorf("error: received non-200 response code: %d", resp.StatusCode)
// // 	}

// // 	// Read and parse the response
// // 	body, err := io.ReadAll(resp.Body)
// // 	if err != nil {
// // 		return 0, fmt.Errorf("error reading response: %v", err)
// // 	}

// // 	var response struct {
// // 		Score float64 `json:"score"`
// // 	}
// // 	if err := json.Unmarshal(body, &response); err != nil {
// // 		return 0, fmt.Errorf("error unmarshaling response: %v", err)
// // 	}

// // 	return response.Score, nil
// // }


// func joinStrings(strs []string, sep string) string {
// 	if len(strs) == 0 {
// 		return ""
// 	}
// 	result := strs[0]
// 	for _, s := range strs[1:] {
// 		result += sep + s
// 	}
// 	return result
// }

