package admin_organization_handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	organization_service "github.com/yca-software/2chi-kit/go-api/internals/services/organization"
	admin_organization_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/admin/organization"
)

type AdminOrganizationHandlerTestSuite struct {
	suite.Suite
	handler                  *admin_organization_handler.Handler
	echo                     *echo.Echo
	mockOrganizationService  *organization_service.MockService
	mockAuditLogService      *audit_log_service.MockService
	services                 *services.Services
	accessInfo               *models.AccessInfo
	orgID                    string
}

func TestAdminOrganizationHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AdminOrganizationHandlerTestSuite))
}

func (s *AdminOrganizationHandlerTestSuite) SetupTest() {
	s.echo = echo.New()
	s.mockOrganizationService = organization_service.NewMockService()
	s.mockAuditLogService = audit_log_service.NewMockService()
	s.services = &services.Services{
		Organization: s.mockOrganizationService,
		AuditLog:     s.mockAuditLogService,
	}
	s.handler = admin_organization_handler.New(s.services)
	group := s.echo.Group("/api/admin")
	s.handler.RegisterEndpoints(group)

	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			Email:   "admin@example.com",
			IsAdmin: true,
		},
	}
	s.orgID = uuid.New().String()
}

func (s *AdminOrganizationHandlerTestSuite) setAccessInfo(c echo.Context) {
	c.Set("accessInfo", s.accessInfo)
}

func (s *AdminOrganizationHandlerTestSuite) TestListOrganizations_Success() {
	expectedOrgs := &organization_service.PaginatedListResponse{}

	s.mockOrganizationService.On("List", mock.AnythingOfType("*organization_service.ListRequest"), s.accessInfo).
		Return(expectedOrgs, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/organization", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.ListOrganizations(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockOrganizationService.AssertExpectations(s.T())
}

func (s *AdminOrganizationHandlerTestSuite) TestGetOrganization_Success() {
	expectedOrg := &models.Organization{
		ID: uuid.MustParse(s.orgID),
	}

	s.mockOrganizationService.On("Get", mock.AnythingOfType("*organization_service.GetRequest"), s.accessInfo).
		Return(expectedOrg, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/organization/"+s.orgID, nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId")
	c.SetParamValues(s.orgID)
	s.handler.GetOrganization(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockOrganizationService.AssertExpectations(s.T())
}

func (s *AdminOrganizationHandlerTestSuite) TestListOrganizationAuditLogs_Success() {
	expectedLogs := &audit_log_service.ListForOrganizationResponse{}

	s.mockAuditLogService.On("ListForOrganization", mock.AnythingOfType("*audit_log_service.ListForOrganizationRequest"), s.accessInfo).
		Return(expectedLogs, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/organization/"+s.orgID+"/audit-log", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId")
	c.SetParamValues(s.orgID)
	s.handler.ListOrganizationAuditLogs(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockAuditLogService.AssertExpectations(s.T())
}

func (s *AdminOrganizationHandlerTestSuite) TestUpdateOrganizationSubscriptionSettings_Success() {
	expectedOrg := &models.Organization{ID: uuid.MustParse(s.orgID)}
	body := []byte(`{"subscriptionType":3,"subscriptionSeats":10}`)

	s.mockOrganizationService.On("AdminUpdateSubscriptionSettings", mock.AnythingOfType("*organization_service.AdminUpdateSubscriptionSettingsRequest"), s.accessInfo).
		Return(expectedOrg, nil)

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/organization/"+s.orgID+"/subscription", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId")
	c.SetParamValues(s.orgID)
	s.handler.UpdateOrganizationSubscriptionSettings(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockOrganizationService.AssertExpectations(s.T())
}

func (s *AdminOrganizationHandlerTestSuite) TestListArchivedOrganizations_Success() {
	expectedOrgs := &organization_service.PaginatedListResponse{}

	s.mockOrganizationService.On("ListArchived", mock.AnythingOfType("*organization_service.ListRequest"), s.accessInfo).
		Return(expectedOrgs, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/organization/archived", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.ListArchivedOrganizations(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockOrganizationService.AssertExpectations(s.T())
}

func (s *AdminOrganizationHandlerTestSuite) TestGetArchivedOrganization_Success() {
	expectedOrg := &models.Organization{ID: uuid.MustParse(s.orgID)}

	s.mockOrganizationService.On("GetArchived", mock.AnythingOfType("*organization_service.GetRequest"), s.accessInfo).
		Return(expectedOrg, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/organization/archived/"+s.orgID, nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId")
	c.SetParamValues(s.orgID)
	s.handler.GetArchivedOrganization(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockOrganizationService.AssertExpectations(s.T())
}

func (s *AdminOrganizationHandlerTestSuite) TestRestoreOrganization_Success() {
	s.mockOrganizationService.On("Restore", mock.AnythingOfType("*organization_service.RestoreRequest"), s.accessInfo).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/organization/archived/"+s.orgID+"/restore", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId")
	c.SetParamValues(s.orgID)
	s.handler.RestoreOrganization(c)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockOrganizationService.AssertExpectations(s.T())
}
