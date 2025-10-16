package cache

import (
	"github.com/redis/go-redis/v9"
)

type RedisCfg struct {
	Addr     string
	Password string
	DB       int
}

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

func NewRedisCluster(cfg RedisClusterCfg) *redis.ClusterClient {
	return redis.NewClusterClient(&redis.ClusterOptions{Addrs: cfg.Addrs, Password: cfg.Password})
}
