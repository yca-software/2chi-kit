package api_key_repository_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/database"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	api_key_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/api_key"
	yca_error "github.com/yca-software/go-common/error"
)

type ApiKeyRepositoryTestSuite struct {
	suite.Suite
	repo api_key_repository.Repository
	db   *sqlx.DB
}

func TestApiKeyRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ApiKeyRepositoryTestSuite))
}

func (s *ApiKeyRepositoryTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}

	testDB, err := database.GetTestDB()
	require.NoError(s.T(), err)
	s.db = testDB.DB()
	s.repo = api_key_repository.New(s.db, nil)
}

func (s *ApiKeyRepositoryTestSuite) SetupTest() {
	// Insert test organizations
	_, err := s.db.Exec(`
		INSERT INTO organizations (id, created_at, name, address, city, zip, country, place_id, geo, timezone, billing_email, custom_subscription, subscription_expires_at, subscription_type, subscription_seats, paddle_customer_id)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', 'Test Org 1', '123 Main St', 'New York', '10001', 'US', 'place_123', ST_MakePoint(-74.006, 40.7128), 'America/New_York', 'billing1@example.com', false, NULL, 1, 10, 'paddle_customer_1'),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 13:00:00+00', 'Test Org 2', '456 Oak Ave', 'Los Angeles', '90001', 'US', 'place_456', ST_MakePoint(-118.2437, 34.0522), 'America/Los_Angeles', 'billing2@example.com', true, '2025-12-31 23:59:59+00', 2, 20, 'paddle_customer_2');
	`)
	require.NoError(s.T(), err)

	// Insert test API keys
	permissions1, _ := json.Marshal([]string{"read", "write"})
	permissions2, _ := json.Marshal([]string{"read"})
	_, err = s.db.Exec(`
		INSERT INTO api_keys (id, created_at, expires_at, name, key_prefix, key_hash, organization_id, permissions)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', '2024-01-16 10:00:00+00', 'Test Key 1', 'test_', 'hash1', '00000001-0000-0000-0000-000000000001', $1),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 13:00:00+00', '2024-01-16 10:00:00+00', 'Test Key 2', 'test_', 'hash2', '00000001-0000-0000-0000-000000000001', $2),
			('00000003-0000-0000-0000-000000000003', '2024-01-15 14:00:00+00', '2024-01-16 10:00:00+00', 'Expired Key', 'test_', 'hash3', '00000002-0000-0000-0000-000000000002', $1);
	`, permissions1, permissions2)
	require.NoError(s.T(), err)
}

func (s *ApiKeyRepositoryTestSuite) TearDownTest() {
	_, _ = s.db.Exec("TRUNCATE TABLE api_keys CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE organizations CASCADE")
}

func (s *ApiKeyRepositoryTestSuite) TestCreate_DuplicateKeyHash() {
	permissions := models.RolePermissions{"read", "write"}
	apiKey := &models.APIKey{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(time.Hour * 24),
		Name:           "Duplicate Key",
		KeyPrefix:      "test_",
		KeyHash:        "hash1",
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		Permissions:    permissions,
	}
	err := s.repo.Create(nil, apiKey)
	require.Error(s.T(), err)
	require.Equal(s.T(), 409, err.(*yca_error.Error).StatusCode)
}

func (s *ApiKeyRepositoryTestSuite) TestCreate_Success() {
	permissions := models.RolePermissions{"read", "write"}
	apiKey := &models.APIKey{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(time.Hour * 24),
		Name:           "New Key",
		KeyPrefix:      "new_",
		KeyHash:        "new_hash",
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		Permissions:    permissions,
	}
	err := s.repo.Create(nil, apiKey)
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID(apiKey.OrganizationID.String(), apiKey.ID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), apiKey.Name, got.Name)
}

func (s *ApiKeyRepositoryTestSuite) TestGetByHash_NotFound() {
	got, err := s.repo.GetByHash("nonexistent_hash")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *ApiKeyRepositoryTestSuite) TestGetByHash_Success() {
	got, err := s.repo.GetByHash("hash1")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), "Test Key 1", got.Name)
}

func (s *ApiKeyRepositoryTestSuite) TestGetByID_NotFound() {
	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", uuid.New().String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *ApiKeyRepositoryTestSuite) TestGetByID_WrongOrganization() {
	got, err := s.repo.GetByID("00000002-0000-0000-0000-000000000002", "00000001-0000-0000-0000-000000000001")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *ApiKeyRepositoryTestSuite) TestGetByID_Success() {
	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), "Test Key 1", got.Name)
}

func (s *ApiKeyRepositoryTestSuite) TestListByOrganizationID() {
	keys, err := s.repo.ListByOrganizationID("00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), keys)
	require.Len(s.T(), *keys, 2)
}

func (s *ApiKeyRepositoryTestSuite) TestDelete_NotFound() {
	err := s.repo.Delete(nil, "00000001-0000-0000-0000-000000000001", uuid.New().String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *ApiKeyRepositoryTestSuite) TestDelete_Success() {
	err := s.repo.Delete(nil, "00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)

	got, err := s.repo.GetByHash("hash1")
	require.Error(s.T(), err)
	require.Nil(s.T(), got)
}

func (s *ApiKeyRepositoryTestSuite) TestCleanupStaleExpired() {
	permissions, _ := json.Marshal([]string{"read"})
	_, err := s.db.Exec(`
		INSERT INTO api_keys (id, created_at, expires_at, name, key_prefix, key_hash, organization_id, permissions)
		VALUES
			('c0000001-0000-0000-0000-000000000001', NOW() - INTERVAL '400 days', NOW() - INTERVAL '35 days', 'Stale Expired', 'st_', 'hash_stale_expired', '00000001-0000-0000-0000-000000000001', $1),
			('c0000002-0000-0000-0000-000000000002', NOW() - INTERVAL '20 days', NOW() - INTERVAL '5 days', 'Recent Expired', 'st_', 'hash_recent_expired', '00000001-0000-0000-0000-000000000001', $1);
	`, permissions)
	require.NoError(s.T(), err)

	err = s.repo.CleanupStaleExpired()
	require.NoError(s.T(), err)

	_, err = s.repo.GetByHash("hash_stale_expired")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)

	got, err := s.repo.GetByHash("hash_recent_expired")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
}
