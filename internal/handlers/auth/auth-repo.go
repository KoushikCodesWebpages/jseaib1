package auth

import (


	"RAAS/core/config"
	"RAAS/internal/dto"
	"RAAS/internal/models"
	"RAAS/utils"

	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo struct {
	DB *mongo.Database
}

func NewUserRepo(db *mongo.Database) *UserRepo {
	return &UserRepo{
		DB: db,
	}
}

func (r *UserRepo) ValidateSeekerSignUpInput(input dto.SeekerSignUpInput) error {
	if input.Email == "" || input.Password == "" || input.Number == "" {
		return fmt.Errorf("all fields are required")
	}
	return nil
}

func (r *UserRepo) CheckDuplicateEmailWithDeleted(email string) (found bool, deleted bool, err error) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    var user models.AuthUser
    err = r.DB.Collection("auth_users").
        FindOne(ctx, bson.M{"email": email}).
        Decode(&user)

    if err != nil {
        if err == mongo.ErrNoDocuments {
            return false, false, nil
        }
        return false, false, err
    }

    if user.IsDeleted {
        return true, true, nil
    }
    return true, false, nil
}

func (r *UserRepo) CheckDuplicatePhoneWithDeleted(phone string) (found bool, deleted bool, err error) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    var user models.AuthUser
    err = r.DB.Collection("auth_users").
        FindOne(ctx, bson.M{"phone": phone}).
        Decode(&user)

    if err != nil {
        if err == mongo.ErrNoDocuments {
            return false, false, nil
        }
        return false, false, err
    }

    if user.IsDeleted {
        return true, true, nil
    }
    return true, false, nil
}


