package organization_member_handler_test

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
	organization_member_service "github.com/yca-software/2chi-kit/go-api/internals/services/organization_member"
	organization_member_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/organization/member"
)

type OrganizationMemberHandlerTestSuite struct {
	suite.Suite
	handler                      *organization_member_handler.Handler
	echo                         *echo.Echo
	mockOrganizationMemberService *organization_member_service.MockService
	services                     *services.Services
	accessInfo                   *models.AccessInfo
	orgID                        string
	memberID                     string
}

func TestOrganizationMemberHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationMemberHandlerTestSuite))
}

func (s *OrganizationMemberHandlerTestSuite) SetupTest() {
	s.echo = echo.New()
	s.mockOrganizationMemberService = organization_member_service.NewMockService()
	s.services = &services.Services{
		OrganizationMember: s.mockOrganizationMemberService,
	}
	s.handler = organization_member_handler.New(s.services)
	group := s.echo.Group("/api/organization/:orgId")
	s.handler.RegisterEndpoints(group)

	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "test@example.com",
		},
	}
	s.orgID = uuid.New().String()
	s.memberID = uuid.New().String()
}

func (s *OrganizationMemberHandlerTestSuite) setAccessInfo(c echo.Context) {
	c.Set("accessInfo", s.accessInfo)
	c.SetParamNames("orgId")
	c.SetParamValues(s.orgID)
}

func (s *OrganizationMemberHandlerTestSuite) TestListOrganizationMembers_Success() {
	expectedMembers := &[]models.OrganizationMemberWithUser{}

	s.mockOrganizationMemberService.On("ListByOrganization", mock.AnythingOfType("*organization_member_service.ListByOrganizationRequest"), s.accessInfo).
		Return(expectedMembers, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/organization/"+s.orgID+"/member", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	s.handler.ListOrganizationMembers(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockOrganizationMemberService.AssertExpectations(s.T())
}

func (s *OrganizationMemberHandlerTestSuite) TestUpdateOrganizationMemberRole_Success() {
	reqBody := organization_member_service.UpdateRequest{
		RoleID: uuid.New().String(),
	}
	body, _ := json.Marshal(reqBody)

	expectedMember := &models.OrganizationMemberWithUser{}

	s.mockOrganizationMemberService.On("Update", mock.AnythingOfType("*organization_member_service.UpdateRequest"), s.accessInfo).
		Return(expectedMember, nil)

	req := httptest.NewRequest(http.MethodPatch, "/api/organization/"+s.orgID+"/member/"+s.memberID+"/role", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId", "memberId")
	c.SetParamValues(s.orgID, s.memberID)
	s.handler.UpdateOrganizationMemberRole(c)

	s.Equal(http.StatusOK, rec.Code)
	s.mockOrganizationMemberService.AssertExpectations(s.T())
}

func (s *OrganizationMemberHandlerTestSuite) TestRemoveOrganizationMember_Success() {
	s.mockOrganizationMemberService.On("Remove", mock.AnythingOfType("*organization_member_service.RemoveRequest"), s.accessInfo).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/organization/"+s.orgID+"/member/"+s.memberID, nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)
	c.SetParamNames("orgId", "memberId")
	c.SetParamValues(s.orgID, s.memberID)
	s.handler.RemoveOrganizationMember(c)

	s.Equal(http.StatusNoContent, rec.Code)
	s.mockOrganizationMemberService.AssertExpectations(s.T())
}
