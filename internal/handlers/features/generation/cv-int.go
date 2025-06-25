package generation

import (

	"fmt"
	"net/http"



	"RAAS/internal/handlers/repository"
	"RAAS/internal/models"
    "log"

	"github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type InternalCVHandler struct{}

func NewInternalCVHandler() *InternalCVHandler {
    return &InternalCVHandler{}
}

// POST /b1/generate-cv
func (h *InternalCVHandler) PostCV(c *gin.Context) {

    db := c.MustGet("db").(*mongo.Database)
    cvColl := db.Collection("cv")
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
            c.JSON(http.StatusOK, existing.CVData)
            return
        }
    }

    // Step 2: Upsert first to validate limits and track usage
	if err := upsertSelectedJobApp(db, userID, req.JobID, "cv", "internal"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
    // 3. Fetch supporting data
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
        experienceSummaryObjs, _ := repository.GetWorkExperience(&seeker)
        certificateObjs, _ := repository.GetCertificates(&seeker)
        languageObjs, _ := repository.GetLanguages(&seeker)
        pastProjects, _ := repository.GetPastProjects(&seeker)
        educationObjs, _ := repository.GetAcademics(&seeker)

    // 1️⃣ Build education strings
	education := []string{}
	for _, e := range educationObjs {
		degree, _ := e["degree"].(string)
		field, _ := e["field_of_study"].(string)
		inst, _ := e["institution"].(string)

		// Handle start_date
		startStr := "Unknown"
		if startRaw, ok := e["start_date"].(primitive.DateTime); ok && !startRaw.Time().IsZero() {
			startStr = startRaw.Time().Format("Jan 2006")
		}
		// Handle end_date
		endStr := "Present"
		if endRaw, ok := e["end_date"].(primitive.DateTime); ok && !endRaw.Time().IsZero() {
			endStr = endRaw.Time().Format("Jan 2006")
		}

		period := fmt.Sprintf("%s – %s", startStr, endStr)
		education = append(education, fmt.Sprintf("%s in %s at %s (%s)", degree, field, inst, period))
	}


    // 2️⃣ Extract cert titles
    certifications := []string{}
    for _, cert := range certificateObjs {
        name, _ := cert["certificate_name"].(string)
        if name != "" {
            certifications = append(certifications, name)
        }
    }

    // 3️⃣ Format language proficiency
    languages := []string{}
    for _, lang := range languageObjs {
        langName, _ := lang["language"].(string)
        proficiency, _ := lang["proficiency"].(string)
        if langName != "" {
            languages = append(languages, fmt.Sprintf("%s: %s", langName, proficiency))
        }
    }
    //  Format experience summary
    experienceSummaries := []string{}
    for _, e := range experienceSummaryObjs {
        // Handle start_date
        startStr := ""
        if startVal, ok := e["start_date"].(primitive.DateTime); ok {
            startStr = startVal.Time().Format("Jan 2006")
        }

        // Handle end_date
        endStr := "Present"
        if endVal, ok := e["end_date"].(primitive.DateTime); ok && !endVal.Time().IsZero() {
            endStr = endVal.Time().Format("Jan 2006")
        }

        // Period string
        period := fmt.Sprintf("%s – %s", startStr, endStr)

        // Extract fields
        position, _ := e["job_title"].(string)
        company, _ := e["company_name"].(string)
        description, _ := e["key_responsibilities"].(string)

        // Final summary
        summary := fmt.Sprintf("%s at %s (%s): %s", position, company, period, description)
        experienceSummaries = append(experienceSummaries, summary)
    }

    // 4. Build payload matching your required structure
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
            "experience_summary": experienceSummaries,
            "past_projects":      pastProjects,
            "certifications":     certifications,
            "languages":          languages,
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

    // 5. Call ML CV API
    cvResp, err := CallCVAPI(payload)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("CV API failed: %v", err)})
        return
    }

    if userID == "" || req.JobID == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user or job ID"})
    return
    }
    // 6. Save generated CV JSON
    _, err = cvColl.InsertOne(c, bson.M{
        "auth_user_id": userID,
        "job_id":       req.JobID,
        "cv_data":      cvResp,
    })
    if err != nil {
    log.Printf("InsertOne error: %v", err)
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save CV data"})
    return
}

    c.JSON(http.StatusOK, gin.H{
        "job_id":  req.JobID,
        "cv_data": cvResp,
    })
}

// GET /b1/generate-cv?job_id=...
func (h *InternalCVHandler) GetCV(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    cvColl := db.Collection("cv")

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

    c.JSON(http.StatusOK, gin.H{
        "job_id"    :  jobID,
        "cv_data"   :  existing.CVData,
    })


}

func (h *InternalCVHandler) PutCV(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    cvColl := db.Collection("cv")
    userID := c.MustGet("userID").(string)

    var req struct {
        JobID  string                 `json:"job_id" binding:"required"`
        CVData map[string]interface{} `json:"cv_data" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
    filter := bson.M{"auth_user_id": userID, "job_id": req.JobID}
    update := bson.M{"$set": bson.M{"cv_data": req.CVData}}

    var updated models.CVData
    if err := cvColl.FindOneAndUpdate(c, filter, update, opts).Decode(&updated); err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "CV not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Update error"})
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "job_id"    :  req.JobID,
        "cv_data"   :  updated.CVData,
    })
}
