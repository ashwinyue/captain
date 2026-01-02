package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/rag/internal/service"
)

type RetrieveHandler struct {
	vectorSearchSvc *service.VectorSearchService
}

func NewRetrieveHandler(vectorSearchSvc *service.VectorSearchService) *RetrieveHandler {
	return &RetrieveHandler{vectorSearchSvc: vectorSearchSvc}
}

type RetrieveRequest struct {
	CollectionID string `json:"collection_id" binding:"required"`
	Query        string `json:"query" binding:"required"`
	TopK         int    `json:"top_k"`
}

type RetrieveDocument struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
	Score    float64                `json:"score"`
}

type RetrieveResponse struct {
	Documents []RetrieveDocument `json:"documents"`
}

func (h *RetrieveHandler) Retrieve(c *gin.Context) {
	var req RetrieveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.TopK == 0 {
		req.TopK = 5
	}

	collectionID, err := uuid.Parse(req.CollectionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection_id"})
		return
	}

	// Use vector search service to search documents
	results, err := h.vectorSearchSvc.Search(c.Request.Context(), &service.VectorSearchRequest{
		Query:        req.Query,
		CollectionID: collectionID,
		TopK:         req.TopK,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	docs := make([]RetrieveDocument, len(results))
	for i, r := range results {
		docs[i] = RetrieveDocument{
			ID:       r.Document.ID.String(),
			Content:  r.Document.Content,
			Metadata: r.Document.Metadata,
			Score:    r.Similarity,
		}
	}

	c.JSON(http.StatusOK, RetrieveResponse{Documents: docs})
}
