package appuser
import (
    "context"
    "log"
    "net/http"
    "time"
    "strings"
    "fmt"
    "math"
    "RAAS/internal/dto"
    "RAAS/internal/handlers/repository"
    "RAAS/internal/models"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)
type SeekerProfileHandler struct{}
func NewSeekerProfileHandler() *SeekerProfileHandler {
    return &SeekerProfileHandler{}
}
func (h *SeekerProfileHandler) GetDashboard(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    userID := c.MustGet("userID").(string)
    var timeline models.UserEntryTimeline
    if err := db.Collection("user_entry_timelines").
        FindOne(c, bson.M{"auth_user_id": userID}).
        Decode(&timeline); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "issue":   "Could not retrieve user progress timeline.",
            "error":   "timeline_fetch_failed",
            "details": err.Error(),
        })
        return
    }
    if !timeline.Completed {
        c.JSON(http.StatusForbidden, gin.H{
            "issue": "Complete your profile setup to access the dashboard.",
            "error": "profile_incomplete",
        })
        return
    }
    seeker := h.fetchSeeker(c, db, userID)
    if seeker == nil {
        c.JSON(http.StatusNotFound, gin.H{
            "issue": "User profile data is missing.",
            "error": "seeker_not_found",
        })
        return
    }
    resp := dto.DashboardResponse{
        InfoBlocks:              h.buildInfo(*seeker, db),
        Profile:                 h.buildFields(*seeker),
        Checklist:               h.buildChecklist(c, db, userID),
        MiniNewJobsResponse:     h.buildMiniJobs(db, userID),
        MiniTestSummaryResponse: h.buildMiniTestSummary(),
    }
    c.JSON(http.StatusOK, gin.H{
        "info_block":   resp.InfoBlocks,
        "profile":      resp.Profile,
        "checklist":    resp.Checklist,
        "new_jobs":     resp.MiniNewJobsResponse,
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



func (h *SeekerProfileHandler) buildInfo(s models.Seeker, db *mongo.Database) dto.InfoBlocks {
    matchColl := db.Collection("match_scores")
    // :one: Calculate the 2-week cutoff date
    twoWeeksAgo := time.Now().Add(-14 * 24 * time.Hour)
    // :two: Build the filter for the last 2 weeks
    filter := bson.M{
        "auth_user_id": s.AuthUserID,
        "created_at":   bson.M{"$gte": twoWeeksAgo},
        "match_score":  bson.M{"$gt": 80},
    }
    // :three: Count matches in that period
    topJobsCount, _ := matchColl.CountDocuments(context.TODO(), filter)
    return dto.InfoBlocks{
        AuthUserID:                 s.AuthUserID,
        TotalApplications:          s.TotalApplications,
        WeeklyAppliedJobs:          s.WeeklyAppliedJobs,
        TopJobs:                    int(topJobsCount),
        SubscriptionTier:           s.SubscriptionTier,
        SubscriptionPeriod:         s.SubscriptionPeriod,
        SubscriptionIntervalStart:  s.SubscriptionIntervalStart,
        SubscriptionIntervalEnd:    s.SubscriptionIntervalEnd,
        ExternalApplications:       s.ExternalApplications,
        InternalApplications:       s.InternalApplications,
        ProficiencyTest:            s.ProficiencyTest,
    }
}


type Limits struct {
    Internal int
    External int
    Tests    int
}

type UsageStats struct {
    Used    int
    Total   int
    Percent int
}

func calculateUsage(used, total int) UsageStats {
    if total == 0 {
        return UsageStats{Used: used, Total: total, Percent: 100}
    }

    if used > total {
        return UsageStats{Used: used, Total: total, Percent: 100}
    }

    remaining := total - used
    percent := int(math.Round(float64(remaining) / float64(total) * 100))
    return UsageStats{Used: used, Total: total, Percent: percent}
}


func getLimits(tier, period string) Limits {
    key := strings.ToLower(tier) + "_" + strings.ToLower(period)
    switch key {
    case "basic_monthly":
        return Limits{150, 25, 3}
    case "advanced_monthly":
        return Limits{240, 35, 5}
    case "premium_monthly":
        return Limits{360, 75, 10}
    case "basic_quarterly":
        return Limits{450, 75, 9}
    case "advanced_quarterly":
        return Limits{720, 105, 15}
    case "premium_quarterly":
        return Limits{1080, 225, 30}
    default: // Free or unrecognized
        return Limits{5, 2, 0}
    }
}

func (h *SeekerProfileHandler) newbuildInfo(s models.Seeker, db *mongo.Database) dto.NewInfoBlocks {
    matchColl := db.Collection("match_scores")

    // Calculate the 2-week cutoff date
    twoWeeksAgo := time.Now().Add(-14 * 24 * time.Hour)

    // Build the filter for high score matches in the last 2 weeks
    filter := bson.M{
        "auth_user_id": s.AuthUserID,
        "created_at":   bson.M{"$gte": twoWeeksAgo},
        "match_score":  bson.M{"$gt": 80},
    }

    // Count matches in that period
    topJobsCount, _ := matchColl.CountDocuments(context.TODO(), filter)

// Get limits based on subscription
limits := getLimits(s.SubscriptionTier, s.SubscriptionPeriod)
// cclear

// Calculate usage
internalUsage := calculateUsage(s.InternalApplications, limits.Internal)
externalUsage := calculateUsage(s.ExternalApplications, limits.External)
testUsage := calculateUsage(s.ProficiencyTest, limits.Tests)

// log.Printf("[DEBUG] Internal used: %d/%d => %d%% left", s.InternalApplications, limits.Internal, internalUsage.Percent)
// log.Printf("[DEBUG] External used: %d/%d => %d%% left", s.ExternalApplications, limits.External, externalUsage.Percent)
// log.Printf("[DEBUG] Tests used: %d/%d => %d%% left", s.ProficiencyTest, limits.Tests, testUsage.Percent)


    return dto.NewInfoBlocks{
        AuthUserID:                s.AuthUserID,
        TotalApplications:         s.TotalApplications,
        WeeklyAppliedJobs:         s.WeeklyAppliedJobs,
        TopJobs:                   int(topJobsCount),
        SubscriptionTier:          s.SubscriptionTier,
        SubscriptionPeriod:        s.SubscriptionPeriod,
        SubscriptionIntervalStart: s.SubscriptionIntervalStart,
        SubscriptionIntervalEnd:   s.SubscriptionIntervalEnd,
        InternalApplications:      s.InternalApplications,
        ExternalApplications:      s.ExternalApplications,
        ProficiencyTest:           s.ProficiencyTest,

        InternalLimit:             limits.Internal,
        ExternalLimit:             limits.External,
        TestLimit:                 limits.Tests,
        InternalRemainingPercent:  internalUsage.Percent,
        ExternalRemainingPercent:  externalUsage.Percent,
        TestRemainingPercent:      testUsage.Percent,
    }
}

func (h *SeekerProfileHandler) buildFields(s models.Seeker) dto.Profile {
    // Calculate profile completion
    completion, _ := repository.CalculateJobProfileCompletion(s)
    return dto.Profile{
        PhotoUrl:           s.PhotoUrl,
        FirstName:          repository.DereferenceString(repository.GetOptionalField(s.PersonalInfo, "first_name")),
        SecondName:         repository.GetOptionalField(s.PersonalInfo, "second_name"),
        ProfileCompletion:  completion,
        PrimaryJobTitle:    s.PrimaryTitle,
        SecondaryJobTitle:  ptrVal(s.SecondaryTitle),
        TertiaryJobTitle:   ptrVal(s.TertiaryTitle),
    }
}
func (h *SeekerProfileHandler) buildChecklist(c context.Context, db *mongo.Database, userID string) (dto.Checklist) {
    // :one: Check if a ProfilePic exists for the user
    profilePicColl := db.Collection("profile_pic")
    count, err := profilePicColl.CountDocuments(c, bson.M{"auth_user_id": userID})
    if err != nil {
        return dto.Checklist{}
    }
    hasProfilePic := count > 0
    // :two: Build the Checklist
    checklist := dto.Checklist{
        ChecklistMultifactorAuth:  true,                // or your existing logic
        ChecklistCVFormatFixed:     true,
        ChecklistCLFormatFixed:     true,
        ChecklistProfileImg:        hasProfilePic,       // :white_check_mark: Set based on DB lookup
        ChecklistDataUsage:         true,
        ChecklistDataTraining:      true,
        ChecklistNumberLock:        true,
        ChecklistDataFinalization:  true,
        ChecklistTerms:             true,
        ChecklistComplete:          true,
    }
    return checklist
}
func (h *SeekerProfileHandler) buildMiniJobs(db *mongo.Database, userID string) dto.MiniNewJobsResponse {
    const maxMiniJobs = 3
    cutoffDate := time.Now().AddDate(0, 0, -14)
    // Step 1: Fetch top match scores for the user
    matchScoreFilter := bson.M{
        "auth_user_id": userID,
    }
    opts := options.Find().
        SetSort(bson.D{{Key: "match_score", Value: -1}}).
        SetLimit(10)
    cursor, err := db.Collection("match_scores").Find(context.TODO(), matchScoreFilter, opts)
    if err != nil {
        fmt.Println("Error fetching match scores:", err)
        return dto.MiniNewJobsResponse{}
    }
    defer cursor.Close(context.TODO())
    var topMatches []models.MatchScore
    if err := cursor.All(context.TODO(), &topMatches); err != nil {
        fmt.Println("Error decoding match scores:", err)
        return dto.MiniNewJobsResponse{}
    }
    // Step 2: Fetch applied job IDs
    appliedJobIDs, _ := repository.FetchAppliedJobIDs(context.TODO(), db.Collection("selected_job_applications"), userID)
    appliedMap := make(map[string]bool)
    for _, id := range appliedJobIDs {
        appliedMap[id] = true
    }
    // Step 3: Filter out applied jobs & fetch job data
    var miniJobs []dto.MiniJob
    for _, match := range topMatches {
        if appliedMap[match.JobID] {
            continue
        }
        // Fetch job
        var job models.Job
        err := db.Collection("jobs").FindOne(context.TODO(), bson.M{
            "job_id":       match.JobID,
            "posted_date": bson.M{"$gte": cutoffDate.Format("2006-01-02")},
        }).Decode(&job)
        if err != nil {
            continue
        }
        miniJobs = append(miniJobs, dto.MiniJob{
            Title:        job.Title,
            Company:      job.Company,
            Location:     job.Location,
            ProfileMatch: int(match.MatchScore),
        })
        if len(miniJobs) >= maxMiniJobs {
            break
        }
    }
    return dto.MiniNewJobsResponse{
        MiniNewJobs: miniJobs,
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






type DashboardV2Handler struct{}

func NewDashboardHandler() *DashboardV2Handler {
    return &DashboardV2Handler{}
}

func (h *DashboardV2Handler) GetStatus(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    userID := c.MustGet("userID").(string)
    complete, err := h.isTimelineComplete(c, db, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "status_check_failed",
            "details": err.Error(),
        })
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "profile_completed": complete,
        "user_id":           userID,
    })
}

func (h *DashboardV2Handler) isTimelineComplete(c *gin.Context, db *mongo.Database, userID string) (bool, error) {
    var timeline models.UserEntryTimeline
    err := db.Collection("user_entry_timelines").
        FindOne(c, bson.M{"auth_user_id": userID}).
        Decode(&timeline)
    if err != nil {
        return false, err
    }
    return timeline.Completed, nil
}

func (h *DashboardV2Handler) withTimelineCheck(c *gin.Context, handlerFunc func()) {
    db := c.MustGet("db").(*mongo.Database)
    userID := c.MustGet("userID").(string)
    complete, err := h.isTimelineComplete(c, db, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "timeline_check_failed",
            "details": err.Error(),
        })
        return
    }
    if !complete {
        c.JSON(http.StatusForbidden, gin.H{
            "error": "profile_incomplete",
            "issue": "Complete your profile setup to access this data.",
        })
        return
    }
    handlerFunc()
}

