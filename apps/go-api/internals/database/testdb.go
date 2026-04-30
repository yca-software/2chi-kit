package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/go-common/logger"
)

const testWaitTimeout = 120 * time.Second

// IntegrationTestContainerStartTimeout bounds image pull + container start for testcontainers.
const IntegrationTestContainerStartTimeout = 8 * time.Minute

// Use postgis/postgis image for PostGIS-enabled integration tests.
// Pin a tag for deterministic CI/local runs.
const IntegrationTestPostgresImage = "postgis/postgis:16-3.4"

// integrationTestContainerName is the Docker container name used with testcontainers reuse so
// every repository test process shares one DB (each package is a separate process, so sync.Once
// alone is not enough). Run repository tests with -p 1 so packages do not mutate the same
// schema concurrently (see Makefile).
const integrationTestContainerName = "2chi-kit-go-api-integration-postgis"

// EnvIntegrationTestDSN skips testcontainers when set (e.g. postgres://user:pass@localhost:5432/testdb?sslmode=disable).
// Migrations still run via AutoMigrate; use a dedicated database.
// (Prefix TWOCHI_KIT_: POSIX env names cannot start with a digit.)
const EnvIntegrationTestDSN = "TWOCHI_KIT_GO_API_INTEGRATION_TEST_DSN"

// EnvIntegrationTestNoContainerReuse forces a fresh ephemeral container per process when set to "1"
// (slow; only for debugging isolation issues).
const EnvIntegrationTestNoContainerReuse = "TWOCHI_KIT_GO_API_INTEGRATION_TEST_NO_CONTAINER_REUSE"

var (
	testOnce     sync.Once
	testInstance *TestDB
	testMu       sync.Mutex
)

type TestDB struct {
	container          testcontainers.Container
	db                 *sqlx.DB
	connStr            string
	terminateContainer bool // false: external DSN or reusable shared container
}

// GetTestDB returns a test database for repository integration tests.
//
// By default starts (or reuses) one shared PostGIS testcontainer (integrationTestContainerName)
// across processes. Set TWOCHI_KIT_GO_API_INTEGRATION_TEST_DSN to use an
// existing server instead (fastest on repeat runs). Repository packages must run with -p 1
// when sharing one database (see Makefile).
func GetTestDB() (*TestDB, error) {
	var err error
	testOnce.Do(func() {
		testInstance, err = setupTestDB()
	})
	return testInstance, err
}

func setupTestDB() (*TestDB, error) {
	if dsn := strings.TrimSpace(os.Getenv(EnvIntegrationTestDSN)); dsn != "" {
		return setupTestDBFromDSN(dsn)
	}

	ctx, cancel := context.WithTimeout(context.Background(), IntegrationTestContainerStartTimeout)
	defer cancel()

	opts := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(1).
					WithStartupTimeout(90*time.Second),
				wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
					return fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable",
						host, port.Port())
				}).WithStartupTimeout(90*time.Second),
			).WithDeadline(testWaitTimeout),
		),
	}
	terminateContainer := true
	if os.Getenv(EnvIntegrationTestNoContainerReuse) != "1" {
		opts = append(opts, testcontainers.WithReuseByName(integrationTestContainerName))
		terminateContainer = false
	}

	// Supply the correct PostGIS image
	container, err := postgres.Run(ctx, IntegrationTestPostgresImage, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	root, err := helpers.ModuleRoot()
	if err != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("module root: %w", err)
	}

	cfg := &PostgreSQLConfig{
		DSN:            connStr,
		AutoMigrate:    true,
		MigrationsPath: filepath.Join(root, "migrations"),
	}

	db, err := InitPostgreSQLClient(cfg, logger.New())
	if err != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = container.Terminate(ctx)
		return nil, err
	}

	return &TestDB{
		container:          container,
		db:                 db,
		connStr:            connStr,
		terminateContainer: terminateContainer,
	}, nil
}

func setupTestDBFromDSN(dsn string) (*TestDB, error) {
	root, err := helpers.ModuleRoot()
	if err != nil {
		return nil, fmt.Errorf("module root: %w", err)
	}

	cfg := &PostgreSQLConfig{
		DSN:            dsn,
		AutoMigrate:    true,
		MigrationsPath: filepath.Join(root, "migrations"),
	}

	db, err := InitPostgreSQLClient(cfg, logger.New())
	if err != nil {
		return nil, err
	}

	return &TestDB{
		container:          nil,
		db:                 db,
		connStr:            dsn,
		terminateContainer: false,
	}, nil
}

func (t *TestDB) DB() *sqlx.DB {
	newDB, err := sqlx.Connect("pgx", t.connStr)
	if err != nil {
		return t.db
	}
	return newDB
}

func (t *TestDB) ConnectionString() string {
	return t.connStr
}

func CleanupTestDB() {
	testMu.Lock()
	defer testMu.Unlock()
	if testInstance != nil {
		if testInstance.db != nil {
			_ = testInstance.db.Close()
		}
		if testInstance.container != nil && testInstance.terminateContainer {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = testInstance.container.Terminate(ctx)
		}
		testInstance = nil
	}
}

func (t *TestDB) NewTestConnection() (*sqlx.DB, error) {
	return sqlx.Connect("pgx", t.connStr)
}
