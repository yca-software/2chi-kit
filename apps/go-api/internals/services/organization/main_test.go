package organization_service_test

import (
	"testing"
	"time"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	audit_log_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/audit_log"
	organization_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization"
	organization_member_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization_member"
	role_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/role"
	team_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/team"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	google_service "github.com/yca-software/2chi-kit/go-api/internals/services/google"
	organization_service "github.com/yca-software/2chi-kit/go-api/internals/services/organization"
	paddle_service "github.com/yca-software/2chi-kit/go-api/internals/services/paddle"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
	yca_repository "github.com/yca-software/go-common/repository"
	yca_validate "github.com/yca-software/go-common/validator"
)

type OrganizationServiceTestSuite struct {
	suite.Suite
	svc           organization_service.Service
	repos         *repositories.Repositories
	orgRepo       *organization_repository.MockRepository
	roleRepo      *role_repository.MockRepository
	teamRepo      *team_repository.MockRepository
	orgMemberRepo *organization_member_repository.MockRepository
	auditLogRepo  *audit_log_repository.MockRepository
	googleSvc     *google_service.MockService
	paddleSvc     *paddle_service.MockPaddleService
	auditLogSvc   *audit_log_service.MockService
	logger        *yca_log.MockLogger
	authorizer    *helpers.Authorizer
	now           time.Time
}

func TestOrganizationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationServiceTestSuite))
}

func (s *OrganizationServiceTestSuite) SetupTest() {
	s.now = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	s.orgRepo = organization_repository.NewMock()
	s.roleRepo = role_repository.NewMock()
	s.teamRepo = team_repository.NewMock()
	s.orgMemberRepo = organization_member_repository.NewMock()
	s.auditLogRepo = audit_log_repository.NewMock()
	s.googleSvc = google_service.NewMockService()
	s.paddleSvc = paddle_service.NewMockPaddleService()
	s.auditLogSvc = audit_log_service.NewMockService()
	s.logger = &yca_log.MockLogger{}

	s.repos = &repositories.Repositories{
		Organization:       s.orgRepo,
		Role:               s.roleRepo,
		Team:               s.teamRepo,
		OrganizationMember: s.orgMemberRepo,
		AuditLog:           s.auditLogRepo,
	}

	s.authorizer = helpers.NewAuthorizer(func() time.Time { return s.now })

	s.svc = organization_service.NewService(&organization_service.Dependencies{
		Validator:       yca_validate.New(),
		Logger:          s.logger,
		Repos:           s.repos,
		Authorizer:      s.authorizer,
		AuditLogService: s.auditLogSvc,
		PaddleService:   s.paddleSvc,
		GoogleService:   s.googleSvc,
		GenerateID:      uuid.NewV7,
		Now:             func() time.Time { return s.now },
	})
}

// --- Create: validations ---

