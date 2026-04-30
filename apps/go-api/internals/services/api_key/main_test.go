package api_key_service_test

import (
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
	api_key_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/api_key"
	organization_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization"
	api_key_service "github.com/yca-software/2chi-kit/go-api/internals/services/api_key"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type ApiKeyServiceTestSuite struct {
	suite.Suite
	svc         api_key_service.Service
	repos       *repositories.Repositories
	orgRepo     *organization_repository.MockRepository
	apiKeyRepo  *api_key_repository.MockRepository
	auditLogSvc *audit_log_service.MockService
	logger      *yca_log.MockLogger
	now         time.Time
	accessInfo  *models.AccessInfo // admin so auth is not tested here
}

func TestApiKeyServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ApiKeyServiceTestSuite))
}

func (s *ApiKeyServiceTestSuite) SetupTest() {
	s.now = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	s.orgRepo = organization_repository.NewMock()
	s.apiKeyRepo = api_key_repository.NewMock()
	s.auditLogSvc = audit_log_service.NewMockService()
	s.logger = &yca_log.MockLogger{}

	s.repos = &repositories.Repositories{
		Organization: s.orgRepo,
		ApiKey:       s.apiKeyRepo,
	}

	// AccessInfo with admin so authorization passes; we only test validations and business logic.
	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			IsAdmin: true,
		},
	}

	s.svc = api_key_service.New(&api_key_service.Dependencies{
		Validator:       yca_validate.New(),
		Repos:           s.repos,
		Authorizer:      helpers.NewAuthorizer(func() time.Time { return s.now }),
		GenerateID:      uuid.NewV7,
		GenerateToken:   func() (string, error) { return "abcdefghijklmnopqrstuvwxyz123456", nil },
		HashToken:       func(token string) string { return "hashed:" + token },
		Now:             func() time.Time { return s.now },
		Logger:          s.logger,
		AuditLogService: s.auditLogSvc,
	})
}

func (s *ApiKeyServiceTestSuite) defaultKeyExpiry() time.Time {
	return s.now.Add(24 * time.Hour)
}

// --- Create: validations ---

