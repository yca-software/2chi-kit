package audit_log_service_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	audit_log_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/audit_log"
	organization_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type AuditLogServiceTestSuite struct {
	suite.Suite
	svc          audit_log_service.Service
	repos        *repositories.Repositories
	orgRepo      *organization_repository.MockRepository
	auditLogRepo *audit_log_repository.MockRepository
	logger       *yca_log.MockLogger
	now          time.Time
	accessInfo   *models.AccessInfo // admin so auth is not tested here
}

func TestAuditLogServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuditLogServiceTestSuite))
}

func (s *AuditLogServiceTestSuite) SetupTest() {
	s.now = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	s.orgRepo = organization_repository.NewMock()
	s.auditLogRepo = audit_log_repository.NewMock()
	s.logger = &yca_log.MockLogger{}

	s.repos = &repositories.Repositories{
		Organization: s.orgRepo,
		AuditLog:     s.auditLogRepo,
	}

	// AccessInfo with admin so authorization passes; we only test validations and business logic.
	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			Email:   "test@example.com",
			IsAdmin: true,
		},
	}

	s.svc = audit_log_service.New(&audit_log_service.Dependencies{
		Validator:  yca_validate.New(),
		Repos:      s.repos,
		Logger:     s.logger,
		GenerateID: uuid.NewV7,
		Now:        func() time.Time { return s.now },
		Authorizer: helpers.NewAuthorizer(func() time.Time { return s.now }),
	})
}

// --- Create: validations ---

