package generation


import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"

    "RAAS/core/config"
    "RAAS/internal/handlers/repository"
    "RAAS/internal/models"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"

)

type InternalCVHandler struct{}

func NewInternalCVHandler() *InternalCVHandler {
    return &InternalCVHandler{}
}

// POST /b1/generate-cv
func (h *InternalCVHandler) PostCV(c *gin.Context) {

    db := c.MustGet("db").(*mongo.Database)
    cvColl := db.Collection("cvs")
    selColl := db.Collection("selected_job_applications")
    jobColl := db.Collection("jobs")
    seekerColl := db.Collection("seekers")
    authUserColl := db.Collection("auth_users")

    userID := c.MustGet("userID").(string)

    var req struct{ JobID string `json:"job_id" binding:"required"` }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid job_id"})
        return
    }

    // 1. Check for existing CV
    var selApp struct{ CVGenerated bool `bson:"cv_generated"` }

    selColl.FindOne(c, bson.M{"auth_user_id": userID, "job_id": req.JobID}).Decode(&selApp)

    if selApp.CVGenerated {
        var existing models.CVData
        if err := cvColl.FindOne(c, bson.M{"auth_user_id": userID, "job_id": req.JobID}).Decode(&existing); err == nil {
            c.JSON(http.StatusOK, existing.CVContent)
            return
        }
    }

    // 2. Fetch supporting data
    var job models.Job
	
    if err := jobColl.FindOne(c, bson.M{"job_id": req.JobID}).Decode(&job); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
        return
    }
    var seeker models.Seeker
    seekerColl.FindOne(c, bson.M{"auth_user_id": userID}).Decode(&seeker)

    var authUser models.AuthUser
    authUserColl.FindOne(c, bson.M{"auth_user_id": userID}).Decode(&authUser)

    pInfo, _ := repository.GetPersonalInfo(&seeker)
    we, _ := repository.GetWorkExperience(&seeker)
    certs, _ := repository.GetCertificates(&seeker)
    langs, _ := repository.GetLanguages(&seeker)
    pastProjects, _ := repository.GetPastProjects(&seeker)
    education, _ := repository.GetAcademics(&seeker)

    // 3. Build payload matching your required structure
    payload := map[string]interface{}{
        "user_details": map[string]interface{}{
            "name":               fmt.Sprintf("%s %s", pInfo.FirstName, *pInfo.SecondName),
            "designation":        seeker.PrimaryTitle,
            "address":            pInfo.City,
            "contact":            authUser.Phone,
            "email":              authUser.Email,
            "portfolio":          pInfo.Portfolio,
            "linkedin":           pInfo.LinkedInProfile,
            "tools":              []string{},           // or populate as needed
            "skills":             seeker.KeySkills,
            "education":          education,
            "experience_summary": we,
            "past_projects":      pastProjects,
            "certifications":     certs,
            "languages":          langs,
        },
        "job_description": map[string]interface{}{
            "job_title":       job.JobTitle,
            "title":           job.Title,
            "company":         job.Company,
            "location":        job.Location,
            "job_type":        job.JobType,
            "link":            job.Link,
            "description":     job.JobDescription,
            "responsibilities": "", // optional
            "qualifications":   "", // optional
            "skills":           job.Skills,
            "benefits":         "", // optional
        },
        "cv_data": map[string]string{"language": "English", "spec": ""},
    }

    // 4. Call ML CV API
    mlResp, err := h.callCVAPI(payload)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("CV API failed: %v", err)})
        return
    }

    // 5. Save generated CV JSON
    _, err = cvColl.InsertOne(c, bson.M{
        "auth_user_id": userID,
        "job_id":       req.JobID,
        "cv_data":      mlResp,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save CV data"})
        return
    }

    // 6. Upsert tracking entry
    if err := upsertSelectedJobApp(selColl, userID, req.JobID, "cv", "internal"); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application status"})
        return
    }

    // 7. Return CV JSON
    c.JSON(http.StatusOK, mlResp)
}

// GET /b1/generate-cv?job_id=...
func (h *InternalCVHandler) GetCV(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    cvColl := db.Collection("cvs")

    userID := c.MustGet("userID").(string)
    jobID := c.Query("job_id")
    if jobID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing job_id query parameter"})
        return
    }

    var existing models.CVData
    err := cvColl.FindOne(c, bson.M{
        "auth_user_id": userID,
        "job_id":       jobID,
    }).Decode(&existing)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "CV not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        }
        return
    }

    c.JSON(http.StatusOK, existing.CVContent)
}

// helper to call ML API
func (h *InternalCVHandler) callCVAPI(payload map[string]interface{}) (map[string]interface{}, error) {
    apiURL, apiKey := config.Cfg.Cloud.CV_Url, config.Cfg.Cloud.GEN_API_KEY

    buf, _ := json.Marshal(payload)
    req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(buf))
    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("CV API error: %s", string(body))
    }

    var out map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        return nil, err
    }

    return out, nil
}
