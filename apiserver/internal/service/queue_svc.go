package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/repository"
)

type QueueService struct {
	repo *repository.QueueRepository
}

func NewQueueService(repo *repository.QueueRepository) *QueueService {
	return &QueueService{repo: repo}
}

func (s *QueueService) ListWaiting(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.VisitorWaitingQueue, int64, error) {
	return s.repo.FindWaiting(ctx, projectID, limit, offset)
}

func (s *QueueService) AddToQueue(ctx context.Context, item *model.VisitorWaitingQueue) error {
	item.Status = model.QueueStatusWaiting
	item.WaitStartedAt = time.Now()
	return s.repo.Create(ctx, item)
}

func (s *QueueService) AssignToStaff(ctx context.Context, projectID, queueID, staffID uuid.UUID) error {
	return s.repo.Assign(ctx, projectID, queueID, staffID)
}

func (s *QueueService) RemoveFromQueue(ctx context.Context, projectID, queueID uuid.UUID, status model.QueueStatus) error {
	return s.repo.UpdateStatus(ctx, projectID, queueID, status)
}

func (s *QueueService) GetQueuePosition(ctx context.Context, projectID, queueID uuid.UUID) (int, error) {
	pos, err := s.repo.GetPosition(ctx, projectID, queueID)
	return int(pos), err
}

func (s *QueueService) GetQueueCount(ctx context.Context, projectID uuid.UUID) (waiting, assigned int64, err error) {
	return s.repo.GetCount(ctx, projectID)
}