func (s *OrganizationServiceTestSuite) TestCreate_Validation_MissingName() {
	req := &organization_service.CreateRequest{
		Name:         "",
		PlaceID:      "place_123",
		BillingEmail: "test@example.com",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "user@example.com",
		},
	}
	resp, err := s.svc.Create(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *OrganizationServiceTestSuite) TestCreate_Validation_MissingPlaceID() {
	req := &organization_service.CreateRequest{
		Name:         "Test Org",
		PlaceID:      "",
		BillingEmail: "test@example.com",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "user@example.com",
		},
	}
	resp, err := s.svc.Create(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *OrganizationServiceTestSuite) TestCreate_Validation_InvalidEmail() {
	req := &organization_service.CreateRequest{
		Name:         "Test Org",
		PlaceID:      "place_123",
		BillingEmail: "not-an-email",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "user@example.com",
		},
	}
	resp, err := s.svc.Create(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Create: business logic ---

func (s *OrganizationServiceTestSuite) TestCreate_APIKeyForbidden() {
	req := &organization_service.CreateRequest{
		Name:         "Test Org",
		PlaceID:      "place_123",
		BillingEmail: "test@example.com",
	}
	accessInfo := &models.AccessInfo{
		ApiKey: &models.APIKey{
			ID: uuid.New(),
		},
	}
	resp, err := s.svc.Create(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(403, e.StatusCode)
		// Just verify it's a forbidden error
		s.NotEmpty(e.ErrorCode)
	}
}

func (s *OrganizationServiceTestSuite) TestCreate_GetLocationDataFails() {
	req := &organization_service.CreateRequest{
		Name:         "Test Org",
		PlaceID:      "place_123",
		BillingEmail: "test@example.com",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "user@example.com",
		},
	}
	tx := yca_repository.NewMockTx()
	s.orgRepo.On("BeginTx").Return(tx, nil)
	s.googleSvc.On("GetLocationData", mock.Anything, "place_123").Return((*models.LocationData)(nil), yca_error.NewInternalServerError(nil, constants.INTERNAL_SERVER_ERROR_CODE, nil))
	tx.On("Rollback").Return(nil).Maybe()

	resp, err := s.svc.Create(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.googleSvc.AssertExpectations(s.T())
	s.orgRepo.AssertExpectations(s.T())
}

func (s *OrganizationServiceTestSuite) TestCreate_PaddleCustomerCreationFails() {
	req := &organization_service.CreateRequest{
		Name:         "Test Org",
		PlaceID:      "place_123",
		BillingEmail: "test@example.com",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Email:  "user@example.com",
		},
	}
	locationData := &models.LocationData{
		Address:  "123 Main St",
		City:     "New York",
		Zip:      "10001",
		Country:  "US",
		PlaceID:  "place_123",
		Geo:      models.Point{},
		Timezone: "America/New_York",
	}
	tx := yca_repository.NewMockTx()
	s.orgRepo.On("BeginTx").Return(tx, nil)
	s.googleSvc.On("GetLocationData", mock.Anything, "place_123").Return(locationData, nil)
	s.paddleSvc.On("CreateCustomer", mock.AnythingOfType("*models.Organization")).Return((*paddle.Customer)(nil), yca_error.NewInternalServerError(nil, constants.INTERNAL_SERVER_ERROR_CODE, nil))
	s.logger.On("Log", mock.Anything).Return()
	tx.On("Rollback").Return(nil).Maybe()

	resp, err := s.svc.Create(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.googleSvc.AssertExpectations(s.T())
	s.paddleSvc.AssertExpectations(s.T())
	s.orgRepo.AssertExpectations(s.T())
}

func (s *OrganizationServiceTestSuite) TestCreate_Success() {
	req := &organization_service.CreateRequest{
		Name:         "Test Org",
		PlaceID:      "place_123",
		BillingEmail: "test@example.com",
	}
	userID := uuid.New()
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Email:  "user@example.com",
		},
	}
	locationData := &models.LocationData{
		Address:  "123 Main St",
		City:     "New York",
		Zip:      "10001",
		Country:  "US",
		PlaceID:  "place_123",
		Geo:      models.Point{},
		Timezone: "America/New_York",
	}
	paddleCustomer := &paddle.Customer{
		ID: "paddle_customer_123",
	}
	tx := yca_repository.NewMockTx()

	s.googleSvc.On("GetLocationData", mock.Anything, "place_123").Return(locationData, nil)
	s.paddleSvc.On("CreateCustomer", mock.AnythingOfType("*models.Organization")).Return(paddleCustomer, nil)
	s.orgRepo.On("BeginTx").Return(tx, nil)
	s.orgRepo.On("Create", tx, mock.AnythingOfType("*models.Organization")).Return(nil)
	s.roleRepo.On("CreateMany", tx, mock.AnythingOfType("*[]models.Role")).Return(nil)
	s.orgMemberRepo.On("Create", tx, mock.AnythingOfType("*models.OrganizationMember")).Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	resp, err := s.svc.Create(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.NotNil(resp.Organization)
	s.NotNil(resp.Roles)
	s.NotNil(resp.Members)
	s.Equal("Test Org", resp.Organization.Name)
	s.Equal("paddle_customer_123", resp.Organization.PaddleCustomerID)
	s.googleSvc.AssertExpectations(s.T())
	s.paddleSvc.AssertExpectations(s.T())
	s.orgRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
	s.orgMemberRepo.AssertExpectations(s.T())
}

// --- Get: validations ---

