package organization_api_key_handler_test

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
	api_key_service "github.com/yca-software/2chi-kit/go-api/internals/services/api_key"
	organization_api_key_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/organization/api_key"
)

type OrganizationApiKeyHandlerTestSuite struct {
	suite.Suite
	handler           *organization_api_key_handler.Handler
	echo              *echo.Echo
	mockApiKeyService *api_key_service.MockService
	services          *services.Services
	accessInfo        *models.AccessInfo
	orgID             string
	apiKeyID          string
}

func TestOrganizationApiKeyHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationApiKeyHandlerTestSuite))
}

func (s *OrganizationApiKeyHandlerTestSuite) SetupTest() {
	s.echo = echo.New()
	s.mockApiKeyService = api_key_service.NewMockService()
	s.services = &services.Services{
		ApiKey: s.mockApiKeyService,
	}
	s.handler = organization_api_key_handler.New(s.services)
	group := s.echo.Group("/api/organization/:orgId")
	s.handler.RegisterEndpoints(group)

	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "test@example.com",
		},
	}
	s.orgID = uuid.New().String()
	s.apiKeyID = uuid.New().String()
}

func (s *OrganizationApiKeyHandlerTestSuite) setAccessInfo(c echo.Context) {
	c.Set("accessInfo", s.accessInfo)
	c.SetParamNames("orgId")
	c.SetParamValues(s.orgID)
}

func (s *OrganizationApiKeyHandlerTestSuite) TestCreateApiKey_Success() {
	reqBody := api_key_service.CreateRequest{
		Name: "Test API Key",
	}
	body, _ := json.Marshal(reqBody)

	expectedApiKey := &api_key_service.CreateResponse{
		ApiKey: &models.APIKey{},
		Secret: "secret-key",
	}

	s.mockApiKeyService.On("Create", mock.AnythingOfType("*api_key_service.CreateRequest"), s.accessInfo).
		Return(expectedApiKey, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/organization/"+s.orgID+"/api-key", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.CreateApiKey(c)

	s.Equal(http.StatusCreated, rec.Code)
	s.mockApiKeyService.AssertExpectations(s.T())
}

func (s *OrganizationApiKeyHandlerTestSuite) TestListApiKeys_Success() {
	expectedApiKeys := &[]models.APIKey{}

	s.mockApiKeyService.On("List", mock.AnythingOfType("*api_key_service.ListRequest"), s.accessInfo).
		Return(expectedApiKeys, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/organization/"+s.orgID+"/api-key", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.ListApiKeys(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockApiKeyService.AssertExpectations(s.T())
}

func (s *OrganizationApiKeyHandlerTestSuite) TestUpdateApiKey_Success() {
	reqBody := api_key_service.UpdateRequest{
		Name:        "Updated API Key",
		Permissions: models.RolePermissions{"org:read", "members:read"},
	}
	body, _ := json.Marshal(reqBody)

	expectedApiKey := &models.APIKey{
		ID:             uuid.MustParse(s.apiKeyID),
		OrganizationID: uuid.MustParse(s.orgID),
		Name:           "Updated API Key",
		Permissions:    models.RolePermissions{"org:read", "members:read"},
	}

	s.mockApiKeyService.On("Update", mock.AnythingOfType("*api_key_service.UpdateRequest"), s.accessInfo).
		Return(expectedApiKey, nil)

	req := httptest.NewRequest(http.MethodPatch, "/api/organization/"+s.orgID+"/api-key/"+s.apiKeyID, bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId", "apiKeyId")
	c.SetParamValues(s.orgID, s.apiKeyID)
	s.handler.UpdateApiKey(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockApiKeyService.AssertExpectations(s.T())
}

func (s *OrganizationApiKeyHandlerTestSuite) TestDeleteApiKey_Success() {
	s.mockApiKeyService.On("Delete", mock.AnythingOfType("*api_key_service.DeleteRequest"), s.accessInfo).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/organization/"+s.orgID+"/api-key/"+s.apiKeyID, nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId", "apiKeyId")
	c.SetParamValues(s.orgID, s.apiKeyID)
	s.handler.DeleteApiKey(c)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockApiKeyService.AssertExpectations(s.T())
}
