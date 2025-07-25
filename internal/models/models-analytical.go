package models

import "time"

type UserAnalytics struct {
	// Identity
	AuthUserID           string     `json:"auth_user_id" bson:"auth_user_id"`
	Email                string     `json:"email" bson:"email"`

	// Login & Sessions
	LoginCount           int        `json:"login_count" bson:"login_count"`
	LastLoginAt          *time.Time `json:"last_login_at,omitempty" bson:"last_login_at,omitempty"`
	LastLoginDevice      *string    `json:"last_login_device,omitempty" bson:"last_login_device,omitempty"`
	LastLoginLocation    *string    `json:"last_login_location,omitempty" bson:"last_login_location,omitempty"`

	TotalSessionTime     int64      `json:"total_session_time" bson:"total_session_time"` // in seconds
	LastSessionTime      int64      `json:"last_session_time" bson:"last_session_time"`   // in seconds
	SessionCount         int        `json:"session_count" bson:"session_count"`

	// Device & Platform
	DevicesUsed          []string   `json:"devices_used,omitempty" bson:"devices_used,omitempty"`         // e.g. ["Chrome", "Android App"]
	OSUsed               []string   `json:"os_used,omitempty" bson:"os_used,omitempty"`                   // e.g. ["Windows", "iOS"]
	AppVersions          []string   `json:"app_versions,omitempty" bson:"app_versions,omitempty"`         // e.g. ["1.0.3", "1.1.0"]
	BrowserUsed          []string   `json:"browser_used,omitempty" bson:"browser_used,omitempty"`

	// Geo
	CountriesUsed        []string   `json:"countries_used,omitempty" bson:"countries_used,omitempty"`     // e.g. ["IN", "US"]
	CityUsed             []string   `json:"city_used,omitempty" bson:"city_used,omitempty"`

	// Feature Usage
	JobSearchCount       int        `json:"job_search_count" bson:"job_search_count"`
	ApplicationsSent     int        `json:"applications_sent" bson:"applications_sent"`
	ResumeUploads        int        `json:"resume_uploads" bson:"resume_uploads"`
	CoverLettersWritten  int        `json:"cover_letters_written" bson:"cover_letters_written"`
	SavedJobs            int        `json:"saved_jobs" bson:"saved_jobs"`
	ProfileViews         int        `json:"profile_views" bson:"profile_views"`
	RecommendationsViewed int       `json:"recommendations_viewed" bson:"recommendations_viewed"`

	// Engagement & Behavior
	SearchQueries        []string   `json:"search_queries,omitempty" bson:"search_queries,omitempty"`
	AvgSessionDuration   float64    `json:"avg_session_duration" bson:"avg_session_duration"` // in seconds
	BounceRate           float64    `json:"bounce_rate,omitempty" bson:"bounce_rate,omitempty"` // 0.0–1.0
	OnboardingCompleted  bool       `json:"onboarding_completed" bson:"onboarding_completed"`
	FirstActionTaken     *string    `json:"first_action_taken,omitempty" bson:"first_action_taken,omitempty"`

	// Marketing Attribution
	SignupSource         *string    `json:"signup_source,omitempty" bson:"signup_source,omitempty"`     // e.g. "Google Ads", "LinkedIn"
	Referrer             *string    `json:"referrer,omitempty" bson:"referrer,omitempty"`
	CampaignID           *string    `json:"campaign_id,omitempty" bson:"campaign_id,omitempty"`
	UTMSource            *string    `json:"utm_source,omitempty" bson:"utm_source,omitempty"`
	UTMMedium            *string    `json:"utm_medium,omitempty" bson:"utm_medium,omitempty"`
	UTMCampaign          *string    `json:"utm_campaign,omitempty" bson:"utm_campaign,omitempty"`

	// Flags
	ABGroup              *string    `json:"ab_group,omitempty" bson:"ab_group,omitempty"`                 // e.g. "A", "B"
	FeedbackSubmitted    bool       `json:"feedback_submitted" bson:"feedback_submitted"`
	HasSubscription      bool       `json:"has_subscription" bson:"has_subscription"`
	NotificationOptIn    bool       `json:"notification_opt_in" bson:"notification_opt_in"`

	// Temporal
	FirstSeenAt          *time.Time `json:"first_seen_at" bson:"first_seen_at,omitempty"`
	LastSeenAt           *time.Time `json:"last_seen_at" bson:"last_seen_at,omitempty"`
	CreatedAt            *time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt            *time.Time `json:"updated_at" bson:"updated_at"`
}