func (s *ApiKeyServiceTestSuite) TestCreate_Validation_MissingOrganizationID() {
	req := &api_key_service.CreateRequest{
		OrganizationID: "",
		Name:           "My Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
		ExpiresAt:      s.defaultKeyExpiry(),
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *ApiKeyServiceTestSuite) TestCreate_Validation_InvalidOrganizationID() {
	req := &api_key_service.CreateRequest{
		OrganizationID: "not-a-uuid",
		Name:           "My Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
		ExpiresAt:      s.defaultKeyExpiry(),
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *ApiKeyServiceTestSuite) TestCreate_Validation_EmptyName() {
	req := &api_key_service.CreateRequest{
		OrganizationID: uuid.New().String(),
		Name:           "",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
		ExpiresAt:      s.defaultKeyExpiry(),
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *ApiKeyServiceTestSuite) TestCreate_Validation_NameTooLong() {
	longName := make([]byte, 256)
	for i := range longName {
		longName[i] = 'a'
	}
	req := &api_key_service.CreateRequest{
		OrganizationID: uuid.New().String(),
		Name:           string(longName),
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
		ExpiresAt:      s.defaultKeyExpiry(),
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *ApiKeyServiceTestSuite) TestCreate_Validation_EmptyPermissions() {
	req := &api_key_service.CreateRequest{
		OrganizationID: uuid.New().String(),
		Name:           "My Key",
		Permissions:    nil,
		ExpiresAt:      s.defaultKeyExpiry(),
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *ApiKeyServiceTestSuite) TestCreate_Validation_MissingExpiresAt() {
	req := &api_key_service.CreateRequest{
		OrganizationID: uuid.New().String(),
		Name:           "My Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
		// ExpiresAt omitted → zero time; validate:"required" must reject before repos run.
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertNotCalled(s.T(), "GetByID")
}

func (s *ApiKeyServiceTestSuite) TestCreate_Validation_InvalidPermission() {
	// Invalid permission is rejected before any repo call.
	req := &api_key_service.CreateRequest{
		OrganizationID: uuid.New().String(),
		Name:           "My Key",
		Permissions:    models.RolePermissions{"invalid:permission"},
		ExpiresAt:      s.defaultKeyExpiry(),
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVALID_API_KEY_PERMISSION_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertNotCalled(s.T(), "GetByID")
}

func (s *ApiKeyServiceTestSuite) TestCreate_Validation_MultipleInvalidPermissions() {
	req := &api_key_service.CreateRequest{
		OrganizationID: uuid.New().String(),
		Name:           "My Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ, "invalid:permission"},
		ExpiresAt:      s.defaultKeyExpiry(),
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVALID_API_KEY_PERMISSION_CODE, e.ErrorCode)
	}
}

func (s *ApiKeyServiceTestSuite) TestCreate_Validation_MultipleValidPermissions() {
	orgID := uuid.New()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, Name: "Acme", SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.apiKeyRepo.On("Create", nil, mock.AnythingOfType("*models.APIKey")).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, s.accessInfo).Return(&models.AuditLog{}, nil)
	s.logger.On("Log", mock.Anything).Maybe()

	req := &api_key_service.CreateRequest{
		OrganizationID: orgID.String(),
		Name:           "My Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ, constants.PERMISSION_API_KEY_READ},
		ExpiresAt:      s.defaultKeyExpiry(),
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(resp.ApiKey.Permissions, 2)
	s.orgRepo.AssertExpectations(s.T())
	s.apiKeyRepo.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestCreate_SetsExpiresAt() {
	orgID := uuid.New()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, Name: "Acme", SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)

	exp := s.now.Add(24 * time.Hour)
	s.apiKeyRepo.On("Create", nil, mock.MatchedBy(func(apiKey *models.APIKey) bool {
		return apiKey.ExpiresAt.Equal(exp)
	})).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, s.accessInfo).Return(&models.AuditLog{}, nil)
	s.logger.On("Log", mock.Anything).Maybe()

	req := &api_key_service.CreateRequest{
		OrganizationID: orgID.String(),
		Name:           "My Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
		ExpiresAt:      exp,
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.ApiKey.ExpiresAt)
	s.True(resp.ApiKey.ExpiresAt.Equal(exp))
	s.orgRepo.AssertExpectations(s.T())
	s.apiKeyRepo.AssertExpectations(s.T())
}

// --- Create: business logic ---

func (s *ApiKeyServiceTestSuite) TestCreate_OrganizationNotFound() {
	orgID := uuid.New().String()
	s.orgRepo.On("GetByID", orgID).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &api_key_service.CreateRequest{
		OrganizationID: orgID,
		Name:           "My Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
		ExpiresAt:      s.defaultKeyExpiry(),
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.MockRepository.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestCreate_GenerateTokenError() {
	orgID := uuid.New()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, Name: "Acme", SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	s.orgRepo.MockRepository.On("GetByID", orgID.String()).Return(org, nil)

	svc := api_key_service.New(&api_key_service.Dependencies{
		Validator:       yca_validate.New(),
		Repos:           s.repos,
		Authorizer:      helpers.NewAuthorizer(func() time.Time { return s.now }),
		GenerateID:      uuid.NewV7,
		GenerateToken:   func() (string, error) { return "", errors.New("token generation failed") },
		HashToken:       func(token string) string { return "hashed:" + token },
		Now:             func() time.Time { return s.now },
		Logger:          s.logger,
		AuditLogService: s.auditLogSvc,
	})

	req := &api_key_service.CreateRequest{
		OrganizationID: orgID.String(),
		Name:           "My Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
		ExpiresAt:      s.defaultKeyExpiry(),
	}
	resp, err := svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.MockRepository.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestCreate_GenerateIDError() {
	orgID := uuid.New()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, Name: "Acme", SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	s.orgRepo.MockRepository.On("GetByID", orgID.String()).Return(org, nil)

	svc := api_key_service.New(&api_key_service.Dependencies{
		Validator:       yca_validate.New(),
		Repos:           s.repos,
		Authorizer:      helpers.NewAuthorizer(func() time.Time { return s.now }),
		GenerateID:      func() (uuid.UUID, error) { return uuid.Nil, errors.New("id generation failed") },
		GenerateToken:   func() (string, error) { return "abcdefghijklmnopqrstuvwxyz123456", nil },
		HashToken:       func(token string) string { return "hashed:" + token },
		Now:             func() time.Time { return s.now },
		Logger:          s.logger,
		AuditLogService: s.auditLogSvc,
	})

	req := &api_key_service.CreateRequest{
		OrganizationID: orgID.String(),
		Name:           "My Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
		ExpiresAt:      s.defaultKeyExpiry(),
	}
	resp, err := svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.MockRepository.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestCreate_Success_ReturnsApiKeyAndSecretWithPrefix() {
	orgID := uuid.New()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Acme",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_PRO,
		SubscriptionExpiresAt: &futureExpiry,
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.apiKeyRepo.On("Create", nil, mock.AnythingOfType("*models.APIKey")).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, s.accessInfo).Return(&models.AuditLog{}, nil)
	s.logger.On("Log", mock.Anything).Maybe()

	req := &api_key_service.CreateRequest{
		OrganizationID: orgID.String(),
		Name:           "My Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
		ExpiresAt:      s.defaultKeyExpiry(),
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.ApiKey)
	s.Equal("My Key", resp.ApiKey.Name)
	s.Equal(orgID, resp.ApiKey.OrganizationID)
	s.Equal(constants.PERMISSION_ORG_READ, resp.ApiKey.Permissions[0])
	s.Equal(constants.API_KEY_PREFIX+"abcdefghijklmnopqrstuvwxyz123456", resp.Secret)
	s.Equal(constants.API_KEY_PREFIX+"abcdefgh", resp.ApiKey.KeyPrefix) // Prefix is sk_ + first 8 chars of raw key
	s.Equal("hashed:abcdefghijklmnopqrstuvwxyz123456", resp.ApiKey.KeyHash)
	s.Equal(s.now, resp.ApiKey.CreatedAt)
	s.orgRepo.AssertExpectations(s.T())
	s.apiKeyRepo.AssertExpectations(s.T())
	s.auditLogSvc.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestCreate_RepoCreateFails() {
	orgID := uuid.New()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, Name: "Acme", SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	s.orgRepo.MockRepository.On("GetByID", orgID.String()).Return(org, nil)
	s.apiKeyRepo.On("Create", nil, mock.AnythingOfType("*models.APIKey")).Return(errors.New("db error"))

	req := &api_key_service.CreateRequest{
		OrganizationID: orgID.String(),
		Name:           "My Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
		ExpiresAt:      s.defaultKeyExpiry(),
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.apiKeyRepo.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestCreate_AuditLogFailureDoesNotFailRequest() {
	orgID := uuid.New()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, Name: "Acme", SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.apiKeyRepo.On("Create", nil, mock.AnythingOfType("*models.APIKey")).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, s.accessInfo).Return((*models.AuditLog)(nil), errors.New("audit log failed"))
	s.logger.On("Log", mock.MatchedBy(func(data yca_log.LogData) bool {
		return data.Level == "error" && data.Message == "Failed to create audit log"
	})).Return()

	req := &api_key_service.CreateRequest{
		OrganizationID: orgID.String(),
		Name:           "My Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
		ExpiresAt:      s.defaultKeyExpiry(),
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.NoError(err) // Audit log failure should not fail the request
	s.Require().NotNil(resp)
	s.auditLogSvc.AssertExpectations(s.T())
	s.logger.AssertExpectations(s.T())
}

// --- List: validations ---

func (s *ApiKeyServiceTestSuite) TestList_Validation_MissingOrganizationID() {
	req := &api_key_service.ListRequest{OrganizationID: ""}
	resp, err := s.svc.List(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *ApiKeyServiceTestSuite) TestList_Validation_InvalidOrganizationID() {
	req := &api_key_service.ListRequest{OrganizationID: "not-a-uuid"}
	resp, err := s.svc.List(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- List: business logic ---

func (s *ApiKeyServiceTestSuite) TestList_OrganizationNotFound() {
	orgID := uuid.New().String()
	s.orgRepo.On("GetByID", orgID).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &api_key_service.ListRequest{OrganizationID: orgID}
	resp, err := s.svc.List(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.MockRepository.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestList_Success_ReturnsApiKeys() {
	orgID := uuid.New()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, Name: "Acme", SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	s.orgRepo.MockRepository.On("GetByID", orgID.String()).Return(org, nil)
	list := &[]models.APIKey{
		{ID: uuid.New(), Name: "Key 1", OrganizationID: orgID},
		{ID: uuid.New(), Name: "Key 2", OrganizationID: orgID},
	}
	s.apiKeyRepo.On("ListByOrganizationID", orgID.String()).Return(list, nil)

	req := &api_key_service.ListRequest{OrganizationID: orgID.String()}
	resp, err := s.svc.List(req, s.accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal(list, resp)
	s.Len(*resp, 2)
	s.orgRepo.AssertExpectations(s.T())
	s.apiKeyRepo.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestList_Success_EmptyList() {
	orgID := uuid.New()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, Name: "Acme", SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	emptyList := &[]models.APIKey{}
	s.apiKeyRepo.On("ListByOrganizationID", orgID.String()).Return(emptyList, nil)

	req := &api_key_service.ListRequest{OrganizationID: orgID.String()}
	resp, err := s.svc.List(req, s.accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(*resp, 0)
	s.orgRepo.AssertExpectations(s.T())
	s.apiKeyRepo.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestList_RepoListFails() {
	orgID := uuid.New()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, Name: "Acme", SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.apiKeyRepo.On("ListByOrganizationID", orgID.String()).Return((*[]models.APIKey)(nil), errors.New("db error"))

	req := &api_key_service.ListRequest{OrganizationID: orgID.String()}
	resp, err := s.svc.List(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.apiKeyRepo.AssertExpectations(s.T())
}

// --- Update: validations ---

func (s *ApiKeyServiceTestSuite) TestUpdate_Validation_MissingOrganizationID() {
	req := &api_key_service.UpdateRequest{
		OrganizationID: "",
		ApiKeyID:       uuid.New().String(),
		Name:           "My Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
	}
	got, err := s.svc.Update(req, s.accessInfo)
	s.Error(err)
	s.Nil(got)
}

func (s *ApiKeyServiceTestSuite) TestUpdate_Validation_InvalidOrganizationID() {
	req := &api_key_service.UpdateRequest{
		OrganizationID: "not-a-uuid",
		ApiKeyID:       uuid.New().String(),
		Name:           "My Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
	}
	got, err := s.svc.Update(req, s.accessInfo)
	s.Error(err)
	s.Nil(got)
}

func (s *ApiKeyServiceTestSuite) TestUpdate_Validation_MissingApiKeyID() {
	req := &api_key_service.UpdateRequest{
		OrganizationID: uuid.New().String(),
		ApiKeyID:       "",
		Name:           "My Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
	}
	got, err := s.svc.Update(req, s.accessInfo)
	s.Error(err)
	s.Nil(got)
}

func (s *ApiKeyServiceTestSuite) TestUpdate_Validation_EmptyName() {
	req := &api_key_service.UpdateRequest{
		OrganizationID: uuid.New().String(),
		ApiKeyID:       uuid.New().String(),
		Name:           "",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
	}
	got, err := s.svc.Update(req, s.accessInfo)
	s.Error(err)
	s.Nil(got)
}

func (s *ApiKeyServiceTestSuite) TestUpdate_Validation_InvalidPermission() {
	req := &api_key_service.UpdateRequest{
		OrganizationID: uuid.New().String(),
		ApiKeyID:       uuid.New().String(),
		Name:           "My Key",
		Permissions:    models.RolePermissions{"invalid:permission"},
	}
	got, err := s.svc.Update(req, s.accessInfo)
	s.Error(err)
	s.Nil(got)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVALID_API_KEY_PERMISSION_CODE, e.ErrorCode)
	}
}

// --- Update: business logic ---

func (s *ApiKeyServiceTestSuite) TestUpdate_OrganizationNotFound() {
	orgID := uuid.New().String()
	apiKeyID := uuid.New().String()
	s.orgRepo.On("GetByID", orgID).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &api_key_service.UpdateRequest{
		OrganizationID: orgID,
		ApiKeyID:       apiKeyID,
		Name:           "Updated Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
	}
	got, err := s.svc.Update(req, s.accessInfo)
	s.Error(err)
	s.Nil(got)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestUpdate_ApiKeyNotFound() {
	orgID := uuid.New()
	apiKeyID := uuid.New().String()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.apiKeyRepo.On("GetByID", orgID.String(), apiKeyID).Return((*models.APIKey)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &api_key_service.UpdateRequest{
		OrganizationID: orgID.String(),
		ApiKeyID:       apiKeyID,
		Name:           "Updated Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
	}
	got, err := s.svc.Update(req, s.accessInfo)
	s.Error(err)
	s.Nil(got)
	s.orgRepo.AssertExpectations(s.T())
	s.apiKeyRepo.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestUpdate_Success() {
	orgID := uuid.New()
	apiKeyID := uuid.New()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	existingKey := &models.APIKey{
		ID:             apiKeyID,
		OrganizationID: orgID,
		Name:           "Old Name",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.apiKeyRepo.On("GetByID", orgID.String(), apiKeyID.String()).Return(existingKey, nil)
	s.apiKeyRepo.On("Update", nil, mock.MatchedBy(func(k *models.APIKey) bool {
		return k.Name == "Updated Key" && len(k.Permissions) == 2
	})).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, s.accessInfo).Return(&models.AuditLog{}, nil)

	req := &api_key_service.UpdateRequest{
		OrganizationID: orgID.String(),
		ApiKeyID:       apiKeyID.String(),
		Name:           "Updated Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ, constants.PERMISSION_MEMBERS_READ},
	}
	got, err := s.svc.Update(req, s.accessInfo)
	s.NoError(err)
	s.Require().NotNil(got)
	s.Equal("Updated Key", got.Name)
	s.Equal(models.RolePermissions{constants.PERMISSION_ORG_READ, constants.PERMISSION_MEMBERS_READ}, got.Permissions)
	s.orgRepo.AssertExpectations(s.T())
	s.apiKeyRepo.AssertExpectations(s.T())
	s.auditLogSvc.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestUpdate_RepoUpdateFails() {
	orgID := uuid.New()
	apiKeyID := uuid.New()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	existingKey := &models.APIKey{
		ID:             apiKeyID,
		OrganizationID: orgID,
		Name:           "Old Name",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.apiKeyRepo.On("GetByID", orgID.String(), apiKeyID.String()).Return(existingKey, nil)
	s.apiKeyRepo.On("Update", nil, mock.Anything).Return(errors.New("db error"))

	req := &api_key_service.UpdateRequest{
		OrganizationID: orgID.String(),
		ApiKeyID:       apiKeyID.String(),
		Name:           "Updated Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
	}
	got, err := s.svc.Update(req, s.accessInfo)
	s.Error(err)
	s.Nil(got)
	s.apiKeyRepo.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestUpdate_AuditLogFailureDoesNotFailRequest() {
	orgID := uuid.New()
	apiKeyID := uuid.New()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	existingKey := &models.APIKey{
		ID:             apiKeyID,
		OrganizationID: orgID,
		Name:           "Old Name",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.apiKeyRepo.On("GetByID", orgID.String(), apiKeyID.String()).Return(existingKey, nil)
	s.apiKeyRepo.On("Update", nil, mock.Anything).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, s.accessInfo).Return((*models.AuditLog)(nil), errors.New("audit log failed"))
	s.logger.On("Log", mock.MatchedBy(func(data yca_log.LogData) bool {
		return data.Level == "error" && data.Message == "Failed to create audit log"
	})).Return()

	req := &api_key_service.UpdateRequest{
		OrganizationID: orgID.String(),
		ApiKeyID:       apiKeyID.String(),
		Name:           "Updated Key",
		Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
	}
	got, err := s.svc.Update(req, s.accessInfo)
	s.NoError(err)
	s.Require().NotNil(got)
	s.apiKeyRepo.AssertExpectations(s.T())
	s.auditLogSvc.AssertExpectations(s.T())
	s.logger.AssertExpectations(s.T())
}

// --- Delete: validations ---

func (s *ApiKeyServiceTestSuite) TestDelete_Validation_MissingOrganizationID() {
	req := &api_key_service.DeleteRequest{
		OrganizationID: "",
		ApiKeyID:       uuid.New().String(),
	}
	err := s.svc.Delete(req, s.accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *ApiKeyServiceTestSuite) TestDelete_Validation_InvalidOrganizationID() {
	req := &api_key_service.DeleteRequest{
		OrganizationID: "not-a-uuid",
		ApiKeyID:       uuid.New().String(),
	}
	err := s.svc.Delete(req, s.accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *ApiKeyServiceTestSuite) TestDelete_Validation_MissingApiKeyID() {
	req := &api_key_service.DeleteRequest{
		OrganizationID: uuid.New().String(),
		ApiKeyID:       "",
	}
	err := s.svc.Delete(req, s.accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *ApiKeyServiceTestSuite) TestDelete_Validation_InvalidApiKeyID() {
	req := &api_key_service.DeleteRequest{
		OrganizationID: uuid.New().String(),
		ApiKeyID:       "not-a-uuid",
	}
	err := s.svc.Delete(req, s.accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Delete: business logic ---

func (s *ApiKeyServiceTestSuite) TestDelete_OrganizationNotFound() {
	orgID := uuid.New().String()
	s.orgRepo.On("GetByID", orgID).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &api_key_service.DeleteRequest{
		OrganizationID: orgID,
		ApiKeyID:       uuid.New().String(),
	}
	err := s.svc.Delete(req, s.accessInfo)
	s.Error(err)
	s.orgRepo.MockRepository.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestDelete_ApiKeyNotFound() {
	orgID := uuid.New()
	apiKeyID := uuid.New().String()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.apiKeyRepo.On("GetByID", orgID.String(), apiKeyID).Return((*models.APIKey)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	// No delete or audit when the key does not exist.

	req := &api_key_service.DeleteRequest{
		OrganizationID: orgID.String(),
		ApiKeyID:       apiKeyID,
	}
	err := s.svc.Delete(req, s.accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
	s.apiKeyRepo.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestDelete_Success() {
	orgID := uuid.New()
	apiKeyUUID := uuid.New()
	apiKeyID := apiKeyUUID.String()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	key := &models.APIKey{ID: apiKeyUUID, Name: "k1"}
	s.apiKeyRepo.On("GetByID", orgID.String(), apiKeyID).Return(key, nil)
	s.apiKeyRepo.On("Delete", nil, orgID.String(), apiKeyID).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, s.accessInfo).Return(&models.AuditLog{}, nil)
	s.logger.On("Log", mock.Anything).Maybe()

	req := &api_key_service.DeleteRequest{
		OrganizationID: orgID.String(),
		ApiKeyID:       apiKeyID,
	}
	err := s.svc.Delete(req, s.accessInfo)
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
	s.apiKeyRepo.AssertExpectations(s.T())
	s.auditLogSvc.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestDelete_RepoDeleteFails() {
	orgID := uuid.New()
	apiKeyUUID := uuid.New()
	apiKeyID := apiKeyUUID.String()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	key := &models.APIKey{ID: apiKeyUUID, Name: "k1"}
	s.apiKeyRepo.On("GetByID", orgID.String(), apiKeyID).Return(key, nil)
	s.apiKeyRepo.On("Delete", nil, orgID.String(), apiKeyID).Return(errors.New("db error"))
	// No audit log should be written if delete fails.
	s.logger.On("Log", mock.Anything).Maybe()

	req := &api_key_service.DeleteRequest{
		OrganizationID: orgID.String(),
		ApiKeyID:       apiKeyID,
	}
	err := s.svc.Delete(req, s.accessInfo)
	s.Error(err)
	s.apiKeyRepo.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestDelete_AuditLogFailureDoesNotFailRequest() {
	orgID := uuid.New()
	apiKeyUUID := uuid.New()
	apiKeyID := apiKeyUUID.String()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: orgID, SubscriptionType: constants.SUBSCRIPTION_TYPE_PRO, SubscriptionExpiresAt: &futureExpiry}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	key := &models.APIKey{ID: apiKeyUUID, Name: "k1"}
	s.apiKeyRepo.On("GetByID", orgID.String(), apiKeyID).Return(key, nil)
	s.apiKeyRepo.On("Delete", nil, orgID.String(), apiKeyID).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, s.accessInfo).Return((*models.AuditLog)(nil), errors.New("audit log failed"))
	s.logger.On("Log", mock.MatchedBy(func(data yca_log.LogData) bool {
		return data.Level == "error" && data.Message == "Failed to create audit log"
	})).Return()

	req := &api_key_service.DeleteRequest{
		OrganizationID: orgID.String(),
		ApiKeyID:       apiKeyID,
	}
	err := s.svc.Delete(req, s.accessInfo)
	s.NoError(err) // Audit log failure should not fail the request
	s.apiKeyRepo.AssertExpectations(s.T())
	s.auditLogSvc.AssertExpectations(s.T())
	s.logger.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestCleanupStaleExpired() {
	s.apiKeyRepo.On("CleanupStaleExpired").Return(nil).Once()
	err := s.svc.CleanupStaleExpired()
	s.NoError(err)
	s.apiKeyRepo.AssertExpectations(s.T())
}

func (s *ApiKeyServiceTestSuite) TestCleanupStaleExpired_RepoError() {
	repoErr := errors.New("cleanup failed")
	s.apiKeyRepo.On("CleanupStaleExpired").Return(repoErr).Once()
	err := s.svc.CleanupStaleExpired()
	s.ErrorIs(err, repoErr)
	s.apiKeyRepo.AssertExpectations(s.T())
}
