package generation

import (

    "RAAS/internal/handlers/repository"
    "RAAS/internal/models"
    "fmt"
  
    "net/http"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type InternalCoverLetterHandler struct{}

func NewInternalCoverLetterHandler() *InternalCoverLetterHandler {
    return &InternalCoverLetterHandler{}
}

func (h *InternalCoverLetterHandler) PostCoverLetter(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	clColl := db.Collection("cover_letters")
	selColl := db.Collection("selected_job_applications")
	jobColl := db.Collection("jobs")
	seekerColl := db.Collection("seekers")
	authUserColl := db.Collection("auth_users")

	userID := c.MustGet("userID").(string)

	var req struct {
		JobID string `json:"job_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing job_id"})
		return
	}

	// Step 1: If already generated, return cached CL
	var selApp struct {
		CoverLetterGenerated bool `bson:"cover_letter_generated"`
	}
	selErr := selColl.FindOne(c, bson.M{"auth_user_id": userID, "job_id": req.JobID}).Decode(&selApp)
	if selErr == nil && selApp.CoverLetterGenerated {
		var existing models.CoverLetterData
		if err := clColl.FindOne(c, bson.M{"auth_user_id": userID, "job_id": req.JobID}).Decode(&existing); err == nil {
			c.JSON(http.StatusOK, existing.CLData)
			return
		}
	}

	// Step 2: Upsert selection record FIRST (this checks quota and deducts if first-time)
	if err := upsertSelectedJobApp(db, userID, req.JobID, "cover_letter", "internal"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Step 3: Fetch job, seeker and user details
	var job models.Job
	if err := jobColl.FindOne(c, bson.M{"job_id": req.JobID}).Decode(&job); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	var seeker models.Seeker
	_ = seekerColl.FindOne(c, bson.M{"auth_user_id": userID}).Decode(&seeker)

	var authUser models.AuthUser
	_ = authUserColl.FindOne(c, bson.M{"auth_user_id": userID}).Decode(&authUser)

	pInfo, _ := repository.GetPersonalInfo(&seeker)
	we, _ := repository.GetWorkExperience(&seeker)
	certs, _ := repository.GetCertificates(&seeker)
	langs, _ := repository.GetLanguages(&seeker)
	pastProjects, _ := repository.GetPastProjects(&seeker)
	education, _ := repository.GetAcademics(&seeker)

	// Step 4: Build ML API payload
	payload := map[string]interface{}{
		"user_details": map[string]interface{}{
			"name":               fmt.Sprintf("%s %s", pInfo.FirstName, *pInfo.SecondName),
			"designation":        seeker.PrimaryTitle,
			"address":            pInfo.City,
			"contact":            authUser.Phone,
			"email":              authUser.Email,
			"portfolio":          /*pInfo.ExternalLinks*/"",
			"linkedin":           pInfo.LinkedInProfile,
			"tools":              "VsCode",
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

	// Step 5: Call ML API
	clResp, err := CallCoverLetterAPI(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ML API failed: %v", err)})
		return
	}

	// Step 6: Save generated cover letter
	_, err = clColl.InsertOne(c, bson.M{
		"auth_user_id": userID,
		"job_id":       req.JobID,
		"cl_data":      clResp,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cover letter"})
		return
	}

	// Step 7: Decrease daily CL quota
	seekerColl.UpdateOne(c, bson.M{"auth_user_id": userID}, bson.M{"$inc": bson.M{"daily_generatable_coverletter": -1}})

	// Step 8: Return response
	c.JSON(http.StatusOK, gin.H{
		"job_id":  req.JobID,
		"cl_data": clResp,
	})
}


func (h *InternalCoverLetterHandler) PutCoverLetter(c *gin.Context) {
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

    c.JSON(http.StatusOK, gin.H{
    "job_id":  req.JobID,
    "cl_data": updated.CLData,
    })
}

func (h *InternalCoverLetterHandler) GetCoverLetter(c *gin.Context) {
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

    c.JSON(http.StatusOK, gin.H{
    "job_id":  jobID,
    "cl_data": stored.CLData,
    })
}


