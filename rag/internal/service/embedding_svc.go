package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pgvector/pgvector-go"
)

// EmbeddingService handles embedding generation
type EmbeddingService struct {
	apiKey     string
	baseURL    string
	model      string
	dimensions int
	httpClient *http.Client
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(apiKey, baseURL, model string, dimensions int) *EmbeddingService {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "text-embedding-3-small"
	}
	if dimensions == 0 {
		dimensions = 1536
	}
	return &EmbeddingService{
		apiKey:     apiKey,
		baseURL:    baseURL,
		model:      model,
		dimensions: dimensions,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// EmbeddingRequest represents the OpenAI embedding API request
type EmbeddingRequest struct {
	Input          interface{} `json:"input"`
	Model          string      `json:"model"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
	Dimensions     int         `json:"dimensions,omitempty"`
}

// EmbeddingResponse represents the OpenAI embedding API response
type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// GenerateEmbedding generates embedding for a single text
func (s *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) (pgvector.Vector, error) {
	embeddings, err := s.GenerateEmbeddings(ctx, []string{text})
	if err != nil {
		return pgvector.Vector{}, err
	}
	if len(embeddings) == 0 {
		return pgvector.Vector{}, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

// GenerateEmbeddings generates embeddings for multiple texts
func (s *EmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([]pgvector.Vector, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	reqBody := EmbeddingRequest{
		Input: texts,
		Model: s.model,
	}
	if s.dimensions > 0 {
		reqBody.Dimensions = s.dimensions
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/embeddings", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var embResp EmbeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to pgvector.Vector
	vectors := make([]pgvector.Vector, len(embResp.Data))
	for _, data := range embResp.Data {
		vectors[data.Index] = pgvector.NewVector(data.Embedding)
	}

	return vectors, nil
}

// GetDimensions returns the embedding dimensions
func (s *EmbeddingService) GetDimensions() int {
	return s.dimensions
}
