package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateAccessToken creates a new JWT access token for a user
func GenerateAccessToken(userID string) (string, error) {
	// Define token expiration
	expireTime := GetTokenExpiration()

	// Create claims
	claims := jwt.MapClaims{
		"id":  userID,
		"exp": expireTime.Unix(),
		"iat": time.Now().Unix(),
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(GetEnv("JWT_SECRET", "your-secret-key")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken creates a new refresh token
func GenerateRefreshToken() (string, string, time.Time) {
	// Generate random token
	b := make([]byte, 32)
	rand.Read(b)
	refreshToken := hex.EncodeToString(b)

	// Hash token for storage
	hash := sha256.Sum256([]byte(refreshToken))
	hashedToken := hex.EncodeToString(hash[:])

	// Set expiration time (7 days)
	expireTime := time.Now().Add(7 * 24 * time.Hour)

	return refreshToken, hashedToken, expireTime
}

// GetTokenExpiration returns the expiration time for access tokens
func GetTokenExpiration() time.Time {
	// Parse the JWT_EXPIRE environment variable with a default of 24 hours
	expireStr := GetEnv("JWT_EXPIRE", "24h")

	// Try to parse duration from environment variable
	duration, err := time.ParseDuration(expireStr)
	if err != nil {
		// Default to 24 hours if parsing fails
		duration = 24 * time.Hour
	}

	return time.Now().Add(duration)
}

// HashString hashes a string using SHA-256
func HashString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
