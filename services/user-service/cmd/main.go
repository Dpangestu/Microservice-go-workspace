package main

import (
	"context"
	"fmt"
	"log"
	"time"

	shcache "bkc_microservice/shared/cache"
	shcfg "bkc_microservice/shared/config"
	shdb "bkc_microservice/shared/database"
	shhttp "bkc_microservice/shared/http"

	appsvc "bkc_microservice/services/user-service/internal/application/services"
	httpif "bkc_microservice/services/user-service/internal/http"
	"bkc_microservice/services/user-service/internal/infrastructure/persistence"
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

	// === Setup Repositories & Services ===
	userRepo := persistence.NewMySQLUserRepository(pool)
	roleRepo := persistence.NewRoleRepository(pool)
	permissionRepo := persistence.NewMySQLPermissionRepository(pool)
	userActivityRepo := persistence.NewMySQLUserActivityRepository(pool)
	profileRepo := persistence.NewMySQLUserProfileRepository(pool)
	settingsRepo := persistence.NewMySQLUserSettingsRepository(pool)
	rpRepo := persistence.NewMySQLRolePermissionsRepository(pool)

	userService := appsvc.NewUserService(
		userRepo, roleRepo, permissionRepo, userActivityRepo,
		profileRepo, settingsRepo, rpRepo,
		rdb,
	)

	// === Setup HTTP Router & Middlewares ===
	router := httpif.NewRouter(userService, rdb)
	handler := shhttp.CORS(shhttp.CorrelationID(shhttp.JSONLogger(router)))

	srv := shhttp.NewServer(shhttp.ServerOptions{
		Addr:         cfg.Server.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	})

	fmt.Printf("[SERVER] user-service listening on %s\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
