package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/rag/internal/service"
)

type FileHandler struct {
	svc *service.FileService
}

func NewFileHandler(svc *service.FileService) *FileHandler {
	return &FileHandler{svc: svc}
}

func (h *FileHandler) List(c *gin.Context) {
	projectID, err := uuid.Parse(c.Query("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	var collectionID *uuid.UUID
	if cid := c.Query("collection_id"); cid != "" {
		id, err := uuid.Parse(cid)
		if err == nil {
			collectionID = &id
		}
	}

	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	files, total, err := h.svc.List(c.Request.Context(), projectID, collectionID, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": files,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *FileHandler) Upload(c *gin.Context) {
	projectID, err := uuid.Parse(c.Query("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	var collectionID *uuid.UUID
	if cid := c.PostForm("collection_id"); cid != "" {
		id, err := uuid.Parse(cid)
		if err == nil {
			collectionID = &id
		}
	}

	uploadedFile, err := h.svc.Upload(
		c.Request.Context(),
		projectID,
		collectionID,
		header.Filename,
		header.Header.Get("Content-Type"),
		header.Size,
		file,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, uploadedFile)
}

func (h *FileHandler) BatchUpload(c *gin.Context) {
	projectID, err := uuid.Parse(c.Query("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	var collectionID *uuid.UUID
	if cid := c.PostForm("collection_id"); cid != "" {
		id, err := uuid.Parse(cid)
		if err == nil {
			collectionID = &id
		}
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files provided"})
		return
	}

	var results []interface{}
	var errors []map[string]interface{}

	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			errors = append(errors, map[string]interface{}{
				"filename": header.Filename,
				"error":    err.Error(),
			})
			continue
		}

		uploadedFile, err := h.svc.Upload(
			c.Request.Context(),
			projectID,
			collectionID,
			header.Filename,
			header.Header.Get("Content-Type"),
			header.Size,
			file,
		)
		file.Close()

		if err != nil {
			errors = append(errors, map[string]interface{}{
				"filename": header.Filename,
				"error":    err.Error(),
			})
			continue
		}

		results = append(results, uploadedFile)
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":          results,
		"success_count": len(results),
		"failed_count":  len(errors),
		"errors":        errors,
	})
}

func (h *FileHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	file, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.JSON(http.StatusOK, file)
}

func (h *FileHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *FileHandler) Download(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	filePath, err := h.svc.GetFilePath(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found on storage"})
		return
	}

	c.FileAttachment(filePath, filepath.Base(filePath))
}

// ListDocuments returns documents for a file
func (h *FileHandler) ListDocuments(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	docs, total, err := h.svc.ListDocuments(c.Request.Context(), id, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": docs,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}
