package auth_service_test

import (
	"errors"
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
	invitation_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/invitation"
	organization_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization"
	organization_member_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization_member"
	user_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user"
	user_email_verification_token_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user_email_verification_token"
	user_password_reset_token_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user_password_reset_token"
	user_refresh_token_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user_refresh_token"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	auth_service "github.com/yca-software/2chi-kit/go-api/internals/services/auth"
	google_service "github.com/yca-software/2chi-kit/go-api/internals/services/google"
	yca_email "github.com/yca-software/go-common/email"
	yca_error "github.com/yca-software/go-common/error"
	yca_translate "github.com/yca-software/go-common/localizer"
	yca_log "github.com/yca-software/go-common/logger"
	yca_password "github.com/yca-software/go-common/password"
	yca_repository "github.com/yca-software/go-common/repository"
	yca_validate "github.com/yca-software/go-common/validator"
)

type AuthServiceTestSuite struct {
	suite.Suite
	svc                        auth_service.Service
	repos                      *repositories.Repositories
	userRepo                   *user_repository.MockRepository
	adminAccessRepo            *admin_access_repository.MockRepository
	orgMemberRepo              *organization_member_repository.MockRepository
	refreshTokenRepo           *user_refresh_token_repository.MockRepository
	emailVerificationTokenRepo *user_email_verification_token_repository.MockRepository
	passwordResetTokenRepo     *user_password_reset_token_repository.MockRepository
	invitationRepo             *invitation_repository.MockRepository
	orgRepo                    *organization_repository.MockRepository
	googleSvc                  *google_service.MockService
	emailSvc                   *yca_email.MockEmailService
	auditLogSvc                *audit_log_service.MockService
	logger                     *yca_log.MockLogger
	translator                 *yca_translate.MockTranslator
	now                        time.Time
}

func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

func (s *AuthServiceTestSuite) SetupTest() {
	s.now = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	s.userRepo = user_repository.NewMock()
	s.adminAccessRepo = admin_access_repository.NewMock()
	s.orgMemberRepo = organization_member_repository.NewMock()
	s.refreshTokenRepo = user_refresh_token_repository.NewMockRepository()
	s.emailVerificationTokenRepo = user_email_verification_token_repository.NewMock()
	s.passwordResetTokenRepo = user_password_reset_token_repository.NewMock()
	s.invitationRepo = invitation_repository.NewMock()
	s.orgRepo = organization_repository.NewMock()
	s.googleSvc = google_service.NewMockService()
	s.emailSvc = yca_email.NewMockEmailService()
	s.auditLogSvc = audit_log_service.NewMockService()
	s.logger = &yca_log.MockLogger{}
	s.translator = yca_translate.NewMockTranslator()
	s.translator.On("TranslateError", mock.Anything, mock.Anything, mock.Anything).Return("").Maybe()

	s.repos = &repositories.Repositories{
		User:                       s.userRepo,
		AdminAccess:                s.adminAccessRepo,
		OrganizationMember:         s.orgMemberRepo,
		UserRefreshToken:           s.refreshTokenRepo,
		UserEmailVerificationToken: s.emailVerificationTokenRepo,
		UserPasswordResetToken:     s.passwordResetTokenRepo,
		Invitation:                 s.invitationRepo,
		Organization:               s.orgRepo,
	}

	s.svc = auth_service.New(&auth_service.Dependencies{
		AccessSecret:         "test-secret-key",
		AccessTTL:            "60",
		RefreshTTL:           "24",
		EmailVerificationTTL: "24",
		AppURL:               "https://example.com",
		Validator:            yca_validate.New(),
		Repos:                s.repos,
		Authorizer:           helpers.NewAuthorizer(func() time.Time { return s.now }),
		GenerateID:           uuid.NewV7,
		GenerateToken:        func() (string, error) { return "test-token-123", nil },
		HashToken:            func(token string) string { return "hashed:" + token },
		Now:                  func() time.Time { return s.now },
		Logger:               s.logger,
		AuditLogService:      s.auditLogSvc,
		GoogleService:        s.googleSvc,
		EmailService:         s.emailSvc,
		Translator:           s.translator,
		PasswordHashFn:       func(password string) (string, error) { return "hashed:" + password, nil },
		PasswordCompareFn:    func(password, encodedHash string) bool { return encodedHash == "hashed:"+password },
	})
}

// --- SignUp: validations ---