func (s *OrganizationServiceTestSuite) TestGet_Validation_InvalidUUID() {
	req := &organization_service.GetRequest{
		OrganizationID: "not-a-uuid",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	resp, err := s.svc.Get(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Get: business logic ---

func (s *OrganizationServiceTestSuite) TestGet_NotFound() {
	orgID := uuid.New()
	req := &organization_service.GetRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.Get(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *OrganizationServiceTestSuite) TestGet_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	org := &models.Organization{
		ID:   orgID,
		Name: "Test Org",
	}
	req := &organization_service.GetRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)

	resp, err := s.svc.Get(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
}

func (s *OrganizationServiceTestSuite) TestGet_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	org := &models.Organization{
		ID:   orgID,
		Name: "Test Org",
	}
	req := &organization_service.GetRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)

	resp, err := s.svc.Get(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal(org.Name, resp.Name)
	s.orgRepo.AssertExpectations(s.T())
}

// --- Update: validations ---

func (s *OrganizationServiceTestSuite) TestUpdate_Validation_MissingName() {
	req := &organization_service.UpdateRequest{
		OrganizationID: uuid.New().String(),
		Name:           "",
		PlaceID:        "place_123",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	resp, err := s.svc.Update(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *OrganizationServiceTestSuite) TestUpdate_Validation_InvalidUUID() {
	req := &organization_service.UpdateRequest{
		OrganizationID: "not-a-uuid",
		Name:           "Updated Org",
		PlaceID:        "place_123",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	resp, err := s.svc.Update(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Update: business logic ---

func (s *OrganizationServiceTestSuite) TestUpdate_NotFound() {
	orgID := uuid.New()
	req := &organization_service.UpdateRequest{
		OrganizationID: orgID.String(),
		Name:           "Updated Org",
		PlaceID:        "place_123",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.Update(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *OrganizationServiceTestSuite) TestUpdate_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                          orgID,
		Name:                        "Test Org",
		SubscriptionType:            constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt:       &expiresAt,
		SubscriptionPaymentInterval: constants.PAYMENT_INTERVAL_MONTHLY,
	}
	req := &organization_service.UpdateRequest{
		OrganizationID: orgID.String(),
		Name:           "Updated Org",
		PlaceID:        "place_123",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)

	resp, err := s.svc.Update(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
}

func (s *OrganizationServiceTestSuite) TestUpdate_Success_WithoutPlaceIDChange() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                          orgID,
		Name:                        "Test Org",
		PlaceID:                     "place_123",
		SubscriptionType:            constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt:       &expiresAt,
		SubscriptionPaymentInterval: constants.PAYMENT_INTERVAL_MONTHLY,
	}
	req := &organization_service.UpdateRequest{
		OrganizationID: orgID.String(),
		Name:           "Updated Org",
		PlaceID:        "place_123",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ORG_WRITE},
				},
			},
		},
	}
	tx := yca_repository.NewMockTx()
	paddleCustomer := &paddle.Customer{ID: "paddle_123"}

	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.paddleSvc.On("UpdateCustomer", mock.AnythingOfType("*models.Organization")).Return(paddleCustomer, nil)
	s.orgRepo.On("BeginTx").Return(tx, nil)
	s.orgRepo.On("Update", tx, mock.AnythingOfType("*models.Organization")).Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	resp, err := s.svc.Update(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal("Updated Org", resp.Name)
	s.orgRepo.AssertExpectations(s.T())
	s.paddleSvc.AssertExpectations(s.T())
}

func (s *OrganizationServiceTestSuite) TestUpdate_Success_WithPlaceIDChange() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                          orgID,
		Name:                        "Test Org",
		PlaceID:                     "place_123",
		SubscriptionType:            constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt:       &expiresAt,
		SubscriptionPaymentInterval: constants.PAYMENT_INTERVAL_MONTHLY,
	}
	req := &organization_service.UpdateRequest{
		OrganizationID: orgID.String(),
		Name:           "Updated Org",
		PlaceID:        "place_456",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ORG_WRITE},
				},
			},
		},
	}
	locationData := &models.LocationData{
		Address:  "456 Oak Ave",
		City:     "Los Angeles",
		Zip:      "90001",
		Country:  "US",
		PlaceID:  "place_456",
		Geo:      models.Point{},
		Timezone: "America/Los_Angeles",
	}
	tx := yca_repository.NewMockTx()
	paddleCustomer := &paddle.Customer{ID: "paddle_123"}

	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.googleSvc.On("GetLocationData", mock.Anything, "place_456").Return(locationData, nil)
	s.paddleSvc.On("UpdateCustomer", mock.AnythingOfType("*models.Organization")).Return(paddleCustomer, nil)
	s.orgRepo.On("BeginTx").Return(tx, nil)
	s.orgRepo.On("Update", tx, mock.AnythingOfType("*models.Organization")).Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	resp, err := s.svc.Update(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal("Updated Org", resp.Name)
	s.Equal("place_456", resp.PlaceID)
	s.orgRepo.AssertExpectations(s.T())
	s.googleSvc.AssertExpectations(s.T())
	s.paddleSvc.AssertExpectations(s.T())
}

// --- Archive: validations ---

