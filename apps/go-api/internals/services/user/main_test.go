package user_service_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	admin_access_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/admin_access"
	organization_member_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization_member"
	user_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user"
	user_service "github.com/yca-software/2chi-kit/go-api/internals/services/user"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
	yca_password "github.com/yca-software/go-common/password"
	yca_validate "github.com/yca-software/go-common/validator"
)

type UserServiceTestSuite struct {
	suite.Suite
	svc        user_service.Service
	repos      *repositories.Repositories
	userRepo   *user_repository.MockRepository
	adminRepo  *admin_access_repository.MockRepository
	orgMemRepo *organization_member_repository.MockRepository
	logger     *yca_log.MockLogger
	authorizer *helpers.Authorizer
	now        time.Time
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) SetupTest() {
	s.now = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	s.userRepo = user_repository.NewMock()
	s.adminRepo = admin_access_repository.NewMock()
	s.orgMemRepo = organization_member_repository.NewMock()
	s.logger = &yca_log.MockLogger{}

	s.repos = &repositories.Repositories{
		User:               s.userRepo,
		AdminAccess:        s.adminRepo,
		OrganizationMember: s.orgMemRepo,
	}

	s.authorizer = helpers.NewAuthorizer(func() time.Time { return s.now })

	s.svc = user_service.NewService(&user_service.Dependencies{
		GenerateID:     uuid.NewV7,
		Now:            func() time.Time { return s.now },
		Validator:      yca_validate.New(),
		Repositories:   s.repos,
		Authorizer:     s.authorizer,
		Logger:         s.logger,
		PasswordHashFn: func(string) (string, error) { return "hashed", nil },
		GenerateToken:  func() (string, error) { return "token", nil },
		HashToken:      func(string) string { return "hash" },
	})
}

// --- Get: validations ---

