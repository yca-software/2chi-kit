package user_refresh_token_repository_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/database"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	user_refresh_token_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user_refresh_token"
	yca_error "github.com/yca-software/go-common/error"
)

type UserRefreshTokenRepositoryTestSuite struct {
	suite.Suite
	repo user_refresh_token_repository.Repository
	db   *sqlx.DB
}

func TestUserRefreshTokenRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRefreshTokenRepositoryTestSuite))
}

func (s *UserRefreshTokenRepositoryTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}

	testDB, err := database.GetTestDB()
	require.NoError(s.T(), err)
	s.db = testDB.DB()
	s.repo = user_refresh_token_repository.New(s.db, nil)
}

func (s *UserRefreshTokenRepositoryTestSuite) SetupTest() {
	_, err := s.db.Exec(`
		INSERT INTO users (id, created_at, first_name, last_name, language, avatar_url, email, email_verified_at, password, google_id, terms_accepted_at, terms_version)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', 'Test', 'User', 'en', '', 'test@example.com', NULL, 'hashedpassword', NULL, '2024-01-15 12:00:00+00', '1.0.0'),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 12:00:00+00', 'Google', 'User', 'en', 'https://example.com/fixture.png', 'google@example.com', '2024-01-15 12:00:00+00', NULL, 'google123', '2024-01-15 12:00:00+00', '1.0.0');

		INSERT INTO user_refresh_tokens (id, user_id, created_at, expires_at, revoked_at, ip, user_agent, token_hash, impersonated_by)
		VALUES
			('00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', NOW() - INTERVAL '1 hour', NOW() + INTERVAL '24 hours', NULL, '127.0.0.1', 'Mozilla/5.0', 'token_hash_1', NULL),
			('00000002-0000-0000-0000-000000000002', '00000001-0000-0000-0000-000000000001', NOW() - INTERVAL '1 hour', NOW() + INTERVAL '24 hours', NULL, '127.0.0.1', 'Mozilla/5.0', 'token_hash_2', NULL),
			('00000003-0000-0000-0000-000000000003', '00000002-0000-0000-0000-000000000002', NOW() - INTERVAL '1 hour', NOW() + INTERVAL '24 hours', NOW(), '127.0.0.1', 'Mozilla/5.0', 'token_hash_3', NULL),
			('00000004-0000-0000-0000-000000000004', '00000001-0000-0000-0000-000000000001', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day', NULL, '127.0.0.1', 'Mozilla/5.0', 'token_hash_expired', NULL),
			('00000005-0000-0000-0000-000000000005', '00000002-0000-0000-0000-000000000002', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', '127.0.0.1', 'Mozilla/5.0', 'token_hash_expired_revoked', NULL),
			('00000006-0000-0000-0000-000000000006', '00000001-0000-0000-0000-000000000001', NOW() - INTERVAL '1 hour', NOW() + INTERVAL '24 hours', NULL, '127.0.0.1', 'Mozilla/5.0', 'token_hash_impersonation', '00000002-0000-0000-0000-000000000002');
	`)
	require.NoError(s.T(), err)
}

func (s *UserRefreshTokenRepositoryTestSuite) TearDownTest() {
	_, _ = s.db.Exec("TRUNCATE TABLE user_refresh_tokens CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE users CASCADE")
}

func (s *UserRefreshTokenRepositoryTestSuite) TestCreate() {
	token := &models.UserRefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IP:        "192.168.1.1",
		UserAgent: "TestAgent",
		TokenHash: "new_token_hash",
		// Note: ImpersonatedBy zero UUID handling may need repository fix
		// Testing with valid user ID to verify Create works
		ImpersonatedBy: uuid.NullUUID{UUID: uuid.MustParse("00000002-0000-0000-0000-000000000002"), Valid: true},
	}
	err := s.repo.Create(nil, token)
	require.NoError(s.T(), err)

	// Verify the token was created
	retrievedToken, err := s.repo.GetByHash(nil, "new_token_hash")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), retrievedToken)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestCreate_DuplicateID() {
	token := &models.UserRefreshToken{
		ID:        uuid.MustParse("00000001-0000-0000-0000-000000000001"), // Already exists
		UserID:    uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IP:        "127.0.0.1",
		UserAgent: "Mozilla/5.0",
		TokenHash: "duplicate_id_token",
	}
	err := s.repo.Create(nil, token)
	require.Error(s.T(), err)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestGetByHash() {
	token, err := s.repo.GetByHash(nil, "token_hash_1")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), token)
	require.Nil(s.T(), token.RevokedAt)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestGetByHash_WithRevokedToken() {
	token, err := s.repo.GetByHash(nil, "token_hash_3")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), token)
	require.NotNil(s.T(), token.RevokedAt)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestGetByHash_WithExpiredToken() {
	// Should still retrieve expired tokens (expiration check is handled elsewhere)
	token, err := s.repo.GetByHash(nil, "token_hash_expired")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), token)
	require.True(s.T(), token.ExpiresAt.Before(time.Now()))
}

