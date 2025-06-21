package dto

type InfoBlocks struct {
    AuthUserID                  string      `json:"auth_user_id"`
    SubscriptionTier            string      `json:"subscription_tier"`
    DailySelectableJobsCount    int         `json:"daily_selectable_jobs_count"`
    DailyGeneratableCV          int         `json:"daily_generatable_cv"`
    DailyGeneratableCoverletter int         `json:"daily_generatable_coverletter"`
    TotalApplications           int         `json:"total_applications"`
    TotalJobsAvailable          int         `json:"total_jobs_available"`
}

type Profile struct {
    FirstName           string           `json:"first_name"`
    SecondName          *string          `json:"second_name,omitempty"`
    ProfileCompletion   int              `json:"profile_completion"`
    PrimaryJobTitle     string           `json:"primary_job_title"`
    SecondaryJobTitle   string           `json:"secondary_job_title"`
    TertiaryJobTitle    string           `json:"tertiary_job_title"`
}

type Checklist struct {
    
    ChecklistMultifactorAuth             bool    `json:"checklist_mfa"`   
    ChecklistCVFormatFixed               bool    `json:"checklist_cv_format_fixed"`
    ChecklistCLFormatFixed               bool    `json:"checklist_cl_format_fixed"`
    ChecklistProfileImg                  bool    `json:"checklist_profile_img"`
    ChecklistDataUsage                   bool    `json:"checklist_data_usage"`
    ChecklistDataTraining                bool    `json:"checklist_data_training"`
    ChecklistNumberLock                  bool    `json:"checklist_number_lock"`
    ChecklistDataFinalization            bool    `json:"checklist_data_finalization"`
    ChecklistTerms                       bool    `json:"checklist_terms"`
    
    ChecklistComplete           bool    `json:"checklist_complete"`
}

type MiniJob struct {
    Title               string      `json:"title"`
    Company             string      `json:"company"`
    Location            string      `json:"location"`
    ProfileMatch        int         `json:"profile_match"`
}

type MiniTest struct {
    Languages           string      `json:"languages"`
    RemainingAttempts   int         `json:"remaining_attempts"`
    Grade               float64     `json:"grade"`
    ProficiencyLevel    string      `json:"proficiency_level"`
}

type MiniNewJobsResponse struct {
    MiniNewJobs []MiniJob `json:"mini_new_jobs"`
}


type MiniTestSummaryResponse struct {
    Tests               []MiniTest  `json:"tests"`
}



type DashboardResponse struct {
    InfoBlocks                  `json:",inline"`
    Profile                     `json:",inline"`
    Checklist                   `json:",inline"`
    MiniNewJobsResponse         `json:"new_jobs"`
    MiniTestSummaryResponse     `json:"test_summary"`
}
