package invitation_repository_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/database"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	invitation_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/invitation"
	yca_error "github.com/yca-software/go-common/error"
)

type InvitationRepositoryTestSuite struct {
	suite.Suite
	repo invitation_repository.Repository
	db   *sqlx.DB
}

func TestInvitationRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(InvitationRepositoryTestSuite))
}

func (s *InvitationRepositoryTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}

	testDB, err := database.GetTestDB()
	require.NoError(s.T(), err)
	s.db = testDB.DB()
	s.repo = invitation_repository.New(s.db, nil)
}

func (s *InvitationRepositoryTestSuite) SetupTest() {
	// Insert test users
	_, err := s.db.Exec(`
		INSERT INTO users (id, created_at, first_name, last_name, language, avatar_url, email, email_verified_at, password, google_id, terms_accepted_at, terms_version)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', 'Admin', 'User', 'en', '', 'admin@example.com', NULL, 'hashedpassword', NULL, '2024-01-15 12:00:00+00', '1.0.0');
	`)
	require.NoError(s.T(), err)

	// Insert test organizations
	_, err = s.db.Exec(`
		INSERT INTO organizations (id, created_at, name, address, city, zip, country, place_id, geo, timezone, billing_email, custom_subscription, subscription_expires_at, subscription_type, subscription_seats, paddle_customer_id)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', 'Test Org 1', '123 Main St', 'New York', '10001', 'US', 'place_123', ST_MakePoint(-74.006, 40.7128), 'America/New_York', 'billing1@example.com', false, NULL, 1, 10, 'paddle_customer_1');
	`)
	require.NoError(s.T(), err)

	// Insert test roles
	_, err = s.db.Exec(`
		INSERT INTO roles (id, created_at, organization_id, name, description, permissions, locked)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', '00000001-0000-0000-0000-000000000001', 'Admin', 'Admin role', '[]'::jsonb, false);
	`)
	require.NoError(s.T(), err)

	// Insert test invitations
	_, err = s.db.Exec(`
		INSERT INTO invitations (id, created_at, expires_at, organization_id, role_id, email, invited_by_id, invited_by_email, token_hash, accepted_at, revoked_at)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', NOW() + INTERVAL '90 days', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', 'invite1@example.com', '00000001-0000-0000-0000-000000000001', 'admin@example.com', 'token_hash_1', NULL, NULL),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 13:00:00+00', NOW() + INTERVAL '90 days', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', 'invite2@example.com', '00000001-0000-0000-0000-000000000001', 'admin@example.com', 'token_hash_2', NULL, NULL),
			('00000003-0000-0000-0000-000000000003', '2024-01-15 14:00:00+00', '2024-02-15 14:00:00+00', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', 'accepted@example.com', '00000001-0000-0000-0000-000000000001', 'admin@example.com', 'token_hash_3', '2024-01-16 10:00:00+00', NULL),
			('00000004-0000-0000-0000-000000000004', '2024-01-15 15:00:00+00', '2024-02-15 15:00:00+00', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', 'revoked@example.com', '00000001-0000-0000-0000-000000000001', 'admin@example.com', 'token_hash_4', NULL, '2024-01-16 10:00:00+00');
	`)
	require.NoError(s.T(), err)
}

func (s *InvitationRepositoryTestSuite) TearDownTest() {
	_, _ = s.db.Exec("TRUNCATE TABLE invitations CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE roles CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE organizations CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE users CASCADE")
}

func (s *InvitationRepositoryTestSuite) TestCreate_DuplicateTokenHash() {
	invitation := &models.Invitation{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		RoleID:         uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		Email:          "new@example.com",
		InvitedByID:    uuid.NullUUID{UUID: uuid.MustParse("00000001-0000-0000-0000-000000000001"), Valid: true},
		InvitedByEmail: "admin@example.com",
		TokenHash:      "token_hash_1",
	}
	err := s.repo.Create(nil, invitation)
	require.Error(s.T(), err)
	require.Equal(s.T(), 409, err.(*yca_error.Error).StatusCode)
}

func (s *InvitationRepositoryTestSuite) TestCreate_Success() {
	invitation := &models.Invitation{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		RoleID:         uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		Email:          "new@example.com",
		InvitedByID:    uuid.NullUUID{UUID: uuid.MustParse("00000001-0000-0000-0000-000000000001"), Valid: true},
		InvitedByEmail: "admin@example.com",
		TokenHash:      "new_token_hash",
	}
	err := s.repo.Create(nil, invitation)
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID(invitation.OrganizationID.String(), invitation.ID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), invitation.Email, got.Email)
}

