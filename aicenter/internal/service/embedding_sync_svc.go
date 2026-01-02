package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/aicenter/internal/model"
	"github.com/tgo/captain/aicenter/pkg/external/rag"
)

// EmbeddingSyncStatus represents sync status
type EmbeddingSyncStatus string

const (
	SyncStatusPending EmbeddingSyncStatus = "pending"
	SyncStatusSuccess EmbeddingSyncStatus = "success"
	SyncStatusFailed  EmbeddingSyncStatus = "failed"
)

// EmbeddingConfig represents embedding configuration to sync
type EmbeddingConfig struct {
	ProjectID uuid.UUID `json:"project_id"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	APIKey    string    `json:"api_key"`
	BaseURL   string    `json:"base_url,omitempty"`
	IsActive  bool      `json:"is_active"`
}

// EmbeddingSyncService handles embedding config sync to RAG service
type EmbeddingSyncService struct {
	db        *gorm.DB
	ragClient *rag.Client
	mu        sync.Mutex
	logger    *slog.Logger
}

// NewEmbeddingSyncService creates a new embedding sync service
func NewEmbeddingSyncService(db *gorm.DB, ragURL string) *EmbeddingSyncService {
	return &EmbeddingSyncService{
		db:        db,
		ragClient: rag.NewClient(ragURL),
		logger:    slog.Default().With("service", "embedding_sync"),
	}
}

// BuildEmbeddingConfigs builds embedding configs from ProjectAIConfig records
func (s *EmbeddingSyncService) BuildEmbeddingConfigs(ctx context.Context, projectIDs []uuid.UUID) ([]EmbeddingConfig, error) {
	var configs []model.ProjectAIConfig
	if err := s.db.WithContext(ctx).Where("project_id IN ?", projectIDs).Find(&configs).Error; err != nil {
		return nil, err
	}

	var result []EmbeddingConfig
	for _, cfg := range configs {
		if cfg.DefaultEmbeddingProviderID == nil || cfg.DefaultEmbeddingModel == "" {
			continue
		}

		// Get provider
		var provider model.LLMProvider
		if err := s.db.WithContext(ctx).First(&provider, "id = ?", cfg.DefaultEmbeddingProviderID).Error; err != nil {
			s.logger.Warn("embedding provider not found",
				"project_id", cfg.ProjectID,
				"provider_id", cfg.DefaultEmbeddingProviderID)
			continue
		}

		if !provider.IsActive {
			s.logger.Warn("embedding provider inactive",
				"project_id", cfg.ProjectID,
				"provider_id", provider.ID)
			continue
		}

		// Map provider kind to RAG provider
		ragProvider := mapProviderForRAG(provider.ProviderKind, provider.Vendor)
		if ragProvider == "" {
			s.logger.Warn("unsupported provider for RAG",
				"project_id", cfg.ProjectID,
				"provider_kind", provider.ProviderKind)
			continue
		}

		result = append(result, EmbeddingConfig{
			ProjectID: cfg.ProjectID,
			Provider:  ragProvider,
			Model:     cfg.DefaultEmbeddingModel,
			APIKey:    provider.APIKey,
			BaseURL:   provider.APIBaseURL,
			IsActive:  true,
		})
	}

	return result, nil
}

// mapProviderForRAG maps internal provider kind to RAG provider
func mapProviderForRAG(providerKind, vendor string) string {
	switch providerKind {
	case "openai":
		return "openai"
	case "openai_compatible", "openai-compatible":
		if vendor == "qwen3" || vendor == "qwen" {
			return "qwen3"
		}
	}
	return ""
}

// SyncWithRetry syncs embedding configs to RAG with retries
func (s *EmbeddingSyncService) SyncWithRetry(ctx context.Context, configs []EmbeddingConfig, maxRetries int, baseDelay time.Duration) error {
	if len(configs) == 0 {
		return nil
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := s.syncToRAG(ctx, configs)
		if err == nil {
			s.logger.Info("embedding configs synced to RAG",
				"count", len(configs))
			return nil
		}

		lastErr = err
		if attempt >= maxRetries {
			break
		}

		delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
		s.logger.Warn("embedding sync attempt failed, retrying",
			"attempt", attempt,
			"delay", delay,
			"error", err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	s.logger.Error("embedding sync failed after retries",
		"attempts", maxRetries,
		"error", lastErr)
	return lastErr
}

// syncToRAG performs the actual sync to RAG service
func (s *EmbeddingSyncService) syncToRAG(ctx context.Context, configs []EmbeddingConfig) error {
	for _, cfg := range configs {
		err := s.ragClient.SyncEmbeddingConfig(ctx, &rag.EmbeddingConfigRequest{
			ProjectID: cfg.ProjectID.String(),
			Provider:  cfg.Provider,
			Model:     cfg.Model,
			APIKey:    cfg.APIKey,
			BaseURL:   cfg.BaseURL,
			IsActive:  cfg.IsActive,
		})
		if err != nil {
			return fmt.Errorf("sync embedding config for project %s: %w", cfg.ProjectID, err)
		}
	}
	return nil
}

// FireAndForgetSync schedules background sync with retry and status updates
func (s *EmbeddingSyncService) FireAndForgetSync(projectIDs []uuid.UUID) {
	go func() {
		ctx := context.Background()
		s.mu.Lock()
		defer s.mu.Unlock()

		if len(projectIDs) == 0 {
			return
		}

		// Build configs
		configs, err := s.BuildEmbeddingConfigs(ctx, projectIDs)
		if err != nil {
			s.logger.Error("failed to build embedding configs", "error", err)
			return
		}

		if len(configs) == 0 {
			return
		}

		// Mark as pending
		s.updateSyncStatus(ctx, projectIDs, SyncStatusPending, "")

		// Sync with retry
		maxRetries := 3
		baseDelay := time.Second

		for attempt := 1; attempt <= maxRetries; attempt++ {
			// Increment attempt count
			s.incrementAttemptCount(ctx, projectIDs)

			err := s.syncToRAG(ctx, configs)
			if err == nil {
				// Success
				s.updateSyncStatus(ctx, projectIDs, SyncStatusSuccess, "")
				s.updateLastSyncAt(ctx, projectIDs)
				s.logger.Info("embedding configs synced to RAG", "count", len(configs))
				return
			}

			if attempt >= maxRetries {
				// Final failure
				s.updateSyncStatus(ctx, projectIDs, SyncStatusFailed, err.Error())
				s.logger.Error("embedding sync failed after retries",
					"attempts", attempt,
					"error", err)
				return
			}

			delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
			s.logger.Warn("embedding sync attempt failed, retrying",
				"attempt", attempt,
				"delay", delay,
				"error", err)
			time.Sleep(delay)
		}
	}()
}

func (s *EmbeddingSyncService) updateSyncStatus(ctx context.Context, projectIDs []uuid.UUID, status EmbeddingSyncStatus, errMsg string) {
	s.db.WithContext(ctx).
		Model(&model.ProjectAIConfig{}).
		Where("project_id IN ?", projectIDs).
		Updates(map[string]interface{}{
			"sync_status": string(status),
			"sync_error":  errMsg,
		})
}

func (s *EmbeddingSyncService) incrementAttemptCount(ctx context.Context, projectIDs []uuid.UUID) {
	s.db.WithContext(ctx).
		Model(&model.ProjectAIConfig{}).
		Where("project_id IN ?", projectIDs).
		UpdateColumn("sync_attempt_count", gorm.Expr("COALESCE(sync_attempt_count, 0) + 1"))
}

func (s *EmbeddingSyncService) updateLastSyncAt(ctx context.Context, projectIDs []uuid.UUID) {
	s.db.WithContext(ctx).
		Model(&model.ProjectAIConfig{}).
		Where("project_id IN ?", projectIDs).
		UpdateColumn("last_sync_at", time.Now())
}

// RetryFailedSyncs retries all failed embedding syncs (called at startup)
func (s *EmbeddingSyncService) RetryFailedSyncs(ctx context.Context) error {
	var configs []model.ProjectAIConfig
	if err := s.db.WithContext(ctx).
		Where("sync_status = ? OR sync_status IS NULL", "failed").
		Where("default_embedding_provider_id IS NOT NULL").
		Where("default_embedding_model != ''").
		Find(&configs).Error; err != nil {
		return err
	}

	if len(configs) == 0 {
		s.logger.Info("no failed embedding syncs to retry")
		return nil
	}

	projectIDs := make([]uuid.UUID, 0, len(configs))
	for _, cfg := range configs {
		projectIDs = append(projectIDs, cfg.ProjectID)
	}

	s.logger.Info("retrying failed embedding syncs", "count", len(projectIDs))
	s.FireAndForgetSync(projectIDs)
	return nil
}
