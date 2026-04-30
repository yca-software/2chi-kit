package support_service_test

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	support_service "github.com/yca-software/2chi-kit/go-api/internals/services/support"
	yca_email "github.com/yca-software/go-common/email"
	yca_error "github.com/yca-software/go-common/error"
	yca_validate "github.com/yca-software/go-common/validator"
)

const testInbox = "support-inbox@example.com"

type SupportServiceTestSuite struct {
	suite.Suite
	svc          support_service.Service
	emailService *yca_email.MockEmailService
}

func TestSupportServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SupportServiceTestSuite))
}

func (s *SupportServiceTestSuite) SetupTest() {
	s.emailService = yca_email.NewMockEmailService()
	s.svc = support_service.New(&support_service.Dependencies{
		SupportInboxEmail: testInbox,
		EmailService:      s.emailService,
		Now:               time.Now,
		Validator:         yca_validate.New(),
	})
}

func accessInfoWithUser(userID uuid.UUID, email string) *models.AccessInfo {
	return &models.AccessInfo{
		IPAddress: "203.0.113.10",
		UserAgent: "suite-test-agent/1.0",
		User: &models.UserAccessInfo{
			UserID: userID,
			Email:  email,
		},
	}
}

func (s *SupportServiceTestSuite) TestSubmit_UnauthorizedWhenNoAccessInfo() {
	err := s.svc.Submit(&support_service.SubmitRequest{
		Subject: "Help",
		Message: "I need help",
	}, nil)
	s.Require().Error(err)
	var appErr *yca_error.Error
	s.Require().ErrorAs(err, &appErr)
	s.Equal(constants.UNAUTHORIZED_CODE, appErr.ErrorCode)
	s.emailService.AssertNotCalled(s.T(), "SendEmail", mock.Anything, mock.Anything, mock.Anything)
}

func (s *SupportServiceTestSuite) TestSubmit_ForbiddenWhenNoUser() {
	accessInfo := &models.AccessInfo{}

	err := s.svc.Submit(&support_service.SubmitRequest{
		Subject: "Help",
		Message: "I need help",
	}, accessInfo)

	s.Require().Error(err)
	var appErr *yca_error.Error
	s.Require().ErrorAs(err, &appErr)
	s.Equal(constants.FORBIDDEN_CODE, appErr.ErrorCode)
	s.emailService.AssertNotCalled(s.T(), "SendEmail", mock.Anything, mock.Anything, mock.Anything)
}

func (s *SupportServiceTestSuite) TestSubmit_Validation_InvalidPageURL() {
	userID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	accessInfo := accessInfoWithUser(userID, "user@example.com")

	err := s.svc.Submit(&support_service.SubmitRequest{
		Subject: "x",
		Message: "y",
		PageURL: "not-a-url",
	}, accessInfo)

	s.Require().Error(err)
	var appErr *yca_error.Error
	s.Require().ErrorAs(err, &appErr)
	s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, appErr.ErrorCode)
	s.emailService.AssertNotCalled(s.T(), "SendEmail", mock.Anything, mock.Anything, mock.Anything)
}

func (s *SupportServiceTestSuite) TestSubmit_InternalServerErrorWhenInboxNotConfigured() {
	s.T().Setenv("SUPPORT_INBOX_EMAIL", "")

	svc := support_service.New(&support_service.Dependencies{
		SupportInboxEmail: "",
		EmailService:      s.emailService,
		Now:               time.Now,
		Validator:         yca_validate.New(),
	})

	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	err := svc.Submit(&support_service.SubmitRequest{
		Subject: "Help",
		Message: "Body",
	}, accessInfoWithUser(userID, "u@example.com"))

	s.Require().Error(err)
	var appErr *yca_error.Error
	s.Require().ErrorAs(err, &appErr)
	s.Equal(constants.INTERNAL_SERVER_ERROR_CODE, appErr.ErrorCode)
	s.emailService.AssertNotCalled(s.T(), "SendEmail", mock.Anything, mock.Anything, mock.Anything)
}

func (s *SupportServiceTestSuite) TestSubmit_Success_SendsToInboxWithPrefixedSubject() {
	userID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	accessInfo := accessInfoWithUser(userID, "reporter@example.com")

	s.emailService.On("SendEmail", testInbox, "[Support] Bug report", mock.MatchedBy(func(body string) bool {
		return strings.Contains(body, "reporter@example.com") &&
			strings.Contains(body, userID.String()) &&
			strings.Contains(body, "Bug report") &&
			strings.Contains(body, "Something broke") &&
			strings.Contains(body, "https://app.example.com/settings") &&
			strings.Contains(body, "Mozilla/5.0") &&
			strings.Contains(body, accessInfo.IPAddress)
	})).Return(nil).Once()

	err := s.svc.Submit(&support_service.SubmitRequest{
		Subject:   "Bug report",
		Message:   "Something broke",
		PageURL:   "https://app.example.com/settings",
		UserAgent: "Mozilla/5.0",
	}, accessInfo)

	s.Require().NoError(err)
	s.emailService.AssertExpectations(s.T())
}

func (s *SupportServiceTestSuite) TestSubmit_Success_DefaultSubjectWhenSubjectBlank() {
	userID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	accessInfo := accessInfoWithUser(userID, "user@example.com")
	expectedSubject := "Support request from user@example.com"

	s.emailService.On("SendEmail", testInbox, expectedSubject, mock.MatchedBy(func(body string) bool {
		return strings.Contains(body, "Message only") && strings.Contains(body, userID.String())
	})).Return(nil).Once()

	err := s.svc.Submit(&support_service.SubmitRequest{
		Subject: "",
		Message: "Message only",
	}, accessInfo)

	s.Require().NoError(err)
	s.emailService.AssertExpectations(s.T())
}

func (s *SupportServiceTestSuite) TestSubmit_SendEmailErrorPropagates() {
	userID := uuid.MustParse("55555555-5555-4555-8555-555555555555")
	sendErr := errors.New("resend unavailable")

	s.emailService.On("SendEmail", mock.Anything, mock.Anything, mock.Anything).Return(sendErr).Once()

	err := s.svc.Submit(&support_service.SubmitRequest{
		Subject: "S",
		Message: "M",
	}, accessInfoWithUser(userID, "e@example.com"))

	s.Require().Error(err)
	s.Require().ErrorIs(err, sendErr)
	s.emailService.AssertExpectations(s.T())
}

func (s *SupportServiceTestSuite) TestSubmit_InboxFromEnvViaDependenciesLikeBootstrap() {
	s.T().Setenv("SUPPORT_INBOX_EMAIL", "from-env@example.com")

	svc := support_service.New(&support_service.Dependencies{
		SupportInboxEmail: strings.TrimSpace(os.Getenv("SUPPORT_INBOX_EMAIL")),
		EmailService:      s.emailService,
		Now:               time.Now,
		Validator:         yca_validate.New(),
	})

	userID := uuid.MustParse("66666666-6666-4666-8666-666666666666")
	s.emailService.On("SendEmail", "from-env@example.com", mock.Anything, mock.Anything).Return(nil).Once()

	err := svc.Submit(&support_service.SubmitRequest{
		Subject: "Env inbox",
		Message: "Hi",
	}, accessInfoWithUser(userID, "u@example.com"))

	s.Require().NoError(err)
	s.emailService.AssertExpectations(s.T())
}
