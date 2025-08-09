package exam

import (
	"RAAS/internal/dto"
	"RAAS/internal/handlers/repository"
	"RAAS/internal/models"

	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)
type ExamPortalHandler struct{}

func NewExamPortalHandler() *ExamPortalHandler {
	return &ExamPortalHandler{}
}
func (h *ExamPortalHandler) GenerateRandomExam(c *gin.Context) {
    var request dto.ExamDTORequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "issue": "invalid_input",
            "error": err.Error(),
        })
        return
    }

    db := c.MustGet("db").(*mongo.Database)

    // ✅ STEP 1: Get current user ID (assumes JWT middleware sets it in context)
    userID := c.MustGet("userID").(string)


    // ✅ STEP 2: Fetch seeker
    seeker, err := repository.GetSeekerData(db, userID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "issue": "seeker_not_found",
            "error": err.Error(),
        })
        return
    }

    // ✅ STEP 3: Access check
    if seeker.SubscriptionTier != "advanced" && seeker.ProficiencyTest <= 0 {
        c.JSON(http.StatusForbidden, gin.H{
            "issue": "upgrade_required",
            "error": "Access restricted. Upgrade to Advanced or complete Proficiency Test.",
        })
        return
    }

    // ✅ STEP 4: Proceed with generating exam
    collection := db.Collection("questions")

    filter := bson.M{"is_active": true}
    if request.Language != nil {
        filter["language"] = *request.Language
    }
    if request.Difficulty != nil {
        filter["difficulty"] = *request.Difficulty
    }

    limit := 10
    if request.TotalQuestions != nil {
        limit = *request.TotalQuestions
    }

    pipeline := mongo.Pipeline{
        {{Key: "$match", Value: filter}},
        {{Key: "$sample", Value: bson.M{"size": limit}}},
    }

    cursor, err := collection.Aggregate(context.TODO(), pipeline)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "issue": "db_aggregation_failed",
            "error": err.Error(),
        })
        return
    }
    defer cursor.Close(context.TODO())

    var questions []models.Question
    if err := cursor.All(context.TODO(), &questions); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "issue": "decode_failed",
            "error": err.Error(),
        })
        return
    }

    if len(questions) < limit {
        c.JSON(http.StatusBadRequest, gin.H{
            "issue": "not_enough_questions_found",
            "error": fmt.Sprintf("Only %d questions found but %d requested", len(questions), limit),
        })
        return
    }

    var questionDTOs []dto.ExamQuestionDTO
    var totalMarks float64
    for _, q := range questions {
        // Create a temporary question object to manipulate
        sanitizedQuestion := q
        
        for i := range sanitizedQuestion.Options {
            // Keep the original OptionID
            // sanitizedQuestion.Options[i].OptionID is already correct
            sanitizedQuestion.Options[i].IsCorrect = false
        }
        
        // Sanitize other sensitive fields
        sanitizedQuestion.CorrectOptionIDs = nil
        sanitizedQuestion.AnswerKey = ""
        sanitizedQuestion.Explanation = ""

        totalMarks += sanitizedQuestion.Marks
        questionDTOs = append(questionDTOs, ToQuestionDTO(sanitizedQuestion))
    }

    finalMarks := totalMarks
    if request.TotalMarks != nil {
        finalMarks = *request.TotalMarks
    }

    duration := 60
    if request.DurationMins != nil {
        duration = *request.DurationMins
    }

    now := time.Now()
    examResp := dto.ExamResponseDTO{
        ExamID:              uuid.NewString(),
        Title:               deref(request.Title, "Exam"),
        Description:         deref(request.Description, ""),
        Questions:           questionDTOs,
        DurationMinutes:     duration,
        TotalMarks:          finalMarks,
        AllowNegativeMark:   true,
        AttemptsAllowed:     1,
        IsPublic:            derefBool(request.IsPublic, false),
        StartsAt:            now.Format(time.RFC3339),
        CreatedAt:           now.Format(time.RFC3339),
        UpdatedAt:           now.Format(time.RFC3339),
    }

    if duration > 0 {
        examResp.EndsAt = now.Add(time.Duration(duration) * time.Minute).Format(time.RFC3339)
    }

    // ✅ Reduce ProficiencyTest count by 1
    _, err = db.Collection("seekers").UpdateOne(
        context.TODO(),
        bson.M{"auth_user_id": userID},
        bson.M{"$inc": bson.M{"proficiency_test": -1}},
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "issue": "failed_to_update_test_quota",
            "error": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, examResp)
}

