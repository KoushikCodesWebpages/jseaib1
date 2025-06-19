package generation

// import (
// 	//"bytes"
// 	"fmt"
// 		"gorm.io/gorm"
// 	"net/http"
// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"
// 	"RAAS/models"
// 	"io"
// )

// // CVHandler struct to hold DB reference (or other dependencies)
// // type CVDownloadHandler struct {
// // 	db *gorm.DB
// // }

// // NewCVHandler creates a new CVHandler with the given DB instance
// func NewCVDownloadHandler(db *gorm.DB) *CVHandler {
// 	return &CVHandler{db: db}
// }

// func (h *CVHandler) GetCVMetadata(c *gin.Context) {
// 	userID := c.MustGet("userID").(uuid.UUID)

// 	var cv models.CV
// 	if err := h.db.Where("auth_user_id = ?", userID).First(&cv).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "CV not found for this user"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"job_id":    cv.JobID,
// 		"cv_url":    cv.CVUrl,
// 		"createdAt": cv.CreatedAt,
// 	})
// }

// type DownloadCVRequest struct {
// 	JobID string `json:"job_id"`
// }

// // GetCV handles the retrieval of a CV by user ID and job ID (POST method)
// func (h *CVHandler) DownloadCV(c *gin.Context) {
// 	// Retrieve the user ID from the context (assuming it's set via authentication middleware)
// 	userID := c.MustGet("userID").(uuid.UUID)

// 	var requestBody DownloadCVRequest

// 	// Bind the request body to the struct
// 	if err := c.ShouldBindJSON(&requestBody); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
// 		return
// 	}

// 	jobID := requestBody.JobID
// 	if jobID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
// 		return
// 	}

// 	var cv models.CV

// 	// Query the CV table for the specified AuthUserID and JobID
// 	if err := h.db.Where("auth_user_id = ? AND job_id = ?", userID, jobID).First(&cv).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "CV not found"})
// 		return
// 	}

// 	if cv.CVUrl == "" {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "CV file URL not found"})
// 		return
// 	}

// 	// Download the CV from the URL
// 	response, err := http.Get(cv.CVUrl)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download the file", "details": err.Error()})
// 		return
// 	}
// 	defer response.Body.Close()

// 	if response.StatusCode != http.StatusOK {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to download the file, received status: %s", response.Status)})
// 		return
// 	}

// 	// Read the file content into memory
// 	fileContent, err := io.ReadAll(response.Body)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the file content", "details": err.Error()})
// 		return
// 	}

// 	// Set headers and send the file in-memory
// 	c.Header("Content-Disposition", "attachment; filename=cv.docx")
// 	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
// 	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", fileContent)
// }
