package middleware

import (
	"time"

	"gotodolist/utils"

	"github.com/gin-gonic/gin"
)

// Logger is a middleware function that logs requests using our custom logger
func Logger() gin.HandlerFunc {
	logger := utils.GetLogger()

	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate request latency
		latency := time.Since(startTime)

		// Log request details
		logger.LogRequest(c, latency)
	}
}
