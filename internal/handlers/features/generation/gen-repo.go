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
    db *mongo.Database,
    userID,
    jobID string,
    genType string,   // "cover_letter", "cv", etc.
    sourceType string, // "internal" | "external"
) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    seekersColl := db.Collection("seekers")
    appsColl := db.Collection("selected_job_applications")

    fieldGen := fmt.Sprintf("%s_generated", genType)
    filter := bson.M{"auth_user_id": userID, "job_id": jobID}

    var existing bson.M
    err := appsColl.FindOne(ctx, filter).Decode(&existing)
    isInsert := err == mongo.ErrNoDocuments

    // For existing documents, also enforce view_link=true when external
    mustSetViewLink := sourceType == "external"

    if isInsert {
        var seeker struct {
            InternalApplications int `bson:"internal_application_count"`
            ExternalApplications int `bson:"external_application_count"`
        }
        if err := seekersColl.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker); err != nil {
            return fmt.Errorf("failed to fetch seeker: %w", err)
        }

        if sourceType == "internal" && seeker.InternalApplications <= 0 {
            return fmt.Errorf("internal application limit exceeded")
        }
        if sourceType == "external" && seeker.ExternalApplications <= 0 {
            return fmt.Errorf("external application limit exceeded")
        }

        session, err := db.Client().StartSession()
        if err != nil {
            return fmt.Errorf("failed to start session: %w", err)
        }
        defer session.EndSession(ctx)

        _, err = session.WithTransaction(ctx,
            func(sc mongo.SessionContext) (interface{}, error) {
                setFields := bson.M{
                    fieldGen:        true,
                    "selected_date": time.Now(),
                }
                onInsert := bson.M{
                    "status":          "pending",
                    "source":          sourceType,
                    "view_link":       mustSetViewLink, // external → true; internal → false
                }
                update := bson.M{
                    "$set":         setFields,
                    "$setOnInsert": onInsert,
                }
                _, err := appsColl.UpdateOne(sc, filter, update, options.Update().SetUpsert(true))
                if err != nil {
                    return nil, err
                }

                decField := "external_application_count"
                if sourceType == "internal" {
                    decField = "internal_application_count"
                }
                if _, err := seekersColl.UpdateOne(sc,
                    bson.M{"auth_user_id": userID, decField: bson.M{"$gt": 0}},
                    bson.M{"$inc": bson.M{decField: -1}},
                ); err != nil {
                    return nil, err
                }

                return nil, nil
            },
        )
        if err != nil {
            return fmt.Errorf("transaction failed: %w", err)
        }
    } else {
        // Existing document: just update gen field, selected_date, and view_link if external
        updateBson := bson.M{
            "$set": bson.M{
                fieldGen:        true,
                "selected_date": time.Now(),
            },
        }
        if mustSetViewLink {
            updateBson["$set"].(bson.M)["view_link"] = true
        }
        if _, err := appsColl.UpdateOne(ctx, filter, updateBson); err != nil {
            return fmt.Errorf("update existing application failed: %w", err)
        }
    }

    return nil
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