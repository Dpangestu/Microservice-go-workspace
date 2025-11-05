package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	userBundleCachePrefix = "user:bundle:"
	userBundleTTL         = 5 * time.Minute
)

// CacheUserBundle menyimpan user bundle ke Redis
func (s *UserService) CacheUserBundle(ctx context.Context, userID string, data map[string]interface{}) error {
	if s.RedisClient == nil {
		return nil // Skip jika Redis tidak tersedia
	}

	key := fmt.Sprintf("%s%s", userBundleCachePrefix, userID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[UserService] Error marshaling cache data: %v", err)
		return nil
	}

	err = s.RedisClient.Set(ctx, key, jsonData, userBundleTTL).Err()
	if err != nil {
		log.Printf("[UserService] Error caching user bundle for %s: %v", userID, err)
		return nil // Non-blocking: jangan fail
	}

	log.Printf("[UserService] Cached user bundle for %s (TTL: %v)", userID, userBundleTTL)
	return nil
}

// GetCachedUserBundle mengambil user bundle dari Redis
func (s *UserService) GetCachedUserBundle(ctx context.Context, userID string) (map[string]interface{}, error) {
	if s.RedisClient == nil {
		return nil, nil // Skip jika Redis tidak tersedia
	}

	key := fmt.Sprintf("%s%s", userBundleCachePrefix, userID)
	val, err := s.RedisClient.Get(ctx, key).Result()

	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		log.Printf("[UserService] Error getting cache for %s: %v", userID, err)
		return nil, nil // Fallback ke database
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		log.Printf("[UserService] Error unmarshaling cache: %v", err)
		return nil, nil // Fallback ke database
	}

	log.Printf("[UserService] Cache hit for user %s", userID)
	return data, nil
}

// InvalidateUserCache menghapus cache user (saat profile update)
func (s *UserService) InvalidateUserCache(ctx context.Context, userID string) error {
	if s.RedisClient == nil {
		return nil
	}

	key := fmt.Sprintf("%s%s", userBundleCachePrefix, userID)
	err := s.RedisClient.Del(ctx, key).Err()
	if err != nil {
		log.Printf("[UserService] Error invalidating cache for %s: %v", userID, err)
		return nil // Non-blocking
	}

	log.Printf("[UserService] Invalidated cache for user %s", userID)
	return nil
}
