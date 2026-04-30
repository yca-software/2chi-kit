package organization_invitation_handler_test

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
	invitation_service "github.com/yca-software/2chi-kit/go-api/internals/services/invitation"
	organization_invitation_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/organization/invitation"
	http_middlewares "github.com/yca-software/2chi-kit/go-api/internals/transport/http/middlewares"
)

type OrganizationInvitationHandlerTestSuite struct {
	suite.Suite
	handler                *organization_invitation_handler.Handler
	echo                   *echo.Echo
	mockInvitationService  *invitation_service.MockService
	services               *services.Services
	accessInfo             *models.AccessInfo
	orgID                   string
	invitationID            string
}

func TestOrganizationInvitationHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationInvitationHandlerTestSuite))
}

func (s *OrganizationInvitationHandlerTestSuite) SetupTest() {
	s.echo = echo.New()
	s.mockInvitationService = invitation_service.NewMockService()
	s.services = &services.Services{
		Invitation: s.mockInvitationService,
	}
	s.handler = organization_invitation_handler.New(s.services, &http_middlewares.RateLimitConfig{})
	group := s.echo.Group("/api/organization/:orgId")
	s.handler.RegisterEndpoints(group)

	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "test@example.com",
		},
	}
	s.orgID = uuid.New().String()
	s.invitationID = uuid.New().String()
}

func (s *OrganizationInvitationHandlerTestSuite) setAccessInfo(c echo.Context) {
	c.Set("accessInfo", s.accessInfo)
	c.SetParamNames("orgId")
	c.SetParamValues(s.orgID)
}

func (s *OrganizationInvitationHandlerTestSuite) TestCreateInvitation_Success() {
	reqBody := invitation_service.CreateRequest{
		Email: "invitee@example.com",
	}
	body, _ := json.Marshal(reqBody)

	expectedInvitation := &invitation_service.CreateResponse{
		Invitation: &models.Invitation{},
	}

	s.mockInvitationService.On("Create", mock.AnythingOfType("*invitation_service.CreateRequest"), s.accessInfo).
		Return(expectedInvitation, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/organization/"+s.orgID+"/invitation", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.CreateInvitation(c)

	s.Equal(http.StatusCreated, rec.Code)
	s.mockInvitationService.AssertExpectations(s.T())
}

func (s *OrganizationInvitationHandlerTestSuite) TestListInvitations_Success() {
	expectedInvitations := &[]models.Invitation{}

	s.mockInvitationService.On("List", mock.AnythingOfType("*invitation_service.ListRequest"), s.accessInfo).
		Return(expectedInvitations, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/organization/"+s.orgID+"/invitation", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.ListInvitations(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockInvitationService.AssertExpectations(s.T())
}

func (s *OrganizationInvitationHandlerTestSuite) TestRevokeInvitation_Success() {
	s.mockInvitationService.On("Revoke", mock.AnythingOfType("*invitation_service.RevokeRequest"), s.accessInfo).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/organization/"+s.orgID+"/invitation/"+s.invitationID, nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId", "invitationId")
	c.SetParamValues(s.orgID, s.invitationID)
	s.handler.RevokeInvitation(c)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockInvitationService.AssertExpectations(s.T())
}
