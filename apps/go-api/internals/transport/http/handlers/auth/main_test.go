package auth_handler_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/metrics"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	auth_service "github.com/yca-software/2chi-kit/go-api/internals/services/auth"
	auth_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/auth"
	http_middlewares "github.com/yca-software/2chi-kit/go-api/internals/transport/http/middlewares"
	yca_error "github.com/yca-software/go-common/error"
)

type AuthHandlerTestSuite struct {
	suite.Suite
	handler         *auth_handler.Handler
	echo            *echo.Echo
	mockAuthService *auth_service.MockService
	services        *services.Services
	middlewares     http_middlewares.Middlewares
}

func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

func (s *AuthHandlerTestSuite) SetupTest() {
	s.echo = echo.New()
	// Set up error handler to properly handle yca_error.Error
	s.echo.HTTPErrorHandler = func(err error, c echo.Context) {
		if httpErr, ok := err.(*yca_error.Error); ok {
			c.JSON(httpErr.StatusCode, httpErr)
			return
		}
		if he, ok := err.(*echo.HTTPError); ok {
			c.JSON(he.Code, he)
			return
		}
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	s.mockAuthService = auth_service.NewMockService()
	s.services = &services.Services{
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

	rateLimitConfig := &http_middlewares.RateLimitConfig{}
	s.handler = auth_handler.New(s.services, s.middlewares, rateLimitConfig)
	s.handler.RegisterEndpoints(s.echo.Group("/api"))
}

func (s *AuthHandlerTestSuite) TestAuthenticateWithPassword_Success() {
	reqBody := auth_service.AuthenticateWithPasswordRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	expectedResp := &auth_service.AuthenticateResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	s.mockAuthService.On("AuthenticateWithPassword", mock.AnythingOfType("*auth_service.AuthenticateWithPasswordRequest")).
		Return(expectedResp, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.mockAuthService.AssertExpectations(s.T())
}

func (s *AuthHandlerTestSuite) TestAuthenticateWithPassword_InvalidRequest() {
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *AuthHandlerTestSuite) TestAuthenticateWithPassword_ServiceError() {
	reqBody := auth_service.AuthenticateWithPasswordRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	var nilResp *auth_service.AuthenticateResponse
	s.mockAuthService.On("AuthenticateWithPassword", mock.AnythingOfType("*auth_service.AuthenticateWithPasswordRequest")).
		Return(nilResp, yca_error.NewUnauthorizedError(nil, "invalid credentials", nil))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	s.Equal(http.StatusUnauthorized, rec.Code)
	s.mockAuthService.AssertExpectations(s.T())
}

func (s *AuthHandlerTestSuite) TestSignUp_Success() {
	reqBody := auth_service.SignUpRequest{
		Email:        "newuser@example.com",
		Password:     "password123",
		FirstName:    "Test",
		LastName:     "User",
		TermsVersion: "1.0.0",
	}
	body, _ := json.Marshal(reqBody)

	expectedResp := &auth_service.SignUpResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	s.mockAuthService.On("SignUp", mock.AnythingOfType("*auth_service.SignUpRequest")).
		Return(expectedResp, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	s.Equal(http.StatusCreated, rec.Code)
	s.mockAuthService.AssertExpectations(s.T())
}

func (s *AuthHandlerTestSuite) TestSignUp_InvalidRequest() {
	req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewReader([]byte("invalid json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *AuthHandlerTestSuite) TestRefreshAccessToken_Success() {
	reqBody := auth_service.RefreshAccessTokenRequest{
		RefreshToken: "refresh-token",
	}
	body, _ := json.Marshal(reqBody)

	expectedResp := &auth_service.RefreshAccessTokenResponse{
		AccessToken: "new-access-token",
	}

	s.mockAuthService.On("RefreshAccessToken", mock.AnythingOfType("*auth_service.RefreshAccessTokenRequest")).
		Return(expectedResp, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.mockAuthService.AssertExpectations(s.T())
}

func (s *AuthHandlerTestSuite) TestForgotPassword_Success() {
	reqBody := auth_service.ForgotPasswordRequest{
		Email: "test@example.com",
	}
	body, _ := json.Marshal(reqBody)

	s.mockAuthService.On("ForgotPassword", mock.AnythingOfType("*auth_service.ForgotPasswordRequest")).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/forgot-password", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockAuthService.AssertExpectations(s.T())
}

func (s *AuthHandlerTestSuite) TestResetPassword_Success() {
	reqBody := auth_service.ResetPasswordRequest{
		Token:    "reset-token",
		Password: "newpassword123",
	}
	body, _ := json.Marshal(reqBody)

	s.mockAuthService.On("ResetPassword", mock.AnythingOfType("*auth_service.ResetPasswordRequest")).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockAuthService.AssertExpectations(s.T())
}

func (s *AuthHandlerTestSuite) TestVerifyEmail_Success() {
	reqBody := auth_service.VerifyEmailRequest{
		Token: "verification-token",
	}
	body, _ := json.Marshal(reqBody)

	s.mockAuthService.On("VerifyEmail", mock.AnythingOfType("*auth_service.VerifyEmailRequest")).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/verify-email", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockAuthService.AssertExpectations(s.T())
}

func (s *AuthHandlerTestSuite) TestAuthenticateWithGoogle_Success() {
	reqBody := auth_service.AuthenticateWithGoogleRequest{
		Code:         "google-code",
		TermsVersion: "1.0.0",
	}
	body, _ := json.Marshal(reqBody)

	expectedResp := &auth_service.AuthenticateResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	s.mockAuthService.On("AuthenticateWithGoogle", mock.AnythingOfType("*auth_service.AuthenticateWithGoogleRequest")).
		Return(expectedResp, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/oauth/google", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.mockAuthService.AssertExpectations(s.T())
}