func (r *UserRepo) CreateSeeker(input dto.SeekerSignUpInput, hashedPassword string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	found, deleted, err := r.CheckDuplicateEmailWithDeleted(input.Email)
	if err != nil {
		return fmt.Errorf("failed checking email: %w", err)
	}
	if found {
		if deleted {
			return fmt.Errorf("email_recently_deleted")
		}
		return fmt.Errorf("email already in use")
	}

	found, deleted, err = r.CheckDuplicatePhoneWithDeleted(input.Number)
	if err != nil {
		return fmt.Errorf("failed checking phone: %w", err)
	}
	if found {
		if deleted {
			return fmt.Errorf("phone_recently_deleted")
		}
		return fmt.Errorf("phone already in use")
	}


	// Proceed with creation...
	authUserID := uuid.New().String()
	token := uuid.New().String()
	now := time.Now()
	authUser := models.AuthUser{
		AuthUserID:          authUserID,                  // generated unique ID
		Email:               input.Email,                 
		Phone:               input.Number,                
		Password:            hashedPassword,              
		Role:                "seeker",                    // or based on business logic
		EmailVerified:       false,                       
		Provider:            "",                          // if oauth, e.g. "google", else “”
		VerificationToken:   token,                       
		ResetTokenExpiry:    nil,                         // no reset token yet
		IsActive:            true,                        
		CreatedBy:           authUserID,                  // or "system" / admin ID
		UpdatedBy:           authUserID,
		CreatedAt:           &now,
		UpdatedAt:           &now,
		LastLoginAt:         nil,                         // will set on login
		PasswordLastUpdated: &now,                        // password now just set
		TwoFactorEnabled:    false,
		TwoFactorSecret:     nil,
		IsDeleted:           false,
		DeletedAt:           nil,
		SignupIP:            nil,                    // capture if available
		LastLoginIP:         nil,
	}
	// Insert AuthUser
	_, err = r.DB.Collection("auth_users").InsertOne(ctx, authUser)
	if err != nil {
		return fmt.Errorf("failed to create auth user: %w", err)
	}

		seeker := models.Seeker{
		AuthUserID:                  authUserID,
		PhotoUrl:                    "", // or set default if needed

		TotalApplications:           0,
		WeeklyAppliedJobs:           0,
		TopJobs:                     0,
		StripeCustomerID: 			"",
		SubscriptionTier:            "free",
		SubscriptionPeriod:          "monthly",
		SubscriptionIntervalStart:   time.Now(),
		SubscriptionIntervalEnd:     now.AddDate(0, 1, 0),

		InternalApplications:        5,
		ExternalApplications:        2,
		ProficiencyTest:            1,

		PersonalInfo:                bson.M{},
		WorkExperiences:             []bson.M{},
		Academics:                   []bson.M{},
		PastProjects:                []bson.M{},
		Certificates:                []bson.M{},
		Languages:                   []bson.M{},
		KeySkills:                   []string{},

		PrimaryTitle:                "",
		SecondaryTitle:              nil,
		TertiaryTitle:               nil,

		CvFormat: 					"modern_deedy",
		ClFormat: 					"cl_format_01",		
		CreatedAt:                   now,
		UpdatedAt:                   now,
	}
	
	_, err = r.DB.Collection("seekers").InsertOne(ctx, seeker)
	if err != nil {
		return fmt.Errorf("failed to create seeker profile: %w", err)
	}

	// Create Timeline
	timeline := models.UserEntryTimeline{
		AuthUserID: authUserID,

		// Compulsory steps
		PersonalInfoCompleted:   false,
		PersonalInfoRequired:    true,

		AcademicsCompleted:      false,
		AcademicsRequired:       true,

		LanguagesCompleted:      false,
		LanguagesRequired:       true,

		JobTitlesCompleted:      false,
		JobTitlesRequired:       true,

		KeySkillsCompleted:      false,
		KeySkillsRequired:       true,

		// Optional steps
		WorkExperiencesCompleted: false,
		WorkExperiencesRequired:  false,

		PastProjectsCompleted:    false,
		PastProjectsRequired:     false,

		CertificatesCompleted:    false,
		CertificatesRequired:     false,

		// Overall
		Completed: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}


	_, err = r.DB.Collection("user_entry_timelines").InsertOne(ctx, timeline)
	if err != nil {
		return fmt.Errorf("user created but failed to create entry timeline: %w", err)
	}


	userPreferences := models.UserPreferences{
		AuthUserID:   authUserID,
		Language:     "english",
		Timezone:     "CET",
		CookiePolicy: false,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	_, err = r.DB.Collection("preferences").InsertOne(ctx, userPreferences)
	if err != nil {
		return fmt.Errorf("user created but failed to create userpreferences: %w", err)
	}

	notifications := models.NotificationSettings{
		AuthUserID:      authUserID,
		Subscription:    false,
		RecommendedJobs: false,
		GermanTest:      false,
		Announcements:   false,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	_, err = r.DB.Collection("notifications").InsertOne(ctx, notifications)
	if err != nil {
		return fmt.Errorf("user created but failed to create notifications: %w", err)
	}



	// Prepare verification email
	verificationLink := fmt.Sprintf("%s/b1/auth/verify-email?token=%s", config.Cfg.Project.FrontendBaseUrl, token)
	emailBody := fmt.Sprintf(`
			<html>
			<body style="font-family: Arial, sans-serif; background-color: #f9f9f9; margin: 0; padding: 0;">
				<div style="max-width: 600px; margin: 40px auto; background: #ffffff; padding: 30px; border-radius: 10px; box-shadow: 0 2px 8px rgba(0,0,0,0.05);">
				<h2 style="color: #4CAF50; text-align: center;">Welcome to JSE AI!</h2>
				<p>Hi %s,</p>
				<p>Thanks for signing up! To get started, please confirm your email address by clicking the button below:</p>
				<div style="text-align: center; margin: 30px 0;">
					<a href="%s" style="background-color: #4CAF50; color: #ffffff; padding: 14px 24px; text-decoration: none; border-radius: 6px; font-weight: bold;">
					Verify Email
					</a>
				</div>
				<p>If you didn’t create this account, you can safely ignore this email.</p>
				<p>Cheers,<br><strong>The Team</strong></p>
				</div>
			</body>
			</html>
			`, input.Email, verificationLink)

	emailCfg := utils.EmailConfig{
		Host:     config.Cfg.Cloud.EmailHost,
		Port:     config.Cfg.Cloud.EmailPort,
		Username: config.Cfg.Cloud.EmailHostUser,
		Password: config.Cfg.Cloud.EmailHostPassword,
		From:     config.Cfg.Cloud.DefaultFromEmail,
		UseTLS:   config.Cfg.Cloud.EmailUseTLS,
	}

	if err := utils.SendEmail(emailCfg, input.Email, "Verify your email", emailBody); err != nil {
		return fmt.Errorf("user created but failed to send verification email: %w", err)
	}

	return nil
}


func (r *UserRepo) AuthenticateUser(ctx context.Context, email, password string) (*models.AuthUser, error) {
    var user models.AuthUser

    err := r.DB.Collection("auth_users").FindOne(ctx, bson.M{"email": email}).Decode(&user)
    if err == mongo.ErrNoDocuments {
        return nil, fmt.Errorf("user_not_found")
    } else if err != nil {
        return nil, fmt.Errorf("db_error: %v", err)
    }

    // ❗ Check for soft-deleted account
    if user.IsDeleted {
        return nil, fmt.Errorf("user_deleted")
    }

    if !user.EmailVerified {
        return nil, fmt.Errorf("email_not_verified")
    }

    // 🧪 Compare password using bcrypt
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
        return nil, fmt.Errorf("invalid_password")
    }

    return &user, nil
}

func (r *UserRepo) FindUserByEmail(ctx context.Context, email string) (*models.AuthUser, error) {
	var user models.AuthUser

	err := r.DB.Collection("auth_users").FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("user not found")
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}
