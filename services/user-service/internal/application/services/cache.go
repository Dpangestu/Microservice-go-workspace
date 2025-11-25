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
func (s *userServiceImpl) CacheUserBundle(ctx context.Context, userID string, data map[string]interface{}) error {
	if s.RedisClient == nil {
		return nil // Skip jika Redis tidak tersedia
	}

	key := userBundleCacheKey(userID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[UserService] Error marshaling cache data: %v", err)
		return err
	}

	err = s.RedisClient.Set(ctx, key, jsonData, userBundleTTL).Err()
	if err != nil {
		log.Printf("[UserService] Error caching user bundle for %s: %v", userID, err)
		// Jangan fail jika cache error, tetap return data
		return nil
	}

	log.Printf("[UserService] Cached user bundle for %s (TTL: %v)", userID, userBundleTTL)
	return nil
}

// GetCachedUserBundle mengambil user bundle dari Redis
func (s *userServiceImpl) GetCachedUserBundle(ctx context.Context, userID string) (map[string]interface{}, error) {
	if s.RedisClient == nil {
		return nil, nil // Skip jika Redis tidak tersedia
	}

	key := userBundleCacheKey(userID)
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

// InvalidateUserCache menghapus cache user (dipanggil saat profile update)
func (s *userServiceImpl) InvalidateUserCache(ctx context.Context, userID string) error {
	if s.RedisClient == nil {
		return nil
	}

	key := userBundleCacheKey(userID)
	err := s.RedisClient.Del(ctx, key).Err()
	if err != nil {
		log.Printf("[UserService] Error invalidating cache for %s: %v", userID, err)
		return nil // Tidak fail jika invalidate gagal
	}

	log.Printf("[UserService] Invalidated cache for user %s", userID)
	return nil
}

func userBundleCacheKey(userID string) string {
	return fmt.Sprintf("%s%s", userBundleCachePrefix, userID)
}
