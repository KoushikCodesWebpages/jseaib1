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
    userID, jobID, genType, sourceType string, // sourceType: "internal" | "external"
) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    appsColl := db.Collection("selected_job_applications")
    seekersColl := db.Collection("seekers")
    jobsColl := db.Collection("jobs") // For internal job lookup

    // Only fetch `company` for internal jobs
    var company string
    if sourceType == "internal" {
        var intJob struct{ Company string `bson:"company"` }
        if err := jobsColl.FindOne(ctx, bson.M{"job_id": jobID}).Decode(&intJob); err != nil {
            return fmt.Errorf("internal job not found: %w", err)
        }
        company = intJob.Company
    }

    fieldGen := fmt.Sprintf("%s_generated", genType)
    filter := bson.M{"auth_user_id": userID, "job_id": jobID}

    existingErr := appsColl.FindOne(ctx, filter).Err()
    isInsert := existingErr == mongo.ErrNoDocuments
    mustSetViewLink := sourceType == "external"

    if isInsert {
        session, err := db.Client().StartSession()
        if err != nil {
            return fmt.Errorf("failed to start session: %w", err)
        }
        defer session.EndSession(ctx)

        _, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
            // Common update fields
            setFields := bson.M{
                fieldGen:        true,
                "selected_date": time.Now(),
            }
            // Only set company if internal
            if sourceType == "internal" {
                setFields["company"] = company
            }

            update := bson.M{
                "$set":         setFields,
                "$setOnInsert": bson.M{
                    "status":    "pending",
                    "source":    sourceType,
                    "view_link": mustSetViewLink,
                },
            }
            opts := options.Update().SetUpsert(true)
            if _, err := appsColl.UpdateOne(sc, filter, update, opts); err != nil {
                return nil, err
            }

            // Decrement seeker application count
            decField := "internal_application_count"
            if sourceType == "external" {
                decField = "external_application_count"
            }
            _, err := seekersColl.UpdateOne(sc,
                bson.M{"auth_user_id": userID, decField: bson.M{"$gt": 0}},
                bson.M{"$inc": bson.M{decField: -1}},
            )
            return nil, err
        })
        if err != nil {
            return fmt.Errorf("transaction failed: %w", err)
        }
    } else {
        update := bson.M{"$set": bson.M{
            fieldGen:        true,
            "selected_date": time.Now(),
        }}
        if sourceType == "internal" {
            update["$set"].(bson.M)["company"] = company
        }
        if mustSetViewLink {
            update["$set"].(bson.M)["view_link"] = true
        }
        if _, err := appsColl.UpdateOne(ctx, filter, update); err != nil {
            return fmt.Errorf("failed updating existing app: %w", err)
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