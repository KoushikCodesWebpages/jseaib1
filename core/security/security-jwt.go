package security

import (
	"RAAS/core/config"

	
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"strings"
	"time"
)

var jwtSecret []byte

// getJWTSecret loads the JWT secret key from the config (ensuring it's only loaded once).
func getJWTSecret() []byte {
	if jwtSecret == nil {
		jwtSecret = []byte(config.Cfg.Project.JWTSecretKey)
	}
	return jwtSecret
}

// CustomClaims struct defines the structure of the JWT claims, now with UserID as string.
type CustomClaims struct {
	UserID string `json:"user_id"` // Change UserID to string
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a signed JWT token using the user's ID (string), email, and role.
// The token expiration time is defined by the config value `AccessTokenLifetime`.
func GenerateJWT(userID string, email, role string) (string, error) {
	// Define the claims with an expiration time based on the config
	claims := CustomClaims{
		UserID: userID, // Use string for user ID
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(config.Cfg.Project.AccessTokenLifetime))),
		},
	}

	// Create the JWT token with claims and signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret()) // Return the signed token
}

// ValidateJWT validates the given JWT token and returns the parsed claims if valid.
func ValidateJWT(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
	})

	// Handle various types of errors related to token validation
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errors.New("token has expired")
			}
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, errors.New("token is malformed")
			}
		}
		return nil, errors.New("invalid token")
	}

	// Check if token is valid
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims from the token
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("could not parse JWT claims")
	}

	return claims, nil
}

// ParseJWTFromHeader extracts and validates the token from the Authorization header.
func ParseJWTFromHeader(authHeader string) (*CustomClaims, error) {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("missing Bearer token")
	}

	// Clean up the token by trimming the "Bearer " prefix
	tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
	return ValidateJWT(tokenStr) // Validate the extracted token
}

// Helper functions for common checks

// IsRole checks if the user's role matches the given role.
func IsRole(claims *CustomClaims, role string) bool {
	return claims.Role == role
}

// GetUserID retrieves the user ID from the claims.
func GetUserID(claims *CustomClaims) string {
	return claims.UserID // Return the userID as string
}

// GetEmail retrieves the email from the claims.
func GetEmail(claims *CustomClaims) string {
	return claims.Email
}
