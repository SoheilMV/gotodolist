package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username           string             `bson:"username" json:"username" binding:"required"`
	Email              string             `bson:"email" json:"email" binding:"required,email"`
	Password           string             `bson:"password" json:"-"`                     // Password is never returned in JSON
	RefreshToken       string             `bson:"refreshToken,omitempty" json:"-"`       // Refresh token hash stored in DB
	RefreshTokenExpire *time.Time         `bson:"refreshTokenExpire,omitempty" json:"-"` // When the refresh token expires
	CreatedAt          time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt          time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// NewUser creates a new user with default values
func NewUser(username, email, hashedPassword string) *User {
	now := time.Now()
	return &User{
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UserResponse is the structure returned when a user is part of a response
// It doesn't include sensitive data like password
type UserResponse struct {
	ID        primitive.ObjectID `json:"id"`
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	CreatedAt time.Time          `json:"createdAt"`
}

// ToResponse converts a User to a UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}
