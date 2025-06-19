package models

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Constants for roles
const (
	RoleAdmin = "admin"
	RoleUser  = "seeker"
)

// ValidateAuthUserRole validates the role of AuthUser
func (u *AuthUser) ValidateAuthUserRole() error {
	if u.Role != RoleAdmin && u.Role != RoleUser {
		return fmt.Errorf("role must be either 'admin' or 'user'")
	}
	return nil
}

// SetID sets the ID for UserEntryTimeline if it's not already set
func (timeline *UserEntryTimeline) SetID() {
	if timeline.ID.IsZero() {
		timeline.ID = primitive.NewObjectID()
	}
}