func (s *OrganizationServiceTestSuite) TestArchive_Validation_InvalidUUID() {
	req := &organization_service.ArchiveRequest{
		OrganizationID: "not-a-uuid",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	err := s.svc.Archive(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Archive: business logic ---

func (s *OrganizationServiceTestSuite) TestArchive_NotFound() {
	orgID := uuid.New()
	req := &organization_service.ArchiveRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	err := s.svc.Archive(req, accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *OrganizationServiceTestSuite) TestArchive_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                          orgID,
		Name:                        "Test Org",
		SubscriptionType:            constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt:       &expiresAt,
		SubscriptionPaymentInterval: constants.PAYMENT_INTERVAL_MONTHLY,
	}
	req := &organization_service.ArchiveRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)

	err := s.svc.Archive(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
}

func (s *OrganizationServiceTestSuite) TestArchive_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                          orgID,
		Name:                        "Test Org",
		PaddleCustomerID:            "paddle_123",
		SubscriptionType:            constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt:       &expiresAt,
		SubscriptionPaymentInterval: constants.PAYMENT_INTERVAL_MONTHLY,
	}
	req := &organization_service.ArchiveRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ORG_DELETE},
				},
			},
		},
	}
	tx := yca_repository.NewMockTx()

	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.orgRepo.On("BeginTx").Return(tx, nil)
	s.orgRepo.On("Archive", tx, orgID.String()).Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	err := s.svc.Archive(req, accessInfo)
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
	s.paddleSvc.AssertExpectations(s.T())
}

func (s *OrganizationServiceTestSuite) TestArchive_Success_WithSubscription() {
	orgID := uuid.New()
	userID := uuid.New()
	subscriptionID := "paddle_sub_123"
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                          orgID,
		Name:                        "Test Org",
		PaddleCustomerID:            "paddle_123",
		PaddleSubscriptionID:        &subscriptionID,
		SubscriptionType:            constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt:       &expiresAt,
		SubscriptionPaymentInterval: constants.PAYMENT_INTERVAL_MONTHLY,
	}
	req := &organization_service.ArchiveRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ORG_DELETE},
				},
			},
		},
	}
	tx := yca_repository.NewMockTx()

	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.orgRepo.On("BeginTx").Return(tx, nil)
	s.orgRepo.On("Archive", tx, orgID.String()).Return(nil)
	s.paddleSvc.On("CancelSubscription", subscriptionID).Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	err := s.svc.Archive(req, accessInfo)
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
	s.paddleSvc.AssertExpectations(s.T())
}

// --- Delete: validations ---

func (s *OrganizationServiceTestSuite) TestDelete_Validation_InvalidUUID() {
	req := &organization_service.DeleteRequest{
		OrganizationID: "not-a-uuid",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Delete: business logic ---

func (s *OrganizationServiceTestSuite) TestDelete_NotAdmin() {
	req := &organization_service.DeleteRequest{
		OrganizationID: uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}

	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
}

func (s *OrganizationServiceTestSuite) TestDelete_NotFound() {
	orgID := uuid.New()
	req := &organization_service.DeleteRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			IsAdmin: true,
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *OrganizationServiceTestSuite) TestDelete_Success() {
	orgID := uuid.New()
	org := &models.Organization{
		ID:               orgID,
		Name:             "Test Org",
		PaddleCustomerID: "paddle_123",
	}
	req := &organization_service.DeleteRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			IsAdmin: true,
		},
	}
	tx := yca_repository.NewMockTx()

	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.orgRepo.On("BeginTx").Return(tx, nil)
	s.orgRepo.On("Delete", tx, orgID.String()).Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	err := s.svc.Delete(req, accessInfo)
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
	s.paddleSvc.AssertExpectations(s.T())
}

func (s *OrganizationServiceTestSuite) TestDelete_Success_WithSubscription() {
	orgID := uuid.New()
	subscriptionID := "paddle_sub_123"
	org := &models.Organization{
		ID:                   orgID,
		Name:                 "Test Org",
		PaddleCustomerID:     "paddle_123",
		PaddleSubscriptionID: &subscriptionID,
	}
	req := &organization_service.DeleteRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			IsAdmin: true,
		},
	}
	tx := yca_repository.NewMockTx()

	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.orgRepo.On("BeginTx").Return(tx, nil)
	s.orgRepo.On("Delete", tx, orgID.String()).Return(nil)
	s.paddleSvc.On("CancelSubscription", subscriptionID).Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	err := s.svc.Delete(req, accessInfo)
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
	s.paddleSvc.AssertExpectations(s.T())
}

// --- Count: business logic ---

func (s *OrganizationServiceTestSuite) TestCount_NotAdmin() {
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}

	count, err := s.svc.Count(accessInfo)
	s.Error(err)
	s.Equal(0, count)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
}

