package generation

// import (
// 	"RAAS/core/config"
// 	"RAAS/internal/handlers/repository"
// 	"RAAS/internal/models"
	

// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// )

// type ResumeHandler struct{}

// func NewResumeHandler() *ResumeHandler {
// 	return &ResumeHandler{}
// }

// func (h *ResumeHandler) PostResume(c *gin.Context) {
// 	db := c.MustGet("db").(*mongo.Database)
// 	jobCollection := db.Collection("jobs")
// 	seekerCollection := db.Collection("seekers")
// 	authUserCollection := db.Collection("auth_users")
// 	selectedJobCollection := db.Collection("selected_jobs")

// 	userID := c.MustGet("userID").(string)

// 	var input struct {
// 		JobID string `json:"job_id" binding:"required"`
// 	}
// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
// 		return
// 	}

// 	var job models.Job
// 	if err := jobCollection.FindOne(c, bson.M{"job_id": input.JobID}).Decode(&job); err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
// 		return
// 	}

// 	var seeker models.Seeker
// 	if err := seekerCollection.FindOne(c, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching seeker data"})
// 		return
// 	}

// 	var authUser models.AuthUser
// 	if err := authUserCollection.FindOne(c, bson.M{"auth_user_id": userID}).Decode(&authUser); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching auth user data"})
// 		return
// 	}

// 	// Check if daily CV quota is exhausted
// 	if seeker.DailyGeneratableCV <= 0 {
// 		c.JSON(http.StatusTooManyRequests, gin.H{
// 			"error":             "CV generation limit reached for today",
// 			"limit_exhausted":   true,
// 			"daily_quota_limit": 0,
// 		})
// 		return
// 	}

// 	// Gather details from seeker
// 	personalInfo, _ := repository.GetPersonalInfo(&seeker)
// 	professionalSummary, _ := repository.GetProfessionalSummary(&seeker)
// 	workExperience, _ := repository.GetWorkExperience(&seeker)
// 	educationObjs, _ := repository.GetEducation(&seeker)
// 	certificateObjs, _ := repository.GetCertificates(&seeker)
// 	languageObjs, _ := repository.GetLanguages(&seeker)
	
// 	// Simplify education
// 	education := []string{}
// 	for _, e := range educationObjs {
// 		degree, _ := e["degree"].(string)
// 		institution, _ := e["institution"].(string)
// 		years, _ := e["years"].(string)
// 		education = append(education, fmt.Sprintf("%s, %s, %s", degree, institution, years))
// 	}

// 	// Simplify certifications
// 	certifications := []string{}
// 	for _, cert := range certificateObjs {
// 		name, _ := cert["certificate_name"].(string)
// 		certifications = append(certifications, name)
// 	}

// 	// Simplify languages
// 	languages := []string{}
// 	for _, lang := range languageObjs {
// 		langName, _ := lang["language"].(string)
// 		proficiency, _ := lang["proficiency"].(string)
// 		languages = append(languages, fmt.Sprintf("%s: %s", langName, proficiency))
// 	}

// 	// Construct the API payload for CV generation
// 	resumeRequest := map[string]interface{}{
// 		"user_details": map[string]interface{}{
// 			"name":               personalInfo.FirstName + " " + *personalInfo.SecondName,
// 			"designation":        seeker.PrimaryTitle,
// 			"address":            personalInfo.Address,
// 			"contact":            authUser.Phone,
// 			"email":              authUser.Email,
// 			"portfolio":          "",
// 			"linkedin":           personalInfo.LinkedInProfile,
// 			"tools":              "", // optional
// 			"skills":             professionalSummary.Skills,
// 			"education":          education,
// 			"experience_summary": workExperience,
// 			"certifications":     certifications,
// 			"languages":          languages,
// 		},
// 		"job_description": map[string]interface{}{
// 			"job_title":        job.Title,
// 			"company":          job.Company,
// 			"location":         job.Location,
// 			"job_type":         job.JobType,
// 			// "job_lang":,
// 			"skills":           job.Skills,
// 			"qualifications":   job.JobDescription,
// 		},
// 		"cv_details":map[string]interface{}{
// 			"lang": "lang",
// 		},
// 	}

// 	// Generate the resume (docxContent)
// 	// Generate the resume (pdfContent)
// 	pdfContent, err := h.GenerateResume(resumeRequest)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Resume generation failed: %v", err)})
// 		return
// 	}

// 	// Reduce the daily generatable CV count
// 	updateData := bson.M{
// 		"$inc": bson.M{"daily_generatable_cv": -1},
// 	}

// 	// Update the user's daily generatable CV count
// 	if _, err := seekerCollection.UpdateOne(c, bson.M{"auth_user_id": userID}, updateData); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating daily CV count"})
// 		return
// 	}

// 	// Mark the job as having the CV generated
// 	_, err = selectedJobCollection.UpdateOne(c, bson.M{
// 		"auth_user_id": userID,
// 		"job_id":       input.JobID,
// 	}, bson.M{
// 		"$set": bson.M{
// 			"cv_generated": true,
// 		},
// 	})
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update selected job status for CV"})
// 		return
// 	}

//     // // Return the resume request data as JSON
//     // c.JSON(http.StatusOK, gin.H{
//     //     "message": "Resume request data generated successfully",
//     //     "resume_request": resumeRequest,
//     // })

// 	// Send the resume back to the user
// 	c.Header("Content-Disposition", "attachment; filename=resume.pdf")
// 	c.Data(http.StatusOK, "application/pdf", pdfContent)

// }

// // Helper function to send POST request to the external resume generation API
// func (h *ResumeHandler) GenerateResume(apiRequestData map[string]interface{}) ([]byte, error) {
// 	// Load environment variables
// 	apiURL := config.Cfg.Cloud.CV_Url
// 	apiKey := config.Cfg.Cloud.GEN_API_KEY

// 	// Marshal resume data to JSON
// 	jsonData, err := json.Marshal(apiRequestData)
// 	if err != nil {
// 		return nil, fmt.Errorf("error marshalling resume data: %v", err)
// 	}

// 	// Create a POST request to the API
// 	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		return nil, fmt.Errorf("error creating request: %v", err)
// 	}

// 	// Set headers
// 	req.Header.Set("Authorization", "Bearer "+apiKey)
// 	req.Header.Set("Content-Type", "application/json")

// 	// Send the request to the resume API
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

// 	// Read the PDF content from the response
// 	pdfFileContent, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("error reading PDF content: %v", err)
// 	}

// 	// Return the PDF file content
// 	return pdfFileContent, nil
// }


