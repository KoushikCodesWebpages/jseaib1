package exam

import (

	"RAAS/internal/models"
	"RAAS/internal/dto"

	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


func (h *ExamPortalHandler) ProcessExamResults(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	userID := c.MustGet("userID").(string)

	var req dto.ExamResultRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Step 1: Collect question IDs
	var questionIDs []string
	for _, a := range req.Answers {
		questionIDs = append(questionIDs, a.QuestionID)
	}

	// Step 2: Fetch questions from DB
	filter := bson.M{"question_id": bson.M{"$in": questionIDs}}
	cursor, err := db.Collection("questions").Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch questions", "details": err.Error()})
		return
	}
	var questions []models.Question
	if err := cursor.All(context.TODO(), &questions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode questions", "details": err.Error()})
		return
	}

	// Step 3: Build lookup maps
	questionMap := make(map[string]models.Question)
	for _, q := range questions {
		questionMap[q.QuestionID] = q
	}
	answerMap := make(map[string]dto.AnswerSubmissionDTO)
	for _, a := range req.Answers {
		answerMap[a.QuestionID] = a
	}

	// Step 4: Scoring
	totalMarks := 0.0
	attempted, correct, wrong, skipped := 0, 0, 0, 0
	negativeMarks := 0.0

	for _, q := range questions {
		totalMarks += q.Marks
		answer, answered := answerMap[q.QuestionID]

		if !answered || answer.OptionID == "" {
			skipped++
			continue
		}

		attempted++
		if len(q.CorrectOptionIDs) == 1 && answer.OptionID == q.CorrectOptionIDs[0] {
			correct++
		} else {
			wrong++
			negativeMarks += q.NegativeMark
		}
	}


	// Step 5: Score and grade
	score := float64(correct)
	finalScore := score - negativeMarks
	if finalScore < 0 {
		finalScore = 0
	}
	percentage := 0.0
	if totalMarks > 0 {
		percentage = (finalScore / totalMarks) * 100
	}

	grade := "F"
	switch {
	case percentage >= 90:
		grade = "A+"
	case percentage >= 75:
		grade = "A"
	case percentage >= 60:
		grade = "B"
	case percentage >= 50:
		grade = "C"
	case percentage >= 40:
		grade = "D"
	}

	// Step 6: Time & duration
	startTime, _ := time.Parse(time.RFC3339, req.StartsAt)
	endTime, _ := time.Parse(time.RFC3339, req.EndsAt)
	duration := int(endTime.Sub(startTime).Minutes())

	// Step 7: Build final result
	result := models.ExamResult{
		AuthUserID:      userID,
		ExamID:          req.ExamID,
		Title:           req.Title,
		Language:        questions[0].Language,
		Level:           questions[0].Difficulty,
		Score:           finalScore,
		Grade:           grade,
		TotalMarks:      totalMarks,
		Percentage:      percentage,
		Attempted:       attempted,
		Correct:         correct,
		Wrong:           wrong,
		Skipped:         skipped,
		NegativeMarks:   negativeMarks,
		SubmittedAt:     endTime,
		DurationMinutes: duration,
		IsPass:          percentage >= 40,
	}

	// Step 8: Save to DB
	_, err = db.Collection("exam_results").InsertOne(context.TODO(), result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store exam result", "details": err.Error()})
		return
	}

	// Step 9: Return result
	c.JSON(http.StatusOK, result)
}

func (h *ExamPortalHandler) GetFilteredExamResults(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)

	examID := c.Query("exam_id")
	authUserID := c.Query("auth_user_id")

	// Build filter
	filter := bson.M{}
	if examID != "" {
		filter["exam_id"] = examID
	}
	if authUserID != "" {
		filter["auth_user_id"] = authUserID
	}

	// Query MongoDB
	cursor, err := db.Collection("exam_results").Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch exam results", "details": err.Error()})
		return
	}
	defer cursor.Close(context.TODO())

	var results []models.ExamResult
	if err := cursor.All(context.TODO(), &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode results", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}


func (h *ExamPortalHandler) GetRecentExamResults(c *gin.Context) {
	authUserID := c.MustGet("userID").(string)
	if authUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"issue": "auth_user_id is required",
			"error": "missing_auth_user_id",
		})
		return
	}

	db := c.MustGet("db").(*mongo.Database)
	collection := db.Collection("exam_results")

	// Build filter and sort
	filter := bson.M{"auth_user_id": authUserID}
	opts := options.Find().
		SetSort(bson.D{{Key: "submitted_at", Value: -1}}).
		SetLimit(5)

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"issue": "Failed to fetch exam results",
			"error": err.Error(),
		})
		return
	}
	defer cursor.Close(context.TODO())

	var results []models.ExamResult
	if err := cursor.All(context.TODO(), &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"issue": "Failed to decode results",
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
	})
}
