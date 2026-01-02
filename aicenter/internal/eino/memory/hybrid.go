package memory

import (
	"context"
	"log"

	"github.com/cloudwego/eino/schema"
)

// HybridStore combines Redis (hot cache) with PostgreSQL (persistent storage)
// Read: Redis first, fallback to PostgreSQL
// Write: Both Redis and PostgreSQL
type HybridStore struct {
	redis    *RedisStore
	postgres *PostgresStore
}

// NewHybridStore creates a hybrid store with Redis cache and PostgreSQL persistence
func NewHybridStore(redis *RedisStore, postgres *PostgresStore) *HybridStore {
	return &HybridStore{
		redis:    redis,
		postgres: postgres,
	}
}

// Write stores messages to both Redis and PostgreSQL
func (s *HybridStore) Write(ctx context.Context, sessionID string, msgs []*schema.Message) error {
	// Write to PostgreSQL first (authoritative source)
	if err := s.postgres.Write(ctx, sessionID, msgs); err != nil {
		return err
	}

	// Write to Redis (cache, ignore errors)
	if s.redis != nil {
		if err := s.redis.Write(ctx, sessionID, msgs); err != nil {
			log.Printf("[HybridStore] Redis write failed (non-fatal): %v", err)
		}
	}

	return nil
}

// Read returns messages, trying Redis first then PostgreSQL
func (s *HybridStore) Read(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	// Try Redis first
	if s.redis != nil {
		msgs, err := s.redis.Read(ctx, sessionID)
		if err == nil && len(msgs) > 0 {
			return msgs, nil
		}
		if err != nil {
			log.Printf("[HybridStore] Redis read failed, falling back to PostgreSQL: %v", err)
		}
	}

	// Fallback to PostgreSQL
	msgs, err := s.postgres.Read(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Warm up Redis cache if we got data from PostgreSQL
	if s.redis != nil && len(msgs) > 0 {
		if err := s.redis.Write(ctx, sessionID, msgs); err != nil {
			log.Printf("[HybridStore] Redis cache warm-up failed (non-fatal): %v", err)
		}
	}

	return msgs, nil
}

// Append adds messages to both stores
func (s *HybridStore) Append(ctx context.Context, sessionID string, msgs ...*schema.Message) error {
	if len(msgs) == 0 {
		return nil
	}

	// Append to PostgreSQL first
	if err := s.postgres.Append(ctx, sessionID, msgs...); err != nil {
		return err
	}

	// Append to Redis (best effort)
	if s.redis != nil {
		if err := s.redis.Append(ctx, sessionID, msgs...); err != nil {
			log.Printf("[HybridStore] Redis append failed (non-fatal): %v", err)
		}
	}

	return nil
}

// Delete removes a session from both stores
func (s *HybridStore) Delete(ctx context.Context, sessionID string) error {
	// Delete from Redis first
	if s.redis != nil {
		if err := s.redis.Delete(ctx, sessionID); err != nil {
			log.Printf("[HybridStore] Redis delete failed (non-fatal): %v", err)
		}
	}

	// Delete from PostgreSQL
	return s.postgres.Delete(ctx, sessionID)
}

// RefreshCache refreshes the Redis TTL for a session
func (s *HybridStore) RefreshCache(ctx context.Context, sessionID string) error {
	if s.redis != nil {
		return s.redis.Refresh(ctx, sessionID)
	}
	return nil
}

// GetWindowedMessages returns the last N messages, using Redis if available
func (s *HybridStore) GetWindowedMessages(ctx context.Context, sessionID string, windowSize int) ([]*schema.Message, error) {
	msgs, err := s.Read(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if len(msgs) <= windowSize {
		return msgs, nil
	}

	return msgs[len(msgs)-windowSize:], nil
}
