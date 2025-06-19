package generation

// import (
//     "RAAS/core/config" 

// 	"bytes"
// 	"encoding/json"
// 	"fmt"

// 	"net/http"

// 	"io"
// )

// type RequestData struct {
// 	Inputs string `json:"inputs"`
// }

// type ResponseData struct {
// 	GeneratedText string `json:"generated_text"`
// }

// type CVInput struct {
// 	Name             string   `json:"name"`
// 	Designation      string   `json:"designation"`
// 	Contact          string   `json:"contact"`
// 	ProfileSummary   string   `json:"profile_summary"`
// 	SkillsAndTools   []string `json:"skills_and_tools"`
// 	Education        []struct {
// 		Years      string   `json:"years"`
// 		Institution string `json:"institution"`
// 		Details    []string `json:"details"`
// 	} `json:"education"`
// 	ExperienceSummary []struct {
// 		Title  string   `json:"title"`
// 		Bullets []string `json:"bullets"`
// 	} `json:"experience_summary"`
// 	Languages []string `json:"languages"`
// }

// // GenerateCVDocx generates a CV document using input data
// func GenerateCVDocx(input CVInput) ([]byte, error) {
//     // Get the API URL and Key from the config
//     apiURL := config.Cfg.Cloud.CV_Url
//     apiKey := config.Cfg.Cloud.GEN_API_KEY

//     // Check if the required fields are present
//     if apiURL == "" || apiKey == "" {
//         return nil, fmt.Errorf("CV_API_URL or COVER_CV_API_KEY is missing in config")
//     }

//     // Marshal the input data into JSON
//     jsonData, err := json.Marshal(input)
//     if err != nil {
//         return nil, fmt.Errorf("error marshaling input data: %w", err)
//     }

//     // Create the request
//     req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
//     if err != nil {
//         return nil, fmt.Errorf("error creating request: %w", err)
//     }

//     // Set headers
//     req.Header.Set("Content-Type", "application/json")
//     req.Header.Set("Authorization", "Bearer "+apiKey)

//     // Send the request
//     client := &http.Client{}
//     resp, err := client.Do(req)
//     if err != nil {
//         return nil, fmt.Errorf("error sending request: %v", err)
//     }
//     defer resp.Body.Close()

//     // Read the response body into a byte buffer
//     var buf bytes.Buffer
//     _, err = io.Copy(&buf, resp.Body)
//     if err != nil {
//         return nil, fmt.Errorf("error reading response body: %v", err)
//     }

//     return buf.Bytes(), nil
// }


