package redis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// HumanSessionKeyPrefix is the prefix for human session keys
	HumanSessionKeyPrefix = "human_session:"

	// DefaultSessionTimeout is the default timeout for human sessions (5 minutes)
	DefaultSessionTimeout = 5 * time.Minute
)

// HumanSessionManager manages human agent sessions with timeout
type HumanSessionManager struct {
	client         *Client
	sessionTimeout time.Duration
	onExpireFunc   func(ctx context.Context, visitorID uuid.UUID) error
}

// NewHumanSessionManager creates a new human session manager
func NewHumanSessionManager(client *Client, timeout time.Duration) *HumanSessionManager {
	if timeout == 0 {
		timeout = DefaultSessionTimeout
	}
	return &HumanSessionManager{
		client:         client,
		sessionTimeout: timeout,
	}
}

// SetOnExpireCallback sets the callback function to be called when a session expires
func (m *HumanSessionManager) SetOnExpireCallback(fn func(ctx context.Context, visitorID uuid.UUID) error) {
	m.onExpireFunc = fn
}

// StartSession starts a human session for a visitor
func (m *HumanSessionManager) StartSession(ctx context.Context, visitorID uuid.UUID, staffID uuid.UUID) error {
	key := m.getKey(visitorID)
	value := staffID.String()
	return m.client.Set(ctx, key, value, m.sessionTimeout)
}

// RefreshSession refreshes the TTL for an existing human session
func (m *HumanSessionManager) RefreshSession(ctx context.Context, visitorID uuid.UUID) error {
	key := m.getKey(visitorID)
	exists, err := m.client.Exists(ctx, key)
	if err != nil {
		return err
	}
	if exists == 0 {
		return nil // Session doesn't exist, nothing to refresh
	}
	return m.client.Expire(ctx, key, m.sessionTimeout)
}

// EndSession ends a human session
func (m *HumanSessionManager) EndSession(ctx context.Context, visitorID uuid.UUID) error {
	key := m.getKey(visitorID)
	return m.client.Del(ctx, key)
}

// IsInHumanSession checks if a visitor is currently in a human session
func (m *HumanSessionManager) IsInHumanSession(ctx context.Context, visitorID uuid.UUID) (bool, error) {
	key := m.getKey(visitorID)
	exists, err := m.client.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// GetStaffID gets the assigned staff ID for a visitor's human session
func (m *HumanSessionManager) GetStaffID(ctx context.Context, visitorID uuid.UUID) (uuid.UUID, error) {
	key := m.getKey(visitorID)
	value, err := m.client.Get(ctx, key)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(value)
}

// StartExpirationMonitor starts a background goroutine to monitor and handle expired sessions
// This uses polling instead of Redis keyspace notifications for simplicity
func (m *HumanSessionManager) StartExpirationMonitor(ctx context.Context, checkInterval time.Duration) {
	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("[HumanSessionManager] Expiration monitor stopped")
				return
			case <-ticker.C:
				m.checkExpiredSessions(ctx)
			}
		}
	}()
	log.Printf("[HumanSessionManager] Expiration monitor started with interval: %v", checkInterval)
}

// checkExpiredSessions checks for sessions that should have expired but still exist in DB
// This is a fallback mechanism; the main expiration is handled by Redis TTL
func (m *HumanSessionManager) checkExpiredSessions(ctx context.Context) {
	// This method is called periodically to handle any edge cases
	// The actual expiration is handled by Redis TTL, but we can do cleanup here if needed
}

// getKey generates the Redis key for a visitor's human session
func (m *HumanSessionManager) getKey(visitorID uuid.UUID) string {
	return fmt.Sprintf("%s%s", HumanSessionKeyPrefix, visitorID.String())
}

// GetAllActiveSessions returns all active human session visitor IDs
func (m *HumanSessionManager) GetAllActiveSessions(ctx context.Context) ([]uuid.UUID, error) {
	pattern := HumanSessionKeyPrefix + "*"
	keys, err := m.client.Keys(ctx, pattern)
	if err != nil {
		return nil, err
	}

	var visitorIDs []uuid.UUID
	for _, key := range keys {
		idStr := strings.TrimPrefix(key, HumanSessionKeyPrefix)
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		visitorIDs = append(visitorIDs, id)
	}
	return visitorIDs, nil
}