func (s *UserServiceTestSuite) TestGet_Validation_InvalidUserID() {
	req := &user_service.GetRequest{UserID: "not-a-uuid"}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New()},
	}
	resp, err := s.svc.Get(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Get: business logic ---

func (s *UserServiceTestSuite) TestGet_UserNotFound() {
	userID := uuid.New()
	req := &user_service.GetRequest{UserID: userID.String()}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New()},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(nil, yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.Get(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.userRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestGet_PermissionDenied() {
	userID := uuid.New()
	otherUserID := uuid.New()
	user := &models.User{ID: userID, Email: "u@example.com"}
	req := &user_service.GetRequest{UserID: userID.String()}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: otherUserID},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)

	resp, err := s.svc.Get(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
	s.userRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestGet_Success() {
	userID := uuid.New()
	user := &models.User{ID: userID, Email: "u@example.com", FirstName: "John", LastName: "Doe"}
	roles := []models.OrganizationMemberWithOrganizationAndRole{}
	req := &user_service.GetRequest{UserID: userID.String()}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: userID},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)
	s.adminRepo.On("GetByUserID", userID.String()).Return((*models.AdminAccess)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	s.orgMemRepo.On("ListByUserIDWithRole", userID.String()).Return(&roles, nil)

	resp, err := s.svc.Get(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal(user, resp.User)
	s.Nil(resp.AdminAccess)
	s.NotNil(resp.Roles)
	s.Len(*resp.Roles, 0)
	s.userRepo.AssertExpectations(s.T())
	s.adminRepo.AssertExpectations(s.T())
	s.orgMemRepo.AssertExpectations(s.T())
}

// --- List: validations ---

func (s *UserServiceTestSuite) TestList_Validation_InvalidLimit() {
	req := &user_service.ListRequest{Limit: 0, Offset: 0}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New()},
	}
	resp, err := s.svc.List(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- List: business logic ---

func (s *UserServiceTestSuite) TestList_NotAdmin() {
	req := &user_service.ListRequest{Limit: 10, Offset: 0}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New()},
	}

	resp, err := s.svc.List(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
}

func (s *UserServiceTestSuite) TestList_Success() {
	userID := uuid.New()
	users := []models.User{
		{ID: userID, Email: "admin@example.com", FirstName: "Admin", LastName: "User"},
	}
	req := &user_service.ListRequest{SearchPhrase: "", Limit: 10, Offset: 0}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: userID, IsAdmin: true},
	}
	s.userRepo.On("Search", "", 11, 0).Return(&users, nil)

	resp, err := s.svc.List(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(resp.Items, 1)
	s.Equal("admin@example.com", resp.Items[0].Email)
	s.False(resp.HasNext)
	s.userRepo.AssertExpectations(s.T())
}

// --- ChangePassword: validations ---

func (s *UserServiceTestSuite) TestChangePassword_Validation_InvalidUserID() {
	req := &user_service.ChangePasswordRequest{
		UserID:          "not-a-uuid",
		CurrentPassword: "old",
		NewPassword:     "newpassword",
	}
	err := s.svc.ChangePassword(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *UserServiceTestSuite) TestChangePassword_Validation_NewPasswordTooShort() {
	req := &user_service.ChangePasswordRequest{
		UserID:          uuid.New().String(),
		CurrentPassword: "oldpass",
		NewPassword:     "short",
	}
	err := s.svc.ChangePassword(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- ChangePassword: business logic ---

func (s *UserServiceTestSuite) TestChangePassword_UserNotFound() {
	userID := uuid.New()
	req := &user_service.ChangePasswordRequest{
		UserID:          userID.String(),
		CurrentPassword: "oldpass",
		NewPassword:     "newpassword",
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(nil, yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	err := s.svc.ChangePassword(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNAUTHORIZED_CODE, e.ErrorCode)
	}
	s.userRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestChangePassword_CurrentPasswordMismatch() {
	userID := uuid.New()
	hashed := "existing-hash"
	user := &models.User{ID: userID, Email: "u@example.com", Password: &hashed}
	req := &user_service.ChangePasswordRequest{
		UserID:          userID.String(),
		CurrentPassword: "wrong",
		NewPassword:     "newpassword",
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)

	err := s.svc.ChangePassword(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.CURRENT_PASSWORD_MISMATCH_CODE, e.ErrorCode)
	}
	s.userRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestChangePassword_Success() {
	userID := uuid.New()
	currentHash, err := yca_password.Hash("correct")
	s.Require().NoError(err)
	user := &models.User{ID: userID, Email: "u@example.com", Password: &currentHash}
	req := &user_service.ChangePasswordRequest{
		UserID:          userID.String(),
		CurrentPassword: "correct",
		NewPassword:     "newpassword",
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)
	s.userRepo.On("Update", nil, mock.MatchedBy(func(u *models.User) bool {
		return u != nil && u.Password != nil && *u.Password == "hashed"
	})).Return(nil)

	err = s.svc.ChangePassword(req)
	s.NoError(err)
	s.userRepo.AssertExpectations(s.T())
}

// --- AcceptTerms: validations ---

func (s *UserServiceTestSuite) TestAcceptTerms_Validation_InvalidUserID() {
	req := &user_service.AcceptTermsRequest{UserID: "not-a-uuid", TermsVersion: "1.0.0"}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New()},
	}
	resp, err := s.svc.AcceptTerms(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *UserServiceTestSuite) TestAcceptTerms_Validation_EmptyTermsVersion() {
	userID := uuid.New()
	req := &user_service.AcceptTermsRequest{UserID: userID.String(), TermsVersion: ""}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: userID},
	}
	resp, err := s.svc.AcceptTerms(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- AcceptTerms: business logic ---

func (s *UserServiceTestSuite) TestAcceptTerms_UserNotFound() {
	userID := uuid.New()
	req := &user_service.AcceptTermsRequest{UserID: userID.String(), TermsVersion: "1.0.0"}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: userID},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(nil, yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.AcceptTerms(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.userRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestAcceptTerms_PermissionDenied() {
	userID := uuid.New()
	otherUserID := uuid.New()
	user := &models.User{ID: userID, Email: "u@example.com"}
	req := &user_service.AcceptTermsRequest{UserID: userID.String(), TermsVersion: "1.0.0"}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: otherUserID},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)

	resp, err := s.svc.AcceptTerms(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.userRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestAcceptTerms_Success() {
	userID := uuid.New()
	user := &models.User{ID: userID, Email: "u@example.com", TermsVersion: "0.9.0"}
	req := &user_service.AcceptTermsRequest{UserID: userID.String(), TermsVersion: "1.0.0"}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: userID},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)
	s.userRepo.On("Update", nil, mock.MatchedBy(func(u *models.User) bool {
		return u != nil && u.TermsVersion == "1.0.0" && u.TermsAcceptedAt.Equal(s.now)
	})).Return(nil)

	resp, err := s.svc.AcceptTerms(req, accessInfo)
	s.NoError(err)
	s.NotNil(resp)
	s.Equal("1.0.0", resp.TermsVersion)
	s.True(resp.TermsAcceptedAt.Equal(s.now))
	s.userRepo.AssertExpectations(s.T())
}
