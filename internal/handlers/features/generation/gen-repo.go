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

	// Check if application already exists
	var existing bson.M
	err := appsColl.FindOne(ctx, filter).Decode(&existing)
	isInsert := err == mongo.ErrNoDocuments

	if isInsert {
		// Fetch seeker and check application limits
		var seeker struct {
			InternalApplications int `bson:"internal_application_count"`
			ExternalApplications int `bson:"external_application_count"`
		}

		err := seekersColl.FindOne(ctx, bson.M{"auth_user_id": userID}).Decode(&seeker)
		if err != nil {
			return fmt.Errorf("failed to fetch seeker: %w", err)
		}

		if sourceType == "internal" && seeker.InternalApplications <= 0 {
			return fmt.Errorf("internal application limit exceeded")
		} else if sourceType == "external" && seeker.ExternalApplications <= 0 {
			return fmt.Errorf("external application limit exceeded")
		}

		// Start atomic transaction
		session, err := db.Client().StartSession()
		if err != nil {
			return fmt.Errorf("failed to start session: %w", err)
		}
		defer session.EndSession(ctx)

		_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
			// Insert application or upsert
			update := bson.M{
				"$set": bson.M{
					fieldGen:        true,
					"selected_date": time.Now(),
				},
				"$setOnInsert": bson.M{
					"view_link": false,
					"status":    "pending",
					"source":    sourceType,
				},
			}
			opts := options.Update().SetUpsert(true)
			_, err := appsColl.UpdateOne(sc, filter, update, opts)
			if err != nil {
				return nil, err
			}

			// Decrement correct field based on sourceType
			var decField string
			if sourceType == "internal" {
				decField = "internal_application_count"
			} else {
				decField = "external_application_count"
			}

			_, err = seekersColl.UpdateOne(sc,
				bson.M{"auth_user_id": userID, decField: bson.M{"$gt": 0}},
				bson.M{"$inc": bson.M{decField: -1}},
			)
			if err != nil {
				return nil, err
			}

			return nil, nil
		})
		if err != nil {
			return fmt.Errorf("transaction failed: %w", err)
		}
	} else {
		// Just update generated field if already exists
		update := bson.M{
			"$set": bson.M{
				fieldGen:        true,
				"selected_date": time.Now(),
			},
		}
		_, err := appsColl.UpdateOne(ctx, filter, update)
		if err != nil {
			return err
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