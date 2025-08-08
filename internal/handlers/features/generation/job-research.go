package generation

import (

	"fmt"
	"net/http"
	"strings"
	"time"


	"RAAS/internal/handlers/repository"
	"RAAS/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
    // "go.mongodb.org/mongo-driver/bson/primitive"
)

type JobResearchHandler struct{}

func NewJobResearchHandler() *JobResearchHandler {
    return &JobResearchHandler{}
}

func (h *JobResearchHandler) PostJobResearch(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)

    selColl := db.Collection("selected_job_applications")
    jobColl := db.Collection("jobs")
    extJobColl := db.Collection("external_jobs")
    seekerColl := db.Collection("seekers")
    authUserColl := db.Collection("auth_users")
    researchColl := db.Collection("job_research_results")

    userID := c.MustGet("userID").(string)

    var req struct {
        JobID string `json:"job_id" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid values"})
        return
    }
    if userID == "" || req.JobID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user or job ID"})
        return
    }

    // 1. Check if research already exists
    var existing models.JobResearchResult
    err := researchColl.FindOne(c, bson.M{
        "auth_user_id": userID,
        "job_id":       req.JobID,
    }).Decode(&existing)
    if err == nil {
        c.JSON(http.StatusOK, existing.Response)
        return
    } else if err != mongo.ErrNoDocuments {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing research"})
        return
    }

    // 2. Check interview status & source
    var selApp models.SelectedJobApplication
    if err := selColl.FindOne(c, bson.M{
        "auth_user_id": userID,
        "job_id":       req.JobID,
    }).Decode(&selApp); err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "No selected job application found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch selected job application"})
        }
        return
    }
    if !strings.EqualFold(selApp.Status, "interview") {
        c.JSON(http.StatusForbidden, gin.H{"error": "Job is not in interview stage"})
        return
    }

    var payload map[string]interface{}

    if strings.EqualFold(selApp.Source, "internal") {
        // INTERNAL — fetch from jobs
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

        payload = map[string]interface{}{
            "company":   job.Company,
            "job_title": job.JobTitle,
            "candidate_profile": map[string]interface{}{
                "name":               fmt.Sprintf("%s %s", pInfo.FirstName, *pInfo.SecondName),
                "designation":        seeker.PrimaryTitle,
                "address":            pInfo.City,
                "contact":            authUser.Phone,
                "email":              authUser.Email,
                "portfolio":          "",
                "linkedin":           pInfo.LinkedInProfile,
                "tools":              []string{},
                "skills":             seeker.KeySkills,
                "education":          educationObjs,
                "experience_summary": experienceSummaryObjs,
                "past_projects":      pastProjects,
                "certifications":     certificateObjs,
                "languages":          languageObjs,
            },
            "job_description": map[string]interface{}{
                "job_title":      job.JobTitle,
                "title":          job.Title,
                "company":        job.Company,
                "location":       job.Location,
                "job_type":       job.JobType,
                "link":           job.Link,
                "description":    job.JobDescription,
                "responsibilities": "",
                "qualifications": "job.Qualifications",
                "skills":         "",
                "benefits":       "",
            },
            "job_link": job.Link,
        }

    } else if strings.EqualFold(selApp.Source, "external") {
        // EXTERNAL — fetch minimal details
        var extJob struct {
            JobID       string    `bson:"job_id"`
            Title       string    `bson:"title"`
            Company     string    `bson:"company"`
            Description string    `bson:"description"`
            JobLang     string    `bson:"job_language"`
            PostedDate  time.Time `bson:"posted_date"`
        }
        if err := extJobColl.FindOne(c, bson.M{"job_id": req.JobID}).Decode(&extJob); err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "External job not found"})
            return
        }

        payload = map[string]interface{}{
            "job_id":       extJob.JobID,
            "title":        extJob.Title,
            "company":      extJob.Company,
            "description":  extJob.Description,
            "job_language": extJob.JobLang,
            "posted_date":  extJob.PostedDate,
        }
    } else {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job source"})
        return
    }

    // // 3. Call Job Research API
    // researchResp, err := CallJobResearchAPI(payload)
    // if err != nil {
    //     c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Job Research API failed: %v", err)})
    //     return
    // }

    // // 4. Save result
    // _, err = researchColl.InsertOne(c, models.JobResearchResult{
    //     AuthUserID:  userID,
    //     JobID:       req.JobID,
    //     Response:    researchResp,
    //     GeneratedAt: time.Now(),
    // })
    // if err != nil {
    //     c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save job research result"})
    //     return
    // }

    // // 5. Return response
    // c.JSON(http.StatusOK, researchResp)
    c.JSON(http.StatusOK, payload)
}
