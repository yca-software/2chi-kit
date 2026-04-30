package user_refresh_token_service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	user_refresh_token_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user_refresh_token"
	user_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user"
	user_refresh_token_service "github.com/yca-software/2chi-kit/go-api/internals/services/user_refresh_token"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type UserRefreshTokenServiceTestSuite struct {
	suite.Suite
	svc       user_refresh_token_service.Service
	repos     *repositories.Repositories
	userRepo  *user_repository.MockRepository
	tokenRepo *user_refresh_token_repository.MockRepository
	logger    *yca_log.MockLogger
	authorizer *helpers.Authorizer
	now       time.Time
}

func TestUserRefreshTokenServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserRefreshTokenServiceTestSuite))
}

func (s *UserRefreshTokenServiceTestSuite) SetupTest() {
	s.now = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	s.userRepo = user_repository.NewMock()
	s.tokenRepo = user_refresh_token_repository.NewMockRepository()
	s.logger = &yca_log.MockLogger{}

	s.repos = &repositories.Repositories{
		User:             s.userRepo,
		UserRefreshToken: s.tokenRepo,
	}

	s.authorizer = helpers.NewAuthorizer(func() time.Time { return s.now })

	s.svc = user_refresh_token_service.NewService(&user_refresh_token_service.Dependencies{
		Now:          func() time.Time { return s.now },
		Validator:    yca_validate.New(),
		Repositories: s.repos,
		Authorizer:   s.authorizer,
		Logger:       s.logger,
		HashToken:    func(t string) string { return "hashed:" + t },
	})
}

// --- Revoke: validations ---

func (s *UserRefreshTokenServiceTestSuite) TestRevoke_Validation_InvalidUserID() {
	req := &user_refresh_token_service.RevokeRequest{
		UserID:         "not-a-uuid",
		RefreshTokenID: uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New()},
	}
	err := s.svc.Revoke(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal("UNPROCESSABLE_ENTITY", e.ErrorCode)
	}
}

// --- Revoke: business logic ---

func (s *UserRefreshTokenServiceTestSuite) TestRevoke_UserNotFound() {
	userID := uuid.New()
	req := &user_refresh_token_service.RevokeRequest{
		UserID:         userID.String(),
		RefreshTokenID: uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New()},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(nil, yca_error.NewNotFoundError(nil, "NOT_FOUND", nil))

	err := s.svc.Revoke(req, accessInfo)
	s.Error(err)
	s.userRepo.AssertExpectations(s.T())
}

func (s *UserRefreshTokenServiceTestSuite) TestRevoke_PermissionDenied() {
	userID := uuid.New()
	otherUserID := uuid.New()
	user := &models.User{ID: userID, Email: "u@example.com"}
	req := &user_refresh_token_service.RevokeRequest{
		UserID:         userID.String(),
		RefreshTokenID: uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: otherUserID},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)

	err := s.svc.Revoke(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal("FORBIDDEN", e.ErrorCode)
	}
	s.userRepo.AssertExpectations(s.T())
}

func (s *UserRefreshTokenServiceTestSuite) TestRevoke_Success() {
	userID := uuid.New()
	tokenID := uuid.New().String()
	user := &models.User{ID: userID, Email: "u@example.com"}
	req := &user_refresh_token_service.RevokeRequest{
		UserID:         userID.String(),
		RefreshTokenID: tokenID,
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: userID},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)
	s.tokenRepo.On("Revoke", nil, userID.String(), tokenID).Return(nil)

	err := s.svc.Revoke(req, accessInfo)
	s.NoError(err)
	s.userRepo.AssertExpectations(s.T())
	s.tokenRepo.AssertExpectations(s.T())
}

// --- RevokeAll: validations ---

func (s *UserRefreshTokenServiceTestSuite) TestRevokeAll_Validation_InvalidUserID() {
	req := &user_refresh_token_service.RevokeAllRequest{UserID: "not-a-uuid"}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New()},
	}
	err := s.svc.RevokeAll(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal("UNPROCESSABLE_ENTITY", e.ErrorCode)
	}
}

// --- RevokeAll: business logic ---

