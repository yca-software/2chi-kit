package user_password_reset_token_repository_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/database"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	user_password_reset_token_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user_password_reset_token"
	yca_error "github.com/yca-software/go-common/error"
)

type UserPasswordResetTokenRepositoryTestSuite struct {
	suite.Suite
	repo user_password_reset_token_repository.Repository
	db   *sqlx.DB
}

func TestUserPasswordResetTokenRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserPasswordResetTokenRepositoryTestSuite))
}

func (s *UserPasswordResetTokenRepositoryTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}

	testDB, err := database.GetTestDB()
	require.NoError(s.T(), err)
	s.db = testDB.DB()
	s.repo = user_password_reset_token_repository.New(s.db, nil)
}

func (s *UserPasswordResetTokenRepositoryTestSuite) SetupTest() {
	_, err := s.db.Exec(`
		INSERT INTO users (id, created_at, first_name, last_name, language, avatar_url, email, email_verified_at, password, google_id, terms_accepted_at, terms_version)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', 'Test', 'User', 'en', '', 'test@example.com', NULL, 'hashedpassword', NULL, '2024-01-15 12:00:00+00', '1.0.0'),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 12:00:00+00', 'Google', 'User', 'en', 'https://example.com/fixture.png', 'google@example.com', '2024-01-15 12:00:00+00', NULL, 'google123', '2024-01-15 12:00:00+00', '1.0.0');

		INSERT INTO user_password_reset_tokens (id, user_id, created_at, expires_at, used_at, token_hash)
		VALUES
			('00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', NOW() + INTERVAL '400 days', NULL, 'token_hash_1'),
			('00000002-0000-0000-0000-000000000002', '00000002-0000-0000-0000-000000000002', '2024-01-15 12:00:00+00', NOW() + INTERVAL '400 days', NULL, 'token_hash_2'),
			('00000003-0000-0000-0000-000000000003', '00000002-0000-0000-0000-000000000002', '2024-01-15 12:00:00+00', NOW() + INTERVAL '400 days', NOW() - INTERVAL '10 days', 'token_hash_3'),
			('00000004-0000-0000-0000-000000000004', '00000001-0000-0000-0000-000000000001', '2023-01-15 12:00:00+00', '2023-12-31 12:00:00+00', NULL, 'token_hash_expired'),
			('00000005-0000-0000-0000-000000000005', '00000002-0000-0000-0000-000000000002', '2023-01-15 12:00:00+00', '2023-12-31 12:00:00+00', '2023-12-30 12:00:00+00', 'token_hash_expired_used');
	`)
	require.NoError(s.T(), err)
}

func (s *UserPasswordResetTokenRepositoryTestSuite) TearDownTest() {
	_, _ = s.db.Exec("TRUNCATE TABLE user_password_reset_tokens CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE users CASCADE")
}

func (s *UserPasswordResetTokenRepositoryTestSuite) TestCreate() {
	token := &models.UserPasswordResetToken{
		ID:        uuid.New(),
		UserID:    uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		TokenHash: "new_token_hash",
	}
	err := s.repo.Create(nil, token)
	require.NoError(s.T(), err)

	// Verify the token was created
	retrievedToken, err := s.repo.GetByHash(nil, "new_token_hash")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), retrievedToken)
}

func (s *UserPasswordResetTokenRepositoryTestSuite) TestCreate_DuplicateID() {
	token := &models.UserPasswordResetToken{
		ID:        uuid.MustParse("00000001-0000-0000-0000-000000000001"), // Already exists
		UserID:    uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		TokenHash: "duplicate_id_token",
	}
	err := s.repo.Create(nil, token)
	require.Error(s.T(), err)
}

func (s *UserPasswordResetTokenRepositoryTestSuite) TestGetByHash() {
	token, err := s.repo.GetByHash(nil, "token_hash_1")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), token)
	require.Nil(s.T(), token.UsedAt)
}

func (s *UserPasswordResetTokenRepositoryTestSuite) TestGetByHash_WithUsedToken() {
	token, err := s.repo.GetByHash(nil, "token_hash_3")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), token)
	require.NotNil(s.T(), token.UsedAt)
}

func (s *UserPasswordResetTokenRepositoryTestSuite) TestGetByHash_WithExpiredToken() {
	// Should still retrieve expired tokens (expiration check is handled elsewhere)
	token, err := s.repo.GetByHash(nil, "token_hash_expired")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), token)
	require.True(s.T(), token.ExpiresAt.Before(time.Now()))
}

func (s *UserPasswordResetTokenRepositoryTestSuite) TestGetByHash_NotFound() {
	token, err := s.repo.GetByHash(nil, "token_hash_4")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), token)
}

func (s *UserPasswordResetTokenRepositoryTestSuite) TestMarkAsUsed() {
	err := s.repo.MarkAsUsed(nil, "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)

	token, err := s.repo.GetByHash(nil, "token_hash_1")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), token)
	require.NotNil(s.T(), token.UsedAt)
}

func (s *UserPasswordResetTokenRepositoryTestSuite) TestMarkAsUsed_NotFound() {
	err := s.repo.MarkAsUsed(nil, "00000099-0000-0000-0000-000000000099")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *UserPasswordResetTokenRepositoryTestSuite) TestMarkAsUsed_AlreadyUsed() {
	// Mark as used again - since used_at is already set, the condition won't match
	// and no rows will be affected, resulting in an error
	err := s.repo.MarkAsUsed(nil, "00000003-0000-0000-0000-000000000003")
	require.Error(s.T(), err)
}

func (s *UserPasswordResetTokenRepositoryTestSuite) TestCleanup() {
	_, err := s.db.Exec(`
		INSERT INTO user_password_reset_tokens (id, user_id, created_at, expires_at, used_at, token_hash)
		VALUES
			('b0000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', NOW() - INTERVAL '40 days', NOW() - INTERVAL '35 days', NULL, 'stale_unused_expired'),
			('b0000002-0000-0000-0000-000000000002', '00000001-0000-0000-0000-000000000001', NOW() - INTERVAL '40 days', NOW() + INTERVAL '1 day', NOW() - INTERVAL '35 days', 'stale_used'),
			('b0000003-0000-0000-0000-000000000003', '00000001-0000-0000-0000-000000000001', NOW() - INTERVAL '10 days', NOW() - INTERVAL '5 days', NULL, 'recent_expired_keep');
	`)
	require.NoError(s.T(), err)

	err = s.repo.Cleanup(nil)
	require.NoError(s.T(), err)

	_, err = s.repo.GetByHash(nil, "stale_unused_expired")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	_, err = s.repo.GetByHash(nil, "stale_used")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)

	tok, err := s.repo.GetByHash(nil, "recent_expired_keep")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), tok)

	_, err = s.repo.GetByHash(nil, "token_hash_1")
	require.NoError(s.T(), err)
}
