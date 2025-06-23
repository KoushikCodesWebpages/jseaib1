package generation

import (
    "RAAS/core/config"
    "RAAS/internal/handlers/repository"
    "RAAS/internal/models"
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type CoverLetterHandler struct{}

func NewCoverLetterHandler() *CoverLetterHandler {
    return &CoverLetterHandler{}
}

func (h *CoverLetterHandler) PostCoverLetter(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    clColl := db.Collection("cover_letters")

    userID := c.MustGet("userID").(string)
    var req struct {
        JobID string `json:"job_id" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing job_id"})
        return
    }

    // 1. Check existing
    var existing models.CoverLetterData
    err := clColl.FindOne(c, bson.M{"auth_user_id": userID, "job_id": req.JobID}).Decode(&existing)
    if err == nil {
        c.JSON(http.StatusOK, existing.CLData)
        return
    }
    if err != mongo.ErrNoDocuments {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "DB lookup error"})
        return
    }

    // 2. Build and send ML API payload
    jobColl := db.Collection("jobs")
    seekerColl := db.Collection("seekers")
    authUserColl := db.Collection("auth_users")
    selectedJobColl := db.Collection("selected_jobs")

    // Fetch data
    var job = new(models.Job)
    if err := jobColl.FindOne(c, bson.M{"job_id": req.JobID}).Decode(job); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
        return
    }
    var seeker = new(models.Seeker)
    _ = seekerColl.FindOne(c, bson.M{"auth_user_id": userID}).Decode(seeker)
    var authUser = new(models.AuthUser)
    _ = authUserColl.FindOne(c, bson.M{"auth_user_id": userID}).Decode(authUser)
    pInfo, _ := repository.GetPersonalInfo(seeker)
    we, _ := repository.GetWorkExperience(seeker)
    certs, _ := repository.GetCertificates(seeker)
    langs, _ := repository.GetLanguages(seeker)
    pastProjects, _ := repository.GetPastProjects(seeker)
    education, _ := repository.GetAcademics(seeker)

    payload := map[string]interface{}{
        "user_details": map[string]interface{}{
            "name":               fmt.Sprintf("%s %s", pInfo.FirstName, *pInfo.SecondName),
            "designation":        seeker.PrimaryTitle,
            "address":            pInfo.City,
            "contact":            authUser.Phone,
            "email":              authUser.Email,
            "portfolio":          pInfo.Portfolio,
            "linkedin":           pInfo.LinkedInProfile,
            "tools":              "",
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
            "responsibilities": "",
            "qualifications":   "",
            "skills":           job.Skills,
            "benefits":         "",
        },
        "cl_data": map[string]string{"language": "English", "spec": ""},
    }

    mlResp, err := h.callCoverLetterAPI(payload)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ML API failed: %v", err)})
        return
    }

    // 3. Save new CL
    _, err = clColl.InsertOne(c, bson.M{
        "auth_user_id": userID,
        "job_id":       req.JobID,
        "cl_data":      mlResp,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Save failed"})
        return
    }

    // Update user limits & job flag
    seekerColl.UpdateOne(c, bson.M{"auth_user_id": userID}, bson.M{"$inc": bson.M{"daily_generatable_coverletter": -1}})
    selectedJobColl.UpdateOne(c,
        bson.M{"auth_user_id": userID, "job_id": req.JobID},
        bson.M{"$set": bson.M{"cover_letter_generated": true}},
    )

    c.JSON(http.StatusOK, mlResp)
}

func (h *CoverLetterHandler) PutCoverLetter(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    clColl := db.Collection("cover_letters")
    userID := c.MustGet("userID").(string)

    var req struct {
        JobID  string                 `json:"job_id" binding:"required"`
        CLData map[string]interface{} `json:"cl_data" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
    filter := bson.M{"auth_user_id": userID, "job_id": req.JobID}
    update := bson.M{"$set": bson.M{"cl_data": req.CLData}}

    var updated models.CoverLetterData
    if err := clColl.FindOneAndUpdate(c, filter, update, opts).Decode(&updated); err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "Cover letter not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Update error"})
        }
        return
    }

    c.JSON(http.StatusOK, updated.CLData)
}

func (h *CoverLetterHandler) GetCoverLetter(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    clColl := db.Collection("cover_letters")
    userID := c.MustGet("userID").(string)

    jobID := c.Query("job_id")
    if jobID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'job_id' query parameter"})
        return
    }

    var stored models.CoverLetterData
    err := clColl.FindOne(c, bson.M{"auth_user_id": userID, "job_id": jobID}).Decode(&stored)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "Cover letter not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
        }
        return
    }

    c.JSON(http.StatusOK, stored.CLData)
}


func (h *CoverLetterHandler) callCoverLetterAPI(payload map[string]interface{}) (map[string]interface{}, error) {
    apiURL, apiKey := config.Cfg.Cloud.CL_Url, config.Cfg.Cloud.GEN_API_KEY

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
        return nil, fmt.Errorf("API error: %s", string(body))
    }

    var out map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        return nil, err
    }
    return out, nil
}
