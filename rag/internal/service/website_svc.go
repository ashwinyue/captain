package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/tgo/captain/rag/internal/model"
	"github.com/tgo/captain/rag/internal/repository"
)

type WebsiteService struct {
	pageRepo     *repository.WebsitePageRepository
	documentRepo *repository.DocumentRepository
}

func NewWebsiteService(pageRepo *repository.WebsitePageRepository, documentRepo *repository.DocumentRepository) *WebsiteService {
	return &WebsiteService{pageRepo: pageRepo, documentRepo: documentRepo}
}

func (s *WebsiteService) ListPages(ctx context.Context, collectionID uuid.UUID, status string, limit, offset int) ([]model.WebsitePage, int64, error) {
	return s.pageRepo.FindByCollectionID(ctx, collectionID, status, limit, offset)
}

func (s *WebsiteService) GetPage(ctx context.Context, id uuid.UUID) (*model.WebsitePage, error) {
	return s.pageRepo.FindByID(ctx, id)
}

func (s *WebsiteService) AddPage(ctx context.Context, page *model.WebsitePage) error {
	page.Status = model.PageStatusPending
	return s.pageRepo.Create(ctx, page)
}

func (s *WebsiteService) DeletePage(ctx context.Context, id uuid.UUID) error {
	return s.pageRepo.Delete(ctx, id)
}

func (s *WebsiteService) RecrawlPage(ctx context.Context, id uuid.UUID) (*model.WebsitePage, error) {
	page, err := s.pageRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	page.Status = model.PageStatusPending
	page.ErrorMessage = ""

	if err := s.pageRepo.Update(ctx, page); err != nil {
		return nil, err
	}

	// TODO: Trigger async crawl

	return page, nil
}

type CrawlDeeperRequest struct {
	MaxDepth   int      `json:"max_depth"`
	MaxPages   int      `json:"max_pages"`
	UrlPattern string   `json:"url_pattern,omitempty"`
	Excludes   []string `json:"excludes,omitempty"`
}

func (s *WebsiteService) CrawlDeeper(ctx context.Context, id uuid.UUID, req *CrawlDeeperRequest) (*model.WebsitePage, error) {
	page, err := s.pageRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: Implement link extraction and queue child pages

	return page, nil
}

type CrawlProgress struct {
	CollectionID uuid.UUID        `json:"collection_id"`
	Total        int64            `json:"total"`
	Pending      int64            `json:"pending"`
	Crawling     int64            `json:"crawling"`
	Success      int64            `json:"success"`
	Failed       int64            `json:"failed"`
	Progress     map[string]int64 `json:"progress"`
}

func (s *WebsiteService) GetProgress(ctx context.Context, collectionID uuid.UUID) (*CrawlProgress, error) {
	progress, err := s.pageRepo.GetCrawlProgress(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	var total int64
	for _, count := range progress {
		total += count
	}

	return &CrawlProgress{
		CollectionID: collectionID,
		Total:        total,
		Pending:      progress["pending"],
		Crawling:     progress["crawling"],
		Success:      progress["success"],
		Failed:       progress["failed"],
		Progress:     progress,
	}, nil
}
