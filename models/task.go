package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Task represents a task in the todo list
type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title" binding:"required"`
	Description string             `bson:"description,omitempty" json:"description"`
	Completed   bool               `bson:"completed" json:"completed"`
	DueDate     *time.Time         `bson:"dueDate,omitempty" json:"dueDate"`
	Priority    string             `bson:"priority" json:"priority"`
	User        primitive.ObjectID `bson:"user" json:"user"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// NewTask creates a new task with default values
func NewTask(title string, userID primitive.ObjectID) *Task {
	now := time.Now()
	return &Task{
		Title:     title,
		Completed: false,
		Priority:  "medium",
		User:      userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
