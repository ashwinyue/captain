package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/rag/internal/model"
)

// VectorSearchService handles vector similarity search
type VectorSearchService struct {
	db               *gorm.DB
	embeddingService *EmbeddingService
}

// NewVectorSearchService creates a new vector search service
func NewVectorSearchService(db *gorm.DB, embeddingService *EmbeddingService) *VectorSearchService {
	return &VectorSearchService{
		db:               db,
		embeddingService: embeddingService,
	}
}

// VectorSearchResult represents a search result with similarity score
type VectorSearchResult struct {
	Document   *model.Document `json:"document"`
	Score      float64         `json:"score"`
	Similarity float64         `json:"similarity"`
}

// VectorSearchRequest represents a vector search request
type VectorSearchRequest struct {
	Query        string    `json:"query"`
	CollectionID uuid.UUID `json:"collection_id"`
	ProjectID    uuid.UUID `json:"project_id"`
	TopK         int       `json:"top_k"`
	Threshold    float64   `json:"threshold"`
}

// Search performs vector similarity search
func (s *VectorSearchService) Search(ctx context.Context, req *VectorSearchRequest) ([]VectorSearchResult, error) {
	if req.TopK <= 0 {
		req.TopK = 10
	}

	// Generate embedding for query
	queryEmbedding, err := s.embeddingService.GenerateEmbedding(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Perform vector similarity search using pgvector
	var results []struct {
		model.Document
		Distance float64 `gorm:"column:distance"`
	}

	// Use cosine distance for similarity search
	query := s.db.WithContext(ctx).
		Table("rag_documents").
		Select("*, embedding <=> ? as distance", queryEmbedding).
		Where("collection_id = ?", req.CollectionID).
		Where("embedding IS NOT NULL").
		Order("distance ASC").
		Limit(req.TopK)

	// Optionally filter by project_id if provided
	if req.ProjectID != uuid.Nil {
		query = query.Where("project_id = ?", req.ProjectID)
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}

	// Convert to VectorSearchResult
	searchResults := make([]VectorSearchResult, 0, len(results))
	for _, r := range results {
		// Convert distance to similarity (1 - distance for cosine)
		similarity := 1 - r.Distance

		// Apply threshold filter
		if req.Threshold > 0 && similarity < req.Threshold {
			continue
		}

		doc := r.Document
		searchResults = append(searchResults, VectorSearchResult{
			Document:   &doc,
			Score:      r.Distance,
			Similarity: similarity,
		})
	}

	return searchResults, nil
}

// IndexDocument generates embedding and stores it for a document
func (s *VectorSearchService) IndexDocument(ctx context.Context, doc *model.Document) error {
	if doc.Content == "" {
		return nil
	}

	embedding, err := s.embeddingService.GenerateEmbedding(ctx, doc.Content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	doc.Embedding = embedding
	if err := s.db.WithContext(ctx).Save(doc).Error; err != nil {
		return fmt.Errorf("failed to save document with embedding: %w", err)
	}

	return nil
}

// IndexDocuments generates embeddings for multiple documents in batch
func (s *VectorSearchService) IndexDocuments(ctx context.Context, docs []*model.Document) error {
	if len(docs) == 0 {
		return nil
	}

	// Extract content
	contents := make([]string, len(docs))
	for i, doc := range docs {
		contents[i] = doc.Content
	}

	// Generate embeddings in batch
	embeddings, err := s.embeddingService.GenerateEmbeddings(ctx, contents)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Update documents with embeddings
	for i, doc := range docs {
		if i < len(embeddings) {
			doc.Embedding = embeddings[i]
		}
	}

	// Save all documents
	for _, doc := range docs {
		if err := s.db.WithContext(ctx).Save(doc).Error; err != nil {
			return fmt.Errorf("failed to save document: %w", err)
		}
	}

	return nil
}

// DeleteDocumentEmbeddings removes embeddings for documents in a collection
func (s *VectorSearchService) DeleteDocumentEmbeddings(ctx context.Context, collectionID uuid.UUID) error {
	return s.db.WithContext(ctx).
		Model(&model.Document{}).
		Where("collection_id = ?", collectionID).
		Update("embedding", nil).Error
}