func (h *DashboardV2Handler) GetInfoBlock(c *gin.Context) {
    h.withTimelineCheck(c, func() {
        db := c.MustGet("db").(*mongo.Database)
        userID := c.MustGet("userID").(string)
        seeker := NewSeekerProfileHandler().fetchSeeker(c, db, userID)
        if seeker == nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "seeker_not_found"})
            return
        }
        info := NewSeekerProfileHandler().buildInfo(*seeker, db)
        c.JSON(http.StatusOK, gin.H{"info_block": info})
    })
}

func (h *DashboardV2Handler) GetNewInfoBlock(c *gin.Context) {
    h.withTimelineCheck(c, func() {
        db := c.MustGet("db").(*mongo.Database)
        userID := c.MustGet("userID").(string)
        seeker := NewSeekerProfileHandler().fetchSeeker(c, db, userID)
        if seeker == nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "seeker_not_found"})
            return
        }
        info := NewSeekerProfileHandler().newbuildInfo(*seeker, db)
        c.JSON(http.StatusOK, gin.H{"info_block": info})
    })
}





func (h *DashboardV2Handler) GetProfile(c *gin.Context) {
    h.withTimelineCheck(c, func() {
        db := c.MustGet("db").(*mongo.Database)
        userID := c.MustGet("userID").(string)
        seeker := NewSeekerProfileHandler().fetchSeeker(c, db, userID)
        if seeker == nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "seeker_not_found"})
            return
        }
        profile := NewSeekerProfileHandler().buildFields(*seeker)
        c.JSON(http.StatusOK, gin.H{"profile": profile})
    })
}

func (h *DashboardV2Handler) GetChecklist(c *gin.Context) {
    h.withTimelineCheck(c, func() {
        db := c.MustGet("db").(*mongo.Database)
        userID := c.MustGet("userID").(string)
        checklist := NewSeekerProfileHandler().buildChecklist(c, db, userID)
        c.JSON(http.StatusOK, gin.H{"checklist": checklist})
    })
}

func (h *DashboardV2Handler) GetMiniNewJobs(c *gin.Context) {
    h.withTimelineCheck(c, func() {
        db := c.MustGet("db").(*mongo.Database)
        userID := c.MustGet("userID").(string)
        jobs := NewSeekerProfileHandler().buildMiniJobs(db, userID)
        c.JSON(http.StatusOK, gin.H{"new_jobs": jobs})
    })
}

func (h *DashboardV2Handler) GetMiniTestSummary(c *gin.Context) {
    h.withTimelineCheck(c, func() {
        summary := NewSeekerProfileHandler().buildMiniTestSummary()
        c.JSON(http.StatusOK, gin.H{"test_summary": summary})
    })
}

