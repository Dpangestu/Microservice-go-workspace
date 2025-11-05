package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/mux"

	mymw "bkc_microservice/services/api-gateway/internal/middleware"
	shcache "bkc_microservice/shared/cache"
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

	// tenantURL, _ := url.Parse(envOr("TENANT_SERVICE_URL", "http://tenant-service:9003"))
	// tenantRP := httputil.NewSingleHostReverseProxy(tenantURL)

	r := mux.NewRouter()

	rl := shmw.RateLimitSlidingWindow(rdb, "rl:gw", 60, time.Minute)

	requireJWT := mymw.RequireJWTWithJWKS(jwks, cfg.JWT.Issuer)
	requireProfile := mymw.RequireScopeFromClaims("profile")

	// /user/me -> proxy ke user-service
	r.Handle("/user/me",
		rl(requireJWT(requireProfile(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// set path
			r.URL.Path = "/me"

			if claims, ok := mymw.ClaimsFromContext(r.Context()); ok && claims != nil {

				log.Printf("Sending request to user-service with headers: X-User-Id=%s, X-Client-Id=%s, X-Tenant-Id=%s, X-Scope=%s",
					claims.UserID, claims.ClientID, claims.TenantID, claims.Scope)

				r.Header.Set("X-User-Id", claims.UserID)
				r.Header.Set("X-Client-Id", claims.ClientID)
				r.Header.Set("X-Tenant-Id", claims.TenantID)
				r.Header.Set("X-Scope", claims.Scope)
			}

			userRP.ServeHTTP(w, r)
		})))),
	).Methods("GET")

	// Health
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	handler := shhttp.CORS(shhttp.CorrelationID(shhttp.JSONLogger(r)))

	srv := shhttp.NewServer(shhttp.ServerOptions{
		Addr:         cfg.Server.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	})

	log.Printf("api-gateway listening on %s, proxy -> %s", srv.Addr, userURL)
	log.Fatal(srv.ListenAndServe())
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func forwardRequestToTenantService(w http.ResponseWriter, r *http.Request, tenantRP *httputil.ReverseProxy) {
	if claims, ok := mymw.ClaimsFromContext(r.Context()); ok && claims != nil {
		r.Header.Set("X-User-Id", claims.UserID)
		r.Header.Set("X-Client-Id", claims.ClientID)
		r.Header.Set("X-Tenant-Id", claims.TenantID)
		r.Header.Set("X-Scope", claims.Scope)
	}

	tenantRP.ServeHTTP(w, r)
}
