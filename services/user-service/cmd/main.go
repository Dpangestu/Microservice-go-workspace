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

	appsvc "bkc_microservice/services/user-service/internal/application/services"
	"bkc_microservice/services/user-service/internal/infrastructure/clients"
	"bkc_microservice/services/user-service/internal/infrastructure/persistence"
	httpif "bkc_microservice/services/user-service/internal/interfaces/http"
	"bkc_microservice/services/user-service/internal/shared"
)

func main() {
	cfg := shcfg.MustLoad()

	pool := shdb.MustNewPool(shdb.DBConfig{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Name:     cfg.DB.Name,
	})
	defer pool.Close()

	rdb := shcache.NewRedis(shcache.RedisCfg{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rdb.Close()

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("[WARNING] Redis connection failed: %v (caching/rate limiting disabled)", err)
		rdb = nil // Fallback
	}
	cancel()

	// === Setup Logger ===
	logger := shared.NewLogger()

	syncCBSURL := cfg.SyncCBSServiceURL
	if syncCBSURL == "" {
		syncCBSURL = "http://sync-cbs-service:9003"
		log.Printf("[WARN] SyncCBSServiceURL not set in config, using default: %s", syncCBSURL)
	} else {
		log.Printf("[INFO] Sync CBS Service URL: %s", syncCBSURL)
	}

	syncCBSClient := clients.NewSyncCBSClient(syncCBSURL)

	// === Setup Repositories ===
	userRepo := persistence.NewMySQLUserRepository(pool)
	roleRepo := persistence.NewMySQLRoleRepository(pool)
	permissionRepo := persistence.NewMySQLPermissionRepository(pool)
	userActivityRepo := persistence.NewMySQLUserActivityRepository(pool)
	profileRepo := persistence.NewMySQLUserProfileRepository(pool)
	settingsRepo := persistence.NewMySQLUserSettingsRepository(pool)
	rpRepo := persistence.NewMySQLRolePermissionsRepository(pool)

	// === Setup Services ===
	// User Service
	userService := appsvc.NewUserService(
		userRepo,
		roleRepo,
		permissionRepo,
		rpRepo,
		userActivityRepo,
		profileRepo,
		settingsRepo,
		rdb,
		syncCBSClient,
		pool,
	)

	// Role Service
	roleService := appsvc.NewRoleService(
		roleRepo,
		// rpRepo,
	)

	// Permission Service
	permService := appsvc.NewPermissionService(
		permissionRepo,
		rpRepo,
	)

	// === Setup HTTP Router & Middlewares ===
	router := httpif.NewRouter(
		userService,
		roleService,
		permService,
		logger,
		rdb,
	)
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
		log.Printf("user-service listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("user server error: %v", err)
		}
	}()

	<-quit
	log.Println("user-service shutting down...")

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("user shutdown error: %v", err)
	}
	if err := pool.Close(); err != nil {
		log.Printf("db close error: %v", err)
	}
	if err := rdb.Close(); err != nil {
		log.Printf("redis close error: %v", err)
	}
	log.Println("user-service stopped cleanly")
}
