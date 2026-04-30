package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/yca-software/2chi-kit/go-api/internals/database"
	"github.com/yca-software/go-common/logger"
)

type RedisTestSuite struct {
	suite.Suite
}

func TestRedisTestSuite(t *testing.T) {
	suite.Run(t, new(RedisTestSuite))
}

func (s *RedisTestSuite) TestInitRedisClient_EmptyDSN() {
	dsn := ""
	log := logger.New()
	client, err := database.InitRedisClient(dsn, log)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), client)
	assert.Contains(s.T(), err.Error(), "failed to parse Redis DSN")
}

func (s *RedisTestSuite) TestInitRedisClient_InvalidDSN() {
	dsn := "invalid-dsn-format"
	log := logger.New()
	client, err := database.InitRedisClient(dsn, log)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), client)
	assert.Contains(s.T(), err.Error(), "failed to parse Redis DSN")
}

func (s *RedisTestSuite) TestInitRedisClient_InvalidConnection() {
	dsn := "redis://invalid-host:6379/0"
	log := logger.New()
	client, err := database.InitRedisClient(dsn, log)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), client)
	assert.Contains(s.T(), err.Error(), "failed to connect to Redis")
}

func (s *RedisTestSuite) TestInitRedisClient_InvalidPort() {
	dsn := "redis://localhost:99999/0" // Invalid port
	log := logger.New()
	client, err := database.InitRedisClient(dsn, log)

	// Should fail on connection attempt
	assert.Error(s.T(), err)
	assert.Nil(s.T(), client)
}

func (s *RedisTestSuite) TestInitRedisClient_MalformedURL() {
	dsn := "://malformed-url"
	log := logger.New()
	client, err := database.InitRedisClient(dsn, log)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), client)
	assert.Contains(s.T(), err.Error(), "failed to parse Redis DSN")
}

// TestInitRedisClient_WithRealRedis spins up a Redis container, calls InitRedisClient,
// and asserts success and that the returned client can Ping and perform Set/Get. Skipped when -short.
func (s *RedisTestSuite) TestInitRedisClient_WithRealRedis() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	redisCtr, err := rediscontainer.Run(ctx, "redis:7-alpine")
	require.NoError(s.T(), err)
	defer func() {
		err := testcontainers.TerminateContainer(redisCtr)
		require.NoError(s.T(), err)
	}()

	connStr, err := redisCtr.ConnectionString(ctx)
	require.NoError(s.T(), err)

	dsn := connStr
	log := logger.New()
	client, err := database.InitRedisClient(dsn, log)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), client)
	defer client.Close()

	// Verify client is usable: Ping and a round-trip Set/Get
	pingCtx, pingCancel := context.WithTimeout(ctx, 2*time.Second)
	defer pingCancel()
	err = client.Ping(pingCtx).Err()
	assert.NoError(s.T(), err)

	setCtx, setCancel := context.WithTimeout(ctx, 2*time.Second)
	defer setCancel()
	key, value := "testkey:init", "testvalue"
	err = client.Set(setCtx, key, value, 10*time.Second).Err()
	assert.NoError(s.T(), err)

	getCtx, getCancel := context.WithTimeout(ctx, 2*time.Second)
	defer getCancel()
	got, err := client.Get(getCtx, key).Result()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value, got)
}
