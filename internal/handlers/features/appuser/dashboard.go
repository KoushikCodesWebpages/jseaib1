package appuser

import (
    "context"
    "log"
    "net/http"
    "time"
    "fmt"

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

// GetDashboard only returns dashboard if profile setup is complete.
func (h *SeekerProfileHandler) GetDashboard(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	userID := c.MustGet("userID").(string)

	// Fetch seeker profile
	seeker := h.fetchSeeker(c, db, userID)
	if seeker == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"issue": "User profile data is missing.",
			"error": "seeker_not_found",
		})
		return
	}

	// ✅ Check UserEntryTimeline for completion
	var timeline models.UserEntryTimeline
	err := db.Collection("user_entry_timelines").FindOne(c, bson.M{
		"auth_user_id": userID,
	}).Decode(&timeline)

	if err != nil {
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

	// ✅ Build dashboard only if profile is complete
	resp := dto.DashboardResponse{
		InfoBlocks:              h.buildInfo(*seeker),
		Profile:                 h.buildFields(*seeker),
		Checklist:               h.buildChecklist(*seeker),
		MiniNewJobsResponse:     h.buildMiniJobs(db,userID),
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

func (h *SeekerProfileHandler) buildInfo(s models.Seeker) dto.InfoBlocks {
	return dto.InfoBlocks{
		AuthUserID:                 s.AuthUserID,

        TotalApplications:          s.TotalApplications,
        WeeklyAppliedJobs:          s.WeeklyAppliedJobs,	
		TopJobs:                    s.TopJobs,

		SubscriptionTier:           s.SubscriptionTier,
        SubscriptionPeriod:         s.SubscriptionPeriod,
		SubscriptionIntervalStart:  s.SubscriptionIntervalStart,
        SubscriptionIntervalEnd:    s.SubscriptionIntervalEnd,
        
		ExternalApplications:       s.ExternalApplications,
		InternalApplications:       s.InternalApplications,
		ProficiencyTest:            s.ProficiencyTest,
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
