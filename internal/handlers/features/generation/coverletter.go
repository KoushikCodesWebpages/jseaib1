package generation

// import (
// 	"RAAS/core/config"
// 	"RAAS/internal/models"
// 	"RAAS/internal/handlers/repository"


// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// )

// type CoverLetterAndCVRequest struct {
// 	JobID string `json:"job_id" binding:"required"`
// }

// type CoverLetterHandler struct{}

// func NewCoverLetterHandler() *CoverLetterHandler {
// 	return &CoverLetterHandler{}
// }

// func (h *CoverLetterHandler) PostCoverLetter(c *gin.Context) {
// 	// Get the database and collections
// 	db := c.MustGet("db").(*mongo.Database)
// 	jobCollection := db.Collection("jobs")
// 	seekerCollection := db.Collection("seekers")
// 	AuthUserCollection := db.Collection("auth_users")
// 	selectedJobCollection := db.Collection("selected_jobs")

// 	// Get the authenticated user's ID
// 	userID := c.MustGet("userID").(string)

// 	// Input: Expect a JobID to retrieve job details
// 	var input struct {
// 		JobID string `json:"job_id"`
// 	}
// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
// 		return
// 	}

// 	// Fetch job details
// 	var job models.Job
// 	if err := jobCollection.FindOne(c, bson.M{"job_id": input.JobID}).Decode(&job); err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
// 		return
// 	}

// 	// Fetch seeker details
// 	var seeker models.Seeker
// 	if err := seekerCollection.FindOne(c, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching seeker data"})
// 		return
// 	}

// 	// Check if daily cover letter quota is exhausted
// 	if seeker.DailyGeneratableCoverletter <= 0 {
// 		c.JSON(http.StatusTooManyRequests, gin.H{
// 			"error":             "Cover letter generation limit reached for today",
// 			"limit_exhausted":   true,
// 			"daily_quota_limit": 0,
// 		})
// 		return
// 	}

// 	// Fetch auth user details
// 	var authuser models.AuthUser
// 	if err := AuthUserCollection.FindOne(c, bson.M{"auth_user_id": userID}).Decode(&authuser); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching authuser data"})
// 		return
// 	}

// 	// Gather details from seeker
// 	personalInfo, _ := repository.GetPersonalInfo(&seeker)
// 	professionalSummary, _ := repository.GetProfessionalSummary(&seeker)
// 	workExperience, _ := repository.GetWorkExperience(&seeker)
// 	education, _ := repository.GetEducation(&seeker)
// 	certificates, _ := repository.GetCertificates(&seeker)
// 	languages, _ := repository.GetLanguages(&seeker)

// 	// Construct the API payload for cover letter generation
// 	apiRequestData := map[string]interface{}{
// 		"user_details": map[string]interface{}{
// 			"name":          fmt.Sprintf("%s %s", personalInfo.FirstName, *personalInfo.SecondName),
// 			"designation":   seeker.PrimaryTitle,
// 			"address":       personalInfo.Address,
// 			"contact":       authuser.Phone,
// 			"email":         authuser.Email,
// 			"portfolio":     "",
// 			"linkedin":      personalInfo.LinkedInProfile,
// 			"tools":         "",
// 			"skills":        professionalSummary.Skills,
// 			"education":     education,
// 			"experience_summary": workExperience,
// 			"certifications": certificates,
// 			"languages":     languages,
// 		},
// 		"job_description": map[string]interface{}{
// 			"job_title":     job.Title,
// 			"company":       job.Company,
// 			"location":      job.Location,
// 			"job_type":      job.JobType,
// 			"responsibilities": "",
// 			"qualifications":   job.JobDescription,
// 			"skills":           job.Skills,
// 			"benefits":         "",
// 		},
// 	}

// 	// Call the external API to generate the cover letter (DOCX)
// 	docxContent, err := h.generateCoverLetter(apiRequestData)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Cover letter generation failed: %v", err)})
// 		return
// 	}

// 	// Reduce the daily generatable cover letter count
// 	updateData := bson.M{
// 		"$inc": bson.M{"daily_generatable_coverletter": -1},
// 	}

// 	// Update the user's daily generatable cover letter count
// 	if _, err := seekerCollection.UpdateOne(c, bson.M{"auth_user_id": userID}, updateData); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating daily cover letter count"})
// 		return
// 	}

// 	_, err = selectedJobCollection.UpdateOne(c, bson.M{
// 		"auth_user_id": userID,
// 		"job_id":       input.JobID,
// 	}, bson.M{
// 		"$set": bson.M{
// 			"cover_letter_generated": true,
// 		},
// 	})
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update selected job status for cover letter"})
// 		return
// 	}

// 	// Set headers and return the .docx file to the user
// 	c.Header("Content-Disposition", "attachment; filename=cover_letter.docx")
// 	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxContent)
// }

// // Helper function to send POST request to the external cover letter generation API
// func (h *CoverLetterHandler) generateCoverLetter(apiRequestData map[string]interface{}) ([]byte, error) {
// 	// Load environment variables
// 	apiURL := config.Cfg.Cloud.CL_Url
// 	apiKey := config.Cfg.Cloud.GEN_API_KEY

// 	// Marshal cover letter data to JSON
// 	jsonData, err := json.Marshal(apiRequestData)
// 	if err != nil {
// 		return nil, fmt.Errorf("error marshalling cover letter data: %v", err)
// 	}

// 	// Create a POST request to the API
// 	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		return nil, fmt.Errorf("error creating request: %v", err)
// 	}

// 	// Set headers
// 	req.Header.Set("Authorization", "Bearer "+apiKey)
// 	req.Header.Set("Content-Type", "application/json")

// 	// Send the request to the cover letter API
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("error sending request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	// Check if the response status is OK
// 	if resp.StatusCode != http.StatusOK {
// 		body, _ := io.ReadAll(resp.Body)
// 		return nil, fmt.Errorf("error response from API: %v", string(body))
// 	}

// 	// Read the DOCX content from the response
// 	docxFileContent, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("error reading DOCX content: %v", err)
// 	}

// 	// Return the DOCX file content
// 	return docxFileContent, nil
// }
