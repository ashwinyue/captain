package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/aicenter/pkg/auth"
)

const (
	ContextKeyTokenInfo = "token_info"
	ContextKeyProjectID = "project_id"
	ContextKeyAPIKey    = "api_key"
)

type AuthMiddleware struct {
	authClient *auth.Client
	isDev      bool
}

func NewAuthMiddleware(authClient *auth.Client, isDev bool) *AuthMiddleware {
	return &AuthMiddleware{
		authClient: authClient,
		isDev:      isDev,
	}
}

// JWTAuth validates JWT Bearer token via apiserver
func (m *AuthMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"error": gin.H{
					"code":    "MISSING_AUTHORIZATION",
					"message": "Authorization header is required",
				},
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(401, gin.H{
				"error": gin.H{
					"code":    "INVALID_AUTHORIZATION",
					"message": "Authorization header must be Bearer token",
				},
			})
			return
		}

		token := parts[1]

		// Validate token via apiserver
		tokenInfo, err := m.authClient.ValidateToken(c.Request.Context(), token)
		if err != nil {
			if err == auth.ErrUnauthorized {
				c.AbortWithStatusJSON(401, gin.H{
					"error": gin.H{
						"code":    "INVALID_TOKEN",
						"message": "Invalid or expired token",
					},
				})
				return
			}
			c.AbortWithStatusJSON(500, gin.H{
				"error": gin.H{
					"code":    "AUTH_SERVICE_ERROR",
					"message": "Authentication service error",
				},
			})
			return
		}

		c.Set(ContextKeyTokenInfo, tokenInfo)
		c.Set(ContextKeyProjectID, tokenInfo.ProjectID.String())
		c.Next()
	}
}

// APIKeyAuth validates API key via apiserver or X-Project-ID header
func (m *AuthMiddleware) APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try API key first
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			// Block dev API key in production
			if apiKey == "dev" && !m.isDev {
				c.AbortWithStatusJSON(403, gin.H{
					"error": gin.H{
						"code":    "DEV_KEY_FORBIDDEN",
						"message": "Development API key not allowed in production",
					},
				})
				return
			}

			// Validate API key via apiserver
			projectInfo, err := m.authClient.ValidateAPIKey(c.Request.Context(), apiKey)
			if err != nil {
				if err == auth.ErrUnauthorized {
					c.AbortWithStatusJSON(401, gin.H{
						"error": gin.H{
							"code":    "INVALID_API_KEY",
							"message": "Invalid API key",
						},
					})
					return
				}
				c.AbortWithStatusJSON(500, gin.H{
					"error": gin.H{
						"code":    "AUTH_SERVICE_ERROR",
						"message": "Authentication service error",
					},
				})
				return
			}

			c.Set(ContextKeyProjectID, projectInfo.ID.String())
			c.Set(ContextKeyAPIKey, apiKey)
			c.Next()
			return
		}

		// Fallback to X-Project-ID header (for internal services)
		projectID := c.GetHeader("X-Project-ID")
		if projectID == "" {
			projectID = c.Query("project_id")
		}

		if projectID == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"error": gin.H{
					"code":    "MISSING_AUTH",
					"message": "X-API-Key or X-Project-ID header is required",
				},
			})
			return
		}

		// Validate UUID format
		if _, err := uuid.Parse(projectID); err != nil {
			c.AbortWithStatusJSON(400, gin.H{
				"error": gin.H{
					"code":    "INVALID_PROJECT_ID",
					"message": "Invalid project ID format",
				},
			})
			return
		}

		c.Set(ContextKeyProjectID, projectID)
		c.Next()
	}
}

// GetProjectID extracts project ID from context
func GetProjectID(c *gin.Context) uuid.UUID {
	projectIDStr := c.GetString(ContextKeyProjectID)
	if projectIDStr == "" {
		return uuid.Nil
	}
	id, _ := uuid.Parse(projectIDStr)
	return id
}

// GetTokenInfo extracts token info from context
func GetTokenInfo(c *gin.Context) *auth.TokenInfo {
	info, exists := c.Get(ContextKeyTokenInfo)
	if !exists {
		return nil
	}
	return info.(*auth.TokenInfo)
}
