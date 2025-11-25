package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	shcache "bkc_microservice/shared/cache"
	shcfg "bkc_microservice/shared/config"
	shdb "bkc_microservice/shared/database"
	shhttp "bkc_microservice/shared/http"

	appsvc "bkc_microservice/services/sync-cbs-service/internal/application/services"
	"bkc_microservice/services/sync-cbs-service/internal/infrastructure/persistence"
	httpif "bkc_microservice/services/sync-cbs-service/internal/interfaces/http"

	// shared "bkc_microservice/services/user-service/internal/shared"
	"bkc_microservice/services/sync-cbs-service/internal/shared"
)

func main() {
	cfg := shcfg.MustLoad()

	// === Setup Database ===
	pool := shdb.MustNewPool(shdb.DBConfig{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Name:     cfg.DB.Name,
	})
	defer pool.Close()

	// === Setup Redis ===
	rdb := shcache.NewRedis(shcache.RedisCfg{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rdb.Close()

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("[WARNING] Redis connection failed: %v (caching disabled)", err)
		rdb = nil // Fallback
	}
	cancel()

	// === Setup Logger ===
	logger := shared.NewLogger()

	// === Setup Repositories ===
	sycroneRepo := persistence.NewMySQLSycroneCoreRepository(pool)

	// === Setup Services ===
	syncService := appsvc.NewSyncService(sycroneRepo)

	// === Setup HTTP Router & Middlewares ===
	router := httpif.NewRouter(syncService, logger, rdb)
	handler := shhttp.CORS(shhttp.CorrelationID(shhttp.JSONLogger(router)))

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
		log.Printf("sync-cbs-service listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("sync-cbs server error: %v", err)
		}
	}()

	<-quit
	log.Println("sync-cbs-service shutting down...")

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("sync-cbs shutdown error: %v", err)
	}
	if err := pool.Close(); err != nil {
		log.Printf("db close error: %v", err)
	}
	if err := rdb.Close(); err != nil {
		log.Printf("redis close error: %v", err)
	}
	log.Println("sync-cbs-service stopped cleanly")
}
