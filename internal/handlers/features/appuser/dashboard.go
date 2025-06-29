package appuser

import (
    "context"
    "log"
    "net/http"
    "time"

    "RAAS/internal/dto"
    "RAAS/internal/handlers/repository"
    "RAAS/internal/models"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type SeekerProfileHandler struct{}

func NewSeekerProfileHandler() *SeekerProfileHandler {
    return &SeekerProfileHandler{}
}

// Single dashboard endpoint using all parts individually
func (h *SeekerProfileHandler) GetDashboard(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    userID := c.MustGet("userID").(string)

    seeker := h.fetchSeeker(c, db, userID)
    if seeker == nil {
        c.JSON(http.StatusNotFound, gin.H{"issue":"User data is not present in the system","error": "seeker not found"})
        return
    }

resp := dto.DashboardResponse{
        InfoBlocks:              	h.buildInfo(*seeker),
        Profile:     				h.buildFields(*seeker),
        Checklist:         			h.buildChecklist(*seeker),
        MiniNewJobsResponse:     	h.buildMiniJobs(),
        MiniTestSummaryResponse: 	h.buildMiniTestSummary(),
    }

    c.JSON(http.StatusOK, gin.H{
        "info_block": resp.InfoBlocks,
        "profile":    resp.Profile,
        "checklist":  resp.Checklist,
        "new_jobs":   resp.MiniNewJobsResponse,
        "test_summary": resp.MiniTestSummaryResponse,

    })
}

// Fetch seeker document
func (h *SeekerProfileHandler) fetchSeeker(c *gin.Context, db *mongo.Database, userID string) *models.Seeker {
    var s models.Seeker
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := db.Collection("seekers").
        FindOne(ctx, bson.M{"auth_user_id": userID}).
        Decode(&s); err != nil {
        log.Printf("Error fetching seeker: %v", err)
        return nil
    }
    return &s
}

// Build the Info block
func (h *SeekerProfileHandler) buildInfo(s models.Seeker) dto.InfoBlocks{
    return dto.InfoBlocks{
        AuthUserID:                  s.AuthUserID,
        SubscriptionTier:            s.SubscriptionTier,
        DailySelectableJobsCount:    s.DailySelectableJobsCount,
        DailyGeneratableCV:          s.DailyGeneratableCV,
        DailyGeneratableCoverletter: s.DailyGeneratableCoverletter,
        TotalApplications:           s.TotalApplications,
        TotalJobsAvailable:          0,
    }
}

// Build the Profile fields
func (h *SeekerProfileHandler) buildFields(s models.Seeker) dto.Profile {
    return dto.Profile{
        FirstName:         repository.DereferenceString(repository.GetOptionalField(s.PersonalInfo, "first_name")),
        SecondName:        repository.GetOptionalField(s.PersonalInfo, "second_name"),
        ProfileCompletion: repository.CalculateProfileCompletion(s),
        PrimaryJobTitle:   s.PrimaryTitle,
        SecondaryJobTitle: ptrVal(s.SecondaryTitle),
        TertiaryJobTitle:  ptrVal(s.TertiaryTitle),
    }
}

// Build the Checklist
func (h *SeekerProfileHandler) buildChecklist(s models.Seeker) dto.Checklist {
    return dto.Checklist{
        ChecklistMultifactorAuth: false,
        ChecklistCVFormatFixed: false,
        ChecklistCLFormatFixed: false,
        ChecklistProfileImg: false,
        ChecklistDataUsage: false,
        ChecklistDataTraining: false,
        ChecklistNumberLock: false,
        ChecklistDataFinalization: false,
        ChecklistTerms: false,
        ChecklistComplete: false,
    }
}

// Static new job list
func (h *SeekerProfileHandler) buildMiniJobs() dto.MiniNewJobsResponse {
    return dto.MiniNewJobsResponse{
        MiniNewJobs: []dto.MiniJob{
            {Title: "Backend Engineer", Company: "TechCorp", Location: "Berlin, Germany", ProfileMatch: 85},
            {Title: "Frontend Developer", Company: "WebWorks", Location: "Bavaria, Germany", ProfileMatch: 78},
            {Title: "DevOps Engineer", Company: "CloudSync Ltd.", Location: "Baden-Württemberg, Germany", ProfileMatch: 92},
        },
    }
}

// Static test summary
func (h *SeekerProfileHandler) buildMiniTestSummary() dto.MiniTestSummaryResponse {
    return dto.MiniTestSummaryResponse{
        Tests: []dto.MiniTest{
            {Languages: "German", RemainingAttempts: 2, Grade: 78.5, ProficiencyLevel: "Intermediate"},
            {Languages: "English", RemainingAttempts: 1, Grade: 92.0, ProficiencyLevel: "Advanced"},
        },
    }
}

func ptrVal(s *string) string {
    if s != nil {
        return *s
    }
    return ""
}
