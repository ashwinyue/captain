package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/tgo/captain/rag/internal/model"
	"github.com/tgo/captain/rag/internal/repository"
)

type CollectionService struct {
	collectionRepo *repository.CollectionRepository
	documentRepo   *repository.DocumentRepository
}

func NewCollectionService(collectionRepo *repository.CollectionRepository, documentRepo *repository.DocumentRepository) *CollectionService {
	return &CollectionService{collectionRepo: collectionRepo, documentRepo: documentRepo}
}

func (s *CollectionService) List(ctx context.Context, projectID uuid.UUID, collectionType string, limit, offset int) ([]model.Collection, int64, error) {
	return s.collectionRepo.FindByProjectID(ctx, projectID, collectionType, limit, offset)
}

func (s *CollectionService) Create(ctx context.Context, collection *model.Collection) error {
	return s.collectionRepo.Create(ctx, collection)
}

func (s *CollectionService) GetByID(ctx context.Context, id uuid.UUID, includeStats bool) (*model.Collection, error) {
	collection, err := s.collectionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if includeStats {
		count, _ := s.documentRepo.CountByCollectionID(ctx, id)
		collection.DocumentCount = int(count)
	}

	return collection, nil
}

func (s *CollectionService) Update(ctx context.Context, collection *model.Collection) error {
	return s.collectionRepo.Update(ctx, collection)
}

func (s *CollectionService) Delete(ctx context.Context, id uuid.UUID) error {
	// Delete associated documents first
	s.documentRepo.DeleteByCollectionID(ctx, id)
	return s.collectionRepo.Delete(ctx, id)
}

type SearchRequest struct {
	Query      string   `json:"query"`
	TopK       int      `json:"top_k"`
	Filters    []string `json:"filters,omitempty"`
	MinScore   float64  `json:"min_score,omitempty"`
	MaxResults int      `json:"max_results,omitempty"`
}

type SearchResult struct {
	Documents []DocumentResult `json:"documents"`
	Total     int              `json:"total"`
}

type DocumentResult struct {
	ID       uuid.UUID `json:"id"`
	Content  string    `json:"content"`
	Score    float64   `json:"score"`
	Metadata JSONMap   `json:"metadata,omitempty"`
}

type JSONMap = model.JSONMap

func (s *CollectionService) SearchDocuments(ctx context.Context, collectionID uuid.UUID, req *SearchRequest) (*SearchResult, error) {
	// TODO: Implement vector search with embedding service
	// For now, return empty results
	return &SearchResult{
		Documents: []DocumentResult{},
		Total:     0,
	}, nil
}

// ListDocuments returns documents for a collection with pagination
func (s *CollectionService) ListDocuments(ctx context.Context, collectionID uuid.UUID, limit, offset int) ([]model.Document, int64, error) {
	return s.documentRepo.FindByCollectionID(ctx, collectionID, limit, offset)
}

// BatchCreate creates multiple collections
func (s *CollectionService) BatchCreate(ctx context.Context, collections []model.Collection) ([]model.Collection, error) {
	for i := range collections {
		if err := s.collectionRepo.Create(ctx, &collections[i]); err != nil {
			return nil, err
		}
	}
	return collections, nil
}
