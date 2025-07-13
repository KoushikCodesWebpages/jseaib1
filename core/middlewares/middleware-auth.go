// middleware/jwt_auth.go
package middleware

import (
    //"RAAS/config"
    "RAAS/core/security"
    "github.com/gin-gonic/gin"
    "net/http"
    "strings"
    "log"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // log.Println("AuthHeaderTypes:", config.Cfg.AuthHeaderTypes)
        // log.Println("JWTSecretKey:", config.Cfg.JWTSecretKey)

        
        authHeader := c.GetHeader("Authorization")
        //log.Println("Authorization Header:", authHeader)

        if authHeader == "" {
            log.Println("Error: Authorization header is missing")
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
            c.Abort()
            return
        }

        parts := strings.Split(authHeader, " ")
        //log.Println("Auth Header Parts:", parts)

        if len(parts) != 2 {
            log.Println("Error: Invalid authorization header format (parts length)")
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
            c.Abort()
            return
        }
        if !strings.EqualFold(parts[0], "Bearer") {
            log.Println("Error: Invalid authorization header type. Expected: Bearer Got:", parts[0])
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
            c.Abort()
            return
        }
        tokenStr := strings.TrimSpace(parts[1])
        //log.Println("Token String:", tokenStr)

        claims, err := security.ValidateJWT(tokenStr)
        if err != nil {
            log.Printf("Token verification error: %v", err)
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        //log.Println("Token is valid. Claims:", claims)

        // Store user info in context
        c.Set("userID", claims.UserID)
        c.Set("email", claims.Email)
        c.Set("role", claims.Role)

        c.Next()
    }
}