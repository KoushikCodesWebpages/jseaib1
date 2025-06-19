package middleware

import (

    "time"

    "github.com/gin-gonic/gin"
    "github.com/ulule/limiter/v3"
    ginmiddleware "github.com/ulule/limiter/v3/drivers/middleware/gin"
    memorystore "github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimiterMiddleware returns a Gin middleware that limits incoming requests.
// limit: maximum number of requests
// period: duration for the limit (e.g., time.Minute)
func RateLimiterMiddleware(limit int64, period time.Duration) gin.HandlerFunc {
    // Define the rate limit
    rate := limiter.Rate{
        Period: period,
        Limit:  limit,
    }

    // Create a new in-memory store
    store := memorystore.NewStore()

    // Create a new limiter instance
    instance := limiter.New(store, rate)

    // Create the Gin middleware
    return ginmiddleware.NewMiddleware(instance)
}
