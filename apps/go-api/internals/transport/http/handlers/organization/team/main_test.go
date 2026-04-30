package organization_team_handler_test

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
	team_service "github.com/yca-software/2chi-kit/go-api/internals/services/team"
	organization_team_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/organization/team"
)

type OrganizationTeamHandlerTestSuite struct {
	suite.Suite
	handler         *organization_team_handler.Handler
	echo            *echo.Echo
	mockTeamService *team_service.MockService
	services        *services.Services
	accessInfo      *models.AccessInfo
	orgID           string
	teamID          string
}

func TestOrganizationTeamHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationTeamHandlerTestSuite))
}

func (s *OrganizationTeamHandlerTestSuite) SetupTest() {
	s.echo = echo.New()
	s.mockTeamService = team_service.NewMockService()
	s.services = &services.Services{
		Team: s.mockTeamService,
	}
	s.handler = organization_team_handler.New(s.services)
	group := s.echo.Group("/api/organization/:orgId")
	s.handler.RegisterEndpoints(group)

	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "test@example.com",
		},
	}
	s.orgID = uuid.New().String()
	s.teamID = uuid.New().String()
}

func (s *OrganizationTeamHandlerTestSuite) setAccessInfo(c echo.Context) {
	c.Set("accessInfo", s.accessInfo)
	c.SetParamNames("orgId")
	c.SetParamValues(s.orgID)
}

func (s *OrganizationTeamHandlerTestSuite) TestCreateTeam_Success() {
	reqBody := team_service.CreateRequest{
		Name: "Test Team",
	}
	body, _ := json.Marshal(reqBody)

	expectedTeam := &models.Team{
		ID:             uuid.MustParse(s.teamID),
		OrganizationID: uuid.MustParse(s.orgID),
		Name:           "Test Team",
	}

	s.mockTeamService.On("Create", mock.AnythingOfType("*team_service.CreateRequest"), s.accessInfo).
		Return(expectedTeam, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/organization/"+s.orgID+"/team", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.CreateTeam(c)

	s.Equal(http.StatusCreated, rec.Code)
	s.mockTeamService.AssertExpectations(s.T())
}

func (s *OrganizationTeamHandlerTestSuite) TestListTeams_Success() {
	expectedTeams := &[]models.Team{}

	s.mockTeamService.On("List", mock.AnythingOfType("*team_service.ListRequest"), s.accessInfo).
		Return(expectedTeams, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/organization/"+s.orgID+"/team", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.ListTeams(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockTeamService.AssertExpectations(s.T())
}

func (s *OrganizationTeamHandlerTestSuite) TestUpdateTeam_Success() {
	reqBody := team_service.UpdateRequest{
		Name: "Updated Team",
	}
	body, _ := json.Marshal(reqBody)

	expectedTeam := &models.Team{
		ID:             uuid.MustParse(s.teamID),
		OrganizationID: uuid.MustParse(s.orgID),
		Name:           "Updated Team",
	}

	s.mockTeamService.On("Update", mock.AnythingOfType("*team_service.UpdateRequest"), s.accessInfo).
		Return(expectedTeam, nil)

	req := httptest.NewRequest(http.MethodPatch, "/api/organization/"+s.orgID+"/team/"+s.teamID, bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId", "teamId")
	c.SetParamValues(s.orgID, s.teamID)
	s.handler.UpdateTeam(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockTeamService.AssertExpectations(s.T())
}

func (s *OrganizationTeamHandlerTestSuite) TestDeleteTeam_Success() {
	s.mockTeamService.On("Delete", mock.AnythingOfType("*team_service.DeleteRequest"), s.accessInfo).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/organization/"+s.orgID+"/team/"+s.teamID, nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId", "teamId")
	c.SetParamValues(s.orgID, s.teamID)
	s.handler.DeleteTeam(c)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockTeamService.AssertExpectations(s.T())
}
