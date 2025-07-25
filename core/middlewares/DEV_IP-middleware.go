package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func IPWhitelistMiddleware(allowedIPs []string) gin.HandlerFunc {
	whitelist := make(map[string]bool)
	for _, ip := range allowedIPs {
		whitelist[ip] = true
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if !whitelist[clientIP] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":      "Access denied",
				"your_ip":    clientIP,
				"whitelisted": allowedIPs,
			})
			return
		}
		c.Next()
	}
}