func (s *AuditLogServiceTestSuite) TestCreate_Validation_MissingOrganizationID() {
	data := json.RawMessage(`{"key":"value"}`)
	req := &audit_log_service.CreateRequest{
		OrganizationID: "",
		Action:         constants.AUDIT_ACTION_TYPE_CREATE,
		ResourceType:   constants.RESOURCE_TYPE_API_KEY,
		ResourceID:     uuid.New().String(),
		Data:           &data,
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuditLogServiceTestSuite) TestCreate_Validation_InvalidOrganizationID() {
	data := json.RawMessage(`{"key":"value"}`)
	req := &audit_log_service.CreateRequest{
		OrganizationID: "not-a-uuid",
		Action:         constants.AUDIT_ACTION_TYPE_CREATE,
		ResourceType:   constants.RESOURCE_TYPE_API_KEY,
		ResourceID:     uuid.New().String(),
		Data:           &data,
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuditLogServiceTestSuite) TestCreate_Validation_MissingAction() {
	data := json.RawMessage(`{"key":"value"}`)
	req := &audit_log_service.CreateRequest{
		OrganizationID: uuid.New().String(),
		Action:         "",
		ResourceType:   constants.RESOURCE_TYPE_API_KEY,
		ResourceID:     uuid.New().String(),
		Data:           &data,
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuditLogServiceTestSuite) TestCreate_Validation_MissingResourceType() {
	data := json.RawMessage(`{"key":"value"}`)
	req := &audit_log_service.CreateRequest{
		OrganizationID: uuid.New().String(),
		Action:         constants.AUDIT_ACTION_TYPE_CREATE,
		ResourceType:   "",
		ResourceID:     uuid.New().String(),
		Data:           &data,
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuditLogServiceTestSuite) TestCreate_Validation_MissingResourceID() {
	data := json.RawMessage(`{"key":"value"}`)
	req := &audit_log_service.CreateRequest{
		OrganizationID: uuid.New().String(),
		Action:         constants.AUDIT_ACTION_TYPE_CREATE,
		ResourceType:   constants.RESOURCE_TYPE_API_KEY,
		ResourceID:     "",
		Data:           &data,
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuditLogServiceTestSuite) TestCreate_Validation_InvalidResourceID() {
	data := json.RawMessage(`{"key":"value"}`)
	req := &audit_log_service.CreateRequest{
		OrganizationID: uuid.New().String(),
		Action:         constants.AUDIT_ACTION_TYPE_CREATE,
		ResourceType:   constants.RESOURCE_TYPE_API_KEY,
		ResourceID:     "not-a-uuid",
		Data:           &data,
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuditLogServiceTestSuite) TestCreate_Validation_MissingData() {
	req := &audit_log_service.CreateRequest{
		OrganizationID: uuid.New().String(),
		Action:         constants.AUDIT_ACTION_TYPE_CREATE,
		ResourceType:   constants.RESOURCE_TYPE_API_KEY,
		ResourceID:     uuid.New().String(),
		Data:           nil,
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Create: business logic ---

func (s *AuditLogServiceTestSuite) TestCreate_GenerateIDError() {
	data := json.RawMessage(`{"key":"value"}`)
	svc := audit_log_service.New(&audit_log_service.Dependencies{
		Validator:  yca_validate.New(),
		Repos:      s.repos,
		Logger:     s.logger,
		GenerateID: func() (uuid.UUID, error) { return uuid.Nil, errors.New("id generation failed") },
		Now:        func() time.Time { return s.now },
		Authorizer: helpers.NewAuthorizer(func() time.Time { return s.now }),
	})

	req := &audit_log_service.CreateRequest{
		OrganizationID: uuid.New().String(),
		Action:         constants.AUDIT_ACTION_TYPE_CREATE,
		ResourceType:   constants.RESOURCE_TYPE_API_KEY,
		ResourceID:     uuid.New().String(),
		Data:           &data,
	}
	resp, err := svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
}

func (s *AuditLogServiceTestSuite) TestCreate_Success_WithUserActor() {
	orgID := uuid.New()
	resourceID := uuid.New()
	userID := uuid.New()
	data := json.RawMessage(`{"key":"value"}`)
	s.auditLogRepo.On("Create", nil, mock.AnythingOfType("*models.AuditLog")).Return(nil)

	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  userID,
			Email:   "user@example.com",
			IsAdmin: false,
		},
	}

	req := &audit_log_service.CreateRequest{
		OrganizationID: orgID.String(),
		Action:         constants.AUDIT_ACTION_TYPE_CREATE,
		ResourceType:   constants.RESOURCE_TYPE_API_KEY,
		ResourceID:     resourceID.String(),
		Data:           &data,
	}
	resp, err := s.svc.Create(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal(orgID, resp.OrganizationID)
	s.Equal(userID, resp.ActorID)
	s.Equal("user@example.com", resp.ActorInfo)
	s.Equal(constants.AUDIT_ACTION_TYPE_CREATE, resp.Action)
	s.Equal(constants.RESOURCE_TYPE_API_KEY, resp.ResourceType)
	s.Equal(resourceID, resp.ResourceID)
	s.Equal(&data, resp.Data)
	s.Equal(s.now, resp.CreatedAt)
	s.auditLogRepo.AssertExpectations(s.T())
}

func (s *AuditLogServiceTestSuite) TestCreate_Success_WithApiKeyActor() {
	orgID := uuid.New()
	resourceID := uuid.New()
	apiKeyID := uuid.New()
	data := json.RawMessage(`{"key":"value"}`)
	s.auditLogRepo.On("Create", nil, mock.AnythingOfType("*models.AuditLog")).Return(nil)

	accessInfo := &models.AccessInfo{
		ApiKey: &models.APIKey{
			ID:        apiKeyID,
			KeyPrefix: "sk_test123",
		},
	}

	req := &audit_log_service.CreateRequest{
		OrganizationID: orgID.String(),
		Action:         constants.AUDIT_ACTION_TYPE_CREATE,
		ResourceType:   constants.RESOURCE_TYPE_API_KEY,
		ResourceID:     resourceID.String(),
		Data:           &data,
	}
	resp, err := s.svc.Create(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal(orgID, resp.OrganizationID)
	s.Equal(apiKeyID, resp.ActorID)
	s.Equal("sk_test123", resp.ActorInfo)
	s.Equal(constants.AUDIT_ACTION_TYPE_CREATE, resp.Action)
	s.Equal(constants.RESOURCE_TYPE_API_KEY, resp.ResourceType)
	s.Equal(resourceID, resp.ResourceID)
	s.Equal(&data, resp.Data)
	s.auditLogRepo.AssertExpectations(s.T())
}

func (s *AuditLogServiceTestSuite) TestCreate_Success_WithImpersonation() {
	orgID := uuid.New()
	resourceID := uuid.New()
	userID := uuid.New()
	impersonatedByID := uuid.New()
	data := json.RawMessage(`{"key":"value"}`)
	s.auditLogRepo.On("Create", nil, mock.AnythingOfType("*models.AuditLog")).Return(nil)

	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:              userID,
			Email:               "user@example.com",
			ImpersonatedBy:      uuid.NullUUID{UUID: impersonatedByID, Valid: true},
			ImpersonatedByEmail: "admin@example.com",
			IsAdmin:             false,
		},
	}

	req := &audit_log_service.CreateRequest{
		OrganizationID: orgID.String(),
		Action:         constants.AUDIT_ACTION_TYPE_CREATE,
		ResourceType:   constants.RESOURCE_TYPE_API_KEY,
		ResourceID:     resourceID.String(),
		Data:           &data,
	}
	resp, err := s.svc.Create(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal(userID, resp.ActorID)
	s.Equal("user@example.com", resp.ActorInfo)
	s.True(resp.ImpersonatedByID.Valid)
	s.Equal(impersonatedByID, resp.ImpersonatedByID.UUID)
	s.Equal("admin@example.com", resp.ImpersonatedByEmail)
	s.auditLogRepo.AssertExpectations(s.T())
}

func (s *AuditLogServiceTestSuite) TestCreate_RepoCreateFails() {
	orgID := uuid.New()
	resourceID := uuid.New()
	data := json.RawMessage(`{"key":"value"}`)
	s.auditLogRepo.On("Create", nil, mock.AnythingOfType("*models.AuditLog")).Return(errors.New("db error"))

	req := &audit_log_service.CreateRequest{
		OrganizationID: orgID.String(),
		Action:         constants.AUDIT_ACTION_TYPE_CREATE,
		ResourceType:   constants.RESOURCE_TYPE_API_KEY,
		ResourceID:     resourceID.String(),
		Data:           &data,
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.auditLogRepo.AssertExpectations(s.T())
}

// --- ListForOrganization: validations ---

func (s *AuditLogServiceTestSuite) TestListForOrganization_Validation_MissingOrganizationID() {
	req := &audit_log_service.ListForOrganizationRequest{
		OrganizationID: "",
		Limit:          10,
		Offset:         0,
	}
	resp, err := s.svc.ListForOrganization(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuditLogServiceTestSuite) TestListForOrganization_Validation_InvalidOrganizationID() {
	req := &audit_log_service.ListForOrganizationRequest{
		OrganizationID: "not-a-uuid",
		Limit:          10,
		Offset:         0,
	}
	resp, err := s.svc.ListForOrganization(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuditLogServiceTestSuite) TestListForOrganization_Validation_InvalidLimit_TooLow() {
	req := &audit_log_service.ListForOrganizationRequest{
		OrganizationID: uuid.New().String(),
		Limit:          0,
		Offset:         0,
	}
	resp, err := s.svc.ListForOrganization(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuditLogServiceTestSuite) TestListForOrganization_Validation_InvalidLimit_TooHigh() {
	req := &audit_log_service.ListForOrganizationRequest{
		OrganizationID: uuid.New().String(),
		Limit:          101,
		Offset:         0,
	}
	resp, err := s.svc.ListForOrganization(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *AuditLogServiceTestSuite) TestListForOrganization_Validation_InvalidOffset_Negative() {
	req := &audit_log_service.ListForOrganizationRequest{
		OrganizationID: uuid.New().String(),
		Limit:          10,
		Offset:         -1,
	}
	resp, err := s.svc.ListForOrganization(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- ListForOrganization: business logic ---

func (s *AuditLogServiceTestSuite) TestListForOrganization_OrganizationNotFound() {
	orgID := uuid.New().String()
	s.orgRepo.On("GetByID", orgID).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &audit_log_service.ListForOrganizationRequest{
		OrganizationID: orgID,
		Limit:          10,
		Offset:         0,
	}
	resp, err := s.svc.ListForOrganization(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *AuditLogServiceTestSuite) TestListForOrganization_Success_ReturnsPaginatedResults() {
	orgID := uuid.New()
	futureTime := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC, // Has audit log feature
		SubscriptionExpiresAt: &futureTime,
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	logs := &[]models.AuditLog{
		{ID: uuid.New(), OrganizationID: orgID, Action: constants.AUDIT_ACTION_TYPE_CREATE},
		{ID: uuid.New(), OrganizationID: orgID, Action: constants.AUDIT_ACTION_TYPE_UPDATE},
	}
	s.auditLogRepo.On("ListByOrganizationID", orgID.String(), mock.MatchedBy(func(f *audit_log_repository.AuditLogFilters) bool {
		return f != nil && f.StartDate != nil
	}), 11, 0).Return(logs, nil)

	req := &audit_log_service.ListForOrganizationRequest{
		OrganizationID: orgID.String(),
		Filters:        nil,
		Limit:          10,
		Offset:         0,
	}
	resp, err := s.svc.ListForOrganization(req, s.accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(resp.Items, 2)
	s.False(resp.HasNext)
	s.orgRepo.AssertExpectations(s.T())
	s.auditLogRepo.AssertExpectations(s.T())
}

func (s *AuditLogServiceTestSuite) TestListForOrganization_Success_WithHasNext() {
	orgID := uuid.New()
	futureTime := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &futureTime,
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	// Return 11 items (limit+1) to trigger HasNext
	logs := make([]models.AuditLog, 11)
	for i := range logs {
		logs[i] = models.AuditLog{ID: uuid.New(), OrganizationID: orgID}
	}
	s.auditLogRepo.On("ListByOrganizationID", orgID.String(), mock.MatchedBy(func(f *audit_log_repository.AuditLogFilters) bool {
		return f != nil && f.StartDate != nil
	}), 11, 0).Return(&logs, nil)

	req := &audit_log_service.ListForOrganizationRequest{
		OrganizationID: orgID.String(),
		Filters:        nil,
		Limit:          10,
		Offset:         0,
	}
	resp, err := s.svc.ListForOrganization(req, s.accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(resp.Items, 10) // Should be truncated to limit
	s.True(resp.HasNext)
	s.orgRepo.AssertExpectations(s.T())
	s.auditLogRepo.AssertExpectations(s.T())
}

func (s *AuditLogServiceTestSuite) TestListForOrganization_Success_WithFilters() {
	orgID := uuid.New()
	futureTime := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &futureTime,
	}
	startDate := s.now.Add(-24 * time.Hour)
	endDate := s.now
	filters := &audit_log_repository.AuditLogFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	logs := &[]models.AuditLog{
		{ID: uuid.New(), OrganizationID: orgID},
	}
	s.auditLogRepo.On("ListByOrganizationID", orgID.String(), filters, 11, 10).Return(logs, nil)

	req := &audit_log_service.ListForOrganizationRequest{
		OrganizationID: orgID.String(),
		Filters:        filters,
		Limit:          10,
		Offset:         10,
	}
	resp, err := s.svc.ListForOrganization(req, s.accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(resp.Items, 1)
	s.False(resp.HasNext)
	s.orgRepo.AssertExpectations(s.T())
	s.auditLogRepo.AssertExpectations(s.T())
}

func (s *AuditLogServiceTestSuite) TestListForOrganization_Success_EmptyList() {
	orgID := uuid.New()
	futureTime := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &futureTime,
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	emptyLogs := &[]models.AuditLog{}
	s.auditLogRepo.On("ListByOrganizationID", orgID.String(), mock.MatchedBy(func(f *audit_log_repository.AuditLogFilters) bool {
		return f != nil && f.StartDate != nil
	}), 11, 0).Return(emptyLogs, nil)

	req := &audit_log_service.ListForOrganizationRequest{
		OrganizationID: orgID.String(),
		Filters:        nil,
		Limit:          10,
		Offset:         0,
	}
	resp, err := s.svc.ListForOrganization(req, s.accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(resp.Items, 0)
	s.False(resp.HasNext)
	s.orgRepo.AssertExpectations(s.T())
	s.auditLogRepo.AssertExpectations(s.T())
}

func (s *AuditLogServiceTestSuite) TestListForOrganization_RepoListFails() {
	orgID := uuid.New()
	futureTime := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &futureTime,
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.auditLogRepo.On("ListByOrganizationID", orgID.String(), mock.MatchedBy(func(f *audit_log_repository.AuditLogFilters) bool {
		return f != nil && f.StartDate != nil
	}), 11, 0).Return((*[]models.AuditLog)(nil), errors.New("db error"))

	req := &audit_log_service.ListForOrganizationRequest{
		OrganizationID: orgID.String(),
		Filters:        nil,
		Limit:          10,
		Offset:         0,
	}
	resp, err := s.svc.ListForOrganization(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.auditLogRepo.AssertExpectations(s.T())
}