func (s *AuthServiceTestSuite) TestSignUp_Validation_MissingEmail() {
	req := &auth_service.SignUpRequest{
		Email:        "",
		Password:     "password123",
		FirstName:    "John",
		LastName:     "Doe",
		Language:     "en",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		TermsVersion: "1.0.0",
	}
	resp, err := s.svc.SignUp(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestSignUp_Validation_InvalidEmail() {
	req := &auth_service.SignUpRequest{
		Email:        "not-an-email",
		Password:     "password123",
		FirstName:    "John",
		LastName:     "Doe",
		Language:     "en",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		TermsVersion: "1.0.0",
	}
	resp, err := s.svc.SignUp(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestSignUp_Validation_MissingPassword() {
	req := &auth_service.SignUpRequest{
		Email:        "test@example.com",
		Password:     "",
		FirstName:    "John",
		LastName:     "Doe",
		Language:     "en",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		TermsVersion: "1.0.0",
	}
	resp, err := s.svc.SignUp(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestSignUp_Validation_MissingFirstName() {
	req := &auth_service.SignUpRequest{
		Email:        "test@example.com",
		Password:     "password123",
		FirstName:    "",
		LastName:     "Doe",
		Language:     "en",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		TermsVersion: "1.0.0",
	}
	resp, err := s.svc.SignUp(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestSignUp_Validation_MissingLastName() {
	req := &auth_service.SignUpRequest{
		Email:        "test@example.com",
		Password:     "password123",
		FirstName:    "John",
		LastName:     "",
		Language:     "en",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		TermsVersion: "1.0.0",
	}
	resp, err := s.svc.SignUp(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestSignUp_Validation_InvalidIPAddress() {
	req := &auth_service.SignUpRequest{
		Email:        "test@example.com",
		Password:     "password123",
		FirstName:    "John",
		LastName:     "Doe",
		Language:     "en",
		IPAddress:    "not-an-ip",
		UserAgent:    "test-agent",
		TermsVersion: "1.0.0",
	}
	resp, err := s.svc.SignUp(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestSignUp_Validation_InvalidTermsVersion() {
	req := &auth_service.SignUpRequest{
		Email:        "test@example.com",
		Password:     "password123",
		FirstName:    "John",
		LastName:     "Doe",
		Language:     "en",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		TermsVersion: "invalid",
	}
	resp, err := s.svc.SignUp(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- SignUp: business logic ---

func (s *AuthServiceTestSuite) TestSignUp_EmailAlreadyInUse() {
	email := "test@example.com"
	existingUser := &models.User{ID: uuid.New(), Email: email}
	s.userRepo.On("GetByEmail", nil, email).Return(existingUser, nil)

	req := &auth_service.SignUpRequest{
		Email:        email,
		Password:     "password123",
		FirstName:    "John",
		LastName:     "Doe",
		Language:     "en",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		TermsVersion: "1.0.0",
	}
	resp, err := s.svc.SignUp(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.EMAIL_ALREADY_IN_USE_CODE, e.ErrorCode)
	}
	s.userRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestSignUp_Success_WithoutInvitation() {
	email := "test@example.com"
	s.userRepo.On("GetByEmail", nil, email).Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	tx := yca_repository.NewMockTx()
	s.userRepo.On("BeginTx").Return(tx, nil)
	s.userRepo.On("Create", tx, mock.AnythingOfType("*models.User")).Return(nil)
	s.refreshTokenRepo.On("Create", tx, mock.AnythingOfType("*models.UserRefreshToken")).Return(nil)
	s.emailVerificationTokenRepo.On("Create", tx, mock.AnythingOfType("*models.UserEmailVerificationToken")).Return(nil)
	s.adminAccessRepo.On("GetByUserID", mock.AnythingOfType("string")).Return((*models.AdminAccess)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	s.orgMemberRepo.On("ListByUserIDWithRole", mock.AnythingOfType("string")).Return(&[]models.OrganizationMemberWithOrganizationAndRole{}, nil)
	s.translator.On("Translate", "en", mock.AnythingOfType("string"), mock.Anything).Return("translated")
	s.emailSvc.On("PrepareEmailBody", "verification", mock.Anything).Return("email-body", nil)
	s.emailSvc.On("SendEmail", email, mock.AnythingOfType("string"), "email-body").Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()

	req := &auth_service.SignUpRequest{
		Email:        email,
		Password:     "password123",
		FirstName:    "John",
		LastName:     "Doe",
		Language:     "en",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		TermsVersion: "1.0.0",
	}
	resp, err := s.svc.SignUp(req)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.NotEmpty(resp.AccessToken)
	s.NotEmpty(resp.RefreshToken)
	s.userRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestSignUp_InvitationEmailMismatch() {
	email := "user@example.com"
	invitedEmail := "invited@example.com"

	s.userRepo.On("GetByEmail", nil, email).Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	tx := yca_repository.NewMockTx()
	s.userRepo.On("BeginTx").Return(tx, nil)
	s.userRepo.On("Create", tx, mock.AnythingOfType("*models.User")).Return(nil)

	invitation := &models.Invitation{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		RoleID:         uuid.New(),
		Email:          invitedEmail,
		ExpiresAt:      s.now.Add(24 * time.Hour),
	}
	tokenHash := "hashed:inv-token"
	s.invitationRepo.On("GetByTokenHash", tokenHash).Return(invitation, nil)
	tx.On("Rollback").Return(nil).Maybe()

	req := &auth_service.SignUpRequest{
		Email:           email,
		Password:        "password123",
		FirstName:       "John",
		LastName:        "Doe",
		Language:        "en",
		IPAddress:       "192.168.1.1",
		UserAgent:       "test-agent",
		TermsVersion:    "1.0.0",
		InvitationToken: "inv-token",
	}
	resp, err := s.svc.SignUp(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVITATION_EMAIL_MISMATCH_CODE, e.ErrorCode)
	}
	s.invitationRepo.AssertExpectations(s.T())
}

// --- AuthenticateWithPassword: validations ---

func (s *AuthServiceTestSuite) TestAuthenticateWithPassword_Validation_MissingEmail() {
	req := &auth_service.AuthenticateWithPasswordRequest{
		Email:     "",
		Password:  "password123",
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}
	resp, err := s.svc.AuthenticateWithPassword(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestAuthenticateWithPassword_Validation_InvalidEmail() {
	req := &auth_service.AuthenticateWithPasswordRequest{
		Email:     "not-an-email",
		Password:  "password123",
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}
	resp, err := s.svc.AuthenticateWithPassword(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestAuthenticateWithPassword_Validation_MissingPassword() {
	req := &auth_service.AuthenticateWithPasswordRequest{
		Email:     "test@example.com",
		Password:  "",
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}
	resp, err := s.svc.AuthenticateWithPassword(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestAuthenticateWithPassword_Validation_InvalidIPAddress() {
	req := &auth_service.AuthenticateWithPasswordRequest{
		Email:     "test@example.com",
		Password:  "password123",
		IPAddress: "not-an-ip",
		UserAgent: "test-agent",
	}
	resp, err := s.svc.AuthenticateWithPassword(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- AuthenticateWithPassword: business logic ---

func (s *AuthServiceTestSuite) TestAuthenticateWithPassword_UserNotFound() {
	email := "test@example.com"
	s.userRepo.On("GetByEmail", nil, email).Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &auth_service.AuthenticateWithPasswordRequest{
		Email:     email,
		Password:  "password123",
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}
	resp, err := s.svc.AuthenticateWithPassword(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.PASSWORD_MISMATCH_CODE, e.ErrorCode)
	}
	s.userRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestAuthenticateWithPassword_NoPassword() {
	email := "test@example.com"
	user := &models.User{
		ID:       uuid.New(),
		Email:    email,
		Password: nil,
	}
	s.userRepo.On("GetByEmail", nil, email).Return(user, nil)

	req := &auth_service.AuthenticateWithPasswordRequest{
		Email:     email,
		Password:  "password123",
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}
	resp, err := s.svc.AuthenticateWithPassword(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.PASSWORD_MISMATCH_CODE, e.ErrorCode)
	}
	s.userRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestAuthenticateWithPassword_PasswordMismatch() {
	email := "test@example.com"
	password := "hashed:wrong-password"
	user := &models.User{
		ID:       uuid.New(),
		Email:    email,
		Password: &password,
	}
	s.userRepo.On("GetByEmail", nil, email).Return(user, nil)

	req := &auth_service.AuthenticateWithPasswordRequest{
		Email:     email,
		Password:  "password123",
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}
	resp, err := s.svc.AuthenticateWithPassword(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.PASSWORD_MISMATCH_CODE, e.ErrorCode)
	}
	s.userRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestAuthenticateWithPassword_Success() {
	email := "test@example.com"
	hashedPassword, err := yca_password.Hash("password123")
	s.Require().NoError(err)
	user := &models.User{
		ID:       uuid.New(),
		Email:    email,
		Password: &hashedPassword,
	}
	s.userRepo.On("GetByEmail", nil, email).Return(user, nil)
	s.adminAccessRepo.On("GetByUserID", user.ID.String()).Return((*models.AdminAccess)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	s.orgMemberRepo.On("ListByUserIDWithRole", user.ID.String()).Return(&[]models.OrganizationMemberWithOrganizationAndRole{}, nil)
	s.refreshTokenRepo.On("Create", nil, mock.AnythingOfType("*models.UserRefreshToken")).Return(nil)

	req := &auth_service.AuthenticateWithPasswordRequest{
		Email:     email,
		Password:  "password123",
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}
	resp, err := s.svc.AuthenticateWithPassword(req)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.NotEmpty(resp.AccessToken)
	s.NotEmpty(resp.RefreshToken)
	s.userRepo.AssertExpectations(s.T())
	s.refreshTokenRepo.AssertExpectations(s.T())
}

// --- AuthenticateWithGoogle: validations ---

func (s *AuthServiceTestSuite) TestAuthenticateWithGoogle_Validation_MissingCode() {
	req := &auth_service.AuthenticateWithGoogleRequest{
		Code:         "",
		TermsVersion: "1.0.0",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		Language:     "en",
	}
	resp, err := s.svc.AuthenticateWithGoogle(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestAuthenticateWithGoogle_Validation_InvalidTermsVersion() {
	req := &auth_service.AuthenticateWithGoogleRequest{
		Code:         "google-code",
		TermsVersion: "invalid",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		Language:     "en",
	}
	resp, err := s.svc.AuthenticateWithGoogle(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestAuthenticateWithGoogle_Validation_InvalidIPAddress() {
	req := &auth_service.AuthenticateWithGoogleRequest{
		Code:         "google-code",
		TermsVersion: "1.0.0",
		IPAddress:    "not-an-ip",
		UserAgent:    "test-agent",
		Language:     "en",
	}
	resp, err := s.svc.AuthenticateWithGoogle(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestAuthenticateWithGoogle_Validation_MissingUserAgent() {
	req := &auth_service.AuthenticateWithGoogleRequest{
		Code:         "google-code",
		TermsVersion: "1.0.0",
		IPAddress:    "192.168.1.1",
		UserAgent:    "",
		Language:     "en",
	}
	resp, err := s.svc.AuthenticateWithGoogle(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestAuthenticateWithGoogle_Validation_InvalidLanguage() {
	req := &auth_service.AuthenticateWithGoogleRequest{
		Code:         "google-code",
		TermsVersion: "1.0.0",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		Language:     "x", // len must be 2
	}
	resp, err := s.svc.AuthenticateWithGoogle(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- AuthenticateWithGoogle: business logic ---

func (s *AuthServiceTestSuite) TestAuthenticateWithGoogle_GetUserInfoFails() {
	var nilGoogleUser *models.GoogleUserInfo
	s.googleSvc.On("GetUserInfo", mock.Anything, "bad-code").Return(nilGoogleUser, yca_error.NewUnauthorizedError(nil, constants.INVALID_TOKEN_CODE, nil))

	req := &auth_service.AuthenticateWithGoogleRequest{
		Code:         "bad-code",
		TermsVersion: "1.0.0",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		Language:     "en",
	}
	resp, err := s.svc.AuthenticateWithGoogle(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVALID_TOKEN_CODE, e.ErrorCode)
	}
	s.googleSvc.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestAuthenticateWithGoogle_EmailNotVerified() {
	googleUser := &models.GoogleUserInfo{
		ID:            "google-123",
		Email:         "test@example.com",
		VerifiedEmail: false,
		Name:          "Test User",
	}
	s.googleSvc.On("GetUserInfo", mock.Anything, "code").Return(googleUser, nil)

	req := &auth_service.AuthenticateWithGoogleRequest{
		Code:         "code",
		TermsVersion: "1.0.0",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		Language:     "en",
	}
	resp, err := s.svc.AuthenticateWithGoogle(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVALID_TOKEN_CODE, e.ErrorCode)
	}
	s.googleSvc.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestAuthenticateWithGoogle_ExistingUserByGoogleID_Success() {
	googleUser := &models.GoogleUserInfo{
		ID:            "google-123",
		Email:         "test@example.com",
		VerifiedEmail: true,
		Name:          "Test User",
	}
	userID := uuid.New()
	googleID := "google-123"
	user := &models.User{
		ID:       userID,
		Email:    "test@example.com",
		GoogleID: &googleID,
	}
	s.googleSvc.On("GetUserInfo", mock.Anything, "code").Return(googleUser, nil)
	s.userRepo.On("GetByGoogleID", nil, "google-123").Return(user, nil)
	tx := yca_repository.NewMockTx()
	s.userRepo.On("BeginTx").Return(tx, nil)
	s.adminAccessRepo.On("GetByUserID", userID.String()).Return((*models.AdminAccess)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	s.orgMemberRepo.On("ListByUserIDWithRole", userID.String()).Return(&[]models.OrganizationMemberWithOrganizationAndRole{}, nil)
	s.refreshTokenRepo.On("Create", tx, mock.AnythingOfType("*models.UserRefreshToken")).Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()

	req := &auth_service.AuthenticateWithGoogleRequest{
		Code:         "code",
		TermsVersion: "1.0.0",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		Language:     "en",
	}
	resp, err := s.svc.AuthenticateWithGoogle(req)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.NotEmpty(resp.AccessToken)
	s.NotEmpty(resp.RefreshToken)
	s.userRepo.AssertExpectations(s.T())
	s.refreshTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestAuthenticateWithGoogle_NewUserWithoutInvitation_Success() {
	googleUser := &models.GoogleUserInfo{
		ID:            "google-456",
		Email:         "new@example.com",
		VerifiedEmail: true,
		Name:          "New User",
		GivenName:     "New",
		FamilyName:    "User",
	}
	s.googleSvc.On("GetUserInfo", mock.Anything, "code").Return(googleUser, nil)
	s.userRepo.On("GetByGoogleID", nil, "google-456").Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	tx := yca_repository.NewMockTx()
	s.userRepo.On("BeginTx").Return(tx, nil)
	s.userRepo.On("GetByEmail", nil, "new@example.com").Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	s.userRepo.On("Create", tx, mock.AnythingOfType("*models.User")).Return(nil)
	s.adminAccessRepo.On("GetByUserID", mock.AnythingOfType("string")).Return((*models.AdminAccess)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	s.orgMemberRepo.On("ListByUserIDWithRole", mock.AnythingOfType("string")).Return(&[]models.OrganizationMemberWithOrganizationAndRole{}, nil)
	s.refreshTokenRepo.On("Create", tx, mock.AnythingOfType("*models.UserRefreshToken")).Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()

	req := &auth_service.AuthenticateWithGoogleRequest{
		Code:         "code",
		TermsVersion: "1.0.0",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		Language:     "en",
	}
	resp, err := s.svc.AuthenticateWithGoogle(req)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.NotEmpty(resp.AccessToken)
	s.NotEmpty(resp.RefreshToken)
	s.userRepo.AssertExpectations(s.T())
	s.refreshTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestAuthenticateWithGoogle_NewUserWithInvitationEmailMismatch() {
	googleUser := &models.GoogleUserInfo{
		ID:            "google-mismatch",
		Email:         "user@example.com",
		VerifiedEmail: true,
		Name:          "Invite User",
	}
	s.googleSvc.On("GetUserInfo", mock.Anything, "code").Return(googleUser, nil)
	s.userRepo.On("GetByGoogleID", nil, "google-mismatch").Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	tx := yca_repository.NewMockTx()
	s.userRepo.On("BeginTx").Return(tx, nil)
	s.userRepo.On("GetByEmail", nil, "user@example.com").Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	s.userRepo.On("Create", tx, mock.AnythingOfType("*models.User")).Return(nil)

	invitation := &models.Invitation{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		RoleID:         uuid.New(),
		Email:          "invited@example.com",
		ExpiresAt:      s.now.Add(24 * time.Hour),
	}
	tokenHash := "hashed:inv-token"
	s.invitationRepo.On("GetByTokenHash", tokenHash).Return(invitation, nil)
	tx.On("Rollback").Return(nil).Maybe()

	req := &auth_service.AuthenticateWithGoogleRequest{
		Code:            "code",
		TermsVersion:    "1.0.0",
		InvitationToken: "inv-token",
		IPAddress:       "192.168.1.1",
		UserAgent:       "test-agent",
		Language:        "en",
	}
	resp, err := s.svc.AuthenticateWithGoogle(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVITATION_EMAIL_MISMATCH_CODE, e.ErrorCode)
	}
	s.invitationRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestAuthenticateWithGoogle_ExistingUserByEmail_LinkGoogle_Success() {
	googleUser := &models.GoogleUserInfo{
		ID:            "google-789",
		Email:         "existing@example.com",
		VerifiedEmail: true,
		Name:          "Existing User",
	}
	userID := uuid.New()
	existingUser := &models.User{
		ID:              userID,
		Email:           "existing@example.com",
		GoogleID:        nil,
		EmailVerifiedAt: nil,
	}
	s.googleSvc.On("GetUserInfo", mock.Anything, "code").Return(googleUser, nil)
	s.userRepo.On("GetByGoogleID", nil, "google-789").Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	tx := yca_repository.NewMockTx()
	s.userRepo.On("BeginTx").Return(tx, nil)
	s.userRepo.On("GetByEmail", nil, "existing@example.com").Return(existingUser, nil)
	s.userRepo.On("Update", tx, mock.AnythingOfType("*models.User")).Return(nil)
	s.adminAccessRepo.On("GetByUserID", userID.String()).Return((*models.AdminAccess)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	s.orgMemberRepo.On("ListByUserIDWithRole", userID.String()).Return(&[]models.OrganizationMemberWithOrganizationAndRole{}, nil)
	s.refreshTokenRepo.On("Create", tx, mock.AnythingOfType("*models.UserRefreshToken")).Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()

	req := &auth_service.AuthenticateWithGoogleRequest{
		Code:         "code",
		TermsVersion: "1.0.0",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
		Language:     "en",
	}
	resp, err := s.svc.AuthenticateWithGoogle(req)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.NotEmpty(resp.AccessToken)
	s.NotEmpty(resp.RefreshToken)
	s.userRepo.AssertExpectations(s.T())
	s.refreshTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestAuthenticateWithGoogle_NewUserWithInvalidInvitationToken() {
	googleUser := &models.GoogleUserInfo{
		ID:            "google-inv",
		Email:         "invite@example.com",
		VerifiedEmail: true,
		Name:          "Invite User",
	}
	s.googleSvc.On("GetUserInfo", mock.Anything, "code").Return(googleUser, nil)
	s.userRepo.On("GetByGoogleID", nil, "google-inv").Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	tx := yca_repository.NewMockTx()
	s.userRepo.On("BeginTx").Return(tx, nil)
	s.userRepo.On("GetByEmail", nil, "invite@example.com").Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	s.userRepo.On("Create", tx, mock.AnythingOfType("*models.User")).Return(nil)
	tokenHash := "hashed:invitation-token"
	s.invitationRepo.On("GetByTokenHash", tokenHash).Return((*models.Invitation)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	tx.On("Rollback").Return(nil).Maybe()

	req := &auth_service.AuthenticateWithGoogleRequest{
		Code:            "code",
		TermsVersion:    "1.0.0",
		InvitationToken: "invitation-token",
		IPAddress:       "192.168.1.1",
		UserAgent:       "test-agent",
		Language:        "en",
	}
	resp, err := s.svc.AuthenticateWithGoogle(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVALID_INVITATION_TOKEN_CODE, e.ErrorCode)
	}
	s.invitationRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestAuthenticateWithGoogle_NewUserWithRevokedInvitation() {
	googleUser := &models.GoogleUserInfo{
		ID:            "google-rev",
		Email:         "revoked@example.com",
		VerifiedEmail: true,
		Name:          "Revoked User",
	}
	revokedAt := s.now.Add(-1 * time.Hour)
	invitation := &models.Invitation{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		RoleID:         uuid.New(),
		Email:          "revoked@example.com",
		RevokedAt:      &revokedAt,
		AcceptedAt:     nil,
		ExpiresAt:      s.now.Add(24 * time.Hour),
	}
	s.googleSvc.On("GetUserInfo", mock.Anything, "code").Return(googleUser, nil)
	s.userRepo.On("GetByGoogleID", nil, "google-rev").Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	tx := yca_repository.NewMockTx()
	s.userRepo.On("BeginTx").Return(tx, nil)
	s.userRepo.On("GetByEmail", nil, "revoked@example.com").Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	s.userRepo.On("Create", tx, mock.AnythingOfType("*models.User")).Return(nil)
	tokenHash := "hashed:inv-token"
	s.invitationRepo.On("GetByTokenHash", tokenHash).Return(invitation, nil)
	tx.On("Rollback").Return(nil).Maybe()

	req := &auth_service.AuthenticateWithGoogleRequest{
		Code:            "code",
		TermsVersion:    "1.0.0",
		InvitationToken: "inv-token",
		IPAddress:       "192.168.1.1",
		UserAgent:       "test-agent",
		Language:        "en",
	}
	resp, err := s.svc.AuthenticateWithGoogle(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVITATION_REVOKED_CODE, e.ErrorCode)
	}
	s.invitationRepo.AssertExpectations(s.T())
}

// --- VerifyEmail: validations ---

func (s *AuthServiceTestSuite) TestVerifyEmail_Validation_MissingToken() {
	req := &auth_service.VerifyEmailRequest{
		Token: "",
	}
	err := s.svc.VerifyEmail(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- VerifyEmail: business logic ---

func (s *AuthServiceTestSuite) TestVerifyEmail_TokenNotFound() {
	token := "test-token"
	tokenHash := "hashed:" + token
	s.emailVerificationTokenRepo.On("GetByHash", nil, tokenHash).Return((*models.UserEmailVerificationToken)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &auth_service.VerifyEmailRequest{
		Token: token,
	}
	err := s.svc.VerifyEmail(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVALID_VERIFICATION_TOKEN_CODE, e.ErrorCode)
	}
	s.emailVerificationTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestVerifyEmail_TokenExpired() {
	token := "test-token"
	tokenHash := "hashed:" + token
	userID := uuid.New()
	expiredTime := s.now.Add(-1 * time.Hour)
	verificationToken := &models.UserEmailVerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: expiredTime,
		UsedAt:    nil,
	}
	s.emailVerificationTokenRepo.On("GetByHash", nil, tokenHash).Return(verificationToken, nil)
	s.userRepo.On("GetByID", nil, userID.String()).Return(&models.User{ID: userID}, nil)

	req := &auth_service.VerifyEmailRequest{
		Token: token,
	}
	err := s.svc.VerifyEmail(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.EXPIRED_VERIFICATION_TOKEN_CODE, e.ErrorCode)
	}
	s.emailVerificationTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestVerifyEmail_TokenAlreadyUsed() {
	token := "test-token"
	tokenHash := "hashed:" + token
	userID := uuid.New()
	usedTime := s.now.Add(-1 * time.Hour)
	verificationToken := &models.UserEmailVerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: s.now.Add(1 * time.Hour),
		UsedAt:    &usedTime,
	}
	s.emailVerificationTokenRepo.On("GetByHash", nil, tokenHash).Return(verificationToken, nil)
	s.userRepo.On("GetByID", nil, userID.String()).Return(&models.User{ID: userID}, nil)

	req := &auth_service.VerifyEmailRequest{
		Token: token,
	}
	err := s.svc.VerifyEmail(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVALID_VERIFICATION_TOKEN_CODE, e.ErrorCode)
	}
	s.emailVerificationTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestVerifyEmail_Success() {
	token := "test-token"
	tokenHash := "hashed:" + token
	userID := uuid.New()
	verificationToken := &models.UserEmailVerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: s.now.Add(1 * time.Hour),
		UsedAt:    nil,
	}
	s.emailVerificationTokenRepo.On("GetByHash", nil, tokenHash).Return(verificationToken, nil)
	s.userRepo.On("GetByID", nil, userID.String()).Return(&models.User{ID: userID}, nil)
	s.emailVerificationTokenRepo.On("MarkAsUsed", nil, verificationToken.ID.String()).Return(nil)

	req := &auth_service.VerifyEmailRequest{
		Token: token,
	}
	err := s.svc.VerifyEmail(req)
	s.NoError(err)
	s.emailVerificationTokenRepo.AssertExpectations(s.T())
}

// --- ForgotPassword: validations ---

func (s *AuthServiceTestSuite) TestForgotPassword_Validation_MissingEmail() {
	req := &auth_service.ForgotPasswordRequest{
		Email:    "",
		Language: "en",
	}
	err := s.svc.ForgotPassword(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestForgotPassword_Validation_InvalidEmail() {
	req := &auth_service.ForgotPasswordRequest{
		Email:    "not-an-email",
		Language: "en",
	}
	err := s.svc.ForgotPassword(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestForgotPassword_Validation_InvalidLanguage() {
	req := &auth_service.ForgotPasswordRequest{
		Email:    "test@example.com",
		Language: "invalid",
	}
	err := s.svc.ForgotPassword(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- ForgotPassword: business logic ---

func (s *AuthServiceTestSuite) TestForgotPassword_UserNotFound() {
	email := "test@example.com"
	s.userRepo.On("GetByEmail", nil, email).Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &auth_service.ForgotPasswordRequest{
		Email:    email,
		Language: "en",
	}
	err := s.svc.ForgotPassword(req)
	s.NoError(err) // Should return nil when user not found (security)
	s.userRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestForgotPassword_Success() {
	email := "test@example.com"
	user := &models.User{
		ID:    uuid.New(),
		Email: email,
	}
	s.userRepo.On("GetByEmail", nil, email).Return(user, nil)
	tx := yca_repository.NewMockTx()
	s.passwordResetTokenRepo.On("BeginTx").Return(tx, nil)
	s.passwordResetTokenRepo.On("Create", tx, mock.AnythingOfType("*models.UserPasswordResetToken")).Return(nil)
	s.translator.On("Translate", "en", mock.AnythingOfType("string"), mock.Anything).Return("translated")
	s.emailSvc.On("PrepareEmailBody", "reset", mock.Anything).Return("email-body", nil)
	s.emailSvc.On("SendEmail", email, mock.AnythingOfType("string"), "email-body").Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()

	req := &auth_service.ForgotPasswordRequest{
		Email:    email,
		Language: "en",
	}
	err := s.svc.ForgotPassword(req)
	s.NoError(err)
	s.userRepo.AssertExpectations(s.T())
	s.passwordResetTokenRepo.AssertExpectations(s.T())
}

// --- ResetPassword: validations ---

func (s *AuthServiceTestSuite) TestResetPassword_Validation_MissingToken() {
	req := &auth_service.ResetPasswordRequest{
		Token:    "",
		Password: "newpassword123",
	}
	err := s.svc.ResetPassword(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestResetPassword_Validation_MissingPassword() {
	req := &auth_service.ResetPasswordRequest{
		Token:    "test-token",
		Password: "",
	}
	err := s.svc.ResetPassword(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestResetPassword_Validation_PasswordTooShort() {
	req := &auth_service.ResetPasswordRequest{
		Token:    "test-token",
		Password: "short",
	}
	err := s.svc.ResetPassword(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- ResetPassword: business logic ---

func (s *AuthServiceTestSuite) TestResetPassword_TokenNotFound() {
	token := "test-token"
	tokenHash := "hashed:" + token
	s.passwordResetTokenRepo.On("GetByHash", nil, tokenHash).Return((*models.UserPasswordResetToken)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &auth_service.ResetPasswordRequest{
		Token:    token,
		Password: "newpassword123",
	}
	err := s.svc.ResetPassword(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVALID_PASSWORD_RESET_TOKEN_CODE, e.ErrorCode)
	}
	s.passwordResetTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestResetPassword_TokenExpired() {
	token := "test-token"
	tokenHash := "hashed:" + token
	userID := uuid.New()
	expiredTime := s.now.Add(-1 * time.Hour)
	resetToken := &models.UserPasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: expiredTime,
		UsedAt:    nil,
	}
	s.passwordResetTokenRepo.On("GetByHash", nil, tokenHash).Return(resetToken, nil)
	s.userRepo.On("GetByID", nil, userID.String()).Return(&models.User{ID: userID}, nil)

	req := &auth_service.ResetPasswordRequest{
		Token:    token,
		Password: "newpassword123",
	}
	err := s.svc.ResetPassword(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.EXPIRED_PASSWORD_RESET_TOKEN_CODE, e.ErrorCode)
	}
	s.passwordResetTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestResetPassword_TokenAlreadyUsed() {
	token := "test-token"
	tokenHash := "hashed:" + token
	userID := uuid.New()
	usedTime := s.now.Add(-1 * time.Hour)
	resetToken := &models.UserPasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: s.now.Add(1 * time.Hour),
		UsedAt:    &usedTime,
	}
	s.passwordResetTokenRepo.On("GetByHash", nil, tokenHash).Return(resetToken, nil)
	s.userRepo.On("GetByID", nil, userID.String()).Return(&models.User{ID: userID}, nil)

	req := &auth_service.ResetPasswordRequest{
		Token:    token,
		Password: "newpassword123",
	}
	err := s.svc.ResetPassword(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVALID_PASSWORD_RESET_TOKEN_CODE, e.ErrorCode)
	}
	s.passwordResetTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestResetPassword_Success() {
	token := "test-token"
	tokenHash := "hashed:" + token
	userID := uuid.New()
	resetToken := &models.UserPasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: s.now.Add(1 * time.Hour),
		UsedAt:    nil,
	}
	user := &models.User{
		ID:       userID,
		Password: nil,
	}
	s.passwordResetTokenRepo.On("GetByHash", nil, tokenHash).Return(resetToken, nil)
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)
	tx := yca_repository.NewMockTx()
	s.userRepo.On("BeginTx").Return(tx, nil)
	s.passwordResetTokenRepo.On("MarkAsUsed", tx, resetToken.ID.String()).Return(nil)
	s.userRepo.On("Update", tx, mock.AnythingOfType("*models.User")).Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()

	req := &auth_service.ResetPasswordRequest{
		Token:    token,
		Password: "newpassword123",
	}
	err := s.svc.ResetPassword(req)
	s.NoError(err)
	s.passwordResetTokenRepo.AssertExpectations(s.T())
	s.userRepo.AssertExpectations(s.T())
}

// --- Logout: validations ---

func (s *AuthServiceTestSuite) TestLogout_Validation_MissingRefreshToken() {
	req := &auth_service.LogoutRequest{
		RefreshToken: "",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	err := s.svc.Logout(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Logout: business logic ---

func (s *AuthServiceTestSuite) TestLogout_TokenNotFound() {
	token := "test-token"
	tokenHash := "hashed:" + token
	s.refreshTokenRepo.On("GetByHash", nil, tokenHash).Return((*models.UserRefreshToken)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &auth_service.LogoutRequest{
		RefreshToken: token,
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	err := s.svc.Logout(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVALID_TOKEN_CODE, e.ErrorCode)
	}
	s.refreshTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestLogout_UserMismatch() {
	token := "test-token"
	tokenHash := "hashed:" + token
	userID := uuid.New()
	otherUserID := uuid.New()
	refreshToken := &models.UserRefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: s.now.Add(1 * time.Hour),
		RevokedAt: nil,
	}
	s.refreshTokenRepo.On("GetByHash", nil, tokenHash).Return(refreshToken, nil)

	req := &auth_service.LogoutRequest{
		RefreshToken: token,
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: otherUserID,
		},
	}
	err := s.svc.Logout(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
	s.refreshTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestLogout_TokenAlreadyRevoked() {
	token := "test-token"
	tokenHash := "hashed:" + token
	userID := uuid.New()
	revokedTime := s.now.Add(-1 * time.Hour)
	refreshToken := &models.UserRefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: s.now.Add(1 * time.Hour),
		RevokedAt: &revokedTime,
	}
	s.refreshTokenRepo.On("GetByHash", nil, tokenHash).Return(refreshToken, nil)

	req := &auth_service.LogoutRequest{
		RefreshToken: token,
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}
	err := s.svc.Logout(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVALID_TOKEN_CODE, e.ErrorCode)
	}
	s.refreshTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestLogout_TokenExpired() {
	token := "test-token"
	tokenHash := "hashed:" + token
	userID := uuid.New()
	expiredTime := s.now.Add(-1 * time.Hour)
	refreshToken := &models.UserRefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: expiredTime,
		RevokedAt: nil,
	}
	s.refreshTokenRepo.On("GetByHash", nil, tokenHash).Return(refreshToken, nil)

	req := &auth_service.LogoutRequest{
		RefreshToken: token,
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}
	err := s.svc.Logout(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.EXPIRED_TOKEN_CODE, e.ErrorCode)
	}
	s.refreshTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestLogout_Success() {
	token := "test-token"
	tokenHash := "hashed:" + token
	userID := uuid.New()
	refreshToken := &models.UserRefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: s.now.Add(1 * time.Hour),
		RevokedAt: nil,
	}
	s.refreshTokenRepo.On("GetByHash", nil, tokenHash).Return(refreshToken, nil)
	s.refreshTokenRepo.On("RevokeByHash", nil, tokenHash).Return(nil)

	req := &auth_service.LogoutRequest{
		RefreshToken: token,
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}
	err := s.svc.Logout(req, accessInfo)
	s.NoError(err)
	s.refreshTokenRepo.AssertExpectations(s.T())
}

// --- RefreshAccessToken: validations ---

func (s *AuthServiceTestSuite) TestRefreshAccessToken_Validation_MissingRefreshToken() {
	req := &auth_service.RefreshAccessTokenRequest{
		RefreshToken: "",
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
	}
	resp, err := s.svc.RefreshAccessToken(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestRefreshAccessToken_Validation_InvalidIPAddress() {
	req := &auth_service.RefreshAccessTokenRequest{
		RefreshToken: "test-token",
		IPAddress:    "not-an-ip",
		UserAgent:    "test-agent",
	}
	resp, err := s.svc.RefreshAccessToken(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- RefreshAccessToken: business logic ---

func (s *AuthServiceTestSuite) TestRefreshAccessToken_TokenNotFound() {
	token := "test-token"
	tokenHash := "hashed:" + token
	s.refreshTokenRepo.On("GetByHash", nil, tokenHash).Return((*models.UserRefreshToken)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &auth_service.RefreshAccessTokenRequest{
		RefreshToken: token,
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
	}
	resp, err := s.svc.RefreshAccessToken(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVALID_TOKEN_CODE, e.ErrorCode)
	}
	s.refreshTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestRefreshAccessToken_UserNotFound() {
	token := "test-token"
	tokenHash := "hashed:" + token
	userID := uuid.New()
	refreshToken := &models.UserRefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: s.now.Add(1 * time.Hour),
		RevokedAt: nil,
	}
	s.refreshTokenRepo.On("GetByHash", nil, tokenHash).Return(refreshToken, nil)
	s.userRepo.On("GetByID", nil, userID.String()).Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &auth_service.RefreshAccessTokenRequest{
		RefreshToken: token,
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
	}
	resp, err := s.svc.RefreshAccessToken(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVALID_TOKEN_CODE, e.ErrorCode)
	}
	s.refreshTokenRepo.AssertExpectations(s.T())
	s.userRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestRefreshAccessToken_TokenRevoked() {
	token := "test-token"
	tokenHash := "hashed:" + token
	userID := uuid.New()
	revokedTime := s.now.Add(-1 * time.Hour)
	refreshToken := &models.UserRefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: s.now.Add(1 * time.Hour),
		RevokedAt: &revokedTime,
	}
	s.refreshTokenRepo.On("GetByHash", nil, tokenHash).Return(refreshToken, nil)
	s.userRepo.On("GetByID", nil, userID.String()).Return(&models.User{ID: userID}, nil)
	s.logger.On("Log", mock.Anything).Return()

	req := &auth_service.RefreshAccessTokenRequest{
		RefreshToken: token,
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
	}
	resp, err := s.svc.RefreshAccessToken(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVALID_TOKEN_CODE, e.ErrorCode)
	}
	s.refreshTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestRefreshAccessToken_TokenExpired() {
	token := "test-token"
	tokenHash := "hashed:" + token
	userID := uuid.New()
	expiredTime := s.now.Add(-1 * time.Hour)
	refreshToken := &models.UserRefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: expiredTime,
		RevokedAt: nil,
	}
	s.refreshTokenRepo.On("GetByHash", nil, tokenHash).Return(refreshToken, nil)
	s.userRepo.On("GetByID", nil, userID.String()).Return(&models.User{ID: userID}, nil)

	req := &auth_service.RefreshAccessTokenRequest{
		RefreshToken: token,
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
	}
	resp, err := s.svc.RefreshAccessToken(req)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.EXPIRED_TOKEN_CODE, e.ErrorCode)
	}
	s.refreshTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestRefreshAccessToken_Success() {
	token := "test-token"
	tokenHash := "hashed:" + token
	userID := uuid.New()
	refreshToken := &models.UserRefreshToken{
		ID:             uuid.New(),
		UserID:         userID,
		ExpiresAt:      s.now.Add(1 * time.Hour),
		RevokedAt:      nil,
		ImpersonatedBy: uuid.NullUUID{Valid: false},
	}
	user := &models.User{
		ID: userID,
	}
	s.refreshTokenRepo.On("GetByHash", nil, tokenHash).Return(refreshToken, nil)
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)
	s.adminAccessRepo.On("GetByUserID", userID.String()).Return((*models.AdminAccess)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	s.orgMemberRepo.On("ListByUserIDWithRole", userID.String()).Return(&[]models.OrganizationMemberWithOrganizationAndRole{}, nil)

	req := &auth_service.RefreshAccessTokenRequest{
		RefreshToken: token,
		IPAddress:    "192.168.1.1",
		UserAgent:    "test-agent",
	}
	resp, err := s.svc.RefreshAccessToken(req)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.NotEmpty(resp.AccessToken)
	s.refreshTokenRepo.AssertExpectations(s.T())
	s.userRepo.AssertExpectations(s.T())
}

// --- ResendVerificationEmail: validations ---

func (s *AuthServiceTestSuite) TestResendVerificationEmail_Validation_MissingUserID() {
	req := &auth_service.ResendVerificationEmailRequest{
		UserID:   "",
		Language: "en",
	}
	err := s.svc.ResendVerificationEmail(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestResendVerificationEmail_Validation_InvalidUserID() {
	req := &auth_service.ResendVerificationEmailRequest{
		UserID:   "not-a-uuid",
		Language: "en",
	}
	err := s.svc.ResendVerificationEmail(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- ResendVerificationEmail: business logic ---

func (s *AuthServiceTestSuite) TestResendVerificationEmail_UserNotFound() {
	userID := uuid.New()
	s.userRepo.On("GetByID", nil, userID.String()).Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &auth_service.ResendVerificationEmailRequest{
		UserID:   userID.String(),
		Language: "en",
	}
	err := s.svc.ResendVerificationEmail(req)
	s.Error(err)
	s.userRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestResendVerificationEmail_AlreadyVerified() {
	userID := uuid.New()
	verifiedTime := s.now.Add(-1 * time.Hour)
	user := &models.User{
		ID:              userID,
		EmailVerifiedAt: &verifiedTime,
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)

	req := &auth_service.ResendVerificationEmailRequest{
		UserID:   userID.String(),
		Language: "en",
	}
	err := s.svc.ResendVerificationEmail(req)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.EMAIL_ALREADY_VERIFIED_CODE, e.ErrorCode)
	}
	s.userRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestResendVerificationEmail_Success() {
	userID := uuid.New()
	email := "test@example.com"
	user := &models.User{
		ID:              userID,
		Email:           email,
		EmailVerifiedAt: nil,
		Language:        "en",
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)
	tx := yca_repository.NewMockTx()
	s.userRepo.On("BeginTx").Return(tx, nil)
	s.emailVerificationTokenRepo.On("Create", tx, mock.AnythingOfType("*models.UserEmailVerificationToken")).Return(nil)
	s.translator.On("Translate", "en", mock.AnythingOfType("string"), mock.Anything).Return("translated")
	s.emailSvc.On("PrepareEmailBody", "verification", mock.Anything).Return("email-body", nil)
	s.emailSvc.On("SendEmail", email, mock.AnythingOfType("string"), "email-body").Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()

	req := &auth_service.ResendVerificationEmailRequest{
		UserID:   userID.String(),
		Language: "en",
	}
	err := s.svc.ResendVerificationEmail(req)
	s.NoError(err)
	s.userRepo.AssertExpectations(s.T())
	s.emailVerificationTokenRepo.AssertExpectations(s.T())
}

// --- Impersonate: validations ---

func (s *AuthServiceTestSuite) TestImpersonate_Validation_InvalidUserID() {
	req := &auth_service.ImpersonateRequest{
		UserID:    "not-a-uuid",
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New(), IsAdmin: true},
	}
	resp, err := s.svc.Impersonate(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestImpersonate_Validation_InvalidIPAddress() {
	req := &auth_service.ImpersonateRequest{
		UserID:    uuid.New().String(),
		IPAddress: "not-an-ip",
		UserAgent: "test-agent",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New(), IsAdmin: true},
	}
	resp, err := s.svc.Impersonate(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestImpersonate_Validation_MissingUserAgent() {
	req := &auth_service.ImpersonateRequest{
		UserID:    uuid.New().String(),
		IPAddress: "192.168.1.1",
		UserAgent: "",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New(), IsAdmin: true},
	}
	resp, err := s.svc.Impersonate(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Impersonate: business logic ---

func (s *AuthServiceTestSuite) TestImpersonate_NotAdmin() {
	req := &auth_service.ImpersonateRequest{
		UserID:    uuid.New().String(),
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New(), IsAdmin: false},
	}
	resp, err := s.svc.Impersonate(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
}

func (s *AuthServiceTestSuite) TestImpersonate_UserNotFound() {
	userID := uuid.New()
	adminID := uuid.New()
	req := &auth_service.ImpersonateRequest{
		UserID:    userID.String(),
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: adminID, IsAdmin: true},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.Impersonate(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.userRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestImpersonate_Success() {
	userID := uuid.New()
	adminID := uuid.New()
	user := &models.User{ID: userID, Email: "target@example.com", FirstName: "Target", LastName: "User"}
	req := &auth_service.ImpersonateRequest{
		UserID:    userID.String(),
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: adminID, IsAdmin: true, Email: "admin@example.com"},
	}
	s.userRepo.On("GetByID", nil, userID.String()).Return(user, nil)
	s.adminAccessRepo.On("GetByUserID", userID.String()).Return((*models.AdminAccess)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	s.orgMemberRepo.On("ListByUserIDWithRole", userID.String()).Return(&[]models.OrganizationMemberWithOrganizationAndRole{}, nil)
	s.refreshTokenRepo.On("Create", nil, mock.AnythingOfType("*models.UserRefreshToken")).Return(nil)

	resp, err := s.svc.Impersonate(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.NotEmpty(resp.AccessToken)
	s.NotEmpty(resp.RefreshToken)
	s.userRepo.AssertExpectations(s.T())
	s.adminAccessRepo.AssertExpectations(s.T())
	s.orgMemberRepo.AssertExpectations(s.T())
	s.refreshTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestCleanupStalePasswordResetTokens() {
	s.passwordResetTokenRepo.On("Cleanup", mock.Anything).Return(nil).Once()
	err := s.svc.CleanupStalePasswordResetTokens()
	s.NoError(err)
	s.passwordResetTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestCleanupStaleEmailVerificationTokens() {
	s.emailVerificationTokenRepo.On("Cleanup", mock.Anything).Return(nil).Once()
	err := s.svc.CleanupStaleEmailVerificationTokens()
	s.NoError(err)
	s.emailVerificationTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestCleanupStalePasswordResetTokens_RepoError() {
	repoErr := errors.New("cleanup failed")
	s.passwordResetTokenRepo.On("Cleanup", mock.Anything).Return(repoErr).Once()
	err := s.svc.CleanupStalePasswordResetTokens()
	s.ErrorIs(err, repoErr)
	s.passwordResetTokenRepo.AssertExpectations(s.T())
}

func (s *AuthServiceTestSuite) TestCleanupStaleEmailVerificationTokens_RepoError() {
	repoErr := errors.New("cleanup failed")
	s.emailVerificationTokenRepo.On("Cleanup", mock.Anything).Return(repoErr).Once()
	err := s.svc.CleanupStaleEmailVerificationTokens()
	s.ErrorIs(err, repoErr)
	s.emailVerificationTokenRepo.AssertExpectations(s.T())
}
