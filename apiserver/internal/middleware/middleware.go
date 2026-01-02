package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/pkg/jwt"
	"gorm.io/gorm"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Project-ID, X-Request-ID, X-User-Language, Accept-Language, X-Platform-API-Key")
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

type AuthMiddleware struct {
	jwtManager *jwt.Manager
	db         *gorm.DB
	isDev      bool
}

func NewAuthMiddleware(jwtManager *jwt.Manager, db *gorm.DB, isDev bool) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager, db: db, isDev: isDev}
}

func (m *AuthMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}
		tokenString := authHeader[7:]
		if strings.HasPrefix(tokenString, "ak_") {
			m.handleAPIKey(c, tokenString)
			return
		}
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		var staff model.Staff
		if err := m.db.Where("id = ? AND is_active = true AND deleted_at IS NULL", claims.UserID).First(&staff).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		c.Set("user", &staff)
		c.Set("user_id", staff.ID)
		c.Set("role", staff.Role)
		if claims.ProjectID != nil {
			c.Set("project_id", claims.ProjectID.String())
		} else if staff.ProjectID != nil {
			c.Set("project_id", staff.ProjectID.String())
		}
		c.Next()
	}
}

func (m *AuthMiddleware) handleAPIKey(c *gin.Context, apiKey string) {
	var project model.Project
	if err := m.db.Where("api_key = ? AND is_active = true AND deleted_at IS NULL", apiKey).First(&project).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		return
	}
	c.Set("project", &project)
	c.Set("project_id", project.ID.String())
	c.Set("api_key", apiKey)
	c.Next()
}

func (m *AuthMiddleware) APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ak_") {
				apiKey = auth[7:]
			}
		}
		if apiKey != "" {
			m.handleAPIKey(c, apiKey)
			return
		}
		m.JWTAuth()(c)
	}
}

func (m *AuthMiddleware) ProjectRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.GetString("project_id")
		if projectID == "" {
			projectID = c.GetHeader("X-Project-ID")
			if projectID == "" {
				projectID = c.Query("project_id")
			}
		}
		if projectID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Project ID required"})
			return
		}
		if _, err := uuid.Parse(projectID); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}
		c.Set("project_id", projectID)
		c.Next()
	}
}
