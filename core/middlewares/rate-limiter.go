package middleware

import (

    "time"
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/ulule/limiter/v3"
    memorystore "github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimiterMiddleware returns a Gin middleware that limits incoming requests.
// limit: maximum number of requests
// period: duration for the limit (e.g., time.Minute)
func RateLimiterMiddleware(limit int64, period time.Duration) gin.HandlerFunc {
    rate := limiter.Rate{
        Period: period,
        Limit:  limit,
    }

    store := memorystore.NewStore()
    instance := limiter.New(store, rate)

    return func(c *gin.Context) {
        context, err := instance.Get(c, c.ClientIP())
        if err != nil {
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                "issue": "Internal rate limit error",
                "error": "rate_limit_error",
            })
            return
        }

        // Set headers (optional)
        c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
        c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
        c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", context.Reset))

        if context.Reached {
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "issue": "Too many requests. Please try again later.",
                "error": "limit_exceeded",
            })
            return
        }

        c.Next()
    }
}

