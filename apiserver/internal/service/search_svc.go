package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/repository"
)

type SearchService struct {
	visitorRepo *repository.VisitorRepository
}

func NewSearchService(visitorRepo *repository.VisitorRepository) *SearchService {
	return &SearchService{visitorRepo: visitorRepo}
}

type SearchScope string

const (
	SearchScopeAll      SearchScope = "all"
	SearchScopeVisitors SearchScope = "visitors"
	SearchScopeMessages SearchScope = "messages"
)

type SearchResult struct {
	Query             string           `json:"query"`
	Scope             SearchScope      `json:"scope"`
	Visitors          []model.Visitor  `json:"visitors"`
	VisitorCount      int              `json:"visitor_count"`
	VisitorPagination SearchPagination `json:"visitor_pagination"`
	Messages          []interface{}    `json:"messages"`
	MessageCount      int              `json:"message_count"`
	MessagePagination SearchPagination `json:"message_pagination"`
}

type SearchPagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

func (s *SearchService) Search(ctx context.Context, projectID uuid.UUID, query string, scope SearchScope, visitorPage, visitorPageSize, messagePage, messagePageSize int) (*SearchResult, error) {
	result := &SearchResult{
		Query:    query,
		Scope:    scope,
		Visitors: []model.Visitor{},
		Messages: []interface{}{},
	}

	// Search visitors
	if scope == SearchScopeAll || scope == SearchScopeVisitors {
		offset := (visitorPage - 1) * visitorPageSize
		visitors, total, err := s.visitorRepo.Search(ctx, projectID, query, visitorPageSize, offset)
		if err != nil {
			return nil, err
		}

		result.Visitors = visitors
		result.VisitorCount = len(visitors)
		result.VisitorPagination = SearchPagination{
			Page:     visitorPage,
			PageSize: visitorPageSize,
			Total:    total,
		}
	} else {
		result.VisitorPagination = SearchPagination{Page: visitorPage, PageSize: visitorPageSize, Total: 0}
	}

	// Messages search would be done via WuKongIM - placeholder for now
	result.MessagePagination = SearchPagination{Page: messagePage, PageSize: messagePageSize, Total: 0}

	return result, nil
}
