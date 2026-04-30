package organization_team_member_handler_test

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
	team_member_service "github.com/yca-software/2chi-kit/go-api/internals/services/team_member"
	organization_team_member_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/organization/team/member"
)

type OrganizationTeamMemberHandlerTestSuite struct {
	suite.Suite
	handler              *organization_team_member_handler.Handler
	echo                 *echo.Echo
	mockTeamMemberService *team_member_service.MockService
	services             *services.Services
	accessInfo           *models.AccessInfo
	orgID                string
	teamID               string
	memberID             string
}

func TestOrganizationTeamMemberHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationTeamMemberHandlerTestSuite))
}

func (s *OrganizationTeamMemberHandlerTestSuite) SetupTest() {
	s.echo = echo.New()
	s.mockTeamMemberService = team_member_service.NewMockService()
	s.services = &services.Services{
		TeamMember: s.mockTeamMemberService,
	}
	s.handler = organization_team_member_handler.New(s.services)
	group := s.echo.Group("/api/organization/:orgId/team/:teamId")
	s.handler.RegisterEndpoints(group)

	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "test@example.com",
		},
	}
	s.orgID = uuid.New().String()
	s.teamID = uuid.New().String()
	s.memberID = uuid.New().String()
}

func (s *OrganizationTeamMemberHandlerTestSuite) setAccessInfo(c echo.Context) {
	c.Set("accessInfo", s.accessInfo)
	c.SetParamNames("orgId", "teamId")
	c.SetParamValues(s.orgID, s.teamID)
}

func (s *OrganizationTeamMemberHandlerTestSuite) TestAddTeamMember_Success() {
	reqBody := team_member_service.AddRequest{
		UserID: uuid.New().String(),
	}
	body, _ := json.Marshal(reqBody)

	expectedMember := &models.TeamMemberWithUser{}

	s.mockTeamMemberService.On("Add", mock.AnythingOfType("*team_member_service.AddRequest"), s.accessInfo).
		Return(expectedMember, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/organization/"+s.orgID+"/team/"+s.teamID+"/member", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.AddTeamMember(c)

	s.Equal(http.StatusCreated, rec.Code)
	s.mockTeamMemberService.AssertExpectations(s.T())
}

func (s *OrganizationTeamMemberHandlerTestSuite) TestListTeamMembers_Success() {
	expectedMembers := &[]models.TeamMemberWithUser{}

	s.mockTeamMemberService.On("ListByTeam", mock.AnythingOfType("*team_member_service.ListByTeamRequest"), s.accessInfo).
		Return(expectedMembers, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/organization/"+s.orgID+"/team/"+s.teamID+"/member", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.ListTeamMembers(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockTeamMemberService.AssertExpectations(s.T())
}

func (s *OrganizationTeamMemberHandlerTestSuite) TestRemoveTeamMember_Success() {
	s.mockTeamMemberService.On("Remove", mock.AnythingOfType("*team_member_service.RemoveRequest"), s.accessInfo).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/organization/"+s.orgID+"/team/"+s.teamID+"/member/"+s.memberID, nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId", "teamId", "memberId")
	c.SetParamValues(s.orgID, s.teamID, s.memberID)
	s.handler.RemoveTeamMember(c)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockTeamMemberService.AssertExpectations(s.T())
}
