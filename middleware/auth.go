package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gotodolist/models"
	"gotodolist/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AuthMiddleware contains the dependencies needed for auth middleware
type AuthMiddleware struct {
	userCollection *mongo.Collection
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(userCollection *mongo.Collection) *AuthMiddleware {
	return &AuthMiddleware{
		userCollection: userCollection,
	}
}

// Protect secures routes that require authentication
func (am *AuthMiddleware) Protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header required",
			})
			c.Abort()
			return
		}

		// Check if the header format is valid
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid authorization format, use Bearer {token}",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate the token
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(utils.GetEnv("JWT_SECRET", "your-secret-key")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Get the user ID from the token
		userIDStr, ok := claims["id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid token payload",
			})
			c.Abort()
			return
		}

		// Convert string ID to ObjectID
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid user ID in token",
			})
			c.Abort()
			return
		}

		// Find the user in the database
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user models.User
		err = am.userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "User not found",
				})
				c.Abort()
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to authenticate user",
			})
			c.Abort()
			return
		}

		// Set user information in the context
		c.Set("user", user)
		c.Set("userId", userID)

		c.Next()
	}
}
