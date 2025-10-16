package http

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

func RateLimitTokenEndpoint(rdb *redis.Client, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip == "" {
				ip = r.Header.Get("X-Forwarded-For")
			}
			if ip == "" {
				ip = "unknown"
			}

			ctx := context.Background()
			key := "rl:token:" + ip
			now := time.Now().UnixNano()
			cut := now - window.Nanoseconds()

			pipe := rdb.TxPipeline()
			pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprint(cut))
			pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
			pipe.Expire(ctx, key, window)
			pipe.ZCard(ctx, key)
			cmds, err := pipe.Exec(ctx)
			if err != nil {
				http.Error(w, "rate limit error", http.StatusInternalServerError)
				return
			}
			zcard := cmds[len(cmds)-1].(*redis.IntCmd).Val()
			if int(zcard) > limit {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
				http.Error(w, "rate_limit_exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
