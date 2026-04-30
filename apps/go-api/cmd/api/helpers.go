package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yca-software/2chi-kit/go-api/internals/database"
)

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getPostgresConfig(moduleRoot string) *database.PostgreSQLConfig {
	dsn := strings.TrimSpace(os.Getenv("POSTGRES_DSN"))
	if dsn == "" {
		panic("POSTGRES_DSN is required but empty or unset")
	}

	autoMigrate := os.Getenv("POSTGRES_AUTO_MIGRATE") == "true"
	migrationsPath := ""
	if autoMigrate {
		if p := strings.TrimSpace(os.Getenv("POSTGRES_MIGRATIONS_PATH")); p != "" {
			abs, err := filepath.Abs(p)
			if err != nil {
				panic("POSTGRES_MIGRATIONS_PATH: " + err.Error())
			}
			migrationsPath = abs
		} else {
			migrationsPath = filepath.Join(moduleRoot, "migrations")
		}
	}

	return &database.PostgreSQLConfig{
		DSN:             dsn,
		MaxOpenConns:    envInt("POSTGRES_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    envInt("POSTGRES_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: envInt("POSTGRES_CONN_MAX_LIFETIME", 5),
		ConnMaxIdleTime: envInt("POSTGRES_CONN_MAX_IDLE_TIME", 10),
		PingTimeout:     envInt("POSTGRES_PING_TIMEOUT", 5),
		AutoMigrate:     autoMigrate,
		MigrationsPath:  migrationsPath,
	}
}
