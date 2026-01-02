package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/tgo/captain/rag/internal/model"
	"github.com/tgo/captain/rag/internal/repository"
)

type QAService struct {
	repo         *repository.QAPairRepository
	docRepo      *repository.DocumentRepository
	embeddingSvc *EmbeddingService
}

func NewQAService(repo *repository.QAPairRepository) *QAService {
	return &QAService{repo: repo}
}

// NewQAServiceWithEmbedding creates QA service with embedding support
func NewQAServiceWithEmbedding(repo *repository.QAPairRepository, docRepo *repository.DocumentRepository, embeddingSvc *EmbeddingService) *QAService {
	return &QAService{
		repo:         repo,
		docRepo:      docRepo,
		embeddingSvc: embeddingSvc,
	}
}

func (s *QAService) List(ctx context.Context, collectionID uuid.UUID, category, status string, limit, offset int) ([]model.QAPair, int64, error) {
	return s.repo.FindByCollectionID(ctx, collectionID, category, status, limit, offset)
}

func (s *QAService) GetByID(ctx context.Context, id uuid.UUID) (*model.QAPair, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *QAService) Create(ctx context.Context, qa *model.QAPair) error {
	if err := s.repo.Create(ctx, qa); err != nil {
		return err
	}
	// Generate embedding for QA pair
	return s.generateEmbeddingForQA(ctx, qa)
}

func (s *QAService) BatchCreate(ctx context.Context, qas []model.QAPair) error {
	if err := s.repo.CreateBatch(ctx, qas); err != nil {
		return err
	}
	// Generate embeddings for all QA pairs
	for i := range qas {
		if err := s.generateEmbeddingForQA(ctx, &qas[i]); err != nil {
			// Log error but continue processing
			fmt.Printf("Failed to generate embedding for QA %s: %v\n", qas[i].ID, err)
		}
	}
	return nil
}

func (s *QAService) Update(ctx context.Context, qa *model.QAPair) error {
	if err := s.repo.Update(ctx, qa); err != nil {
		return err
	}
	// Re-generate embedding for updated QA pair
	return s.generateEmbeddingForQA(ctx, qa)
}

func (s *QAService) Delete(ctx context.Context, id uuid.UUID) error {
	// Delete associated document first
	if s.docRepo != nil {
		// Find and delete document by QA pair ID (stored in metadata)
		s.docRepo.DeleteByQAPairID(ctx, id)
	}
	return s.repo.Delete(ctx, id)
}

// generateEmbeddingForQA creates a document with embedding for a QA pair
func (s *QAService) generateEmbeddingForQA(ctx context.Context, qa *model.QAPair) error {
	if s.embeddingSvc == nil || s.docRepo == nil {
		return nil // Skip if embedding service not configured
	}

	// Combine question and answer for embedding
	content := fmt.Sprintf("问题: %s\n答案: %s", qa.Question, qa.Answer)

	// Generate embedding
	embedding, err := s.embeddingSvc.GenerateEmbedding(ctx, content)
	if err != nil {
		// Mark as failed
		qa.Status = model.QAStatusInactive
		s.repo.Update(ctx, qa)
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Create or update document
	doc := &model.Document{
		ProjectID:    qa.ProjectID,
		CollectionID: qa.CollectionID,
		Content:      content,
		Embedding:    embedding,
		Metadata: model.JSONMap{
			"qa_pair_id": qa.ID.String(),
			"question":   qa.Question,
			"type":       "qa",
		},
	}

	if err := s.docRepo.CreateOrUpdateByQAPairID(ctx, qa.ID, doc); err != nil {
		return err
	}

	// Update QA pair with vector_id and status
	qa.VectorID = doc.ID.String()
	qa.Status = model.QAStatusActive
	return s.repo.Update(ctx, qa)
}

func (s *QAService) ListCategories(ctx context.Context, projectID uuid.UUID, collectionID *uuid.UUID) ([]string, error) {
	return s.repo.ListCategories(ctx, projectID, collectionID)
}

// GetStats returns QA pair statistics for a collection
func (s *QAService) GetStats(ctx context.Context, collectionID uuid.UUID) (*repository.QAStats, error) {
	return s.repo.GetStats(ctx, collectionID)
}
