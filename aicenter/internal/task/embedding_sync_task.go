package task

import (
	"context"

	"github.com/tgo/captain/aicenter/internal/service"
)

// EmbeddingSyncRetryTask retries failed embedding syncs
type EmbeddingSyncRetryTask struct {
	syncService *service.EmbeddingSyncService
}

// NewEmbeddingSyncRetryTask creates a new embedding sync retry task
func NewEmbeddingSyncRetryTask(syncService *service.EmbeddingSyncService) *EmbeddingSyncRetryTask {
	return &EmbeddingSyncRetryTask{
		syncService: syncService,
	}
}

func (t *EmbeddingSyncRetryTask) Name() string {
	return "embedding_sync_retry"
}

func (t *EmbeddingSyncRetryTask) Run(ctx context.Context) error {
	return t.syncService.RetryFailedSyncs(ctx)
}