func (s *UserRefreshTokenServiceTestSuite) TestRevokeAll_PermissionDenied() {
	userID := uuid.New()
	otherUserID := uuid.New()
	user := &models.User{ID: userID, Email: "u@example.com"}
	req := &user_refresh_token_service.RevokeAllRequest{UserID: userID.String()}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: otherUserID},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)

	err := s.svc.RevokeAll(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal("FORBIDDEN", e.ErrorCode)
	}
	s.userRepo.AssertExpectations(s.T())
}

func (s *UserRefreshTokenServiceTestSuite) TestRevokeAll_Success() {
	userID := uuid.New()
	user := &models.User{ID: userID, Email: "u@example.com"}
	req := &user_refresh_token_service.RevokeAllRequest{UserID: userID.String()}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: userID},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)
	s.tokenRepo.On("RevokeAll", nil, userID.String()).Return(nil)

	err := s.svc.RevokeAll(req, accessInfo)
	s.NoError(err)
	s.userRepo.AssertExpectations(s.T())
	s.tokenRepo.AssertExpectations(s.T())
}

func (s *UserRefreshTokenServiceTestSuite) TestRevokeAll_WithKeepRefreshToken_Success() {
	userID := uuid.New()
	tokID := uuid.New()
	user := &models.User{ID: userID, Email: "u@example.com"}
	raw := "refresh-raw-value"
	req := &user_refresh_token_service.RevokeAllRequest{UserID: userID.String(), KeepRefreshToken: raw}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: userID},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)
	s.tokenRepo.On("GetByHash", nil, "hashed:"+raw).Return(&models.UserRefreshToken{
		ID:        tokID,
		UserID:    userID,
		ExpiresAt: s.now.Add(time.Hour),
	}, nil)
	s.tokenRepo.On("RevokeAllExcept", nil, userID.String(), tokID.String()).Return(nil)

	err := s.svc.RevokeAll(req, accessInfo)
	s.NoError(err)
	s.userRepo.AssertExpectations(s.T())
	s.tokenRepo.AssertExpectations(s.T())
}

// --- ListActive: validations ---

func (s *UserRefreshTokenServiceTestSuite) TestListActive_Validation_InvalidUserID() {
	req := &user_refresh_token_service.ListActiveRequest{UserID: "not-a-uuid"}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New()},
	}
	resp, err := s.svc.ListActive(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal("UNPROCESSABLE_ENTITY", e.ErrorCode)
	}
}

// --- ListActive: business logic ---

func (s *UserRefreshTokenServiceTestSuite) TestListActive_PermissionDenied() {
	userID := uuid.New()
	otherUserID := uuid.New()
	user := &models.User{ID: userID, Email: "u@example.com"}
	req := &user_refresh_token_service.ListActiveRequest{UserID: userID.String()}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: otherUserID},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)

	resp, err := s.svc.ListActive(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal("FORBIDDEN", e.ErrorCode)
	}
	s.userRepo.AssertExpectations(s.T())
}

func (s *UserRefreshTokenServiceTestSuite) TestListActive_Success() {
	userID := uuid.New()
	user := &models.User{ID: userID, Email: "u@example.com"}
	tokens := []models.UserRefreshToken{
		{ID: uuid.New(), UserID: userID, IP: "1.2.3.4", UserAgent: "test"},
	}
	req := &user_refresh_token_service.ListActiveRequest{UserID: userID.String()}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: userID},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)
	s.tokenRepo.On("GetActiveByUserID", nil, userID.String()).Return(&tokens, nil)

	resp, err := s.svc.ListActive(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(*resp, 1)
	s.Equal(userID, (*resp)[0].UserID)
	s.userRepo.AssertExpectations(s.T())
	s.tokenRepo.AssertExpectations(s.T())
}

func (s *UserRefreshTokenServiceTestSuite) TestCleanupStaleUnused() {
	s.tokenRepo.On("CleanupStaleUnused", mock.Anything).Return(nil).Once()
	err := s.svc.CleanupStaleUnused()
	s.NoError(err)
	s.tokenRepo.AssertExpectations(s.T())
}

func (s *UserRefreshTokenServiceTestSuite) TestCleanupStaleUnused_RepoError() {
	repoErr := errors.New("cleanup failed")
	s.tokenRepo.On("CleanupStaleUnused", mock.Anything).Return(repoErr).Once()
	err := s.svc.CleanupStaleUnused()
	s.ErrorIs(err, repoErr)
	s.tokenRepo.AssertExpectations(s.T())
}
