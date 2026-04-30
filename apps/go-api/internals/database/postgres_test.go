package database_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/yca-software/2chi-kit/go-api/internals/database"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/go-common/logger"
)

type PostgresTestSuite struct {
	suite.Suite
	originalDSN string
}

func TestPostgresTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}

func (s *PostgresTestSuite) SetupTest() {
	s.originalDSN = os.Getenv("DATABASE_POSTGRES_DSN")
}

func (s *PostgresTestSuite) TearDownTest() {
	if s.originalDSN != "" {
		os.Setenv("DATABASE_POSTGRES_DSN", s.originalDSN)
	} else {
		os.Unsetenv("DATABASE_POSTGRES_DSN")
	}
}

func (s *PostgresTestSuite) TestInitPostgreSQLClient_InvalidDSN() {
	root, err := helpers.ModuleRoot()
	require.NoError(s.T(), err)

	cfg := &database.PostgreSQLConfig{
		DSN:            "invalid-dsn",
		AutoMigrate:    true,
		MigrationsPath: filepath.Join(root, "migrations"),
	}

	log := logger.New()
	client, err := database.InitPostgreSQLClient(cfg, log)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), client)
}

func (s *PostgresTestSuite) TestInitPostgreSQLClient_MissingMigrationsPath() {
	cfg := &database.PostgreSQLConfig{
		DSN:            "postgres://user:pass@localhost/db?sslmode=disable",
		AutoMigrate:    true,
		MigrationsPath: "",
	}

	log := logger.New()
	client, err := database.InitPostgreSQLClient(cfg, log)

	assert.Error(s.T(), err)
	assert.ErrorContains(s.T(), err, "MigrationsPath")
	assert.Nil(s.T(), client)
}

func (s *PostgresTestSuite) TestInitPostgreSQLClient_AutoMigrateDisabled() {
	cfg := &database.PostgreSQLConfig{
		DSN:         "postgres://user:pass@localhost/db?sslmode=disable",
		AutoMigrate: false,
	}

	log := logger.New()
	client, err := database.InitPostgreSQLClient(cfg, log)

	// Should fail due to invalid DSN, but migration path shouldn't matter
	assert.Error(s.T(), err)
	assert.Nil(s.T(), client)
}

// TestInitPostgreSQLClient_WithRealPostgresAndMigrations spins up a Postgres container,
// runs migrations via InitPostgreSQLClient, and asserts success. Skipped when -short.
func (s *PostgresTestSuite) TestInitPostgreSQLClient_WithRealPostgresAndMigrations() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.IntegrationTestContainerStartTimeout)
	defer cancel()

	postgresContainer, err := postgres.Run(ctx,
		database.IntegrationTestPostgresImage,
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(1).
					WithStartupTimeout(90*time.Second),
				wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
					return fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable",
						host, port.Port())
				}).WithStartupTimeout(90*time.Second),
			).WithDeadline(120*time.Second),
		),
	)
	require.NoError(s.T(), err)
	defer func() {
		err := testcontainers.TerminateContainer(postgresContainer)
		require.NoError(s.T(), err)
	}()

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(s.T(), err)

	root, err := helpers.ModuleRoot()
	require.NoError(s.T(), err)

	cfg := &database.PostgreSQLConfig{
		DSN:            connStr,
		AutoMigrate:    true,
		MigrationsPath: filepath.Join(root, "migrations"),
	}

	log := logger.New()
	client, err := database.InitPostgreSQLClient(cfg, log)

	assert.NoError(s.T(), err)
	require.NotNil(s.T(), client)
	_ = client.Close()
}
