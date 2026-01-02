package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tgo/captain/apiserver/internal/pkg/rag"
)

type RAGHandler struct {
	client *rag.Client
}

func NewRAGHandler(client *rag.Client) *RAGHandler {
	return &RAGHandler{client: client}
}

func (h *RAGHandler) respond(c *gin.Context, data []byte, statusCode int, err error) {
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.Data(statusCode, "application/json", data)
}

func (h *RAGHandler) getQueryParams(c *gin.Context, keys ...string) map[string]string {
	params := make(map[string]string)
	for _, key := range keys {
		if val := c.Query(key); val != "" {
			params[key] = val
		}
	}
	return params
}

// Collections

func (h *RAGHandler) ListCollections(c *gin.Context) {
	projectID := c.GetString("project_id")
	params := h.getQueryParams(c, "display_name", "collection_type", "tags", "limit", "offset")
	data, status, err := h.client.ListCollections(c.Request.Context(), projectID, params)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) CreateCollection(c *gin.Context) {
	projectID := c.GetString("project_id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.CreateCollection(c.Request.Context(), projectID, body)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) GetCollection(c *gin.Context) {
	projectID := c.GetString("project_id")
	collectionID := c.Param("id")
	includeStats := c.Query("include_stats") == "true"
	data, status, err := h.client.GetCollection(c.Request.Context(), projectID, collectionID, includeStats)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) UpdateCollection(c *gin.Context) {
	projectID := c.GetString("project_id")
	collectionID := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.UpdateCollection(c.Request.Context(), projectID, collectionID, body)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) DeleteCollection(c *gin.Context) {
	projectID := c.GetString("project_id")
	collectionID := c.Param("id")
	data, status, err := h.client.DeleteCollection(c.Request.Context(), projectID, collectionID)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) SearchCollectionDocuments(c *gin.Context) {
	projectID := c.GetString("project_id")
	collectionID := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.SearchCollectionDocuments(c.Request.Context(), projectID, collectionID, body)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) ListCollectionPages(c *gin.Context) {
	projectID := c.GetString("project_id")
	collectionID := c.Param("id")
	params := h.getQueryParams(c, "status", "limit", "offset")
	data, status, err := h.client.ListWebsitePages(c.Request.Context(), projectID, collectionID, params)
	h.respond(c, data, status, err)
}

// Files

func (h *RAGHandler) ListFiles(c *gin.Context) {
	projectID := c.GetString("project_id")
	params := h.getQueryParams(c, "collection_id", "status", "content_type", "uploaded_by", "tags", "limit", "offset")
	data, status, err := h.client.ListFiles(c.Request.Context(), projectID, params)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) UploadFile(c *gin.Context) {
	projectID := c.GetString("project_id")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	params := make(map[string]string)
	if collectionID := c.PostForm("collection_id"); collectionID != "" {
		params["collection_id"] = collectionID
	}
	if description := c.PostForm("description"); description != "" {
		params["description"] = description
	}
	if language := c.PostForm("language"); language != "" {
		params["language"] = language
	}
	if tags := c.PostForm("tags"); tags != "" {
		params["tags"] = tags
	}

	data, status, err := h.client.UploadFile(c.Request.Context(), projectID, header.Filename, file, header.Header.Get("Content-Type"), params)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) GetFile(c *gin.Context) {
	projectID := c.GetString("project_id")
	fileID := c.Param("id")
	data, status, err := h.client.GetFile(c.Request.Context(), projectID, fileID)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) DeleteFile(c *gin.Context) {
	projectID := c.GetString("project_id")
	fileID := c.Param("id")
	data, status, err := h.client.DeleteFile(c.Request.Context(), projectID, fileID)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) DownloadFile(c *gin.Context) {
	projectID := c.GetString("project_id")
	fileID := c.Param("id")
	resp, err := h.client.DownloadFile(c.Request.Context(), projectID, fileID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// Copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}
	c.Writer.WriteHeader(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

// Website Pages

func (h *RAGHandler) ListWebsitePages(c *gin.Context) {
	projectID := c.GetString("project_id")
	collectionID := c.Query("collection_id")
	params := h.getQueryParams(c, "status", "depth", "parent_page_id", "tree_depth", "limit", "offset")
	data, status, err := h.client.ListWebsitePages(c.Request.Context(), projectID, collectionID, params)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) GetWebsitePage(c *gin.Context) {
	projectID := c.GetString("project_id")
	pageID := c.Param("id")
	data, status, err := h.client.GetWebsitePage(c.Request.Context(), projectID, pageID)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) AddWebsitePage(c *gin.Context) {
	projectID := c.GetString("project_id")
	collectionID := c.Query("collection_id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.AddWebsitePage(c.Request.Context(), projectID, collectionID, body)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) DeleteWebsitePage(c *gin.Context) {
	projectID := c.GetString("project_id")
	pageID := c.Param("id")
	data, status, err := h.client.DeleteWebsitePage(c.Request.Context(), projectID, pageID)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) RecrawlWebsitePage(c *gin.Context) {
	projectID := c.GetString("project_id")
	pageID := c.Param("id")
	data, status, err := h.client.RecrawlWebsitePage(c.Request.Context(), projectID, pageID)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) CrawlDeeperFromPage(c *gin.Context) {
	projectID := c.GetString("project_id")
	pageID := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.CrawlDeeperFromPage(c.Request.Context(), projectID, pageID, body)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) GetCrawlProgress(c *gin.Context) {
	projectID := c.GetString("project_id")
	collectionID := c.Query("collection_id")
	data, status, err := h.client.GetCrawlProgress(c.Request.Context(), projectID, collectionID)
	h.respond(c, data, status, err)
}

// QA Pairs

func (h *RAGHandler) ListQAPairs(c *gin.Context) {
	projectID := c.GetString("project_id")
	collectionID := c.Param("id")
	params := h.getQueryParams(c, "limit", "offset", "category", "status")
	data, status, err := h.client.ListQAPairs(c.Request.Context(), projectID, collectionID, params)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) CreateQAPair(c *gin.Context) {
	projectID := c.GetString("project_id")
	collectionID := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.CreateQAPair(c.Request.Context(), projectID, collectionID, body)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) BatchCreateQAPairs(c *gin.Context) {
	projectID := c.GetString("project_id")
	collectionID := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.BatchCreateQAPairs(c.Request.Context(), projectID, collectionID, body)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) ImportQAPairs(c *gin.Context) {
	projectID := c.GetString("project_id")
	collectionID := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.ImportQAPairs(c.Request.Context(), projectID, collectionID, body)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) GetQAPair(c *gin.Context) {
	projectID := c.GetString("project_id")
	qaPairID := c.Param("id")
	data, status, err := h.client.GetQAPair(c.Request.Context(), projectID, qaPairID)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) UpdateQAPair(c *gin.Context) {
	projectID := c.GetString("project_id")
	qaPairID := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.UpdateQAPair(c.Request.Context(), projectID, qaPairID, body)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) DeleteQAPair(c *gin.Context) {
	projectID := c.GetString("project_id")
	qaPairID := c.Param("id")
	data, status, err := h.client.DeleteQAPair(c.Request.Context(), projectID, qaPairID)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) ListQACategories(c *gin.Context) {
	projectID := c.GetString("project_id")
	collectionID := c.Query("collection_id")
	data, status, err := h.client.ListQACategories(c.Request.Context(), projectID, collectionID)
	h.respond(c, data, status, err)
}

func (h *RAGHandler) GetQAStats(c *gin.Context) {
	projectID := c.GetString("project_id")
	collectionID := c.Param("id")
	data, status, err := h.client.GetQAStats(c.Request.Context(), projectID, collectionID)
	h.respond(c, data, status, err)
}
