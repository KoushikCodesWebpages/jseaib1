package exam

import (
    "context"
    "net/http"
    "time"
	"strconv"

	"RAAS/internal/models"  // Update with your module path
    "RAAS/internal/dto"     // Place your DTOs here

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    // "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

)

type QuestionsHandler struct{}

func NewQuestionsHandler() *QuestionsHandler {
	return &QuestionsHandler{}
}

func (h *QuestionsHandler) PostQuestion(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	qcoll:= db.Collection("questions")

	var body dto.CreateQuestionDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"issue": "invalid request","error":err.Error()})
		return
	}

	q := models.Question{
		QuestionID:       body.QuestionID,
		Type:             body.Type,
		Question:         body.Question,
		Options:          mapOptionDTOs(body.Options),
		Difficulty:       body.Difficulty,
		Language:         body.Language,
		RandomizeOptions: body.RandomizeOptions,
		Title:            body.Title,
		Description:      body.Description,
		CorrectOptionIDs: body.CorrectOptionIDs,
		AnswerKey:        body.AnswerKey,
		Marks:            body.Marks,
		NegativeMark:     body.NegativeMark,
		Tags:             body.Tags,
		Category:         body.Category,
		SubCategory:      body.SubCategory,
		Attachments:      body.Attachments,
		Explanation:      body.Explanation,
		IsActive:         body.IsActive,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	
	_, err := qcoll.InsertOne(context.Background(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Question Already Exists","error":err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"issue": "question created", "id": q.QuestionID})
}


func (h *QuestionsHandler) GetQuestions(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	collection := db.Collection("questions")

	// Pagination
	pagination := c.MustGet("pagination").(gin.H)
	offset := pagination["offset"].(int)
	limit := pagination["limit"].(int)

	// Filters
	filter := bson.M{}
	if qType := c.Query("type"); qType != "" {
		filter["type"] = qType
	}
	if diff := c.Query("difficulty"); diff != "" {
		filter["difficulty"] = diff
	}
	if lang := c.Query("language"); lang != "" {
		filter["language"] = bson.M{"$regex": lang, "$options": "i"}
	}

	// Count total
	total, err := collection.CountDocuments(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Error counting questions","error":err.Error()})
		return
	}

	if total == 0 {
		c.JSON(http.StatusOK, gin.H{
			"pagination": gin.H{
				"total":    0,
				"next":     "",
				"prev":     "",
				"current":  1,
				"per_page": limit,
			},
			"questions": []models.Question{},
		})
		return
	}

	// Find with sort and pagination
	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(c, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Error fetching questions","error":err.Error()})
		return
	}
	defer cursor.Close(c)

	var questions []models.Question
	if err := cursor.All(c, &questions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Error decoding questions","error":err.Error()})
		return
	}

	// Pagination URLs
	nextPage := ""
	prevPage := ""
	if offset+limit < int(total) {
		nextPage = "/b1/api/questions?offset=" + strconv.Itoa(offset+limit) + "&limit=" + strconv.Itoa(limit)
	}
	if offset > 0 {
		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		prevPage = "/b1/api/questions?offset=" + strconv.Itoa(prevOffset) + "&limit=" + strconv.Itoa(limit)
	}

	c.JSON(http.StatusOK, gin.H{
		"pagination": gin.H{
			"total":    total,
			"next":     nextPage,
			"prev":     prevPage,
			"current":  (offset / limit) + 1,
			"per_page": limit,
		},
		"questions": questions,
	})
}





func (h *QuestionsHandler) UpdateQuestion(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	questionID := c.Param("question_id")

	var updateDTO dto.UpdateQuestionDTO
	if err := c.ShouldBindJSON(&updateDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"issue": "Invalid request body","error":err.Error()})
		return
	}

	update := bson.M{}
	if updateDTO.Title != nil {
		update["title"] = *updateDTO.Title
	}
	if updateDTO.Description != nil {
		update["description"] = *updateDTO.Description
	}
	if updateDTO.Type != nil {
		update["type"] = *updateDTO.Type
	}
	if updateDTO.Question != nil {
		update["question"] = *updateDTO.Question
	}
	if updateDTO.Options != nil {
		opts := make([]models.Option, len(*updateDTO.Options))
		for i, o := range *updateDTO.Options {
			opts[i] = models.Option{
				OptionID:        o.OptionID,
				Text:      o.Text,
				Media:     o.Media,
				IsCorrect: o.IsCorrect,
			}
		}
		update["options"] = opts
	}
	if updateDTO.CorrectOptionIDs != nil {
		update["correct_option_ids"] = *updateDTO.CorrectOptionIDs
	}
	if updateDTO.AnswerKey != nil {
		update["answer_key"] = *updateDTO.AnswerKey
	}
	if updateDTO.Marks != nil {
		update["marks"] = *updateDTO.Marks
	}
	if updateDTO.NegativeMark != nil {
		update["negative_mark"] = *updateDTO.NegativeMark
	}
	if updateDTO.Difficulty != nil {
		update["difficulty"] = *updateDTO.Difficulty
	}
	if updateDTO.Language != nil {
		update["language"] = *updateDTO.Language
	}
	if updateDTO.Tags != nil {
		update["tags"] = *updateDTO.Tags
	}
	if updateDTO.Category != nil {
		update["category"] = *updateDTO.Category
	}
	if updateDTO.SubCategory != nil {
		update["sub_category"] = *updateDTO.SubCategory
	}
	if updateDTO.RandomizeOptions != nil {
		update["randomize_options"] = *updateDTO.RandomizeOptions
	}
	if updateDTO.Attachments != nil {
		update["attachments"] = *updateDTO.Attachments
	}
	if updateDTO.Explanation != nil {
		update["explanation"] = *updateDTO.Explanation
	}
	if updateDTO.IsActive != nil {
		update["is_active"] = *updateDTO.IsActive
	}

	update["updated_at"] = time.Now()

	result, err := db.Collection("questions").UpdateOne(
		c,
		bson.M{"question_id": questionID},
		bson.M{"$set": update},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to update question","error":err.Error()})
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"issue": "Question not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"issue": "Question updated successfully"})
}

func (h *QuestionsHandler) PatchQuestion(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	questionID := c.Param("question_id")

	var patchDTO dto.UpdateQuestionDTO
	if err := c.ShouldBindJSON(&patchDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"issue": "Invalid patch data","error":err.Error()})
		return
	}

	update := bson.M{}
	if patchDTO.Title != nil {
		update["title"] = *patchDTO.Title
	}
	if patchDTO.Description != nil {
		update["description"] = *patchDTO.Description
	}
	if patchDTO.Type != nil {
		update["type"] = *patchDTO.Type
	}
	if patchDTO.Question != nil {
		update["question"] = *patchDTO.Question
	}
	if patchDTO.Options != nil {
		opts := make([]models.Option, len(*patchDTO.Options))
		for i, o := range *patchDTO.Options {
			opts[i] = models.Option{
				OptionID:        o.OptionID,
				Text:      o.Text,
				Media:     o.Media,
				IsCorrect: o.IsCorrect,
			}
		}
		update["options"] = opts
	}
	if patchDTO.CorrectOptionIDs != nil {
		update["correct_option_ids"] = *patchDTO.CorrectOptionIDs
	}
	if patchDTO.AnswerKey != nil {
		update["answer_key"] = *patchDTO.AnswerKey
	}
	if patchDTO.Marks != nil {
		update["marks"] = *patchDTO.Marks
	}
	if patchDTO.NegativeMark != nil {
		update["negative_mark"] = *patchDTO.NegativeMark
	}
	if patchDTO.Difficulty != nil {
		update["difficulty"] = *patchDTO.Difficulty
	}
	if patchDTO.Language != nil {
		update["language"] = *patchDTO.Language
	}
	if patchDTO.Tags != nil {
		update["tags"] = *patchDTO.Tags
	}
	if patchDTO.Category != nil {
		update["category"] = *patchDTO.Category
	}
	if patchDTO.SubCategory != nil {
		update["sub_category"] = *patchDTO.SubCategory
	}
	if patchDTO.RandomizeOptions != nil {
		update["randomize_options"] = *patchDTO.RandomizeOptions
	}
	if patchDTO.Attachments != nil {
		update["attachments"] = *patchDTO.Attachments
	}
	if patchDTO.Explanation != nil {
		update["explanation"] = *patchDTO.Explanation
	}
	if patchDTO.IsActive != nil {
		update["is_active"] = *patchDTO.IsActive
	}

	if len(update) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"issue": "No fields provided for patch"})
		return
	}

	update["updated_at"] = time.Now()

	result, err := db.Collection("questions").UpdateOne(
		c,
		bson.M{"question_id": questionID},
		bson.M{"$set": update},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to patch question","error":err.Error()})
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"issue": "Question not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"issue": "Question patched successfully"})
}

func (h *QuestionsHandler) DeleteQuestion(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	questionID := c.Param("question_id")

	result, err := db.Collection("questions").DeleteOne(c, bson.M{"question_id": questionID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to delete question","error":err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"issue": "Question not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"issue": "Question deleted successfully"})
}
