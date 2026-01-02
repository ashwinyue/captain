package service

import (
	"context"
	"log"
	"time"

	"github.com/tgo/captain/apiserver/internal/model"
	"gorm.io/gorm"
)

// QueueCleanupService handles queue expiry cleanup
type QueueCleanupService struct {
	db              *gorm.DB
	expiryDuration  time.Duration
	cleanupInterval time.Duration
}

// NewQueueCleanupService creates a new queue cleanup service
func NewQueueCleanupService(db *gorm.DB, expiryMinutes, cleanupIntervalMinutes int) *QueueCleanupService {
	if expiryMinutes <= 0 {
		expiryMinutes = 30 // Default 30 minutes expiry
	}
	if cleanupIntervalMinutes <= 0 {
		cleanupIntervalMinutes = 5 // Default 5 minutes cleanup interval
	}
	return &QueueCleanupService{
		db:              db,
		expiryDuration:  time.Duration(expiryMinutes) * time.Minute,
		cleanupInterval: time.Duration(cleanupIntervalMinutes) * time.Minute,
	}
}

// Start starts the background cleanup loop
func (s *QueueCleanupService) Start(ctx context.Context) {
	log.Printf("[QueueCleanup] Starting cleanup service (expiry: %v, interval: %v)", s.expiryDuration, s.cleanupInterval)

	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	// Run initial cleanup
	s.cleanup(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[QueueCleanup] Stopping cleanup service")
			return
		case <-ticker.C:
			s.cleanup(ctx)
		}
	}
}

// cleanup performs the actual cleanup of expired queue entries
func (s *QueueCleanupService) cleanup(ctx context.Context) {
	expiryTime := time.Now().Add(-s.expiryDuration)

	// Find expired waiting queue entries
	var expiredEntries []model.VisitorWaitingQueue
	if err := s.db.WithContext(ctx).
		Where("status = ? AND wait_started_at < ?", model.QueueStatusWaiting, expiryTime).
		Limit(100).
		Find(&expiredEntries).Error; err != nil {
		log.Printf("[QueueCleanup] Error finding expired entries: %v", err)
		return
	}

	if len(expiredEntries) == 0 {
		return
	}

	log.Printf("[QueueCleanup] Processing %d expired entries", len(expiredEntries))

	for _, entry := range expiredEntries {
		// Update queue entry status to timeout
		if err := s.db.WithContext(ctx).Model(&model.VisitorWaitingQueue{}).
			Where("id = ?", entry.ID).
			Update("status", model.QueueStatusTimeout).Error; err != nil {
			log.Printf("[QueueCleanup] Error updating queue entry %s: %v", entry.ID, err)
			continue
		}

		// Update visitor status from QUEUED to CLOSED
		if err := s.db.WithContext(ctx).Model(&model.Visitor{}).
			Where("id = ? AND service_status = ?", entry.VisitorID, model.VisitorStatusQueued).
			Updates(map[string]interface{}{
				"service_status": model.VisitorStatusClosed,
				"ai_enabled":     true, // Re-enable AI
			}).Error; err != nil {
			log.Printf("[QueueCleanup] Error updating visitor %s: %v", entry.VisitorID, err)
			continue
		}

		log.Printf("[QueueCleanup] Expired queue entry %s for visitor %s (waited %d seconds)",
			entry.ID, entry.VisitorID, int(time.Since(entry.WaitStartedAt).Seconds()))
	}

	log.Printf("[QueueCleanup] Cleanup complete, processed %d entries", len(expiredEntries))
}
