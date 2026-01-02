package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/service"
)

type SearchHandler struct {
	svc *service.SearchService
}

func NewSearchHandler(svc *service.SearchService) *SearchHandler {
	return &SearchHandler{svc: svc}
}

func (h *SearchHandler) Search(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	query := c.Query("q")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query (q) is required"})
		return
	}

	scope := service.SearchScope(c.DefaultQuery("scope", "all"))
	visitorPage, _ := strconv.Atoi(c.DefaultQuery("visitor_page", "1"))
	visitorPageSize, _ := strconv.Atoi(c.DefaultQuery("visitor_page_size", "10"))
	messagePage, _ := strconv.Atoi(c.DefaultQuery("message_page", "1"))
	messagePageSize, _ := strconv.Atoi(c.DefaultQuery("message_page_size", "20"))

	result, err := h.svc.Search(c.Request.Context(), projectID, query, scope, visitorPage, visitorPageSize, messagePage, messagePageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
