package generation

import (
    "context"
    "time"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "fmt"
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
