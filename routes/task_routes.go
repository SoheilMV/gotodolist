package routes

import (
	"gotodolist/controllers"
	"gotodolist/middleware"

	"github.com/gin-gonic/gin"
)

// SetupTaskRoutes configures the task routes
func SetupTaskRoutes(router *gin.Engine, taskController *controllers.TaskController, authMiddleware *middleware.AuthMiddleware) {
	tasks := router.Group("/tasks")

	// Apply auth middleware to all task routes
	tasks.Use(authMiddleware.Protect())

	{
		tasks.GET("/", taskController.GetTasks)
		tasks.GET("/:id", taskController.GetTask)
		tasks.POST("/", taskController.CreateTask)
		tasks.PUT("/:id", taskController.UpdateTask)
		tasks.DELETE("/:id", taskController.DeleteTask)
	}
}
