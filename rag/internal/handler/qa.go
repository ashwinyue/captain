package handler

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/rag/internal/model"
	"github.com/tgo/captain/rag/internal/service"
)

type QAHandler struct {
	svc *service.QAService
}

func NewQAHandler(svc *service.QAService) *QAHandler {
	return &QAHandler{svc: svc}
}

func (h *QAHandler) List(c *gin.Context) {
	collectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection_id"})
		return
	}

	category := c.Query("category")
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	qas, total, err := h.svc.List(c.Request.Context(), collectionID, category, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": qas,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *QAHandler) Create(c *gin.Context) {
	projectID, err := uuid.Parse(c.Query("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	collectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection_id"})
		return
	}

	var qa model.QAPair
	if err := c.ShouldBindJSON(&qa); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	qa.ProjectID = projectID
	qa.CollectionID = collectionID

	if err := h.svc.Create(c.Request.Context(), &qa); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, qa)
}

func (h *QAHandler) BatchCreate(c *gin.Context) {
	projectID, err := uuid.Parse(c.Query("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	collectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection_id"})
		return
	}

	var req struct {
		QAPairs []model.QAPair `json:"qa_pairs"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i := range req.QAPairs {
		req.QAPairs[i].ProjectID = projectID
		req.QAPairs[i].CollectionID = collectionID
	}

	if err := h.svc.BatchCreate(c.Request.Context(), req.QAPairs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"created": len(req.QAPairs)})
}

// ImportRequest for JSON body import
type ImportRequest struct {
	Format string `json:"format"` // "json" or "csv"
	Data   string `json:"data"`   // raw data string
}

func (h *QAHandler) Import(c *gin.Context) {
	projectID, err := uuid.Parse(c.Query("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	collectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection_id"})
		return
	}

	var qaPairs []model.QAPair

	// Check content type to determine how to parse the request
	contentType := c.GetHeader("Content-Type")

	// Try JSON body first (frontend sends {"format": "json", "data": "..."})
	if contentType == "application/json" {
		var req ImportRequest
		if err := c.ShouldBindJSON(&req); err == nil && req.Data != "" {
			// Parse the data field based on format
			if req.Format == "json" {
				if err := json.Unmarshal([]byte(req.Data), &qaPairs); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON data: " + err.Error()})
					return
				}
			} else if req.Format == "csv" {
				qaPairs = parseCSVData(req.Data)
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported format: " + req.Format})
				return
			}

			if len(qaPairs) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "no valid QA pairs found in data"})
				return
			}

			// Set project and collection IDs
			for i := range qaPairs {
				qaPairs[i].ProjectID = projectID
				qaPairs[i].CollectionID = collectionID
			}

			if err := h.svc.BatchCreate(c.Request.Context(), qaPairs); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, gin.H{"imported": len(qaPairs)})
			return
		}
	}

	// Fall back to file upload
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required or invalid JSON body"})
		return
	}
	defer file.Close()

	// Determine format from content type or file extension
	fileContentType := header.Header.Get("Content-Type")
	isJSON := fileContentType == "application/json" ||
		(len(header.Filename) > 5 && header.Filename[len(header.Filename)-5:] == ".json")

	if isJSON {
		// Parse JSON format: [{"question": "...", "answer": "...", "category": "..."}]
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&qaPairs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON format: " + err.Error()})
			return
		}
	} else {
		// Parse CSV format: question,answer,category
		reader := csv.NewReader(file)

		// Skip header if exists
		header, err := reader.Read()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read CSV: " + err.Error()})
			return
		}

		// Check if first row is header
		isHeader := false
		for _, h := range header {
			if h == "question" || h == "answer" || h == "category" {
				isHeader = true
				break
			}
		}

		if !isHeader {
			// First row is data, parse it
			if len(header) >= 2 {
				qa := model.QAPair{
					Question: header[0],
					Answer:   header[1],
				}
				if len(header) >= 3 {
					qa.Category = header[2]
				}
				qaPairs = append(qaPairs, qa)
			}
		}

		// Read remaining rows
		for {
			record, err := reader.Read()
			if err != nil {
				break
			}
			if len(record) >= 2 {
				qa := model.QAPair{
					Question: record[0],
					Answer:   record[1],
				}
				if len(record) >= 3 {
					qa.Category = record[2]
				}
				qaPairs = append(qaPairs, qa)
			}
		}
	}

	if len(qaPairs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid QA pairs found in file"})
		return
	}

	// Set project and collection IDs
	for i := range qaPairs {
		qaPairs[i].ProjectID = projectID
		qaPairs[i].CollectionID = collectionID
	}

	if err := h.svc.BatchCreate(c.Request.Context(), qaPairs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"imported": len(qaPairs),
		"message":  "QA pairs imported successfully",
	})
}

// parseCSVData parses CSV data string into QA pairs
func parseCSVData(data string) []model.QAPair {
	var qaPairs []model.QAPair
	reader := csv.NewReader(strings.NewReader(data))

	// Read header or first row
	header, err := reader.Read()
	if err != nil {
		return qaPairs
	}

	// Check if first row is header
	isHeader := false
	for _, h := range header {
		if h == "question" || h == "answer" || h == "category" {
			isHeader = true
			break
		}
	}

	if !isHeader && len(header) >= 2 {
		qa := model.QAPair{
			Question: header[0],
			Answer:   header[1],
		}
		if len(header) >= 3 {
			qa.Category = header[2]
		}
		qaPairs = append(qaPairs, qa)
	}

	// Read remaining rows
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		if len(record) >= 2 {
			qa := model.QAPair{
				Question: record[0],
				Answer:   record[1],
			}
			if len(record) >= 3 {
				qa.Category = record[2]
			}
			qaPairs = append(qaPairs, qa)
		}
	}

	return qaPairs
}

func (h *QAHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	qa, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "qa pair not found"})
		return
	}

	c.JSON(http.StatusOK, qa)
}

func (h *QAHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	qa, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "qa pair not found"})
		return
	}

	if err := c.ShouldBindJSON(qa); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Update(c.Request.Context(), qa); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, qa)
}

func (h *QAHandler) Delete(c *gin.Context) {
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

func (h *QAHandler) ListCategories(c *gin.Context) {
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

	categories, err := h.svc.ListCategories(c.Request.Context(), projectID, collectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// Stats returns QA pair statistics for a collection
func (h *QAHandler) Stats(c *gin.Context) {
	collectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection_id"})
		return
	}

	stats, err := h.svc.GetStats(c.Request.Context(), collectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
