package user_repository_test

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
	user_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user"
	yca_error "github.com/yca-software/go-common/error"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	repo user_repository.Repository
	db   *sqlx.DB
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}

	testDB, err := database.GetTestDB()
	require.NoError(s.T(), err)
	s.db = testDB.DB()
	s.repo = user_repository.New(s.db, nil)
}

func (s *UserRepositoryTestSuite) SetupTest() {
	_, err := s.db.Exec(`
		INSERT INTO users (id, created_at, first_name, last_name, language, avatar_url, email, email_verified_at, password, google_id, terms_accepted_at, terms_version)
		VALUES
			('00000001-0000-0000-0000-000000000001', '2024-01-15 12:00:00+00', 'Test', 'User', 'en', '', 'test@example.com', NULL, 'hashedpassword', NULL, '2024-01-15 12:00:00+00', '1.0.0'),
			('00000002-0000-0000-0000-000000000002', '2024-01-15 12:00:00+00', 'Google', 'User', 'en', 'https://example.com/fixture.png', 'google@example.com', '2024-01-15 12:00:00+00', NULL, 'google123', '2024-01-15 12:00:00+00', '1.0.0'),
			('00000003-0000-0000-0000-000000000003', '2024-01-15 13:00:00+00', 'Alice', 'Alpha', 'en', '', 'alice@example.com', NULL, 'alicepass', NULL, '2024-01-15 13:00:00+00', '1.0.0'),
			('00000004-0000-0000-0000-000000000004', '2024-01-15 13:30:00+00', 'Bob', 'Bravo', 'fr', '', 'bob@example.com', NULL, 'bobpass', NULL, '2024-01-15 13:30:00+00', '1.0.0'),
			('00000005-0000-0000-0000-000000000005', '2024-01-15 14:00:00+00', 'Carol', 'Charlie', 'es', '', 'carol@example.com', NULL, NULL, 'carolgoogle', '2024-01-15 14:00:00+00', '1.0.0');
	`)
	require.NoError(s.T(), err)
}

func (s *UserRepositoryTestSuite) TearDownTest() {
	_, _ = s.db.Exec("TRUNCATE TABLE admin_access CASCADE")
	_, _ = s.db.Exec("TRUNCATE TABLE users CASCADE")
}

func (s *UserRepositoryTestSuite) TestCreate_DuplicateEmail() {
	password := "hashedpassword"
	user := &models.User{
		ID:              uuid.New(),
		CreatedAt:       time.Now(),
		FirstName:       "Test",
		LastName:        "User",
		Language:        "en",
		AvatarURL:       "",
		Email:           "test@example.com",
		Password:        &password,
		TermsAcceptedAt: time.Now(),
		TermsVersion:    "1.0.0",
	}
	err := s.repo.Create(nil, user)
	require.Error(s.T(), err)
	require.Equal(s.T(), 409, err.(*yca_error.Error).StatusCode)
}

func (s *UserRepositoryTestSuite) TestCreate_DuplicateGoogleID() {
	googleID := "google123"
	user := &models.User{
		ID:              uuid.New(),
		CreatedAt:       time.Now(),
		FirstName:       "Google",
		LastName:        "User",
		Language:        "en",
		AvatarURL:       "",
		Email:           "google@example.com",
		GoogleID:        &googleID,
		TermsAcceptedAt: time.Now(),
		TermsVersion:    "1.0.0",
	}
	err := s.repo.Create(nil, user)
	require.Error(s.T(), err)
	require.Equal(s.T(), 409, err.(*yca_error.Error).StatusCode)
}

