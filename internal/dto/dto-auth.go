package dto

import (
    "RAAS/internal/models"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type SeekerSignUpInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    Number   string `json:"number" binding:"required,min=10,max=15"`
}

type LoginInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// AuthUserMinimal represents minimal user details for response
type AuthUserMinimal struct {
    Email         string `json:"email"`
    EmailVerified bool   `json:"emailVerified"`
    Provider      string `json:"provider"`
    Number        string `json:"number" binding:"required,len=10"`
}

// SeekerResponse represents the response structure for Seeker details
type SeekerResponse struct {
    ID               primitive.ObjectID `json:"id"`        // Use primitive.ObjectID for Seeker ID in MongoDB
    AuthUserID       string              `json:"authUserId"` // UUID as string for AuthUserID
    AuthUser         AuthUserMinimal     `json:"authUser"`
    SubscriptionTier string              `json:"subscriptionTier"`
}

func SeekerProfileResponse(seeker models.Seeker) SeekerResponse {
    return SeekerResponse{
        ID:               seeker.ID,                 // ID as primitive.ObjectID
        AuthUserID:       seeker.AuthUserID, // UUID as string for AuthUserID
        SubscriptionTier: seeker.SubscriptionTier,
    }
}
