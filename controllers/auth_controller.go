package controllers

import (
	"context"
	"net/http"
	"time"

	"gotodolist/models"
	"gotodolist/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// AuthController handles authentication-related operations
type AuthController struct {
	userCollection *mongo.Collection
	logger         *utils.Logger
}

// NewAuthController creates a new auth controller
func NewAuthController(userCollection *mongo.Collection) *AuthController {
	return &AuthController{
		userCollection: userCollection,
		logger:         utils.GetLogger(),
	}
}

// Register handles user registration
func (ac *AuthController) Register(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var input struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		ac.logger.Warning("Registration failed: Invalid input data")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid input data",
		})
		return
	}

	// Check if username or email already exists
	existingUser := ac.userCollection.FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"username": input.Username},
			{"email": input.Email},
		},
	})

	if existingUser.Err() == nil {
		ac.logger.Warning("Registration failed: Username or email already in use: " + input.Email)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Username or email already in use",
		})
		return
	}

	if existingUser.Err() != mongo.ErrNoDocuments {
		ac.logger.Error("Registration failed: Database error while checking existing users")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to check existing users",
		})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		ac.logger.Error("Registration failed: Password hashing error")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to process password",
		})
		return
	}

	// Create the user
	user := models.NewUser(input.Username, input.Email, string(hashedPassword))

	result, err := ac.userCollection.InsertOne(ctx, user)
	if err != nil {
		ac.logger.Error("Registration failed: Database error while creating user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create user",
		})
		return
	}

	// Get the inserted ID
	user.ID = result.InsertedID.(primitive.ObjectID)

	// Generate tokens and send response
	if err := ac.sendTokenResponse(c, user); err != nil {
		ac.logger.Error("Registration failed: Error sending token response: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate authentication tokens",
		})
		return
	}

	ac.logger.Success("User registered successfully: " + user.Username + " (" + user.Email + ")")
}

// Login handles user login
func (ac *AuthController) Login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		ac.logger.Warning("Login failed: Invalid input data")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid input data",
		})
		return
	}

	// Find the user
	var user models.User
	err := ac.userCollection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ac.logger.Warning("Login failed: Invalid credentials for email: " + input.Email)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid credentials",
			})
			return
		}
		ac.logger.Error("Login failed: Database error while finding user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to find user",
		})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		ac.logger.Warning("Login failed: Invalid password for user: " + user.Email)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid credentials",
		})
		return
	}

	// Generate tokens and send response
	if err := ac.sendTokenResponse(c, &user); err != nil {
		ac.logger.Error("Login failed: Error sending token response: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate authentication tokens",
		})
		return
	}

	ac.logger.Success("User logged in successfully: " + user.Username + " (" + user.Email + ")")
}

// Logout handles user logout
func (ac *AuthController) Logout(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user ID from context
	userID, exists := c.Get("userId")
	if !exists {
		ac.logger.Warning("Logout failed: User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Not authenticated",
		})
		return
	}

	// Clear refresh token in database
	_, err := ac.userCollection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{
			"refreshToken":       nil,
			"refreshTokenExpire": nil,
		}},
	)

	if err != nil {
		ac.logger.Error("Logout failed: Error updating user record: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to complete logout",
		})
		return
	}

	ac.logger.Info("User logged out successfully: " + userID.(primitive.ObjectID).Hex())
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// RefreshToken handles token refresh
func (ac *AuthController) RefreshToken(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var input struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		ac.logger.Warning("Token refresh failed: Invalid input data")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Refresh token is required",
		})
		return
	}

	// Hash the provided token to check against database
	hashedToken := utils.HashString(input.RefreshToken)

	// Find user with matching refresh token that hasn't expired
	var user models.User
	err := ac.userCollection.FindOne(ctx, bson.M{
		"refreshToken":       hashedToken,
		"refreshTokenExpire": bson.M{"$gt": time.Now()},
	}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ac.logger.Warning("Token refresh failed: Invalid or expired refresh token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or expired refresh token",
			})
			return
		}
		ac.logger.Error("Token refresh failed: Database error: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to validate refresh token",
		})
		return
	}

	// Generate new tokens and send response
	if err := ac.sendTokenResponse(c, &user); err != nil {
		ac.logger.Error("Token refresh failed: Error sending token response: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate authentication tokens",
		})
		return
	}

	ac.logger.Info("Tokens refreshed successfully for user: " + user.Username)
}

// GetMe retrieves the authenticated user's information
func (ac *AuthController) GetMe(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		ac.logger.Warning("GetMe failed: User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Not authenticated",
		})
		return
	}

	userObj, ok := user.(models.User)
	if !ok {
		ac.logger.Error("GetMe failed: Type assertion error for user object")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get user data",
		})
		return
	}

	ac.logger.Debug("User retrieved their profile: " + userObj.Username)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    userObj.ToResponse(),
	})
}

// sendTokenResponse generates access and refresh tokens and sends the response
func (ac *AuthController) sendTokenResponse(c *gin.Context, user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Generate access token
	accessToken, err := utils.GenerateAccessToken(user.ID.Hex())
	if err != nil {
		return err
	}

	// Generate refresh token
	refreshToken, hashedRefreshToken, expireTime := utils.GenerateRefreshToken()

	// Update user with new refresh token
	update := bson.M{
		"$set": bson.M{
			"refreshToken":       hashedRefreshToken,
			"refreshTokenExpire": expireTime,
		},
	}

	_, err = ac.userCollection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		return err
	}

	// Send response
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"token":        accessToken,
		"refreshToken": refreshToken,
		"user":         user.ToResponse(),
	})

	return nil
}
