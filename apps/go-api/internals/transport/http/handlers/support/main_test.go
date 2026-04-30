package support_handler_test

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
	support_service "github.com/yca-software/2chi-kit/go-api/internals/services/support"
	support_handler "github.com/yca-software/2chi-kit/go-api/internals/transport/http/handlers/support"
)

type SupportHandlerTestSuite struct {
	suite.Suite
	handler            *support_handler.Handler
	echo               *echo.Echo
	mockSupportService *support_service.MockService
	services           *services.Services
	accessInfo         *models.AccessInfo
}

func TestSupportHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SupportHandlerTestSuite))
}

func (s *SupportHandlerTestSuite) SetupTest() {
	s.echo = echo.New()
	s.mockSupportService = support_service.NewMockService()
	s.services = &services.Services{
		Support: s.mockSupportService,
	}
	s.handler = support_handler.New(s.services, nil, nil)

	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "user@example.com",
		},
	}
}

func (s *SupportHandlerTestSuite) setAccessInfo(c echo.Context) {
	c.Set("accessInfo", s.accessInfo)
}

func (s *SupportHandlerTestSuite) TestSubmit_Success() {
	reqBody := support_service.SubmitRequest{
		Subject: "Bug report",
		Message: "Something is broken",
		PageURL: "https://app.example.com/settings",
	}
	body, _ := json.Marshal(reqBody)

	s.mockSupportService.
		On("Submit", mock.AnythingOfType("*support_service.SubmitRequest"), s.accessInfo).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/support", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)

	err := s.handler.Submit(c)
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, rec.Code)
	s.mockSupportService.AssertExpectations(s.T())
}

func (s *SupportHandlerTestSuite) TestSubmit_EmptySubjectFillsDefault() {
	reqBody := map[string]string{
		"message": "Message only",
	}
	body, _ := json.Marshal(reqBody)

	s.mockSupportService.
		On("Submit", mock.MatchedBy(func(r *support_service.SubmitRequest) bool {
			return r.Subject == "(no subject)" && r.Message == "Message only"
		}), s.accessInfo).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/support", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	s.setAccessInfo(c)

	err := s.handler.Submit(c)
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, rec.Code)
	s.mockSupportService.AssertExpectations(s.T())
}
