package organization_member_repository_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/database"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	organization_member_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization_member"
	yca_error "github.com/yca-software/go-common/error"
)

type OrganizationMemberRepositoryTestSuite struct {
	suite.Suite
	repo organization_member_repository.Repository
	db   *sqlx.DB
}

func TestOrganizationMemberRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationMemberRepositoryTestSuite))
}

func (s *OrganizationMemberRepositoryTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}

	testDB, err := database.GetTestDB()
	require.NoError(s.T(), err)
	s.db = testDB.DB()
	s.repo = organization_member_repository.New(s.db, nil)
}

func (s *OrganizationMemberRepositoryTestSuite) SetupTest() {
	// Insert test organizations
	_, err := s.db.Exec(`
		INSERT INTO organizations (id, created_at, name, address, city, zip, country, place_id, geo, timezone, billing_email, custom_subscription, subscription_expires_at, subscription_type, subscription_seats, paddle_customer_id)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', 'Test Org 1', '123 Main St', 'New York', '10001', 'US', 'place_123', ST_MakePoint(-74.006, 40.7128), 'America/New_York', 'billing1@example.com', false, NULL, 1, 10, 'paddle_customer_1'),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 13:00:00+00', 'Test Org 2', '456 Oak Ave', 'Los Angeles', '90001', 'US', 'place_456', ST_MakePoint(-118.2437, 34.0522), 'America/Los_Angeles', 'billing2@example.com', true, '2025-12-31 23:59:59+00', 2, 20, 'paddle_customer_2');
	`)
	require.NoError(s.T(), err)

	// Insert test users
	_, err = s.db.Exec(`
		INSERT INTO users (id, created_at, first_name, last_name, language, avatar_url, email, email_verified_at, password, google_id, terms_accepted_at, terms_version)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', 'John', 'Doe', 'en', '', 'john@example.com', NULL, 'hashedpassword', NULL, '2024-01-15 12:00:00+00', '1.0.0'),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 13:00:00+00', 'Jane', 'Smith', 'en', '', 'jane@example.com', NULL, 'hashedpassword', NULL, '2024-01-15 13:00:00+00', '1.0.0');
	`)
	require.NoError(s.T(), err)

	// Insert test roles
	permissions, _ := json.Marshal([]string{"read", "write"})
	_, err = s.db.Exec(`
		INSERT INTO roles (id, created_at, organization_id, name, description, permissions, locked)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', '00000001-0000-0000-0000-000000000001', 'Admin', 'Admin role', $1, false),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 13:00:00+00', '00000001-0000-0000-0000-000000000001', 'User', 'User role', $1, false);
	`, permissions)
	require.NoError(s.T(), err)

	// Insert test organization members
	_, err = s.db.Exec(`
		INSERT INTO organization_members (id, created_at, organization_id, user_id, role_id)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001'),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 13:00:00+00', '00000001-0000-0000-0000-000000000001', '00000002-0000-0000-0000-000000000002', '00000002-0000-0000-0000-000000000002');
	`)
	require.NoError(s.T(), err)
}

func (s *OrganizationMemberRepositoryTestSuite) TearDownTest() {
	_, _ = s.db.Exec("TRUNCATE TABLE organization_members CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE roles CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE users CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE organizations CASCADE")
}

func (s *OrganizationMemberRepositoryTestSuite) TestCreate_DuplicateUserOrg() {
	member := &models.OrganizationMember{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		UserID:         uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		RoleID:         uuid.MustParse("00000001-0000-0000-0000-000000000001"),
	}
	err := s.repo.Create(nil, member)
	require.Error(s.T(), err)
	require.Equal(s.T(), 409, err.(*yca_error.Error).StatusCode)
}

func (s *OrganizationMemberRepositoryTestSuite) TestCreate_Success() {
	member := &models.OrganizationMember{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		OrganizationID: uuid.MustParse("00000002-0000-0000-0000-000000000002"),
		UserID:         uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		RoleID:         uuid.MustParse("00000001-0000-0000-0000-000000000001"),
	}
	err := s.repo.Create(nil, member)
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID(member.OrganizationID.String(), member.ID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), member.UserID, got.UserID)
}

func (s *OrganizationMemberRepositoryTestSuite) TestGetByID_NotFound() {
	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", uuid.New().String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *OrganizationMemberRepositoryTestSuite) TestGetByID_Success() {
	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), uuid.MustParse("00000001-0000-0000-0000-000000000001"), got.UserID)
}

