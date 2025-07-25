package generation

import (
 
    "RAAS/internal/handlers/repository"
    "RAAS/internal/models"

    "fmt"
    "net/http"
    "time"
    "log"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type ExternalJobCVNCLGenerator struct{}

func NewExternalJobCVNCLGenerator() *ExternalJobCVNCLGenerator{
	return &ExternalJobCVNCLGenerator{}
}

func (h *ExternalJobCVNCLGenerator) PostExternalCVNCL(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    cvColl := db.Collection("cv")
    clColl := db.Collection("cover_letters")
    selColl := db.Collection("selected_job_applications")
    seekerColl := db.Collection("seekers")
    authUserColl := db.Collection("auth_users")
    extJobColl := db.Collection("external_jobs")

    userID := c.MustGet("userID").(string)

    var req struct {
        JobID          string `json:"job_id" binding:"required"`
        Company        string `json:"company" binding:"required"`
        JobTitle       string `json:"job_title" binding:"required"`
        JobLink        string `json:"link" binding:"required"`
        JobDescription string `json:"description" binding:"required"`
        JobLang        string `json:"job_language" binding:"required"`
        ClFormat       string `json:"cl_format" binding:"required"`
        CvFormat       string `json:"cv_format" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
        return
    }

    // Step 1: Check cached generation
    var selApp struct {
        CoverLetterGenerated bool `bson:"cover_letter_generated"`
        CvGenerated          bool `bson:"cv_generated"`
    }
    selErr := selColl.FindOne(c, bson.M{"auth_user_id": userID, "job_id": req.JobID}).Decode(&selApp)
    if selErr == nil && selApp.CoverLetterGenerated && selApp.CvGenerated {
        // both exist → fetch and return
        var clExisting models.CoverLetterData
        var cvExisting models.CVData
        if err := clColl.FindOne(c, bson.M{"auth_user_id": userID, "job_id": req.JobID}).Decode(&clExisting); err == nil {
            if err2 := cvColl.FindOne(c, bson.M{"auth_user_id": userID, "job_id": req.JobID}).Decode(&cvExisting); err2 == nil {
                c.JSON(http.StatusOK, gin.H{
                    "cl_data": clExisting.CLData,
                    "cv_data": cvExisting.CVData,
                })
                return
            }
        }
    }

    if err := upsertSelectedJobApp(db, userID, req.JobID, "cover_letter", "external"); err != nil {
        c.JSON(http.StatusForbidden, gin.H{"error": err.Error(),"issue":"Limit Exceeded"})
        return
    }
    if err := upsertSelectedJobApp(db, userID, req.JobID, "cv", "external"); err != nil {
        c.JSON(http.StatusForbidden, gin.H{"error": err.Error(),"issue":"Limit Exceeded"})
        return
    }


	//Step 2: Obtain Data for processing.
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

	// Shared user details
    userDetails := map[string]interface{}{
			"name":               fmt.Sprintf("%s %s", pInfo.FirstName, *pInfo.SecondName),
            "designation":        seeker.PrimaryTitle,
            "address":            pInfo.City,
            "contact":            authUser.Phone,
            "email":              authUser.Email,
            "portfolio":          /*pInfo.ExternalLinks*/"",
            "linkedin":           pInfo.LinkedInProfile,
            "tools":              []string{},
                      
            "skills":             seeker.KeySkills,
            "education":          educationObjs,
            "experience_summary": experienceSummaryObjs,
            "past_projects":      pastProjects,
            "certifications":     certificateObjs,
            "languages":          languageObjs,	
    }

	jobDesc := map[string]interface{}{
        "company":         req.Company,
        "job_title":       req.JobTitle,
        "link":            req.JobLink,
        "description":     req.JobDescription,
        "location":        "",
        "job_type":        "",
        "responsibilities": []string{},
        "qualifications":   []string{},
        "skills":           []string{},
        "benefits":         []string{},
    }

	// Step 3: call ML APIs
    cvPayload := map[string]interface{}{
        "user_details":    userDetails,
        "job_description": jobDesc,
        "cv_data":         map[string]string{"language": req.JobLang, "spec": ""},
    }
    cvResp, cvErr := CallCVAPI(cvPayload)
    if cvErr != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "CV generation failed: " + cvErr.Error()})
        return
    }

	clPayload := map[string]interface{}{
        "user_details":    userDetails,
        "job_description": jobDesc,
        "cl_data":         map[string]string{"language": req.JobLang, "spec": ""},
    }
    clResp, clErr := CallCoverLetterAPI(clPayload)
    if clErr != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Cover letter generation failed: " + clErr.Error()})
        return
    }

	// Step 4: Save to MongoDB
    cvColl.InsertOne(c, bson.M{"auth_user_id": userID, "job_id": req.JobID, "cv_data": cvResp, "cv_format":req.CvFormat,})
    clColl.InsertOne(c, bson.M{"auth_user_id": userID, "job_id": req.JobID, "cl_data": clResp,"cl_format":req.ClFormat,})

    

	// limit decrement
    seekerColl.UpdateOne(c, bson.M{"auth_user_id":userID}, bson.M{"$inc": bson.M{"daily_generatable_coverletter": -1}})
     // Step 6: Upsert into external_jobs model
    now := time.Now()
    filterJob := bson.M{"job_id": req.JobID}
    updateJob := bson.M{
        "$setOnInsert": bson.M{
            "job_id":       req.JobID,
            "title":        req.JobTitle,
            "company":      req.Company,
            "description":  req.JobDescription,
            "job_language": req.JobLang,
            "posted_date":  now,
        },
    }
    opts := options.Update().SetUpsert(true)
    if _, err := extJobColl.UpdateOne(c, filterJob, updateJob, opts); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save external job"})
        return
    }


    // 🔁 Step 7: Also update company in selected_job_applications (external)
    _, err := selColl.UpdateOne(
        c,
        bson.M{"auth_user_id": userID, "job_id": req.JobID, "source": "external"},
        bson.M{"$set": bson.M{"company": req.Company}},
    )
    if err != nil {
        log.Printf("❌ Failed to update company in selected_job_applications: %v\n", err)
        // optionally continue or ignore
    }

    // Step 6: response
    c.JSON(http.StatusOK, gin.H{
		"job_id": req.JobID,
        "cv_data": cvResp,
        "cl_data": clResp,
        "cl_format":req.ClFormat,
        "cv_format":req.CvFormat,
    })
}

// GET /b1/external/generate?job_id=...
func (h *ExternalJobCVNCLGenerator) GetExternalCVNCL(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    cvColl := db.Collection("cv")
    clColl := db.Collection("cover_letters")
    selColl := db.Collection("selected_job_applications")

    userID := c.MustGet("userID").(string)
    jobID := c.Query("job_id")
    if jobID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing job_id"})
        return
    }

    // Validate record exists
    var selApp struct {
        CvGenerated          bool `bson:"cv_generated"`
        CoverLetterGenerated bool `bson:"cover_letter_generated"`
    }
    if err := selColl.FindOne(c, bson.M{"auth_user_id": userID, "job_id": jobID}).Decode(&selApp); err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "No data found for this job"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Lookup error"})
        }
        return
    }

    // Fetch documents
    var cv models.CVData
    _ = cvColl.FindOne(c, bson.M{"auth_user_id": userID, "job_id": jobID}).Decode(&cv)

    var cl models.CoverLetterData
    _ = clColl.FindOne(c, bson.M{"auth_user_id": userID, "job_id": jobID}).Decode(&cl)

    c.JSON(http.StatusOK, gin.H{
		"job_id": jobID,
        "cv_data": cv.CVData,
        "cl_data": cl.CLData,
        "cl_format":cl.ClFormat,
        "cv_format":cv.CvFormat,
    })
}

func (h *ExternalJobCVNCLGenerator) PutCoverLetter(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    clColl := db.Collection("cover_letters")
    userID := c.MustGet("userID").(string)

    var req struct {
        JobID       string                  `json:"job_id" binding:"required"`
        CLData      map[string]interface{}  `json:"cl_data" binding:"required"`
        ClFormat    string                  `json:"cl_format" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
    filter := bson.M{"auth_user_id": userID, "job_id": req.JobID}
    update := bson.M{"$set": bson.M{"cl_data": req.CLData,"cl_format":req.ClFormat}}

    var updated models.CoverLetterData
    err := clColl.FindOneAndUpdate(c, filter, update, opts).Decode(&updated)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "Cover letter not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Update error"})
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{"job_id":req.JobID, "cl_data":updated.CLData, "cl_format": req.ClFormat })
}

func (h *ExternalJobCVNCLGenerator) PutCV(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    cvColl := db.Collection("cv")
    userID := c.MustGet("userID").(string)

    var req struct {
        JobID       string                  `json:"job_id" binding:"required"`
        CVData      map[string]interface{}  `json:"cv_data" binding:"required"`
        CvFormat    string                  `json:"cv_format" binding:"required"`

    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
    filter := bson.M{"auth_user_id": userID, "job_id": req.JobID}
    update := bson.M{"$set": bson.M{"cv_data": req.CVData, "cv_format":req.CvFormat}}

    var updated models.CVData
    err := cvColl.FindOneAndUpdate(c, filter, update, opts).Decode(&updated)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "CV not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Update error"})
        }
        return
    }

	c.JSON(http.StatusOK, gin.H{
		"job_id": req.JobID,
        "cv_data": updated.CVData,
        "cv_format": req.CvFormat,
    })
}