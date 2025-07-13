	package repository

	// import (
	// 	"fmt"
	// 	"go.mongodb.org/mongo-driver/bson"
	// 	"go.mongodb.org/mongo-driver/mongo"
	// 	"go.mongodb.org/mongo-driver/mongo/options"
	// 	"strconv"
	// 	"github.com/gin-gonic/gin"
	// )

	// // Paginate handles pagination logic (counting documents, calculating links, etc.)
	// func Paginate(c *gin.Context, db *mongo.Database, filter bson.M, route string) (string, string, int64, error) {
	// 	page := c.DefaultQuery("page", "1")
	// 	limit := c.DefaultQuery("limit", "10")

	// 	// Convert page and limit to integers
	// 	pageInt, err := strconv.Atoi(page)
	// 	if err != nil || pageInt <= 0 {
	// 		pageInt = 1
	// 	}

	// 	limitInt, err := strconv.Atoi(limit)
	// 	if err != nil || limitInt <= 0 {
	// 		limitInt = 10
	// 	}

	// 	// Calculate the skip value (used for pagination in the query)
	// 	skip := (pageInt - 1) * limitInt

	// 	// Get total count of matching documents for pagination
	// 	totalCount, err := db.Collection("jobs").CountDocuments(c, filter) // You can use different collection if needed
	// 	if err != nil {
	// 		return "", "", 0, err
	// 	}

	// 	// Pagination links
	// 	var nextLink, prevLink string
	// 	pageInt64 := int64(pageInt)
	// 	limitInt64 := int64(limitInt)
	// 	if pageInt64*limitInt64 < totalCount {
	// 		nextLink = fmt.Sprintf("%s?page=%d&limit=%d", route, pageInt+1, limitInt)
	// 	}

	// 	if pageInt > 1 {
	// 		prevLink = fmt.Sprintf("%s?page=%d&limit=%d", route, pageInt-1, limitInt)
	// 	}

	// 	return nextLink, prevLink, totalCount, nil
	// }
