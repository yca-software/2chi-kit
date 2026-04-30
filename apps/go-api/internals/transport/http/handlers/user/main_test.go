package user_handler_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/metrics"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	auth_service "github.com/yca-software/2chi-kit/go-api/internals/services/auth"
	user_service "github.com/yca-software/2chi-kit/go-api/internals/services/user"
	user_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/user"
	http_middlewares "github.com/yca-software/2chi-kit/go-api/internals/transport/http/middlewares"
)

type UserHandlerTestSuite struct {
	suite.Suite
	handler         *user_handler.Handler
	echo            *echo.Echo
	mockUserService *user_service.MockService
	mockAuthService *auth_service.MockService
	services        *services.Services
	middlewares     http_middlewares.Middlewares
	accessInfo      *models.AccessInfo
}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}

func (s *UserHandlerTestSuite) SetupTest() {
	s.echo = echo.New()
	s.mockUserService = user_service.NewMockService()
	s.mockAuthService = auth_service.NewMockService()
	s.services = &services.Services{
		User: s.mockUserService,
		Auth: s.mockAuthService,
	}
	mwareCfg := &http_middlewares.MiddlewaresConfig{AccessSecret: "test-secret"}
	m, _ := metrics.New("test")
	repos := &repositories.Repositories{}
	hashToken := func(raw string) string {
		h := sha256.Sum256([]byte(raw))
		return hex.EncodeToString(h[:])
	}
	s.middlewares = http_middlewares.NewMiddlewares(mwareCfg, repos, m, hashToken)

	s.handler = user_handler.New(s.services, s.middlewares, &http_middlewares.RateLimitConfig{})
	s.handler.RegisterEndpoints(s.echo.Group("/api"))

	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "test@example.com",
		},
	}
}

func (s *UserHandlerTestSuite) setAccessInfo(c echo.Context) {
	c.Set("accessInfo", s.accessInfo)
}

func (s *UserHandlerTestSuite) TestGetCurrentUser_Success() {
	expectedUser := &user_service.GetResponse{
		User: &models.User{
			ID:    s.accessInfo.User.UserID,
			Email: "test@example.com",
		},
	}

	s.mockUserService.On("Get", mock.AnythingOfType("*user_service.GetRequest"), s.accessInfo).
		Return(expectedUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.GetCurrentUser(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockUserService.AssertExpectations(s.T())
}

func (s *UserHandlerTestSuite) TestUpdateProfile_Success() {
	reqBody := user_service.UpdateProfileRequest{
		FirstName: "Updated",
		LastName:  "Name",
	}
	body, _ := json.Marshal(reqBody)

	expectedUser := &models.User{
		ID:        s.accessInfo.User.UserID,
		Email:     "test@example.com",
		FirstName: "Updated",
		LastName:  "Name",
	}

	s.mockUserService.On("UpdateProfile", mock.AnythingOfType("*user_service.UpdateProfileRequest"), s.accessInfo).
		Return(expectedUser, nil)

	req := httptest.NewRequest(http.MethodPatch, "/api/user/profile", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.UpdateProfile(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockUserService.AssertExpectations(s.T())
}

func (s *UserHandlerTestSuite) TestChangePassword_Success() {
	reqBody := user_service.ChangePasswordRequest{
		CurrentPassword: "oldpassword",
		NewPassword:     "newpassword123",
	}
	body, _ := json.Marshal(reqBody)

	s.mockUserService.On("ChangePassword", mock.AnythingOfType("*user_service.ChangePasswordRequest")).
		Return(nil)

	req := httptest.NewRequest(http.MethodPatch, "/api/user/password", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.ChangePassword(c)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockUserService.AssertExpectations(s.T())
}

func (s *UserHandlerTestSuite) TestUpdateProfileLanguage_Success() {
	reqBody := user_service.UpdateLanguageRequest{
		Language: "en",
	}
	body, _ := json.Marshal(reqBody)

	expectedUser := &models.User{
		ID:       s.accessInfo.User.UserID,
		Language: "en",
	}

	s.mockUserService.On("UpdateLanguage", mock.AnythingOfType("*user_service.UpdateLanguageRequest"), s.accessInfo).
		Return(expectedUser, nil)

	req := httptest.NewRequest(http.MethodPatch, "/api/user/language", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.UpdateProfileLanguage(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockUserService.AssertExpectations(s.T())
}

func (s *UserHandlerTestSuite) TestResendVerificationEmail_Success() {
	s.mockAuthService.On("ResendVerificationEmail", mock.AnythingOfType("*auth_service.ResendVerificationEmailRequest")).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/user/resend-verification-email", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.ResendVerificationEmail(c)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockAuthService.AssertExpectations(s.T())
}

func (s *UserHandlerTestSuite) TestAcceptTerms_Success() {
	reqBody := user_service.AcceptTermsRequest{
		TermsVersion: "1.0.0",
	}
	body, _ := json.Marshal(reqBody)

	expectedUser := &models.User{
		ID:              s.accessInfo.User.UserID,
		Email:           "test@example.com",
		TermsVersion:    "1.0.0",
		TermsAcceptedAt: time.Now().UTC(),
	}

	s.mockUserService.On("AcceptTerms", mock.AnythingOfType("*user_service.AcceptTermsRequest"), s.accessInfo).
		Return(expectedUser, nil)

	req := httptest.NewRequest(http.MethodPatch, "/api/user/terms", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.AcceptTerms(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockUserService.AssertExpectations(s.T())
}
