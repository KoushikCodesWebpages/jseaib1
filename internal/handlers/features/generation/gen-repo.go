package generation

import (
    "context"
    "time"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "fmt"

    "bytes"
    "encoding/json"
    
    "io"
    "net/http"

    "RAAS/core/config"

    
)

func upsertSelectedJobApp(
    coll *mongo.Collection,
    userID,
    jobID string,
    genType string, // "cover_letter", "cv", etc.
    sourceType string, // "internal" | "external"
) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // dynamic field name for generated type
    fieldGen := fmt.Sprintf("%s_generated", genType)

    filter := bson.M{"auth_user_id": userID, "job_id": jobID}
    update := bson.M{
        "$set": bson.M{
            fieldGen:       true,
            "selected_date": time.Now(),
        },
        "$setOnInsert": bson.M{
            "view_link": false,
            "status":    "pending",
            "source":     sourceType,
        },
    }
    opts := options.Update().SetUpsert(true)
    _, err := coll.UpdateOne(ctx, filter, update, opts)
    return err
}

func callAPI(apiURL, apiKey string, payload map[string]interface{}) (map[string]interface{}, error) {
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

// CallCoverLetterAPI wraps callAPI for cover letters
func CallCoverLetterAPI(payload map[string]interface{}) (map[string]interface{}, error) {
    return callAPI(config.Cfg.Cloud.CL_Url, config.Cfg.Cloud.GEN_API_KEY, payload)
}

// CallCVAPI wraps callAPI for CVs
func CallCVAPI(payload map[string]interface{}) (map[string]interface{}, error) {
    return callAPI(config.Cfg.Cloud.CV_Url, config.Cfg.Cloud.GEN_API_KEY, payload)
}