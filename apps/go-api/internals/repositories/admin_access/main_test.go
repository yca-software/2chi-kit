package admin_access_repository_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/database"
	admin_access_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/admin_access"
	yca_error "github.com/yca-software/go-common/error"
)

type AdminAccessRepositoryTestSuite struct {
	suite.Suite
	repo admin_access_repository.Repository
	db   *sqlx.DB
}

func TestAdminAccessRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(AdminAccessRepositoryTestSuite))
}

func (s *AdminAccessRepositoryTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}

	testDB, err := database.GetTestDB()
	require.NoError(s.T(), err)
	s.db = testDB.DB()
	s.repo = admin_access_repository.New(s.db, nil)
}

func (s *AdminAccessRepositoryTestSuite) SetupTest() {
	_, err := s.db.Exec(`
		INSERT INTO users (id, created_at, first_name, last_name, language, avatar_url, email, email_verified_at, password, google_id, terms_accepted_at, terms_version)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', 'Test', 'User', 'en', '', 'test@example.com', NULL, 'hashedpassword', NULL, '2024-01-15 12:00:00+00', '1.0.0'),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 12:00:00+00', 'Google', 'User', 'en', 'https://example.com/fixture.png', 'google@example.com', '2024-01-15 12:00:00+00', NULL, 'google123', '2024-01-15 12:00:00+00', '1.0.0');

		INSERT INTO admin_access (user_id, created_at)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00');
	`)
	require.NoError(s.T(), err)
}

func (s *AdminAccessRepositoryTestSuite) TearDownTest() {
	_, _ = s.db.Exec("TRUNCATE TABLE admin_access CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE users CASCADE")
}

func (s *AdminAccessRepositoryTestSuite) TestGetByUserID_UserAndAdminExists() {
	userID := uuid.MustParse("00000001-0000-0000-0000-000000000001")
	adminAccess, err := s.repo.GetByUserID(userID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), adminAccess)
}

func (s *AdminAccessRepositoryTestSuite) TestGetByUserID_UserExistsButNotAdmin() {
	userID := uuid.MustParse("00000002-0000-0000-0000-000000000002")
	adminAccess, err := s.repo.GetByUserID(userID.String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), adminAccess)
}

func (s *AdminAccessRepositoryTestSuite) TestGetByUserID_NoUser() {
	userID := uuid.MustParse("00000003-0000-0000-0000-000000000003")
	adminAccess, err := s.repo.GetByUserID(userID.String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), adminAccess)
}
