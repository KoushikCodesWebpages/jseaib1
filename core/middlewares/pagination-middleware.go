package middleware

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

// PaginationMiddleware handles pagination logic using offset and limit
func PaginationMiddleware(c *gin.Context) {
	// Get the "offset" and "limit" query parameters
	offset := c.DefaultQuery("offset", "0") // Default to 0 if not provided
	limit := c.DefaultQuery("limit", "20")  // Default to 10 if not provided

	// Convert offset and limit to integers
	offsetInt, err := strconv.Atoi(offset)
	if err != nil || offsetInt < 0 {
		offsetInt = 0 // Default to offset 0 if invalid
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		limitInt = 10 // Default to limit of 10 if invalid
	}

	// Set a max limit to avoid fetching too many records at once
	const maxLimit = 100
	if limitInt > maxLimit {
		limitInt = maxLimit // Cap the limit to maxLimit if exceeded
	}

	// Store pagination information in context
	c.Set("pagination", gin.H{
		"offset": offsetInt,
		"limit":  limitInt,
	})

	// Proceed to the next handler
	c.Next()
}
