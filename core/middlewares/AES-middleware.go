package middleware

import (
	"RAAS/core/security"

	"bytes"
	"fmt"
	"io"
	"net/http"
 // Import your encryption functions
	"github.com/gin-gonic/gin"
)

// DecryptRequestMiddleware is a Gin middleware that will decrypt incoming data.
func DecryptRequestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = c.GetRawData() // Get raw data from the request
		}

		// Decrypt the request body (if it's encrypted)
		decryptedData, err := security.DecryptData(string(requestBody))
		if err != nil {
			// Handle decryption error, send a response with status 400 Bad Request
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to decrypt data: %v", err)})
			c.Abort()
			return
		}

		// Replace the request body with the decrypted data
		c.Request.Body = io.NopCloser(bytes.NewReader(decryptedData))

		// Continue to the next handler
		c.Next()
	}
}

// EncryptResponseMiddleware is a Gin middleware that will encrypt outgoing responses.
func EncryptResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a response writer to capture the response
		writer := &responseWriter{ResponseWriter: c.Writer}
		c.Writer = writer

		// After the request is completed, encrypt the response data
		c.Next()

		// Encrypt the response data (if it's necessary)
		if c.Writer.Status() == http.StatusOK {
			encryptedResponse, err := security.EncryptData(writer.body)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to encrypt data: %v", err)})
				return
			}

			// Send encrypted data as response
			c.Data(http.StatusOK, "application/json", []byte(encryptedResponse))
		}
	}
}

// responseWriter is a custom implementation of http.ResponseWriter to capture response data.
type responseWriter struct {
	gin.ResponseWriter
	body []byte
}

// Write captures the response body.
func (rw *responseWriter) Write(p []byte) (n int, err error) {
	rw.body = append(rw.body, p...) // Ensure that data is appended, not overwritten.
	return rw.ResponseWriter.Write(p)
}
