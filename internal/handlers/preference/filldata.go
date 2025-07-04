package preference

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"RAAS/core/config"
)

type DataExtractionHandler struct{}

// NewDataExtractionHandler creates a new instance of DataExtractionHandler
func NewDataExtractionHandler() *DataExtractionHandler {
	return &DataExtractionHandler{}
}

// ExtractFromPDF handles the file upload and forwards the PDF to the resume API
func (h *DataExtractionHandler) ExtractFromPDF(c *gin.Context) {
	// Parse the uploaded file
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid file: " + err.Error()})
		return
	}

	// Open the file
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file: " + err.Error()})
		return
	}
	defer file.Close()

	// Call the external resume API
	response, err := CallResumeAPI(file, fileHeader.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Resume API error: " + err.Error()})
		return
	}

	// Return the API response as JSON
	c.JSON(http.StatusOK, response)
}

// UploadPDFToResumeAPI uploads the given file to the resume extractor API
func callResumeAPI(file multipart.File, filename string) (map[string]interface{}, error) {
	// Prepare multipart/form-data body
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close writer: %w", err)
	}

	// Prepare HTTP request
	req, err := http.NewRequest("POST", config.Cfg.Cloud.DataExtractionAPI, &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.Cfg.Cloud.GEN_API_KEY)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result, nil
}

// CallResumeAPI handles external usage, wraps `callResumeAPI`
func CallResumeAPI(file multipart.File, filename string) (map[string]interface{}, error) {
	return callResumeAPI(file, filename)
}

