package generation

import (

	"RAAS/core/config" 
    
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"net/http"
)

type CoverLetterInput struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	RecipientTitle  string `json:"recipient_title"`
	CompanyName     string `json:"company_name"`
	CompanyLocation string `json:"company_location"`
	Body            string `json:"body"`
	Closing         string `json:"closing"`
}


// GenerateCoverLetterDocx generates a cover letter document using input data
func GenerateCoverLetterDocx(input CoverLetterInput, config *config.Config) ([]byte, error) {
    // Get the API URL and Key from the config
    apiURL := config.Cloud.CL_Url
    apiKey := config.Cloud.GEN_API_KEY

    // Check if the required fields are present
    if apiURL == "" || apiKey == "" {
        return nil, fmt.Errorf("COVER_LETTER_API_URL or COVER_LETTER_API_KEY is missing in config")
    }

    // Marshal the input data into JSON
    jsonData, err := json.Marshal(input)
    if err != nil {
        return nil, fmt.Errorf("error marshaling input data: %v", err)
    }

    // Create the request
    req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("error creating request: %v", err)
    }

    // Set headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+apiKey)

    // Send the request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error sending request: %v", err)
    }
    defer resp.Body.Close()

    // Read the response body into a byte buffer
    var buf bytes.Buffer
    _, err = io.Copy(&buf, resp.Body)
    if err != nil {
        return nil, fmt.Errorf("error reading response body: %v", err)
    }

    return buf.Bytes(), nil
}