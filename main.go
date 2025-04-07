package main

import (
	"os"
	"time"

	"gotodolist/configs"
	"gotodolist/controllers"
	"gotodolist/middleware"
	"gotodolist/routes"
	"gotodolist/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables
	utils.LoadEnv()

	// Set Gin mode
	mode := utils.GetEnv("GIN_MODE", "debug")
	gin.SetMode(mode)

	// Initialize logger
	logPath := utils.GetEnv("LOG_FILE", "logs/app.log")
	logger, err := utils.InitLogger(logPath)
	if err != nil {
		// If file logger initialization fails, use the default logger
		logger = utils.GetLogger()
	}
	defer logger.Close()

	// Log application startup
	logger.Info("Starting Todolist API application")
	logger.Info("Running in " + mode + " mode")

	// Initialize Gin router (without default logger)
	router := gin.New()

	// Use our custom logger and recovery middleware
	router.Use(middleware.Logger())
	router.Use(gin.Recovery())

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{utils.GetEnv("CORS_ORIGIN", "*")},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Connect to MongoDB
	mongoURI := utils.GetEnv("MONGO_URI", "mongodb://localhost:27017")
	client := configs.ConnectDB(mongoURI)
	logger.Success("Connected to MongoDB")

	// Initialize collections
	dbName := utils.GetEnv("DB_NAME", "todolist")
	tasksCollection := configs.GetCollection(client, "tasks", dbName)
	usersCollection := configs.GetCollection(client, "users", dbName)

	// Initialize controllers
	taskController := controllers.NewTaskController(tasksCollection)
	authController := controllers.NewAuthController(usersCollection)

	// Initialize middlewares
	authMiddleware := middleware.NewAuthMiddleware(usersCollection)

	// Setup routes
	routes.SetupTaskRoutes(router, taskController, authMiddleware)
	routes.SetupAuthRoutes(router, authController, authMiddleware)
	logger.Info("Routes initialized successfully")

	// Setup Swagger documentation
	router.GET("/api-docs/*any", middleware.Swagger())

	// Define health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "up",
			"timestamp": time.Now(),
		})
	})

	// Default welcome route
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Todolist API. Visit /api-docs for documentation.",
		})
	})

	// Start the server
	port := utils.GetEnv("PORT", "8080")
	logger.Info("Server running on port " + port)

	if err := router.Run(":" + port); err != nil {
		logger.Error("Failed to start server: " + err.Error())
		os.Exit(1)
	}
}