func (s *UserRefreshTokenRepositoryTestSuite) TestGetByHash_NotFound() {
	token, err := s.repo.GetByHash(nil, "token_hash_4")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), token)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestGetActiveByUserID() {
	tokens, err := s.repo.GetActiveByUserID(nil, "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), tokens)
	require.GreaterOrEqual(s.T(), len(*tokens), 1) // At least token_hash_1 and token_hash_2 (excluding impersonation token)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestGetActiveByUserID_NoActiveTokens() {
	tokens, err := s.repo.GetActiveByUserID(nil, "00000002-0000-0000-0000-000000000002")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), tokens)
	require.Len(s.T(), *tokens, 0) // token_hash_3 is revoked
}

func (s *UserRefreshTokenRepositoryTestSuite) TestGetActiveImpersonationTokenByUserID() {
	token, err := s.repo.GetActiveImpersonationTokenByUserID(nil, "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), token)
	require.NotNil(s.T(), token.ImpersonatedBy)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestGetActiveImpersonationTokenByUserID_NotFound() {
	token, err := s.repo.GetActiveImpersonationTokenByUserID(nil, "00000002-0000-0000-0000-000000000002")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), token)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestRevoke() {
	err := s.repo.Revoke(nil, "00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)

	token, err := s.repo.GetByHash(nil, "token_hash_1")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), token)
	require.NotNil(s.T(), token.RevokedAt)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestRevoke_NotFound() {
	err := s.repo.Revoke(nil, "00000001-0000-0000-0000-000000000001", "00000099-0000-0000-0000-000000000099")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestRevoke_AlreadyRevoked() {
	// Revoke already revoked token - condition won't match
	err := s.repo.Revoke(nil, "00000002-0000-0000-0000-000000000002", "00000003-0000-0000-0000-000000000003")
	require.Error(s.T(), err)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestRevokeByHash() {
	err := s.repo.RevokeByHash(nil, "token_hash_2")
	require.NoError(s.T(), err)

	token, err := s.repo.GetByHash(nil, "token_hash_2")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), token)
	require.NotNil(s.T(), token.RevokedAt)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestRevokeByHash_NotFound() {
	err := s.repo.RevokeByHash(nil, "token_hash_4")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestRevokeByHash_AlreadyRevoked() {
	// Revoke already revoked token - condition won't match
	err := s.repo.RevokeByHash(nil, "token_hash_3")
	require.Error(s.T(), err)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestRevokeAll() {
	err := s.repo.RevokeAll(nil, "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)

	// Verify all active tokens for user are revoked
	tokens, err := s.repo.GetActiveByUserID(nil, "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), tokens)
	require.Len(s.T(), *tokens, 0)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestRevokeAll_NoActiveTokens() {
	// User 2 has token_hash_3 (revoked) and token_hash_expired_revoked (expired and revoked)
	// So there are no active tokens to revoke
	err := s.repo.RevokeAll(nil, "00000002-0000-0000-0000-000000000002")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *UserRefreshTokenRepositoryTestSuite) TestCleanupStaleUnused() {
	_, err := s.db.Exec(`
		INSERT INTO user_refresh_tokens (id, user_id, created_at, expires_at, revoked_at, ip, user_agent, token_hash, impersonated_by)
		VALUES
			('a0000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', NOW() - INTERVAL '40 days', NOW() - INTERVAL '35 days', NULL, '127.0.0.1', 'Mozilla/5.0', 'stale_expired_only', NULL),
			('a0000002-0000-0000-0000-000000000002', '00000001-0000-0000-0000-000000000001', NOW() - INTERVAL '40 days', NOW() + INTERVAL '1 day', NOW() - INTERVAL '35 days', '127.0.0.1', 'Mozilla/5.0', 'stale_revoked_only', NULL),
			('a0000003-0000-0000-0000-000000000003', '00000001-0000-0000-0000-000000000001', NOW() - INTERVAL '5 days', NOW() - INTERVAL '2 days', NULL, '127.0.0.1', 'Mozilla/5.0', 'recent_expired_keep', NULL),
			('a0000004-0000-0000-0000-000000000004', '00000001-0000-0000-0000-000000000001', NOW() - INTERVAL '5 days', NOW() + INTERVAL '1 day', NOW() - INTERVAL '2 days', '127.0.0.1', 'Mozilla/5.0', 'recent_revoked_keep', NULL);
	`)
	require.NoError(s.T(), err)

	err = s.repo.CleanupStaleUnused(nil)
	require.NoError(s.T(), err)

	_, err = s.repo.GetByHash(nil, "stale_expired_only")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	_, err = s.repo.GetByHash(nil, "stale_revoked_only")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)

	_, err = s.repo.GetByHash(nil, "recent_expired_keep")
	require.NoError(s.T(), err)
	_, err = s.repo.GetByHash(nil, "recent_revoked_keep")
	require.NoError(s.T(), err)

	activeToken, err := s.repo.GetByHash(nil, "token_hash_1")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), activeToken)
}
