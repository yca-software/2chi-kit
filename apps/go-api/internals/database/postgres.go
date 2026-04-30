package database

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	yca_database "github.com/yca-software/go-common/database"
	"github.com/yca-software/go-common/logger"
)

type PostgreSQLConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	ConnMaxIdleTime int
	PingTimeout     int
	AutoMigrate     bool
	// MigrationsPath is the absolute or relative path to the golang-migrate file source directory.
	// Required when AutoMigrate is true.
	MigrationsPath string
}

func InitPostgreSQLClient(cfg *PostgreSQLConfig, log logger.Logger) (*sqlx.DB, error) {
	if cfg.AutoMigrate && strings.TrimSpace(cfg.MigrationsPath) == "" {
		return nil, fmt.Errorf("PostgreSQLConfig.MigrationsPath is required when AutoMigrate is true")
	}

	dbClient, err := yca_database.NewSQLClient(yca_database.SQLClientConfig{
		DriverName:      "pgx",
		DSN:             cfg.DSN,
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: time.Duration(cfg.ConnMaxLifetime) * time.Minute,
		ConnMaxIdleTime: time.Duration(cfg.ConnMaxIdleTime) * time.Minute,
		PingTimeout:     time.Duration(cfg.PingTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}

	if cfg.AutoMigrate {
		if err := runPostgresMigrations(cfg.DSN, cfg.MigrationsPath, log); err != nil {
			return nil, fmt.Errorf("failed to run postgres migrations: %w", err)
		}
	}

	return dbClient, nil
}

func runPostgresMigrations(dsn, migrationsPath string, log logger.Logger) error {
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get migrations path: %w", err)
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", absPath),
		dsn,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Log(logger.LogData{
				Level:   "info",
				Message: "Database migrations are up to date",
			})
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Log(logger.LogData{
		Level:   "info",
		Message: "Database migrations completed successfully",
	})

	return nil
}
