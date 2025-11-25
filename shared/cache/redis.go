package cache

import (
	"github.com/redis/go-redis/v9"
)

type RedisCfg struct {
	Addr     string
	Password string
	DB       int
}

// NewRedis mengembalikan *redis.Client langsung dari library resmi.
// Fallback/degraded mode DIHANDLE di level service (main.go) via Ping() + logging.
// Package ini tidak melakukan ping atau panic, supaya bisa dipakai fleksibel.
func NewRedis(cfg RedisCfg) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

type RedisClusterCfg struct {
	Addrs    []string
	Password string
}

// NewRedisCluster mengembalikan *redis.ClusterClient untuk mode cluster.
// Sama seperti NewRedis, health check dilakukan oleh pemanggil (service main.go).
func NewRedisCluster(cfg RedisClusterCfg) *redis.ClusterClient {
	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    cfg.Addrs,
		Password: cfg.Password,
	})
}
