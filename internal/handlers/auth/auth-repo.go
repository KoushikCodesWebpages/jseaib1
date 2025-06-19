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

func (r *UserRepo) CheckDuplicateEmailOrPhone(email, phone string) (bool, bool, error) {
	var user models.AuthUser

	filter := bson.M{
		"$or": []bson.M{
			{"email": email},
			{"phone": phone},
		},
	}

	err := r.DB.Collection("auth_users").FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, false, nil
		}
		return false, false, fmt.Errorf("failed to check email or phone: %w", err)
	}

	emailExists := user.Email == email
	phoneExists := user.Phone == phone

	return emailExists, phoneExists, nil
}

func (r *UserRepo) CreateSeeker(input dto.SeekerSignUpInput, hashedPassword string) error {
	// Check for duplicate email or phone
	emailTaken, phoneTaken, err := r.CheckDuplicateEmailOrPhone(input.Email, input.Number)
	if err != nil {
		return fmt.Errorf("error checking for duplicates: %w", err)
	}
	if emailTaken {
		return fmt.Errorf("email is already taken")
	}
	if phoneTaken {
		return fmt.Errorf("phone number is already taken")
	}

	authUserID := uuid.New().String()
	token := uuid.New().String()

	// Create AuthUser
	now := time.Now()

	authUser := models.AuthUser{
		AuthUserID:        	authUserID,
		Email:             	input.Email,
		Password:          	hashedPassword,
		Phone:             	input.Number,
		Role:              	"seeker",
		EmailVerified:     	false,
		VerificationToken: 	token,
		IsActive:         	true,
		CreatedBy:         	authUserID,
		UpdatedBy:         	authUserID,
		CreatedAt:			&now,
		UpdatedAt:         	&now,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert AuthUser
	_, err = r.DB.Collection("auth_users").InsertOne(ctx, authUser)
	if err != nil {
		return fmt.Errorf("failed to create auth user: %w", err)
	}

	// Create Seeker
	seeker := models.Seeker{
		AuthUserID:                  	authUserID,
		SubscriptionTier:            	"free",
		DailySelectableJobsCount:    	10,
		DailyGeneratableCV:          	100,
		DailyGeneratableCoverletter: 	100,
		TotalApplications:           	0,

		PersonalInfo:                	bson.M{},
		WorkExperiences:             	[]bson.M{},
		Academics:                  	[]bson.M{},
		PastProjects:         			[]bson.M{},
		Certificates:               	[]bson.M{},
		Languages:                 		[]bson.M{},
		KeySkills: 						[]string{}  ,
		PrimaryTitle:                	"",
		SecondaryTitle:              	nil,
		TertiaryTitle:               	nil,
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
		return nil, errors.New("user not found")
	} else if err != nil {
		return nil, err
	}

	if !user.EmailVerified {
		return nil, errors.New("email not verified")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("incorrect password")
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
