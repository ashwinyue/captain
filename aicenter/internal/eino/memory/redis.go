package memory

import (
	"context"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
)

// RedisStore persists messages in Redis with optional TTL
type RedisStore struct {
	cli *redis.Client
	ttl time.Duration // 0 means no expiration
}

// RedisStoreConfig configures the Redis store
type RedisStoreConfig struct {
	// Client is the Redis client
	Client *redis.Client
	// TTL is the expiration time for session keys (0 = no expiration)
	TTL time.Duration
}

// NewRedisStore creates a Redis-backed memory store
func NewRedisStore(cfg *RedisStoreConfig) *RedisStore {
	ttl := cfg.TTL
	if ttl == 0 {
		ttl = 30 * time.Minute // Default 30 minutes
	}
	return &RedisStore{
		cli: cfg.Client,
		ttl: ttl,
	}
}

// NewRedisStoreFromURL creates a Redis store from connection URL
func NewRedisStoreFromURL(redisURL string, ttl time.Duration) (*RedisStore, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	cli := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cli.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return NewRedisStore(&RedisStoreConfig{
		Client: cli,
		TTL:    ttl,
	}), nil
}

func (s *RedisStore) sessionKey(sessionID string) string {
	return "memory:session:" + sessionID
}

// Write stores messages for a session (replaces existing)
func (s *RedisStore) Write(ctx context.Context, sessionID string, msgs []*schema.Message) error {
	b, err := EncodeMessages(msgs)
	if err != nil {
		return err
	}
	return s.cli.Set(ctx, s.sessionKey(sessionID), b, s.ttl).Err()
}

// Read returns messages for a session
func (s *RedisStore) Read(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	res, err := s.cli.Get(ctx, s.sessionKey(sessionID)).Bytes()
	if err == redis.Nil {
		return nil, nil // Session not found
	}
	if err != nil {
		return nil, err
	}
	return DecodeMessages(res)
}

// Append adds messages to a session
func (s *RedisStore) Append(ctx context.Context, sessionID string, msgs ...*schema.Message) error {
	if len(msgs) == 0 {
		return nil
	}

	// Read existing messages
	existing, err := s.Read(ctx, sessionID)
	if err != nil {
		return err
	}

	// Append new messages
	existing = append(existing, msgs...)

	// Write back
	return s.Write(ctx, sessionID, existing)
}

// Delete removes a session's messages
func (s *RedisStore) Delete(ctx context.Context, sessionID string) error {
	return s.cli.Del(ctx, s.sessionKey(sessionID)).Err()
}

// Refresh extends the TTL for a session
func (s *RedisStore) Refresh(ctx context.Context, sessionID string) error {
	return s.cli.Expire(ctx, s.sessionKey(sessionID), s.ttl).Err()
}

// GetTTL returns the remaining TTL for a session
func (s *RedisStore) GetTTL(ctx context.Context, sessionID string) (time.Duration, error) {
	return s.cli.TTL(ctx, s.sessionKey(sessionID)).Result()
}

// Close closes the Redis client connection
func (s *RedisStore) Close() error {
	return s.cli.Close()
}
