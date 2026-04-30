package team_repository_test

import (
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
	team_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/team"
	yca_error "github.com/yca-software/go-common/error"
)

type TeamRepositoryTestSuite struct {
	suite.Suite
	repo team_repository.Repository
	db   *sqlx.DB
}

func TestTeamRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(TeamRepositoryTestSuite))
}

func (s *TeamRepositoryTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}

	testDB, err := database.GetTestDB()
	require.NoError(s.T(), err)
	s.db = testDB.DB()
	s.repo = team_repository.New(s.db, nil)
}

func (s *TeamRepositoryTestSuite) SetupTest() {
	// Insert test organizations
	_, err := s.db.Exec(`
		INSERT INTO organizations (id, created_at, name, address, city, zip, country, place_id, geo, timezone, billing_email, custom_subscription, subscription_expires_at, subscription_type, subscription_seats, paddle_customer_id)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', 'Test Org 1', '123 Main St', 'New York', '10001', 'US', 'place_123', ST_MakePoint(-74.006, 40.7128), 'America/New_York', 'billing1@example.com', false, NULL, 1, 10, 'paddle_customer_1');
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
}

func (s *TeamRepositoryTestSuite) TearDownTest() {
	_, _ = s.db.Exec("TRUNCATE TABLE teams CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE organizations CASCADE")
}

func (s *TeamRepositoryTestSuite) TestCreate_Success() {
	team := &models.Team{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		Name:           "New Team",
		Description:    "New team description",
	}
	err := s.repo.Create(nil, team)
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID(team.OrganizationID.String(), team.ID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), team.Name, got.Name)
}

func (s *TeamRepositoryTestSuite) TestGetByID_NotFound() {
	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", uuid.New().String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *TeamRepositoryTestSuite) TestGetByID_Success() {
	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), "Team Alpha", got.Name)
}

func (s *TeamRepositoryTestSuite) TestUpdate_NotFound() {
	team := &models.Team{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		Name:           "Updated Team",
		Description:    "Updated description",
	}
	err := s.repo.Update(nil, team)
	require.Error(s.T(), err)
	var ycaErr *yca_error.Error
	require.True(s.T(), errors.As(err, &ycaErr), "error should be a yca_error.Error")
	require.Equal(s.T(), http.StatusNotFound, ycaErr.StatusCode)
}

func (s *TeamRepositoryTestSuite) TestUpdate_Success() {
	team := &models.Team{
		ID:             uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		CreatedAt:      time.Now(),
		OrganizationID: uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		Name:           "Updated Team Alpha",
		Description:    "Updated alpha description",
	}
	err := s.repo.Update(nil, team)
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID(team.OrganizationID.String(), team.ID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	require.Equal(s.T(), "Updated Team Alpha", got.Name)
}

func (s *TeamRepositoryTestSuite) TestDelete_NotFound() {
	err := s.repo.Delete(nil, "00000001-0000-0000-0000-000000000001", uuid.New().String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *TeamRepositoryTestSuite) TestDelete_Success() {
	err := s.repo.Delete(nil, "00000001-0000-0000-0000-000000000001", "00000002-0000-0000-0000-000000000002")
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID("00000001-0000-0000-0000-000000000001", "00000002-0000-0000-0000-000000000002")
	require.Error(s.T(), err)
	require.Nil(s.T(), got)
}

func (s *TeamRepositoryTestSuite) TestListByOrganizationID() {
	teams, err := s.repo.ListByOrganizationID("00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), teams)
	require.Len(s.T(), *teams, 2)
}
