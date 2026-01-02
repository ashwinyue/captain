package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		log.Printf("[HTTP] %s %s %d %v", method, path, statusCode, latency)
	}
}

func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization, X-API-Key, X-Request-ID, X-Project-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

// ProjectID extracts and validates project ID from request
func ProjectID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get project ID from header or query
		projectID := c.GetHeader("X-Project-ID")
		if projectID == "" {
			projectID = c.Query("project_id")
		}

		if projectID == "" {
			c.AbortWithStatusJSON(400, gin.H{
				"error": gin.H{
					"code":    "MISSING_PROJECT_ID",
					"message": "X-Project-ID header or project_id query parameter is required",
				},
			})
			return
		}

		// Validate UUID format
		if _, err := uuid.Parse(projectID); err != nil {
			c.AbortWithStatusJSON(400, gin.H{
				"error": gin.H{
					"code":    "INVALID_PROJECT_ID",
					"message": "Invalid project_id format, must be a valid UUID",
				},
			})
			return
		}

		c.Set("project_id", projectID)
		c.Next()
	}
}

// APIKeyAuth validates API key authentication
func APIKeyAuth(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Also check Authorization header
			auth := c.GetHeader("Authorization")
			if len(auth) > 7 && auth[:7] == "Bearer " {
				apiKey = auth[7:]
			}
		}

		// For now, just set the API key if present
		// TODO: Validate API key against database
		if apiKey != "" {
			c.Set("api_key", apiKey)
		}

		c.Next()
	}
}
