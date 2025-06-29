package repository

import (
	"RAAS/internal/models"
	// "RAAS/utils"
	// "RAAS/internal/dto"
	// "fmt"
	// "log"
	// "time"
	"context"
	"errors"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)







func IsChecklistComplete(seeker models.Seeker) bool {
	return len(seeker.PersonalInfo) > 0 &&
		len(seeker.WorkExperiences) > 0 &&
		len(seeker.Academics) > 0 &&
		len(seeker.PastProjects) > 0 &&
		len(seeker.Languages) > 0 &&
		len(seeker.Certificates) > 0 &&
		(seeker.PrimaryTitle != "" ||
		 (seeker.SecondaryTitle != nil && *seeker.SecondaryTitle != "") ||
		 (seeker.TertiaryTitle != nil && *seeker.TertiaryTitle != "")) &&
		len(seeker.KeySkills) > 0
}

// FindSeekerByUserID is a global utility function to find a Seeker by userID in MongoDB
func FindSeekerByUserID(collection *mongo.Collection, userID uuid.UUID) (*models.Seeker, error) {
	var seeker models.Seeker
	filter := bson.M{"auth_user_id": userID}
	err := collection.FindOne(context.Background(), filter).Decode(&seeker)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("seeker not found")
		}
		return nil, err
	}
	return &seeker, nil
}

func IsFieldFilled(personalInfo bson.M) bool {
	// Check if the bson.M map is empty
	return len(personalInfo) > 0
}

func DereferenceString(str *string) string {
	if str != nil {
		return *str
	}
	return "" // Return an empty string if the pointer is nil
}


// Helper function to get optional fields
func GetOptionalField(info bson.M, field string) *string {
	if val, ok := info[field]; ok && val != nil {
		v := val.(string)
		return &v
	}
	return nil
}

func GetNextSequence(db *mongo.Database, name string) (uint, error) {
	var result struct {
		SequenceValue uint `bson:"sequence_value"`
	}

	err := db.Collection("counters").FindOneAndUpdate(
		context.TODO(),
		bson.M{"_id": name},
		bson.M{"$inc": bson.M{"sequence_value": 1}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&result)

	return result.SequenceValue, err
}

// CalculateProfileCompletion computes profile completion percentage and returns missing fields
func CalculateProfileCompletion(seeker models.Seeker) (int, []string) {
	completion := 0
	var missing []string

	// Personal Info: Only first_name
	if seeker.PersonalInfo != nil {
		if val, ok := seeker.PersonalInfo["first_name"]; ok && val != "" {
			completion += 10
		} else {
			missing = append(missing, "first_name")
		}
	} else {
		missing = append(missing, "first_name")
	}

	// Key Skills
	if len(seeker.KeySkills) > 0 {
		completion += 20
	} else {
		missing = append(missing, "key_skills")
	}

	// Work Experiences
	if len(seeker.WorkExperiences) > 0 {
		completion += 20
	} else {
		missing = append(missing, "work_experiences")
	}

	// Certificates
	if len(seeker.Certificates) > 0 {
		completion += 20
	} else {
		missing = append(missing, "certificates")
	}

	// Preferred Job Title
	if seeker.PrimaryTitle != "" {
		completion += 20
	} else {
		missing = append(missing, "primary_title")
	}

	// Subscription Tier
	if seeker.SubscriptionTier != "" {
		completion += 10
	} else {
		missing = append(missing, "subscription_tier")
	}

	// Cap at 100
	if completion > 100 {
		completion = 100
	}

	return completion, missing
}