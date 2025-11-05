package config

import (
	"fmt"
	"os"
	"strings"
)

func LoadForService(serviceName string) *Config {
	cfg := MustLoad()

	// normalize service name jadi db_name
	dbNameEnvKey := fmt.Sprintf("DB_NAME_%s", strings.ToUpper(strings.ReplaceAll(serviceName, "-", "_")))
	if v := os.Getenv(dbNameEnvKey); v != "" {
		cfg.DB.Name = v
	} else {
		// <serviceName>_db
		cfg.DB.Name = fmt.Sprintf("%s_db", strings.ReplaceAll(serviceName, "-", "_"))
	}

	return cfg
}
