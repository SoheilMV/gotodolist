package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"gotodolist/models"
	"gotodolist/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TaskController handles task-related operations
type TaskController struct {
	collection *mongo.Collection
}

// NewTaskController creates a new task controller
func NewTaskController(collection *mongo.Collection) *TaskController {
	return &TaskController{
		collection: collection,
	}
}

// GetTasks retrieves all tasks for the authenticated user
func (tc *TaskController) GetTasks(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user ID from context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Parse query parameters for filtering, sorting and pagination
	completed := c.Query("completed")
	priority := c.Query("priority")
	sortField := c.Query("sort")
	sortDir := utils.GetQueryDefault(c, "sortDir", "asc")
	page, _ := strconv.Atoi(utils.GetQueryDefault(c, "page", "1"))
	limit, _ := strconv.Atoi(utils.GetQueryDefault(c, "limit", "10"))

	// Ensure page and limit are valid
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Calculate skip for pagination
	skip := (page - 1) * limit

	// Build query
	query := bson.M{"user": userID}

	// Add filters if provided
	if completed != "" {
		query["completed"] = completed == "true"
	}

	if priority != "" {
		query["priority"] = priority
	}

	// Build sort options
	findOptions := options.Find()

	// Apply sorting
	if sortField != "" {
		var sortOrder int
		if sortDir == "desc" {
			sortOrder = -1
		} else {
			sortOrder = 1
		}

		// Use the sortOrder variable
		findOptions.SetSort(bson.M{sortField: sortOrder})
	} else {
		// Default sort by createdAt
		findOptions.SetSort(bson.M{"createdAt": -1})
	}

	// Apply pagination
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	// Count total documents for pagination
	total, err := tc.collection.CountDocuments(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to count tasks",
		})
		return
	}

	// Execute query with options
	cursor, err := tc.collection.Find(ctx, query, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch tasks",
		})
		return
	}
	defer cursor.Close(ctx)

	var tasks []models.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to parse tasks",
		})
		return
	}

	// Pagination result
	totalPages := (int(total) + limit - 1) / limit
	pagination := gin.H{
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"pagination": pagination,
		"count":      len(tasks),
		"data":       tasks,
	})
}

// GetTask retrieves a single task by ID
func (tc *TaskController) GetTask(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user ID from context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid task ID format",
		})
		return
	}

	var task models.Task
	err = tc.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Task not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch task",
		})
		return
	}

	// Check if the task belongs to the user
	if task.User != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to access this task",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    task,
	})
}

// CreateTask creates a new task
func (tc *TaskController) CreateTask(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user ID from context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	var input struct {
		Title       string     `json:"title" binding:"required"`
		Description string     `json:"description"`
		Completed   bool       `json:"completed"`
		DueDate     *time.Time `json:"dueDate"`
		Priority    string     `json:"priority"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid input data",
		})
		return
	}

	// Validate priority if provided
	if input.Priority != "" && input.Priority != "low" && input.Priority != "medium" && input.Priority != "high" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Priority must be one of: low, medium, high",
		})
		return
	}

	// Create a new task
	task := models.NewTask(input.Title, userID.(primitive.ObjectID))
	task.Description = input.Description
	task.Completed = input.Completed
	task.DueDate = input.DueDate

	if input.Priority != "" {
		task.Priority = input.Priority
	}

	result, err := tc.collection.InsertOne(ctx, task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create task",
		})
		return
	}

	// Get the created task to return
	task.ID = result.InsertedID.(primitive.ObjectID)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    task,
	})
}

// UpdateTask updates an existing task
func (tc *TaskController) UpdateTask(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user ID from context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid task ID format",
		})
		return
	}

	var input struct {
		Title       string     `json:"title"`
		Description string     `json:"description"`
		Completed   bool       `json:"completed"`
		DueDate     *time.Time `json:"dueDate"`
		Priority    string     `json:"priority"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid input data",
		})
		return
	}

	// Validate priority if provided
	if input.Priority != "" && input.Priority != "low" && input.Priority != "medium" && input.Priority != "high" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Priority must be one of: low, medium, high",
		})
		return
	}

	// Get the existing task first
	var existingTask models.Task
	err = tc.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&existingTask)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Task not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch task",
		})
		return
	}

	// Check if the task belongs to the user
	if existingTask.User != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to update this task",
		})
		return
	}

	// Prepare update data
	updateSet := bson.M{
		"updatedAt": time.Now(),
	}

	// Only update fields that were provided
	if input.Title != "" {
		updateSet["title"] = input.Title
	}
	if input.Description != "" {
		updateSet["description"] = input.Description
	}
	updateSet["completed"] = input.Completed
	if input.DueDate != nil {
		updateSet["dueDate"] = input.DueDate
	}
	if input.Priority != "" {
		updateSet["priority"] = input.Priority
	}

	_, err = tc.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updateSet},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update task",
		})
		return
	}

	// Get the updated task
	var updatedTask models.Task
	err = tc.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&updatedTask)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve updated task",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updatedTask,
	})
}

// DeleteTask deletes a task
func (tc *TaskController) DeleteTask(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user ID from context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid task ID format",
		})
		return
	}

	// Get the task first to check ownership
	var task models.Task
	err = tc.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Task not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch task",
		})
		return
	}

	// Check if the task belongs to the user
	if task.User != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to delete this task",
		})
		return
	}

	_, err = tc.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete task",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{},
	})
}
