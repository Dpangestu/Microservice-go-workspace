package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	mymw "bkc_microservice/services/api-gateway/internal/middleware"
	shcache "bkc_microservice/shared/cache"
	shcb "bkc_microservice/shared/circuitbreaker"
	shcfg "bkc_microservice/shared/config"
	shhttp "bkc_microservice/shared/http"
	shmw "bkc_microservice/shared/middleware"
	shsec "bkc_microservice/shared/security"
)

func main() {
	cfg := shcfg.MustLoad()

	rdb := shcache.NewRedis(shcache.RedisCfg{
		Addr:     envOr("REDIS_ADDR", "redis:6379"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	userURL, _ := url.Parse(envOr("USER_SERVICE_URL", "http://user-service:9002"))
	userRP := httputil.NewSingleHostReverseProxy(userURL)

	jwksURL := envOr("AUTH_JWKS_URL", "http://auth-service:9001/oauth/jwks")
	jwks := shsec.NewJWKSCache(jwksURL, 5*time.Minute)

	userRP.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Error proxying to user-service: %v", err)
		http.Error(w, "service unavailable", http.StatusBadGateway)
	}

	// Circuit breaker config
	cbConfig := shcb.Config{
		FailureThreshold:    5,
		SuccessThreshold:    2,
		Timeout:             30 * time.Second,
		HalfOpenMaxRequests: 3,
	}
	userCB := shcb.NewCircuitBreaker(cbConfig)

	r := mux.NewRouter()

	rl := shmw.RateLimitSlidingWindow(rdb, "rl:gw", 60, time.Minute)

	requireJWT := mymw.RequireJWTWithJWKS(jwks, cfg.JWT.Issuer)
	requireProfile := mymw.RequireScopeFromClaims("profile")

	// ===== HEALTH CHECK (NO PROXY) =====
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
		// FIX: Jangan proxy healthz ke backend
	}).Methods("GET")

	// ===== USER/ME ROUTE =====
	r.Handle("/user/me",
		rl(requireJWT(requireProfile(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract claims dari context
			claims, ok := mymw.ClaimsFromContext(r.Context())
			if !ok || claims == nil {
				log.Printf("Claims not found in context")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Check circuit breaker BEFORE proxying
			if userCB.IsOpen() {
				log.Printf("Circuit breaker OPEN - rejecting request")
				http.Error(w, "service unavailable - circuit breaker open", http.StatusServiceUnavailable)
				userCB.RecordFailure()
				return
			}

			// Set request path dan headers
			r.URL.Path = "/me"
			r.Header.Set("X-User-Id", claims.UserID)
			r.Header.Set("X-Client-Id", claims.ClientID)
			r.Header.Set("X-Tenant-Id", claims.TenantID)
			r.Header.Set("X-Scope", claims.Scope)

			log.Printf("Sending request to user-service with headers: X-User-Id=%s, X-Client-Id=%s, X-Tenant-Id=%s, X-Scope=%s",
				claims.UserID, claims.ClientID, claims.TenantID, claims.Scope)

			// Custom response writer untuk track status
			statusCode := http.StatusOK
			rw := &statusCapture{
				ResponseWriter: w,
				statusCode:     &statusCode,
			}

			// FIX: HANYA proxy sekali
			userRP.ServeHTTP(rw, r)

			// Record circuit breaker result AFTER proxy
			if statusCode >= 500 {
				userCB.RecordFailure()
				log.Printf("Request failed with status %d - recording failure", statusCode)
			} else if statusCode == http.StatusServiceUnavailable {
				userCB.RecordFailure()
				log.Printf("Service unavailable - recording failure")
			} else {
				userCB.RecordSuccess()
				log.Printf("Request success - %s", userCB.String())
			}
		})))),
	).Methods("GET")

	// Apply middleware stack
	handler := shhttp.CORS(shhttp.CorrelationID(shhttp.JSONLogger(r)))

	srv := shhttp.NewServer(shhttp.ServerOptions{
		Addr:         cfg.Server.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("api-gateway listening on %s, proxy -> %s", srv.Addr, userURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("gateway server error: %v", err)
		}
	}()

	<-quit
	log.Println("\n api-gateway shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("gateway shutdown error: %v", err)
	}
	if err := rdb.Close(); err != nil {
		log.Printf("redis close error: %v", err)
	}
	log.Println("api-gateway stopped cleanly")
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// statusCapture wraps ResponseWriter untuk track status code
type statusCapture struct {
	http.ResponseWriter
	statusCode *int
	written    bool
}

func (sc *statusCapture) WriteHeader(code int) {
	if !sc.written {
		*sc.statusCode = code
		sc.written = true
		sc.ResponseWriter.WriteHeader(code)
	}
}
