package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

// HTTPRetrieverConfig HTTP 检索器配置
type HTTPRetrieverConfig struct {
	BaseURL        string  // RAG 服务地址
	CollectionID   string  // 知识库 ID
	TopK           int     // 返回文档数量
	ScoreThreshold float64 // 分数阈值
}

// HTTPRetriever 基于 HTTP 调用 RAG 服务的检索器
// 实现 eino retriever.Retriever 接口
type HTTPRetriever struct {
	config *HTTPRetrieverConfig
	client *http.Client
}

// NewHTTPRetriever 创建 HTTP 检索器
func NewHTTPRetriever(config *HTTPRetrieverConfig) (*HTTPRetriever, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if config.CollectionID == "" {
		return nil, fmt.Errorf("collection ID is required")
	}
	if config.TopK <= 0 {
		config.TopK = 5
	}

	return &HTTPRetriever{
		config: config,
		client: &http.Client{},
	}, nil
}

// retrieveRequest RAG 检索请求
type retrieveRequest struct {
	CollectionID string `json:"collection_id"`
	Query        string `json:"query"`
	TopK         int    `json:"top_k"`
}

// retrieveResponse RAG 检索响应
type retrieveResponse struct {
	Documents []retrieveDocument `json:"documents"`
}

type retrieveDocument struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Retrieve 实现 retriever.Retriever 接口
func (r *HTTPRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	log.Printf("[HTTPRetriever] Retrieving: %s (collection=%s, topK=%d)",
		truncate(query, 30), r.config.CollectionID, r.config.TopK)

	// 构建请求
	reqBody := retrieveRequest{
		CollectionID: r.config.CollectionID,
		Query:        query,
		TopK:         r.config.TopK,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// 发送请求
	url := strings.TrimSuffix(r.config.BaseURL, "/") + "/retrieve"
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("retrieve failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result retrieveResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// 转换为 eino schema.Document
	docs := make([]*schema.Document, 0, len(result.Documents))
	for _, doc := range result.Documents {
		// 过滤低分文档
		if r.config.ScoreThreshold > 0 && doc.Score < r.config.ScoreThreshold {
			continue
		}

		einoDoc := &schema.Document{
			ID:       doc.ID,
			Content:  doc.Content,
			MetaData: doc.Metadata,
		}
		einoDoc.WithScore(doc.Score)
		docs = append(docs, einoDoc)
	}

	log.Printf("[HTTPRetriever] Retrieved %d documents", len(docs))
	return docs, nil
}

// GetType 返回检索器类型
func (r *HTTPRetriever) GetType() string {
	return "HTTPRetriever"
}

// IsCallbacksEnabled 是否启用回调
func (r *HTTPRetriever) IsCallbacksEnabled() bool {
	return true
}