func (s *UserRepositoryTestSuite) TestCreate_Success() {
	password := "hashedpassword"
	user := &models.User{
		ID:              uuid.New(),
		CreatedAt:       time.Now(),
		FirstName:       "Test",
		LastName:        "User",
		Language:        "en",
		AvatarURL:       "ccc",
		Email:           "test12341234@example.com",
		Password:        &password,
		TermsAcceptedAt: time.Now(),
		TermsVersion:    "1.0.0",
	}
	err := s.repo.Create(nil, user)
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID(nil, user.ID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
}

func (s *UserRepositoryTestSuite) TestGetByEmail_NotFound() {
	got, err := s.repo.GetByEmail(nil, "nonexistent@example.com")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *UserRepositoryTestSuite) TestGetByEmail_Success() {
	got, err := s.repo.GetByEmail(nil, "test@example.com")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
}

func (s *UserRepositoryTestSuite) TestGetByID_NotFound() {
	got, err := s.repo.GetByID(nil, uuid.New().String())
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *UserRepositoryTestSuite) TestGetByID_Success() {
	got, err := s.repo.GetByID(nil, "00000001-0000-0000-0000-000000000001")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
}

func (s *UserRepositoryTestSuite) TestGetByGoogleID_NotFound() {
	got, err := s.repo.GetByGoogleID(nil, "goo")
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
	require.Nil(s.T(), got)
}

func (s *UserRepositoryTestSuite) TestGetByGoogleID_Success() {
	got, err := s.repo.GetByGoogleID(nil, "google123")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
}

func (s *UserRepositoryTestSuite) TestUpdate_NotFound() {
	password := "hashedpassword"
	user := &models.User{
		ID:              uuid.New(),
		CreatedAt:       time.Now(),
		FirstName:       "Test",
		LastName:        "User",
		Language:        "en",
		AvatarURL:       "",
		Email:           "test@example.com",
		Password:        &password,
		TermsAcceptedAt: time.Now(),
		TermsVersion:    "1.0.0",
	}
	err := s.repo.Update(nil, user)
	require.Error(s.T(), err)
	var ycaErr *yca_error.Error
	require.True(s.T(), errors.As(err, &ycaErr), "error should be a yca_error.Error")
	require.Equal(s.T(), http.StatusNotFound, ycaErr.StatusCode)
}

func (s *UserRepositoryTestSuite) TestUpdate_Success() {
	password := "hashedpassword"
	user := &models.User{
		ID:              uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		CreatedAt:       time.Now(),
		FirstName:       "Test",
		LastName:        "User",
		Language:        "en",
		AvatarURL:       "",
		Email:           "test@example.com",
		Password:        &password,
		TermsAcceptedAt: time.Now(),
		TermsVersion:    "1.0.0",
	}
	err := s.repo.Update(nil, user)
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID(nil, user.ID.String())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
}

func (s *UserRepositoryTestSuite) TestDelete_NotFound() {
	user := &models.User{
		ID:              uuid.New(),
		CreatedAt:       time.Now(),
		FirstName:       "Test",
		LastName:        "User",
		Language:        "en",
		AvatarURL:       "",
		Email:           "testcjwec@example.com",
		TermsAcceptedAt: time.Now(),
		TermsVersion:    "1.0.0",
	}
	err := s.repo.Delete(nil, user)
	require.Error(s.T(), err)
	require.Equal(s.T(), 404, err.(*yca_error.Error).StatusCode)
}

func (s *UserRepositoryTestSuite) TestDelete_Success() {
	user := &models.User{
		ID:              uuid.MustParse("00000001-0000-0000-0000-000000000001"),
		CreatedAt:       time.Now(),
		FirstName:       "Test",
		LastName:        "User",
		Language:        "en",
		AvatarURL:       "",
		Email:           "test@example.com",
		TermsAcceptedAt: time.Now(),
		TermsVersion:    "1.0.0",
	}
	err := s.repo.Delete(nil, user)
	require.NoError(s.T(), err)

	got, err := s.repo.GetByID(nil, user.ID.String())
	require.Error(s.T(), err)
	require.Nil(s.T(), got)
}

func (s *UserRepositoryTestSuite) TestSearch_Empty() {
	users, err := s.repo.Search("", 10, 0)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), users)
	require.Len(s.T(), *users, 5)
}

func (s *UserRepositoryTestSuite) TestSearch_WithPhrase() {
	users, err := s.repo.Search("Test", 10, 0)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), users)
	require.Len(s.T(), *users, 1)
	require.Equal(s.T(), "Test", (*users)[0].FirstName)
}

func (s *UserRepositoryTestSuite) TestCount() {
	count, err := s.repo.Count()
	require.NoError(s.T(), err)
	require.Equal(s.T(), 5, count)
}
