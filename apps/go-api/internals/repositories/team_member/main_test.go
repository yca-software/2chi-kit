package team_member_repository_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/database"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	team_member_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/team_member"
	yca_error "github.com/yca-software/go-common/error"
)

type TeamMemberRepositoryTestSuite struct {
	suite.Suite
	repo team_member_repository.Repository
	db   *sqlx.DB
}

func TestTeamMemberRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(TeamMemberRepositoryTestSuite))
}

func (s *TeamMemberRepositoryTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}

	testDB, err := database.GetTestDB()
	require.NoError(s.T(), err)
	s.db = testDB.DB()
	s.repo = team_member_repository.New(s.db, nil)
}

func (s *TeamMemberRepositoryTestSuite) SetupTest() {
	// Insert test organizations
	_, err := s.db.Exec(`
		INSERT INTO organizations (id, created_at, name, address, city, zip, country, place_id, geo, timezone, billing_email, custom_subscription, subscription_expires_at, subscription_type, subscription_seats, paddle_customer_id)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', 'Test Org 1', '123 Main St', 'New York', '10001', 'US', 'place_123', ST_MakePoint(-74.006, 40.7128), 'America/New_York', 'billing1@example.com', false, NULL, 1, 10, 'paddle_customer_1');
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

	// Insert test teams
	_, err = s.db.Exec(`
		INSERT INTO teams (id, created_at, organization_id, name, description)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', '00000001-0000-0000-0000-000000000001', 'Team Alpha', 'Alpha team description'),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 13:00:00+00', '00000001-0000-0000-0000-000000000001', 'Team Beta', 'Beta team description');
	`)
	require.NoError(s.T(), err)

	// Insert test team members
	_, err = s.db.Exec(`
		INSERT INTO team_members (id, created_at, organization_id, team_id, user_id)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001'),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 13:00:00+00', '00000001-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', '00000002-0000-0000-0000-000000000002'),
			('00000003-0000-0000-0000-000000000003', '2024-01-15 14:00:00+00', '00000001-0000-0000-0000-000000000001', '00000002-0000-0000-0000-000000000002', '00000001-0000-0000-0000-000000000001');
	`)
	require.NoError(s.T(), err)
}

func (s *TeamMemberRepositoryTestSuite) TearDownTest() {
	_, _ = s.db.Exec("TRUNCATE TABLE team_members CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE teams CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE users CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE organizations CASCADE")
}

func (s *TeamMemberRepositoryTestSuite) TestCreate_DuplicateUserTeam() {
	member := &models.TeamMember{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		TeamID:         uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		UserID:         uuid.MustParse("00000001-0000-0000-0000-000000000001"),
	}
	err := s.repo.Create(nil, member)
	require.Error(s.T(), err)
	require.Equal(s.T(), 409, err.(*yca_error.Error).StatusCode)
}

func (s *TeamMemberRepositoryTestSuite) TestCreate_Success() {
	member := &models.TeamMember{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		TeamID:         uuid.MustParse("00000002-0000-0000-0000-000000000002"),
		UserID:         uuid.MustParse("00000002-0000-0000-0000-000000000002"),
	}
	err := s.repo.Create(nil, member)
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID(member.OrganizationID.String(), member.ID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), member.UserID, got.UserID)
}

func (s *TeamMemberRepositoryTestSuite) TestGetByID_NotFound() {
	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", uuid.New().String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *TeamMemberRepositoryTestSuite) TestGetByID_Success() {
	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), uuid.MustParse("00000001-0000-0000-0000-000000000001"), got.UserID)
}

func (s *TeamMemberRepositoryTestSuite) TestGetByIDWithUser_Success() {
	got, err := s.repo.GetByIDWithUser(nil, "00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), "john@example.com", got.UserEmail)
	require.Equal(s.T(), "John", got.UserFirstName)
	require.Equal(s.T(), "Doe", got.UserLastName)
}

func (s *TeamMemberRepositoryTestSuite) TestListByUserID() {
	members, err := s.repo.ListByUserID("00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), members)
	require.Len(s.T(), *members, 2)
	require.Equal(s.T(), "Team Alpha", (*members)[0].TeamName)
}

func (s *TeamMemberRepositoryTestSuite) TestListByTeamID() {
	members, err := s.repo.ListByTeamID("00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), members)
	require.Len(s.T(), *members, 2)
}

func (s *TeamMemberRepositoryTestSuite) TestListByOrganizationID() {
	members, err := s.repo.ListByOrganizationID("00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), members)
	require.Len(s.T(), *members, 3)
}

func (s *TeamMemberRepositoryTestSuite) TestDelete_NotFound() {
	err := s.repo.Delete(nil, "00000001-0000-0000-0000-000000000001", uuid.New().String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *TeamMemberRepositoryTestSuite) TestDelete_Success() {
	err := s.repo.Delete(nil, "00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.Error(s.T(), err)
	require.Nil(s.T(), got)
}
