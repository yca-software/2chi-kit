package user_refresh_token_handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	user_refresh_token_service "github.com/yca-software/2chi-kit/go-api/internals/services/user_refresh_token"
	user_refresh_token_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/user/refresh_token"
)

type UserRefreshTokenHandlerTestSuite struct {
	suite.Suite
	handler                    *user_refresh_token_handler.Handler
	echo                       *echo.Echo
	mockUserRefreshTokenService *user_refresh_token_service.MockService
	services                   *services.Services
	accessInfo                 *models.AccessInfo
}

func TestUserRefreshTokenHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserRefreshTokenHandlerTestSuite))
}

func (s *UserRefreshTokenHandlerTestSuite) SetupTest() {
	s.echo = echo.New()
	s.mockUserRefreshTokenService = user_refresh_token_service.NewMockService()
	s.services = &services.Services{
		UserRefreshToken: s.mockUserRefreshTokenService,
	}
	s.handler = user_refresh_token_handler.NewHandler(s.services)
	group := s.echo.Group("/api/user")
	s.handler.RegisterRoutes(group)

	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "test@example.com",
		},
	}
}

func (s *UserRefreshTokenHandlerTestSuite) setAccessInfo(c echo.Context) {
	c.Set("accessInfo", s.accessInfo)
}

func (s *UserRefreshTokenHandlerTestSuite) TestListActiveRefreshTokens_Success() {
	expectedTokens := &[]models.UserRefreshToken{}

	s.mockUserRefreshTokenService.On("ListActive", mock.AnythingOfType("*user_refresh_token_service.ListActiveRequest"), s.accessInfo).
		Return(expectedTokens, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/token", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.ListActiveRefreshTokens(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockUserRefreshTokenService.AssertExpectations(s.T())
}

func (s *UserRefreshTokenHandlerTestSuite) TestRevokeAllRefreshTokens_Success() {
	s.mockUserRefreshTokenService.On("RevokeAll", mock.AnythingOfType("*user_refresh_token_service.RevokeAllRequest"), s.accessInfo).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/token", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.RevokeAllRefreshTokens(c)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockUserRefreshTokenService.AssertExpectations(s.T())
}

func (s *UserRefreshTokenHandlerTestSuite) TestRevokeRefreshToken_Success() {
	tokenID := uuid.New().String()
	s.mockUserRefreshTokenService.On("Revoke", mock.AnythingOfType("*user_refresh_token_service.RevokeRequest"), s.accessInfo).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/token/"+tokenID, nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("tokenId")
	c.SetParamValues(tokenID)
	s.handler.RevokeRefreshToken(c)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockUserRefreshTokenService.AssertExpectations(s.T())
}
