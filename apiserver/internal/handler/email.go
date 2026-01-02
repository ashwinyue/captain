package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tgo/captain/apiserver/internal/service"
)

type EmailHandler struct {
	svc *service.EmailService
}

func NewEmailHandler(svc *service.EmailService) *EmailHandler {
	return &EmailHandler{svc: svc}
}

func (h *EmailHandler) TestConnection(c *gin.Context) {
	var req service.EmailTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.svc.TestConnection(&req)
	c.JSON(http.StatusOK, result)
}
