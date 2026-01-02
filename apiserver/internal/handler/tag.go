package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/service"
)

type TagHandler struct {
	svc *service.TagService
}

func NewTagHandler(svc *service.TagService) *TagHandler {
	return &TagHandler{svc: svc}
}

func (h *TagHandler) List(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	category := c.Query("category")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	tags, total, err := h.svc.List(c.Request.Context(), projectID, category, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tags, "total": total})
}

func (h *TagHandler) Get(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	tag, err := h.svc.GetByID(c.Request.Context(), projectID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}
	c.JSON(http.StatusOK, tag)
}

func (h *TagHandler) Create(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))

	var tag model.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tag.ProjectID = projectID

	if err := h.svc.Create(c.Request.Context(), &tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tag)
}

func (h *TagHandler) Update(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	tag, err := h.svc.GetByID(c.Request.Context(), projectID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}
	if err := c.ShouldBindJSON(tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Update(c.Request.Context(), tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tag)
}

func (h *TagHandler) Delete(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	if err := h.svc.Delete(c.Request.Context(), projectID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AddVisitorTagRequest represents visitor tag add request
type AddVisitorTagRequest struct {
	VisitorID string `json:"visitor_id" binding:"required"`
	TagID     string `json:"tag_id" binding:"required"`
}

// AddVisitorTag adds a tag to a visitor
func (h *TagHandler) AddVisitorTag(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))

	var req AddVisitorTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	visitorID, err := uuid.Parse(req.VisitorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visitor_id"})
		return
	}
	tagID, err := uuid.Parse(req.TagID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag_id"})
		return
	}

	if err := h.svc.AddToVisitor(c.Request.Context(), projectID, visitorID, tagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// RemoveVisitorTag removes a tag from a visitor
func (h *TagHandler) RemoveVisitorTag(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))

	var req AddVisitorTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	visitorID, err := uuid.Parse(req.VisitorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visitor_id"})
		return
	}
	tagID, err := uuid.Parse(req.TagID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag_id"})
		return
	}

	if err := h.svc.RemoveFromVisitor(c.Request.Context(), projectID, visitorID, tagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
