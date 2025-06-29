package oauth

// import (
//     "context"
//     "time"

//     "github.com/google/uuid"
//     "go.mongodb.org/mongo-driver/bson"
//     "go.mongodb.org/mongo-driver/mongo"

//     "RAAS/internal/models"
// )

// type GoogleUserRepo struct {
//     coll *mongo.Collection
// }

// func NewGoogleUserRepo(db *mongo.Database) *GoogleUserRepo {
//     return &GoogleUserRepo{coll: db.Collection("auth_users")}
// }

// func (r *GoogleUserRepo) GetOrCreateFromGoogle(email, name, picture, number string) (*models.AuthUser, error) {
//     ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//     defer cancel()

//     // 1️⃣ Look up existing user
//     var user models.AuthUser
//     err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&user)
//     if err != nil && err != mongo.ErrNoDocuments {
//         return nil, err
//     }

//     now := time.Now()
//     if err == mongo.ErrNoDocuments {
//         // 🆕 New user: create auth_user
//         authUserID := uuid.NewString()
//         user = models.AuthUser{
//             AuthUserID:    authUserID,
//             Email:         email,
//             Phone:         number,
//             Role:          "seeker",
//             EmailVerified: true,
//             IsActive:      true,
//             CreatedAt:     &now,
//             UpdatedAt:     &now,
//         }
//         if _, err := r.coll.InsertOne(ctx, user); err != nil {
//             return nil, err
//         }

//         // 🛠 Create associated seeker profile
//         seeker := models.Seeker{
//             AuthUserID:                  authUserID,
//             SubscriptionTier:            "free",
//             DailySelectableJobsCount:    10,
//             DailyGeneratableCV:          100,
//             DailyGeneratableCoverletter: 100,
//             TotalApplications:           0,
//             PersonalInfo:                bson.M{},
//             WorkExperiences:             []bson.M{},
//             Academics:                   []bson.M{},
//             PastProjects:                []bson.M{},
//             Certificates:                []bson.M{},
//             Languages:                   []bson.M{},
//             KeySkills:                   []string{},
//             PrimaryTitle:                "",
//         }
//         if _, err := r.coll.Database().
//             Collection("seekers").
//             InsertOne(ctx, seeker); err != nil {
//             return nil, err
//         }

//         // 📅 Create onboarding timeline
//         timeline := models.UserEntryTimeline{
//             AuthUserID:               authUserID,
//             PersonalInfoRequired:     true,
//             AcademicsRequired:        true,
//             LanguagesRequired:        true,
//             JobTitlesRequired:        true,
//             KeySkillsRequired:        true,
//             PersonalInfoCompleted:    false,
//             AcademicsCompleted:       false,
//             LanguagesCompleted:       false,
//             JobTitlesCompleted:       false,
//             KeySkillsCompleted:       false,
//             WorkExperiencesCompleted: false,
//             PastProjectsCompleted:    false,
//             CertificatesCompleted:    false,
//             Completed:                false,
//             CreatedAt:                now,
//             UpdatedAt:                now,
//         }
//         if _, err := r.coll.Database().
//             Collection("user_entry_timelines").
//             InsertOne(ctx, timeline); err != nil {
//             return nil, err
//         }

//         return &user, nil
//     }

//     // 2️⃣ Existing user: update profile info
//     _, err = r.coll.UpdateOne(ctx, bson.M{"email": email},
//         bson.M{"$set": bson.M{"name": name, "picture_url": picture, "updated_at": now}})
//     if err != nil {
//         return nil, err
//     }

//     // Refresh user struct
//     if err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
//         return nil, err
//     }
//     return &user, nil
// }