func (s *OrganizationMemberRepositoryTestSuite) TestGetByUserIDAndOrganizationID_NotFound() {
	got, err := s.repo.GetByUserIDAndOrganizationID(uuid.New().String(), "00000001-0000-0000-0000-000000000001")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *OrganizationMemberRepositoryTestSuite) TestGetByUserIDAndOrganizationID_Success() {
	got, err := s.repo.GetByUserIDAndOrganizationID("00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), uuid.MustParse("00000001-0000-0000-0000-000000000001"), got.ID)
}

func (s *OrganizationMemberRepositoryTestSuite) TestGetByIDWithUser_Success() {
	got, err := s.repo.GetByIDWithUser("00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), "john@example.com", got.UserEmail)
	require.Equal(s.T(), "John", got.UserFirstName)
	require.Equal(s.T(), "Doe", got.UserLastName)
}

func (s *OrganizationMemberRepositoryTestSuite) TestGetByIDWithRole_Success() {
	got, err := s.repo.GetByIDWithRole("00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), "Test Org 1", got.OrganizationName)
	require.Equal(s.T(), "Admin", got.RoleName)
}

func (s *OrganizationMemberRepositoryTestSuite) TestListByUserID() {
	members, err := s.repo.ListByUserID("00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), members)
	require.Len(s.T(), *members, 1)
	require.Equal(s.T(), "Test Org 1", (*members)[0].OrganizationName)
}

func (s *OrganizationMemberRepositoryTestSuite) TestListByUserIDWithRole() {
	members, err := s.repo.ListByUserIDWithRole("00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), members)
	require.Len(s.T(), *members, 1)
	require.Equal(s.T(), "Test Org 1", (*members)[0].OrganizationName)
	require.Equal(s.T(), "Admin", (*members)[0].RoleName)
}

func (s *OrganizationMemberRepositoryTestSuite) TestListByOrganizationID() {
	members, err := s.repo.ListByOrganizationID("00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), members)
	require.Len(s.T(), *members, 2)
}

func (s *OrganizationMemberRepositoryTestSuite) TestListUserEmailsForRole_Admin() {
	emails, err := s.repo.ListUserEmailsForRole(
		"00000001-0000-0000-0000-000000000001",
		"00000001-0000-0000-0000-000000000001",
	)
	require.NoError(s.T(), err)
	require.Equal(s.T(), []string{"john@example.com"}, emails)
}

func (s *OrganizationMemberRepositoryTestSuite) TestListUserEmailsForRole_User() {
	emails, err := s.repo.ListUserEmailsForRole(
		"00000001-0000-0000-0000-000000000001",
		"00000002-0000-0000-0000-000000000002",
	)
	require.NoError(s.T(), err)
	require.Equal(s.T(), []string{"jane@example.com"}, emails)
}

func (s *OrganizationMemberRepositoryTestSuite) TestListUserEmailsForRole_Empty() {
	emails, err := s.repo.ListUserEmailsForRole(
		"00000002-0000-0000-0000-000000000002",
		"00000001-0000-0000-0000-000000000001",
	)
	require.NoError(s.T(), err)
	require.Empty(s.T(), emails)
}

func (s *OrganizationMemberRepositoryTestSuite) TestUpdate_NotFound() {
	member := &models.OrganizationMember{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		UserID:         uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		RoleID:         uuid.MustParse("00000002-0000-0000-0000-000000000002"),
	}
	err := s.repo.Update(nil, member)
	require.Error(s.T(), err)
	var ycaErr *yca_error.Error
	require.True(s.T(), errors.As(err, &ycaErr), "error should be a yca_error.Error")
	require.Equal(s.T(), http.StatusNotFound, ycaErr.StatusCode)
}

func (s *OrganizationMemberRepositoryTestSuite) TestUpdate_Success() {
	member := &models.OrganizationMember{
		ID:             uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		CreatedAt:      time.Now(),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		UserID:         uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		RoleID:         uuid.MustParse("00000002-0000-0000-0000-000000000002"),
	}
	err := s.repo.Update(nil, member)
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID(member.OrganizationID.String(), member.ID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), uuid.MustParse("00000002-0000-0000-0000-000000000002"), got.RoleID)
}

func (s *OrganizationMemberRepositoryTestSuite) TestDelete_NotFound() {
	err := s.repo.Delete(nil, "00000001-0000-0000-0000-000000000001", uuid.New().String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *OrganizationMemberRepositoryTestSuite) TestDelete_Success() {
	err := s.repo.Delete(nil, "00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.Error(s.T(), err)
	require.Nil(s.T(), got)
}
