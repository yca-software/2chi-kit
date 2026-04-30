package organization_role_handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	role_service "github.com/yca-software/2chi-kit/go-api/internals/services/role"
	organization_role_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/organization/role"
)

type OrganizationRoleHandlerTestSuite struct {
	suite.Suite
	handler         *organization_role_handler.Handler
	echo            *echo.Echo
	mockRoleService *role_service.MockService
	services        *services.Services
	accessInfo      *models.AccessInfo
	orgID           string
	roleID          string
}

func TestOrganizationRoleHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationRoleHandlerTestSuite))
}

func (s *OrganizationRoleHandlerTestSuite) SetupTest() {
	s.echo = echo.New()
	s.mockRoleService = role_service.NewMockService()
	s.services = &services.Services{
		Role: s.mockRoleService,
	}
	s.handler = organization_role_handler.New(s.services)
	group := s.echo.Group("/api/organization/:orgId")
	s.handler.RegisterEndpoints(group)

	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "test@example.com",
		},
	}
	s.orgID = uuid.New().String()
	s.roleID = uuid.New().String()
}

func (s *OrganizationRoleHandlerTestSuite) setAccessInfo(c echo.Context) {
	c.Set("accessInfo", s.accessInfo)
	c.SetParamNames("orgId")
	c.SetParamValues(s.orgID)
}

func (s *OrganizationRoleHandlerTestSuite) TestCreateRole_Success() {
	reqBody := role_service.CreateRequest{
		Name: "Test Role",
	}
	body, _ := json.Marshal(reqBody)

	expectedRole := &models.Role{
		ID:             uuid.MustParse(s.roleID),
		OrganizationID: uuid.MustParse(s.orgID),
		Name:           "Test Role",
	}

	s.mockRoleService.On("Create", mock.AnythingOfType("*role_service.CreateRequest"), s.accessInfo).
		Return(expectedRole, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/organization/"+s.orgID+"/role", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.CreateRole(c)

	s.Equal(http.StatusCreated, rec.Code)
	s.mockRoleService.AssertExpectations(s.T())
}

func (s *OrganizationRoleHandlerTestSuite) TestListRoles_Success() {
	expectedRoles := &[]models.Role{}

	s.mockRoleService.On("List", mock.AnythingOfType("*role_service.ListRequest"), s.accessInfo).
		Return(expectedRoles, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/organization/"+s.orgID+"/role", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.ListRoles(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockRoleService.AssertExpectations(s.T())
}

func (s *OrganizationRoleHandlerTestSuite) TestUpdateRole_Success() {
	reqBody := role_service.UpdateRequest{
		Name: "Updated Role",
	}
	body, _ := json.Marshal(reqBody)

	expectedRole := &models.Role{
		ID:             uuid.MustParse(s.roleID),
		OrganizationID: uuid.MustParse(s.orgID),
		Name:           "Updated Role",
	}

	s.mockRoleService.On("Update", mock.AnythingOfType("*role_service.UpdateRequest"), s.accessInfo).
		Return(expectedRole, nil)

	req := httptest.NewRequest(http.MethodPatch, "/api/organization/"+s.orgID+"/role/"+s.roleID, bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId", "roleId")
	c.SetParamValues(s.orgID, s.roleID)
	s.handler.UpdateRole(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockRoleService.AssertExpectations(s.T())
}

func (s *OrganizationRoleHandlerTestSuite) TestDeleteRole_Success() {
	s.mockRoleService.On("Delete", mock.AnythingOfType("*role_service.DeleteRequest"), s.accessInfo).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/organization/"+s.orgID+"/role/"+s.roleID, nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId", "roleId")
	c.SetParamValues(s.orgID, s.roleID)
	s.handler.DeleteRole(c)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockRoleService.AssertExpectations(s.T())
}
