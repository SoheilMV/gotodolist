package routes

import (
	"gotodolist/controllers"
	"gotodolist/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes configures the authentication routes
func SetupAuthRoutes(router *gin.Engine, authController *controllers.AuthController, authMiddleware *middleware.AuthMiddleware) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/refresh-token", authController.RefreshToken)

		// Protected routes
		auth.POST("/logout", authMiddleware.Protect(), authController.Logout)
		auth.GET("/me", authMiddleware.Protect(), authController.GetMe)
	}
}
