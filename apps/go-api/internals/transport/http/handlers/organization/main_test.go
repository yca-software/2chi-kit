package organization_handler_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/metrics"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	"github.com/yca-software/2chi-kit/go-api/internals/services"
	organization_service "github.com/yca-software/2chi-kit/go-api/internals/services/organization"
	organization_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/organization"
	http_middlewares "github.com/yca-software/2chi-kit/go-api/internals/transport/http/middlewares"
)

type OrganizationHandlerTestSuite struct {
	suite.Suite
	handler                 *organization_handler.Handler
	echo                    *echo.Echo
	mockOrganizationService *organization_service.MockService
	services                *services.Services
	middlewares             http_middlewares.Middlewares
	accessInfo              *models.AccessInfo
	orgID                   string
}

func TestOrganizationHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationHandlerTestSuite))
}

func (s *OrganizationHandlerTestSuite) SetupTest() {
	s.echo = echo.New()
	s.mockOrganizationService = organization_service.NewMockService()
	s.services = &services.Services{
		Organization: s.mockOrganizationService,
	}

	mwareCfg := &http_middlewares.MiddlewaresConfig{AccessSecret: "test-secret"}
	m, _ := metrics.New("test")
	repos := &repositories.Repositories{}
	hashToken := func(raw string) string {
		h := sha256.Sum256([]byte(raw))
		return hex.EncodeToString(h[:])
	}
	s.middlewares = http_middlewares.NewMiddlewares(mwareCfg, repos, m, hashToken)

	s.handler = organization_handler.New(s.services, s.middlewares, &http_middlewares.RateLimitConfig{})
	s.handler.RegisterEndpoints(s.echo.Group("/api"))

	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "test@example.com",
		},
	}
	s.orgID = uuid.New().String()
}

func (s *OrganizationHandlerTestSuite) setAccessInfo(c echo.Context) {
	c.Set("accessInfo", s.accessInfo)
}

func (s *OrganizationHandlerTestSuite) TestCreateOrganization_Success() {
	reqBody := organization_service.CreateRequest{
		Name: "Test Organization",
	}
	body, _ := json.Marshal(reqBody)

	expectedOrg := &organization_service.CreateResponse{
		Organization: &models.Organization{
			ID:   uuid.New(),
			Name: "Test Organization",
		},
	}

	s.mockOrganizationService.On("Create", mock.AnythingOfType("*organization_service.CreateRequest"), s.accessInfo).
		Return(expectedOrg, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/organization", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.CreateOrganization(c)

	s.Equal(http.StatusCreated, rec.Code)
	s.mockOrganizationService.AssertExpectations(s.T())
}

func (s *OrganizationHandlerTestSuite) TestGetOrganization_Success() {
	expectedOrg := &models.Organization{
		ID:   uuid.MustParse(s.orgID),
		Name: "Test Organization",
	}

	s.mockOrganizationService.On("Get", mock.AnythingOfType("*organization_service.GetRequest"), s.accessInfo).
		Return(expectedOrg, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/organization/"+s.orgID, nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId")
	c.SetParamValues(s.orgID)
	s.handler.GetOrganization(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockOrganizationService.AssertExpectations(s.T())
}

func (s *OrganizationHandlerTestSuite) TestUpdateOrganization_Success() {
	reqBody := organization_service.UpdateRequest{
		Name: "Updated Organization",
	}
	body, _ := json.Marshal(reqBody)

	expectedOrg := &models.Organization{
		ID:   uuid.MustParse(s.orgID),
		Name: "Updated Organization",
	}

	s.mockOrganizationService.On("Update", mock.AnythingOfType("*organization_service.UpdateRequest"), s.accessInfo).
		Return(expectedOrg, nil)

	req := httptest.NewRequest(http.MethodPatch, "/api/organization/"+s.orgID, bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId")
	c.SetParamValues(s.orgID)
	s.handler.UpdateOrganization(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockOrganizationService.AssertExpectations(s.T())
}

func (s *OrganizationHandlerTestSuite) TestDeleteOrganization_Success() {
	s.mockOrganizationService.On("Delete", mock.AnythingOfType("*organization_service.DeleteRequest"), s.accessInfo).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/organization/"+s.orgID, nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId")
	c.SetParamValues(s.orgID)
	s.handler.DeleteOrganization(c)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockOrganizationService.AssertExpectations(s.T())
}