func (s *InvitationRepositoryTestSuite) TestGetByID_NotFound() {
	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", uuid.New().String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *InvitationRepositoryTestSuite) TestGetByID_Success() {
	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), "invite1@example.com", got.Email)
}

func (s *InvitationRepositoryTestSuite) TestGetByTokenHash_NotFound() {
	got, err := s.repo.GetByTokenHash("nonexistent_token")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *InvitationRepositoryTestSuite) TestGetByTokenHash_Success() {
	got, err := s.repo.GetByTokenHash("token_hash_1")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), "invite1@example.com", got.Email)
}

func (s *InvitationRepositoryTestSuite) TestUpdate_NotFound() {
	invitation := &models.Invitation{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		RoleID:         uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		Email:          "test@example.com",
		InvitedByID:    uuid.NullUUID{UUID: uuid.MustParse("00000001-0000-0000-0000-000000000001"), Valid: true},
		InvitedByEmail: "admin@example.com",
		TokenHash:      "test_token",
		AcceptedAt:     nil,
		RevokedAt:      nil,
	}
	err := s.repo.Update(nil, invitation)
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *InvitationRepositoryTestSuite) TestUpdate_AlreadyAccepted() {
	now := time.Now()
	invitation := &models.Invitation{
		ID:             uuid.MustParse("00000003-0000-0000-0000-000000000003"),
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		RoleID:         uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		Email:          "accepted@example.com",
		InvitedByID:    uuid.NullUUID{UUID: uuid.MustParse("00000001-0000-0000-0000-000000000001"), Valid: true},
		InvitedByEmail: "admin@example.com",
		TokenHash:      "token_hash_3",
		AcceptedAt:     &now,
		RevokedAt:      nil,
	}
	err := s.repo.Update(nil, invitation)
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *InvitationRepositoryTestSuite) TestUpdate_Success() {
	now := time.Now()
	invitation := &models.Invitation{
		ID:             uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		RoleID:         uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		Email:          "invite1@example.com",
		InvitedByID:    uuid.NullUUID{UUID: uuid.MustParse("00000001-0000-0000-0000-000000000001"), Valid: true},
		InvitedByEmail: "admin@example.com",
		TokenHash:      "token_hash_1",
		AcceptedAt:     &now,
		RevokedAt:      nil,
	}
	err := s.repo.Update(nil, invitation)
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.NotNil(s.T(), got.AcceptedAt)
}

func (s *InvitationRepositoryTestSuite) TestListByOrganizationID() {
	invitations, err := s.repo.ListByOrganizationID("00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), invitations)
	require.Len(s.T(), *invitations, 2) // Only non-accepted, non-revoked
}

func (s *InvitationRepositoryTestSuite) TestCleanupStale() {
	_, err := s.db.Exec(`
		INSERT INTO invitations (id, created_at, expires_at, organization_id, role_id, email, invited_by_id, invited_by_email, token_hash, accepted_at, revoked_at)
		VALUES
			('c0000001-0000-0000-0000-000000000001', NOW() - INTERVAL '40 days', NOW() - INTERVAL '20 days', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', 'cleanup_a@example.com', '00000001-0000-0000-0000-000000000001', 'admin@example.com', 'cleanup_stale_accepted', NOW() - INTERVAL '35 days', NULL),
			('c0000002-0000-0000-0000-000000000002', NOW() - INTERVAL '40 days', NOW() + INTERVAL '10 days', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', 'cleanup_r@example.com', '00000001-0000-0000-0000-000000000001', 'admin@example.com', 'cleanup_stale_revoked', NULL, NOW() - INTERVAL '35 days'),
			('c0000003-0000-0000-0000-000000000003', NOW() - INTERVAL '120 days', NOW() - INTERVAL '95 days', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', 'cleanup_p@example.com', '00000001-0000-0000-0000-000000000001', 'admin@example.com', 'cleanup_stale_pending_expired', NULL, NULL),
			('c0000004-0000-0000-0000-000000000004', NOW() - INTERVAL '10 days', NOW() - INTERVAL '5 days', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', 'cleanup_recent@example.com', '00000001-0000-0000-0000-000000000001', 'admin@example.com', 'cleanup_recent_pending_expired', NULL, NULL);
	`)
	require.NoError(s.T(), err)

	err = s.repo.CleanupStale()
	require.NoError(s.T(), err)

	for _, h := range []string{"cleanup_stale_accepted", "cleanup_stale_revoked", "cleanup_stale_pending_expired"} {
		_, err := s.repo.GetByTokenHash(h)
		require.Error(s.T(), err, h)
		require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	}

	got, err := s.repo.GetByTokenHash("cleanup_recent_pending_expired")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
}
