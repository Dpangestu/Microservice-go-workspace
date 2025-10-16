package config

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

type ServerCfg struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DBcfg struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type JWTcfg struct {
	PrivateKeyPath string
	PublicKeyPath  string
	Issuer         string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	AuthCodeTTL    time.Duration
}

type Config struct {
	Server ServerCfg
	DB     DBcfg
	JWT    JWTcfg
	Env    string
}

func tryLoadDotEnv() {
	if os.Getenv("ENV_LOADED") == "1" {
		return
	}
	dir, _ := os.Getwd()
	for i := 0; i < 6; i++ {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			_ = godotenv.Load(envPath)
			os.Setenv("ENV_LOADED", "1")
			return
		}
		dir = filepath.Dir(dir)
	}
}

func resolvePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	cwd, _ := os.Getwd()
	for i := 0; i < 6; i++ {
		try := filepath.Join(cwd, p)
		if _, err := os.Stat(try); err == nil {
			return try
		}
		cwd = filepath.Dir(cwd)
	}
	return p
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func MustLoad() *Config {
	tryLoadDotEnv()

	readTimeout := parseDurOr(getEnv("SERVER_READ_TIMEOUT", "10s"), 10*time.Second)
	writeTimeout := parseDurOr(getEnv("SERVER_WRITE_TIMEOUT", "10s"), 10*time.Second)
	idleTimeout := parseDurOr(getEnv("SERVER_IDLE_TIMEOUT", "60s"), 60*time.Second)

	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "auth_db")

	accessTTL := parseDurOr(getEnv("OAUTH2_ACCESS_TOKEN_EXPIRATION", "15m"), 15*time.Minute)
	refreshTTL := parseDurOr(getEnv("OAUTH2_REFRESH_TOKEN_EXPIRATION", "720h"), 720*time.Hour)
	authCodeTTL := parseDurOr(getEnv("OAUTH2_AUTH_CODE_EXPIRATION", "10m"), 10*time.Minute)

	return &Config{
		Env: getEnv("APP_ENV", "local"),
		Server: ServerCfg{
			Addr:         getEnv("SERVER_PORT", ":8080"),
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
		DB: DBcfg{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPass,
			Name:     dbName,
		},
		JWT: JWTcfg{
			PrivateKeyPath: resolvePath(getEnv("JWT_PRIVATE_KEY_PATH", "./keys/private.pem")),
			PublicKeyPath:  resolvePath(getEnv("JWT_PUBLIC_KEY_PATH", "./keys/public.pem")),
			Issuer:         getEnv("JWT_ISSUER", "auth-service"),
			AccessTTL:      accessTTL,
			RefreshTTL:     refreshTTL,
			AuthCodeTTL:    authCodeTTL,
		},
	}
}

func parseDurOr(s string, def time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("invalid duration %q, use %s", s, def)
		return def
	}
	return d
}
