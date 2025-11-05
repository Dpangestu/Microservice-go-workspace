package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	meEndpointRateLimitPrefix = "rl:user:me:"
)

func RateLimitTokenEndpoint(rdb *redis.Client, limit int, window time.Duration) func(http.Handler) http.Handler {
	return RateLimitSlidingWindow(rdb, "rl:token", limit, window)
}

func RateLimitSlidingWindow(rdb *redis.Client, prefix string, limit int, window time.Duration) func(http.Handler) http.Handler {
	if rdb == nil || limit <= 0 || window <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if ip == "" {
				ip = "unknown"
			}
			now := time.Now()
			nowNS := now.UnixNano()
			cut := nowNS - window.Nanoseconds()

			key := fmt.Sprintf("%s:%s", prefix, ip)

			ctx := r.Context()
			pipe := rdb.TxPipeline()
			pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprint(cut))
			pipe.ZAdd(ctx, key, redis.Z{Score: float64(nowNS), Member: nowNS})
			pipe.Expire(ctx, key, window)
			cardCmd := pipe.ZCard(ctx, key)

			if _, err := pipe.Exec(ctx); err != nil {
				http.Error(w, "rate limit error", http.StatusInternalServerError)
				return
			}

			count := int(cardCmd.Val())
			remaining := limit - count
			if remaining < 0 {
				remaining = 0
			}
			w.Header().Set("X-RateLimit-Limit", fmt.Sprint(limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprint(remaining))
			w.Header().Set("X-RateLimit-Window", window.String())

			if count > limit {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
				http.Error(w, "rate_limit_exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RateLimitTokenPerClient(rdb *redis.Client, limit int, window time.Duration) func(http.Handler) http.Handler {
	if rdb == nil || limit <= 0 || window <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = r.ParseForm()
			clientID := r.FormValue("client_id")
			if strings.TrimSpace(clientID) == "" {
				ip, _, _ := net.SplitHostPort(r.RemoteAddr)
				if ip == "" {
					ip = r.Header.Get("X-Forwarded-For")
				}
				if ip == "" {
					ip = "unknown"
				}
				clientID = "ip:" + ip
			}

			ctx := r.Context()
			key := "rl:token:client:" + clientID
			now := time.Now().UnixNano()
			cut := now - window.Nanoseconds()

			pipe := rdb.TxPipeline()
			pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprint(cut))
			pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
			pipe.Expire(ctx, key, window)
			card := pipe.ZCard(ctx, key)

			if _, err := pipe.Exec(ctx); err != nil {
				http.Error(w, "rate_limit_error", http.StatusInternalServerError)
				return
			}
			if int(card.Val()) > limit {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
				http.Error(w, "rate_limit_exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if p := strings.Split(xff, ","); len(p) > 0 {
			return strings.TrimSpace(p[0])
		}
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return strings.TrimSpace(xr)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateLimitUserMeEndpoint middleware untuk rate limit /user/me per user
func RateLimitUserMeEndpoint(rdb *redis.Client, limit int, window time.Duration) func(http.Handler) http.Handler {
	if rdb == nil || limit <= 0 || window <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Ambil user ID dari header (set by gateway)
			userID := r.Header.Get("X-User-Id")
			if userID == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			key := fmt.Sprintf("%s%s", meEndpointRateLimitPrefix, userID)

			now := time.Now().UnixNano()
			cut := now - window.Nanoseconds()

			// Gunakan sliding window (sama dengan RateLimitSlidingWindow di shared)
			pipe := rdb.TxPipeline()
			pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprint(cut))
			pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
			pipe.Expire(ctx, key, window)
			cardCmd := pipe.ZCard(ctx, key)

			if _, err := pipe.Exec(ctx); err != nil {
				log.Printf("[RateLimit] Error on rate limit check: %v", err)
				// Fallback: jangan block jika Redis error
				next.ServeHTTP(w, r)
				return
			}

			count := int(cardCmd.Val())
			remaining := limit - count

			if remaining < 0 {
				remaining = 0
			}

			// Set response headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

			// Check limit
			if count > limit {
				log.Printf("[RateLimit] User %s exceeded rate limit (%d/%d requests)", userID, count, limit)
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
				http.Error(w, "rate_limit_exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
