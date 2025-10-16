package session

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Manager struct {
	rdb *redis.Client
}

func NewManager(rdb *redis.Client) *Manager {
	return &Manager{rdb: rdb}
}

func (m *Manager) key(userID, token string) string {
	return fmt.Sprintf("session:%s:%s", userID, token)
}

// Create session (store active session in Redis)
func (m *Manager) Create(ctx context.Context, userID, token string, ttl time.Duration) error {
	key := m.key(userID, token)
	return m.rdb.Set(ctx, key, "active", ttl).Err()
}

// Check if a session is active
func (m *Manager) IsActive(ctx context.Context, userID, token string) bool {
	key := m.key(userID, token)
	val, err := m.rdb.Get(ctx, key).Result()
	return err == nil && val == "active"
}

// Revoke a single session
func (m *Manager) Revoke(ctx context.Context, userID, token string) error {
	key := m.key(userID, token)
	return m.rdb.Del(ctx, key).Err()
}

// Revoke all sessions for a user
func (m *Manager) RevokeAll(ctx context.Context, userID string) error {
	iter := m.rdb.Scan(ctx, 0, fmt.Sprintf("session:%s:*", userID), 0).Iterator()
	for iter.Next(ctx) {
		_ = m.rdb.Del(ctx, iter.Val()).Err()
	}
	return iter.Err()
}
