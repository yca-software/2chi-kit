package admin_user_handler_test

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
	auth_service "github.com/yca-software/2chi-kit/go-api/internals/services/auth"
	user_service "github.com/yca-software/2chi-kit/go-api/internals/services/user"
	admin_user_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/admin/user"
)

type AdminUserHandlerTestSuite struct {
	suite.Suite
	handler         *admin_user_handler.Handler
	echo            *echo.Echo
	mockUserService *user_service.MockService
	mockAuthService *auth_service.MockService
	services        *services.Services
	accessInfo      *models.AccessInfo
	userID          string
}

func TestAdminUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AdminUserHandlerTestSuite))
}

func (s *AdminUserHandlerTestSuite) SetupTest() {
	s.echo = echo.New()
	s.mockUserService = user_service.NewMockService()
	s.mockAuthService = auth_service.NewMockService()
	s.services = &services.Services{
		User: s.mockUserService,
		Auth: s.mockAuthService,
	}
	s.handler = admin_user_handler.New(s.services)
	group := s.echo.Group("/api/admin")
	s.handler.RegisterEndpoints(group)

	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			Email:   "admin@example.com",
			IsAdmin: true,
		},
	}
	s.userID = uuid.New().String()
}

func (s *AdminUserHandlerTestSuite) setAccessInfo(c echo.Context) {
	c.Set("accessInfo", s.accessInfo)
}

func (s *AdminUserHandlerTestSuite) TestListUsers_Success() {
	expectedUsers := &user_service.PaginatedListResponse{}

	s.mockUserService.On("List", mock.AnythingOfType("*user_service.ListRequest"), s.accessInfo).
		Return(expectedUsers, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/user", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.ListUsers(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockUserService.AssertExpectations(s.T())
}

func (s *AdminUserHandlerTestSuite) TestGetUser_Success() {
	expectedUser := &user_service.GetResponse{
		User: &models.User{
			ID: uuid.MustParse(s.userID),
		},
	}

	s.mockUserService.On("Get", mock.AnythingOfType("*user_service.GetRequest"), s.accessInfo).
		Return(expectedUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/user/"+s.userID, nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("userId")
	c.SetParamValues(s.userID)
	s.handler.GetUser(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockUserService.AssertExpectations(s.T())
}

func (s *AdminUserHandlerTestSuite) TestImpersonateUser_Success() {
	expectedResp := &auth_service.AuthenticateResponse{
		AccessToken:  "impersonated-token",
		RefreshToken: "refresh-token",
	}

	s.mockAuthService.On("Impersonate", mock.AnythingOfType("*auth_service.ImpersonateRequest"), s.accessInfo).
		Return(expectedResp, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/user/"+s.userID+"/impersonate", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("userId")
	c.SetParamValues(s.userID)
	s.handler.ImpersonateUser(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockAuthService.AssertExpectations(s.T())
}

func (s *AdminUserHandlerTestSuite) TestDeleteUser_Success() {
	s.mockUserService.On("Delete", mock.AnythingOfType("*user_service.DeleteRequest"), s.accessInfo).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/user/"+s.userID, nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("userId")
	c.SetParamValues(s.userID)
	s.handler.DeleteUser(c)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockUserService.AssertExpectations(s.T())
}
