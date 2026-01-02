package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// PaginatedResponse matches tgo-ai Python API format
type PaginatedResponse struct {
	Data       interface{}         `json:"data"`
	Pagination *PaginationMetadata `json:"pagination"`
}

// PaginationMetadata matches tgo-ai Python API format
type PaginationMetadata struct {
	Total   int64 `json:"total"`
	Limit   int   `json:"limit"`
	Offset  int   `json:"offset"`
	HasNext bool  `json:"has_next"`
	HasPrev bool  `json:"has_prev"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func List(c *gin.Context, data interface{}, total int64, limit, offset int) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Data: data,
		Pagination: &PaginationMetadata{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasNext: int64(offset+limit) < total,
			HasPrev: offset > 0,
		},
	})
}

func Error(c *gin.Context, statusCode int, code, message string, details interface{}) {
	c.JSON(statusCode, Response{
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "VALIDATION_ERROR", message, nil)
}

func NotFound(c *gin.Context, resource string) {
	Error(c, http.StatusNotFound, resource+"_NOT_FOUND", resource+" not found", nil)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "AUTHENTICATION_FAILED", message, nil)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "ACCESS_DENIED", message, nil)
}

func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message, nil)
}