func (s *OrganizationServiceTestSuite) TestCount_Success() {
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			IsAdmin: true,
		},
	}
	s.orgRepo.On("Count").Return(5, nil)

	count, err := s.svc.Count(accessInfo)
	s.NoError(err)
	s.Equal(5, count)
	s.orgRepo.AssertExpectations(s.T())
}

// --- List: validations ---

func (s *OrganizationServiceTestSuite) TestList_Validation_InvalidLimit() {
	req := &organization_service.ListRequest{
		Limit:  0,
		Offset: 0,
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	resp, err := s.svc.List(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *OrganizationServiceTestSuite) TestList_Validation_LimitTooHigh() {
	req := &organization_service.ListRequest{
		Limit:  101,
		Offset: 0,
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	resp, err := s.svc.List(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *OrganizationServiceTestSuite) TestList_Validation_NegativeOffset() {
	req := &organization_service.ListRequest{
		Limit:  10,
		Offset: -1,
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	resp, err := s.svc.List(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- List: business logic ---

func (s *OrganizationServiceTestSuite) TestList_NotAdmin() {
	req := &organization_service.ListRequest{
		Limit:  10,
		Offset: 0,
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}

	resp, err := s.svc.List(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
}

func (s *OrganizationServiceTestSuite) TestList_Success() {
	req := &organization_service.ListRequest{
		Limit:        10,
		Offset:       0,
		SearchPhrase: "test",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			IsAdmin: true,
		},
	}
	orgs := []models.Organization{
		{ID: uuid.New(), Name: "Test Org 1"},
		{ID: uuid.New(), Name: "Test Org 2"},
	}
	s.orgRepo.On("Search", "test", 11, 0).Return(&orgs, nil)

	resp, err := s.svc.List(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(resp.Items, 2)
	s.False(resp.HasNext)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *OrganizationServiceTestSuite) TestList_Success_WithHasNext() {
	req := &organization_service.ListRequest{
		Limit:  10,
		Offset: 0,
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			IsAdmin: true,
		},
	}
	orgs := make([]models.Organization, 11)
	for i := range orgs {
		orgs[i] = models.Organization{ID: uuid.New(), Name: "Test Org"}
	}
	s.orgRepo.On("Search", "", 11, 0).Return(&orgs, nil)

	resp, err := s.svc.List(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(resp.Items, 10)
	s.True(resp.HasNext)
	s.orgRepo.AssertExpectations(s.T())
}

// --- AdminUpdateSubscriptionSettings ---

func (s *OrganizationServiceTestSuite) TestAdminUpdateSubscriptionSettings_NotAdmin() {
	req := &organization_service.AdminUpdateSubscriptionSettingsRequest{
		OrganizationID: uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New()},
	}
	resp, err := s.svc.AdminUpdateSubscriptionSettings(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
}

func (s *OrganizationServiceTestSuite) TestAdminUpdateSubscriptionSettings_InvalidUUID() {
	req := &organization_service.AdminUpdateSubscriptionSettingsRequest{
		OrganizationID: "not-a-uuid",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New(), IsAdmin: true},
	}
	resp, err := s.svc.AdminUpdateSubscriptionSettings(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *OrganizationServiceTestSuite) TestAdminUpdateSubscriptionSettings_OrgNotFound() {
	orgID := uuid.New()
	req := &organization_service.AdminUpdateSubscriptionSettingsRequest{
		OrganizationID:   orgID.String(),
		SubscriptionType: ptr(2),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New(), IsAdmin: true},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.AdminUpdateSubscriptionSettings(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *OrganizationServiceTestSuite) TestAdminUpdateSubscriptionSettings_Success() {
	orgID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID: orgID, Name: "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &expiresAt,
	}
	customSub := true
	req := &organization_service.AdminUpdateSubscriptionSettingsRequest{
		OrganizationID:     orgID.String(),
		CustomSubscription: &customSub,
		SubscriptionType:   ptr(3),
		SubscriptionSeats:  ptr(10),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New(), IsAdmin: true},
	}
	tx := yca_repository.NewMockTx()

	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.orgRepo.On("BeginTx").Return(tx, nil)
	s.orgRepo.On("Update", tx, mock.AnythingOfType("*models.Organization")).Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()

	resp, err := s.svc.AdminUpdateSubscriptionSettings(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.True(resp.CustomSubscription)
	s.Equal(3, resp.SubscriptionType)
	s.Equal(10, resp.SubscriptionSeats)
	s.orgRepo.AssertExpectations(s.T())
}

func ptr(i int) *int { return &i }
