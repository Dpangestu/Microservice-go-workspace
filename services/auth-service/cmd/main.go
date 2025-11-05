package main

import (
	"fmt"
	"log"
	"os"

	shcache "bkc_microservice/shared/cache"
	shcfg "bkc_microservice/shared/config"
	shdb "bkc_microservice/shared/database"
	shhttp "bkc_microservice/shared/http"
	smfa "bkc_microservice/shared/mfa"
	shsec "bkc_microservice/shared/security"
	session "bkc_microservice/shared/session"

	appsvc "bkc_microservice/services/auth-service/internal/application/services"
	"bkc_microservice/services/auth-service/internal/infrastructure/persistence"
	httpif "bkc_microservice/services/auth-service/internal/interfaces/http"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := shcfg.MustLoad()

	// ctx := context.Background()
	// pool := shdb.MustNewPool(ctx, cfg.DB.URL)
	// defer pool.Close()

	// shdb.MustRunMigrations(cfg.DB.URL, "./migrations")

	pool := shdb.MustNewPool(shdb.DBConfig{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Name:     cfg.DB.Name,
	})
	defer pool.Close()

	// dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
	// 	cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)
	// shdb.MustRunMigrations(dsn, "./migrations")

	activeKid := os.Getenv("JWT_ACTIVE_KID")
	if activeKid == "" {
		activeKid = "auth-k1"
	}
	keys := map[string]struct{ PrivatePath, PublicPath string }{
		"auth-k1": {cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath},
	}
	keystore := shsec.MustLoadKeyStore(activeKid, cfg.JWT.Issuer, keys)

	var rdbIface interface{}
	if os.Getenv("REDIS_MODE") == "cluster" {
		rdbIface = shcache.NewRedisCluster(shcache.RedisClusterCfg{
			Addrs:    []string{envOr("REDIS_CLUSTER_ADDRS", "redis-node1:6379,redis-node2:6379")},
			Password: os.Getenv("REDIS_PASSWORD"),
		})
	} else {
		rdbIface = shcache.NewRedis(shcache.RedisCfg{
			Addr:     envOr("REDIS_ADDR", "redis:6379"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0,
		})
	}
	rdb, _ := rdbIface.(*redis.Client)

	userRepo := persistence.NewMySQLUserRepo(pool)
	clientRepo := persistence.NewMySQLClientRepo(pool)
	codeRepo := persistence.NewMySQLAuthCodeRepo(pool)
	tokenRepo := persistence.NewMySQLTokenRepo(pool)

	fmt.Println("userRepo:", userRepo)
	fmt.Println("clientRepo:", clientRepo)
	fmt.Println("codeRepo:", codeRepo)
	fmt.Println("tokenRepo:", tokenRepo)

	authSvc := appsvc.NewAuthService(appsvc.Dep{
		UserRepo:       userRepo,
		ClientRepo:     clientRepo,
		CodeRepo:       codeRepo,
		TokenRepo:      tokenRepo,
		KeyStore:       keystore,
		RDB:            rdb,
		SessionManager: session.NewManager(rdb),
		MFAService:     smfa.NewService(&smfa.TOTPService{}, smfa.NewOTPService(rdb)),
		AccessTTL:      cfg.JWT.AccessTTL,
		RefreshTTL:     cfg.JWT.RefreshTTL,
		CodeTTL:        cfg.JWT.AuthCodeTTL,
		UserServiceURL: cfg.UserServiceURL,
	})

	r := httpif.NewRouter(authSvc)

	handler := shhttp.CORS(shhttp.CorrelationID(shhttp.JSONLogger(r)))
	srv := shhttp.NewServer(shhttp.ServerOptions{
		Addr:         cfg.Server.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	})

	fmt.Printf("auth-service listening on %s\n", srv.Addr)

	log.Fatal(srv.ListenAndServe())
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
