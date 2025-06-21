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
    ChecklistPersonalInfo       bool    `json:"checklist_personal_info"`
    ChecklistWorkExperience     bool    `json:"checklist_work_experience"`
    ChecklistAcademics          bool    `json:"checklist_academics"`
    ChecklistPastProjects       bool    `json:"checklist_past_projects"`
    ChecklistLanguages          bool    `json:"checklist_languages"`
    ChecklistCertifications     bool    `json:"checklist_certifications"`
    ChecklistJobTitles          bool    `json:"checklist_job_titles"`
    ChecklistKeySkills          bool    `json:"checklist_key_skills"`
    
    MultifactorAuth             bool    `json:"mfa"`   
    CVFormatFixed               bool    `json:"cv_format_fixed"`
    CLFormatFixed               bool    `json:"cl_format_fixed"`
    ProfileImg                  bool    `json:"profile_img"`
    DataUsage                   bool    `json:"data_usage"`
    DataTraining                bool    `json:"data_training"`
    NumberLock                  bool    `json:"number_lock"`
    DataFinalization            bool    `json:"data_finalization"`
    Terms                       bool    `json:"terms"`
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
