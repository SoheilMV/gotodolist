package utils

import (
	"github.com/gin-gonic/gin"
)

// GetQueryDefault gets a query parameter or returns a default value
func GetQueryDefault(c *gin.Context, key, defaultValue string) string {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	return value
}
