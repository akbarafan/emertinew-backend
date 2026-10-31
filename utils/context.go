package utils

import (
	"github.com/gin-gonic/gin"
)

// GetUserIDFromContext extracts user ID from JWT context
func GetUserIDFromContext(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}

	// Try to convert to int (JWT claims default)
	if id, ok := userID.(int); ok {
		return uint(id)
	}

	// Try to convert to uint
	if id, ok := userID.(uint); ok {
		return id
	}

	// Try to convert from float64 (some JWT libraries)
	if id, ok := userID.(float64); ok {
		return uint(id)
	}

	return 0
}

// GetClientIP extracts client IP from request
func GetClientIP(c *gin.Context) string {
	// Try X-Forwarded-For first (proxy/load balancer)
	ip := c.GetHeader("X-Forwarded-For")
	if ip != "" {
		return ip
	}

	// Try X-Real-IP
	ip = c.GetHeader("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Fallback to RemoteAddr
	return c.ClientIP()
}

// GetUserAgent extracts user agent from request
func GetUserAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
}
