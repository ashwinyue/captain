package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/pkg/redis"
	"github.com/tgo/captain/apiserver/internal/pkg/wukongim"
	"github.com/tgo/captain/apiserver/internal/repository"
)

const (
	// DefaultHumanSessionTimeout is the default timeout for human sessions
	DefaultHumanSessionTimeout = 5 * time.Minute

	// SessionCheckInterval is how often to check for expired sessions
	SessionCheckInterval = 30 * time.Second
)

// HumanSessionService manages human agent session timeouts
type HumanSessionService struct {
	sessionManager *redis.HumanSessionManager
	visitorRepo    *repository.VisitorRepository
	wukongClient   *wukongim.Client
	timeout        time.Duration
}

// NewHumanSessionService creates a new human session service
func NewHumanSessionService(
	redisClient *redis.Client,
	visitorRepo *repository.VisitorRepository,
	wukongClient *wukongim.Client,
	timeout time.Duration,
) *HumanSessionService {
	if timeout == 0 {
		timeout = DefaultHumanSessionTimeout
	}

	sessionManager := redis.NewHumanSessionManager(redisClient, timeout)

	svc := &HumanSessionService{
		sessionManager: sessionManager,
		visitorRepo:    visitorRepo,
		wukongClient:   wukongClient,
		timeout:        timeout,
	}

	return svc
}

// Start begins monitoring for expired sessions
func (s *HumanSessionService) Start(ctx context.Context) {
	// Start the expiration check loop
	go s.expirationCheckLoop(ctx)
	log.Printf("[HumanSessionService] Started with timeout: %v", s.timeout)
}

// OnVisitorTransferToHuman is called when a visitor is transferred to human agent
func (s *HumanSessionService) OnVisitorTransferToHuman(ctx context.Context, visitorID, staffID uuid.UUID) error {
	log.Printf("[HumanSessionService] Starting human session for visitor %s with staff %s", visitorID, staffID)
	return s.sessionManager.StartSession(ctx, visitorID, staffID)
}

// OnVisitorMessage is called when a visitor sends a message (refreshes TTL)
func (s *HumanSessionService) OnVisitorMessage(ctx context.Context, visitorID uuid.UUID) error {
	// Check if visitor is in human session
	inSession, err := s.sessionManager.IsInHumanSession(ctx, visitorID)
	if err != nil {
		return err
	}
	if !inSession {
		return nil // Not in human session, nothing to refresh
	}

	log.Printf("[HumanSessionService] Refreshing session for visitor %s", visitorID)
	return s.sessionManager.RefreshSession(ctx, visitorID)
}

// OnVisitorEndHumanSession is called when a human session ends manually
func (s *HumanSessionService) OnVisitorEndHumanSession(ctx context.Context, visitorID uuid.UUID) error {
	log.Printf("[HumanSessionService] Ending human session for visitor %s", visitorID)
	return s.sessionManager.EndSession(ctx, visitorID)
}

// IsInHumanSession checks if a visitor is in a human session
func (s *HumanSessionService) IsInHumanSession(ctx context.Context, visitorID uuid.UUID) (bool, error) {
	return s.sessionManager.IsInHumanSession(ctx, visitorID)
}

// expirationCheckLoop periodically checks for visitors who should be disconnected
func (s *HumanSessionService) expirationCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(SessionCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[HumanSessionService] Expiration check loop stopped")
			return
		case <-ticker.C:
			s.checkAndDisconnectExpiredSessions(ctx)
		}
	}
}

// checkAndDisconnectExpiredSessions checks DB for visitors marked as human but not in Redis
func (s *HumanSessionService) checkAndDisconnectExpiredSessions(ctx context.Context) {
	// Get all visitors currently marked as in human mode from DB
	visitors, err := s.visitorRepo.FindByServiceStatus(ctx, "active")
	if err != nil {
		log.Printf("[HumanSessionService] Failed to get active visitors: %v", err)
		return
	}

	for _, visitor := range visitors {
		// Check if they have an active Redis session
		inSession, err := s.sessionManager.IsInHumanSession(ctx, visitor.ID)
		if err != nil {
			log.Printf("[HumanSessionService] Failed to check session for visitor %s: %v", visitor.ID, err)
			continue
		}

		// If not in Redis but marked as active in DB, disconnect them
		if !inSession && !visitor.AIEnabled {
			log.Printf("[HumanSessionService] Disconnecting expired session for visitor %s", visitor.ID)
			s.disconnectVisitor(ctx, visitor.ID, visitor.ProjectID)
		}
	}
}

// disconnectVisitor disconnects a visitor from human mode and notifies them
func (s *HumanSessionService) disconnectVisitor(ctx context.Context, visitorID, projectID uuid.UUID) {
	// Update DB to reset to AI mode
	if err := s.visitorRepo.ResetToAIMode(ctx, visitorID); err != nil {
		log.Printf("[HumanSessionService] Failed to reset visitor %s to AI mode: %v", visitorID, err)
		return
	}

	// Send notification to visitor via WuKongIM
	s.sendDisconnectNotification(ctx, visitorID, projectID)

	log.Printf("[HumanSessionService] Visitor %s disconnected from human mode", visitorID)
}

// sendDisconnectNotification sends a notification to the visitor that they've been disconnected
func (s *HumanSessionService) sendDisconnectNotification(ctx context.Context, visitorID, projectID uuid.UUID) {
	if s.wukongClient == nil {
		log.Printf("[HumanSessionService] WuKongIM client not configured, skipping notification")
		return
	}

	// Get visitor's channel info
	visitor, err := s.visitorRepo.FindByID(ctx, visitorID)
	if err != nil {
		log.Printf("[HumanSessionService] Failed to get visitor %s: %v", visitorID, err)
		return
	}

	// Construct the notification message
	message := "由于长时间未回复，已切换为 AI 智能客服为您服务。如需人工服务，请说\"转人工\"。"

	// Send via WuKongIM
	channelID := visitor.ExternalID // The visitor's channel ID
	if channelID == "" {
		log.Printf("[HumanSessionService] Visitor %s has no external_id, skipping notification", visitorID)
		return
	}

	// Send system message
	err = s.wukongClient.SendSystemMessage(ctx, channelID, 251, message)
	if err != nil {
		log.Printf("[HumanSessionService] Failed to send disconnect notification to visitor %s: %v", visitorID, err)
		return
	}

	log.Printf("[HumanSessionService] Sent disconnect notification to visitor %s", visitorID)
}
