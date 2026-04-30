package auth_service

import (
	"github.com/stretchr/testify/mock"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
)

type MockService struct {
	mock.Mock
}

func NewMockService() *MockService {
	return &MockService{
		mock.Mock{},
	}
}

func (m *MockService) AuthenticateWithGoogle(req *AuthenticateWithGoogleRequest) (*AuthenticateResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthenticateResponse), args.Error(1)
}

func (m *MockService) AuthenticateWithPassword(req *AuthenticateWithPasswordRequest) (*AuthenticateResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthenticateResponse), args.Error(1)
}

func (m *MockService) ForgotPassword(req *ForgotPasswordRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockService) Logout(req *LogoutRequest, accessInfo *models.AccessInfo) error {
	args := m.Called(req, accessInfo)
	return args.Error(0)
}

func (m *MockService) RefreshAccessToken(req *RefreshAccessTokenRequest) (*RefreshAccessTokenResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*RefreshAccessTokenResponse), args.Error(1)
}

func (m *MockService) ResetPassword(req *ResetPasswordRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockService) SignUp(req *SignUpRequest) (*SignUpResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*SignUpResponse), args.Error(1)
}

func (m *MockService) VerifyEmail(req *VerifyEmailRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockService) ResendVerificationEmail(req *ResendVerificationEmailRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockService) Impersonate(req *ImpersonateRequest, accessInfo *models.AccessInfo) (*AuthenticateResponse, error) {
	args := m.Called(req, accessInfo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthenticateResponse), args.Error(1)
}

func (m *MockService) CleanupStalePasswordResetTokens() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockService) CleanupStaleEmailVerificationTokens() error {
	args := m.Called()
	return args.Error(0)
}
