package role_service_test

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
	audit_log_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/audit_log"
	organization_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization"
	organization_member_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization_member"
	role_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/role"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	role_service "github.com/yca-software/2chi-kit/go-api/internals/services/role"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type RoleServiceTestSuite struct {
	suite.Suite
	svc          role_service.Service
	repos        *repositories.Repositories
	orgRepo      *organization_repository.MockRepository
	memberRepo   *organization_member_repository.MockRepository
	roleRepo     *role_repository.MockRepository
	auditLogRepo *audit_log_repository.MockRepository
	auditLogSvc  *audit_log_service.MockService
	logger       *yca_log.MockLogger
	authorizer   *helpers.Authorizer
	now          time.Time
}

func TestRoleServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RoleServiceTestSuite))
}

func (s *RoleServiceTestSuite) SetupTest() {
	s.now = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	s.orgRepo = organization_repository.NewMock()
	s.memberRepo = organization_member_repository.NewMock()
	s.roleRepo = role_repository.NewMock()
	s.auditLogRepo = audit_log_repository.NewMock()
	s.auditLogSvc = audit_log_service.NewMockService()
	s.logger = &yca_log.MockLogger{}

	s.repos = &repositories.Repositories{
		Organization:       s.orgRepo,
		OrganizationMember: s.memberRepo,
		Role:               s.roleRepo,
		AuditLog:           s.auditLogRepo,
	}

	s.authorizer = helpers.NewAuthorizer(func() time.Time { return s.now })

	s.svc = role_service.NewService(&role_service.Dependencies{
		GenerateID:      uuid.NewV7,
		Now:             func() time.Time { return s.now },
		Validator:       yca_validate.New(),
		Repositories:    s.repos,
		Authorizer:      s.authorizer,
		Logger:          s.logger,
		AuditLogService: s.auditLogSvc,
	})
}

func (s *RoleServiceTestSuite) newSubscribedOrg(id uuid.UUID) *models.Organization {
	expiresAt := s.now.Add(24 * time.Hour)
	return &models.Organization{
		ID:                    id,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_PRO,
		SubscriptionExpiresAt: &expiresAt,
	}
}

// --- Create: validations ---

func (s *RoleServiceTestSuite) TestCreate_Validation_InvalidOrganizationID() {
	req := &role_service.CreateRequest{
		OrganizationID: "not-a-uuid",
		Name:           "Test Role",
		Description:    "Test description",
		Permissions:    models.RolePermissions{"read", "write"},
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	resp, err := s.svc.Create(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *RoleServiceTestSuite) TestCreate_Validation_MissingName() {
	req := &role_service.CreateRequest{
		OrganizationID: uuid.New().String(),
		Name:           "",
		Description:    "Test description",
		Permissions:    models.RolePermissions{"read", "write"},
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	resp, err := s.svc.Create(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *RoleServiceTestSuite) TestCreate_Validation_MissingPermissions() {
	req := &role_service.CreateRequest{
		OrganizationID: uuid.New().String(),
		Name:           "Test Role",
		Description:    "Test description",
		Permissions:    nil,
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
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

func (s *RoleServiceTestSuite) TestCreate_OrganizationNotFound() {
	orgID := uuid.New()
	req := &role_service.CreateRequest{
		OrganizationID: orgID.String(),
		Name:           "Test Role",
		Description:    "Test description",
		Permissions:    models.RolePermissions{"read", "write"},
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.Create(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *RoleServiceTestSuite) TestCreate_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	org := s.newSubscribedOrg(orgID)
	req := &role_service.CreateRequest{
		OrganizationID: orgID.String(),
		Name:           "Test Role",
		Description:    "Test description",
		Permissions:    models.RolePermissions{"read", "write"},
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)

	resp, err := s.svc.Create(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
}

func (s *RoleServiceTestSuite) TestCreate_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	org := s.newSubscribedOrg(orgID)
	req := &role_service.CreateRequest{
		OrganizationID: orgID.String(),
		Name:           "  Test Role  ",
		Description:    "  Test description  ",
		Permissions:    models.RolePermissions{"read", "write"},
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ROLE_WRITE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.roleRepo.On("Create", nil, mock.AnythingOfType("*models.Role")).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	resp, err := s.svc.Create(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal("Test Role", resp.Name)               // Should be trimmed
	s.Equal("Test description", resp.Description) // Should be trimmed
	s.Equal(req.Permissions, resp.Permissions)
	s.orgRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
}

// --- List: validations ---

func (s *RoleServiceTestSuite) TestList_Validation_InvalidUUID() {
	req := &role_service.ListRequest{
		OrganizationID: "not-a-uuid",
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

func (s *RoleServiceTestSuite) TestList_OrganizationNotFound() {
	orgID := uuid.New()
	req := &role_service.ListRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.List(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *RoleServiceTestSuite) TestList_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	org := s.newSubscribedOrg(orgID)
	req := &role_service.ListRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)

	resp, err := s.svc.List(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
}

func (s *RoleServiceTestSuite) TestList_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	org := s.newSubscribedOrg(orgID)
	req := &role_service.ListRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ROLE_READ},
				},
			},
		},
	}
	roles := []models.Role{
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Name:           "Role 1",
			Permissions:    models.RolePermissions{"read"},
		},
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Name:           "Role 2",
			Permissions:    models.RolePermissions{"read", "write"},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.roleRepo.On("ListByOrganizationID", orgID.String()).Return(&roles, nil)

	resp, err := s.svc.List(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(*resp, 2)
	s.Equal(roles[0].Name, (*resp)[0].Name)
	s.orgRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
}

// --- Update: validations ---

func (s *RoleServiceTestSuite) TestUpdate_Validation_InvalidOrganizationID() {
	req := &role_service.UpdateRequest{
		OrganizationID: "not-a-uuid",
		RoleID:         uuid.New().String(),
		Name:           "Updated Role",
		Description:    "Updated description",
		Permissions:    models.RolePermissions{"read"},
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

func (s *RoleServiceTestSuite) TestUpdate_Validation_InvalidRoleID() {
	req := &role_service.UpdateRequest{
		OrganizationID: uuid.New().String(),
		RoleID:         "not-a-uuid",
		Name:           "Updated Role",
		Description:    "Updated description",
		Permissions:    models.RolePermissions{"read"},
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

func (s *RoleServiceTestSuite) TestUpdate_Validation_MissingName() {
	req := &role_service.UpdateRequest{
		OrganizationID: uuid.New().String(),
		RoleID:         uuid.New().String(),
		Name:           "",
		Description:    "Updated description",
		Permissions:    models.RolePermissions{"read"},
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

func (s *RoleServiceTestSuite) TestUpdate_Validation_MissingPermissions() {
	req := &role_service.UpdateRequest{
		OrganizationID: uuid.New().String(),
		RoleID:         uuid.New().String(),
		Name:           "Updated Role",
		Description:    "Updated description",
		Permissions:    nil,
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

func (s *RoleServiceTestSuite) TestUpdate_OrganizationNotFound() {
	orgID := uuid.New()
	req := &role_service.UpdateRequest{
		OrganizationID: orgID.String(),
		RoleID:         uuid.New().String(),
		Name:           "Updated Role",
		Description:    "Updated description",
		Permissions:    models.RolePermissions{"read"},
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

func (s *RoleServiceTestSuite) TestUpdate_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	org := s.newSubscribedOrg(orgID)
	req := &role_service.UpdateRequest{
		OrganizationID: orgID.String(),
		RoleID:         uuid.New().String(),
		Name:           "Updated Role",
		Description:    "Updated description",
		Permissions:    models.RolePermissions{"read"},
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

func (s *RoleServiceTestSuite) TestUpdate_RoleNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	org := s.newSubscribedOrg(orgID)
	req := &role_service.UpdateRequest{
		OrganizationID: orgID.String(),
		RoleID:         uuid.New().String(),
		Name:           "Updated Role",
		Description:    "Updated description",
		Permissions:    models.RolePermissions{"read"},
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ROLE_WRITE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.roleRepo.On("GetByID", orgID.String(), req.RoleID).Return((*models.Role)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.Update(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
}

func (s *RoleServiceTestSuite) TestUpdate_RoleLocked() {
	orgID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()
	org := s.newSubscribedOrg(orgID)
	role := &models.Role{
		ID:             roleID,
		OrganizationID: orgID,
		Name:           "Locked Role",
		Locked:         true,
	}
	req := &role_service.UpdateRequest{
		OrganizationID: orgID.String(),
		RoleID:         roleID.String(),
		Name:           "Updated Role",
		Description:    "Updated description",
		Permissions:    models.RolePermissions{"read"},
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ROLE_WRITE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.roleRepo.On("GetByID", orgID.String(), roleID.String()).Return(role, nil)

	resp, err := s.svc.Update(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.ROLE_LOCKED_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
}

func (s *RoleServiceTestSuite) TestUpdate_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()
	org := s.newSubscribedOrg(orgID)
	role := &models.Role{
		ID:             roleID,
		OrganizationID: orgID,
		Name:           "Old Role",
		Description:    "Old description",
		Permissions:    models.RolePermissions{"read"},
		Locked:         false,
	}
	req := &role_service.UpdateRequest{
		OrganizationID: orgID.String(),
		RoleID:         roleID.String(),
		Name:           "  Updated Role  ",
		Description:    "  Updated description  ",
		Permissions:    models.RolePermissions{"read", "write"},
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ROLE_WRITE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.roleRepo.On("GetByID", orgID.String(), roleID.String()).Return(role, nil)
	s.roleRepo.On("Update", nil, mock.AnythingOfType("*models.Role")).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	resp, err := s.svc.Update(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal("Updated Role", resp.Name)               // Should be trimmed
	s.Equal("Updated description", resp.Description) // Should be trimmed
	s.Equal(req.Permissions, resp.Permissions)
	s.orgRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
}

func (s *RoleServiceTestSuite) accessInfoWithRoleDelete(orgID uuid.UUID, userID uuid.UUID) *models.AccessInfo {
	return &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ROLE_DELETE},
				},
			},
		},
	}
}

// --- Delete: validations ---

func (s *RoleServiceTestSuite) TestDelete_Validation_InvalidOrganizationID() {
	req := &role_service.DeleteRequest{
		OrganizationID: "not-a-uuid",
		RoleID:         uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{User: &models.UserAccessInfo{UserID: uuid.New()}}
	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *RoleServiceTestSuite) TestDelete_Validation_InvalidRoleID() {
	req := &role_service.DeleteRequest{
		OrganizationID: uuid.New().String(),
		RoleID:         "not-a-uuid",
	}
	accessInfo := &models.AccessInfo{User: &models.UserAccessInfo{UserID: uuid.New()}}
	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Delete: business logic ---

func (s *RoleServiceTestSuite) TestDelete_OrganizationNotFound() {
	orgID := uuid.New()
	roleID := uuid.New().String()
	req := &role_service.DeleteRequest{OrganizationID: orgID.String(), RoleID: roleID}
	accessInfo := &models.AccessInfo{User: &models.UserAccessInfo{UserID: uuid.New()}}
	s.orgRepo.On("GetByID", orgID.String()).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *RoleServiceTestSuite) TestDelete_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	org := s.newSubscribedOrg(orgID)
	req := &role_service.DeleteRequest{OrganizationID: orgID.String(), RoleID: uuid.New().String()}
	accessInfo := &models.AccessInfo{User: &models.UserAccessInfo{UserID: userID}}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)

	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
}

func (s *RoleServiceTestSuite) TestDelete_RoleNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	roleIDStr := uuid.New().String()
	org := s.newSubscribedOrg(orgID)
	req := &role_service.DeleteRequest{OrganizationID: orgID.String(), RoleID: roleIDStr}
	accessInfo := s.accessInfoWithRoleDelete(orgID, userID)
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.roleRepo.On("GetByID", orgID.String(), roleIDStr).Return((*models.Role)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
}

func (s *RoleServiceTestSuite) TestDelete_RoleLocked() {
	orgID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()
	org := s.newSubscribedOrg(orgID)
	role := &models.Role{
		ID:             roleID,
		OrganizationID: orgID,
		Name:           "Locked Role",
		Locked:         true,
	}
	req := &role_service.DeleteRequest{OrganizationID: orgID.String(), RoleID: roleID.String()}
	accessInfo := s.accessInfoWithRoleDelete(orgID, userID)
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.roleRepo.On("GetByID", orgID.String(), roleID.String()).Return(role, nil)

	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.ROLE_LOCKED_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
	s.memberRepo.AssertNotCalled(s.T(), "ListUserEmailsForRole")
	s.roleRepo.AssertNotCalled(s.T(), "Delete")
}

func (s *RoleServiceTestSuite) TestDelete_RoleHasMembers() {
	orgID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()
	org := s.newSubscribedOrg(orgID)
	role := &models.Role{
		ID:             roleID,
		OrganizationID: orgID,
		Name:           "In Use",
		Locked:         false,
	}
	req := &role_service.DeleteRequest{OrganizationID: orgID.String(), RoleID: roleID.String()}
	accessInfo := s.accessInfoWithRoleDelete(orgID, userID)
	emails := []string{"alpha@example.com", "beta@example.com"}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.roleRepo.On("GetByID", orgID.String(), roleID.String()).Return(role, nil)
	s.memberRepo.On("ListUserEmailsForRole", orgID.String(), roleID.String()).Return(emails, nil)

	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.ROLE_HAS_MEMBERS_CODE, e.ErrorCode)
		s.Equal(409, e.StatusCode)
		extra, ok := e.Extra.(map[string]any)
		s.True(ok)
		s.Equal(emails, extra["memberEmails"])
	}
	s.orgRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
	s.memberRepo.AssertExpectations(s.T())
	s.roleRepo.AssertNotCalled(s.T(), "Delete")
}

func (s *RoleServiceTestSuite) TestDelete_RepoDeleteFails() {
	orgID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()
	org := s.newSubscribedOrg(orgID)
	role := &models.Role{
		ID:             roleID,
		OrganizationID: orgID,
		Name:           "Deletable",
		Description:    "desc",
		Permissions:    models.RolePermissions{"read"},
		Locked:         false,
	}
	req := &role_service.DeleteRequest{OrganizationID: orgID.String(), RoleID: roleID.String()}
	accessInfo := s.accessInfoWithRoleDelete(orgID, userID)
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.roleRepo.On("GetByID", orgID.String(), roleID.String()).Return(role, nil)
	s.memberRepo.On("ListUserEmailsForRole", orgID.String(), roleID.String()).Return([]string{}, nil)
	s.roleRepo.On("Delete", nil, orgID.String(), roleID.String()).Return(errors.New("db error"))

	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
	s.memberRepo.AssertExpectations(s.T())
	s.auditLogSvc.AssertNotCalled(s.T(), "Create")
}

func (s *RoleServiceTestSuite) TestDelete_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()
	org := s.newSubscribedOrg(orgID)
	role := &models.Role{
		ID:             roleID,
		OrganizationID: orgID,
		Name:           "To Delete",
		Description:    "Role description",
		Permissions:    models.RolePermissions{"read", "write"},
		Locked:         false,
	}
	req := &role_service.DeleteRequest{OrganizationID: orgID.String(), RoleID: roleID.String()}
	accessInfo := s.accessInfoWithRoleDelete(orgID, userID)
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.roleRepo.On("GetByID", orgID.String(), roleID.String()).Return(role, nil)
	s.memberRepo.On("ListUserEmailsForRole", orgID.String(), roleID.String()).Return([]string{}, nil)
	s.roleRepo.On("Delete", nil, orgID.String(), roleID.String()).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	err := s.svc.Delete(req, accessInfo)
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
	s.memberRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
}

func (s *RoleServiceTestSuite) TestDelete_AuditLogFailureStillSucceeds() {
	orgID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()
	org := s.newSubscribedOrg(orgID)
	role := &models.Role{
		ID:             roleID,
		OrganizationID: orgID,
		Name:           "To Delete",
		Description:    "d",
		Permissions:    models.RolePermissions{"read"},
		Locked:         false,
	}
	req := &role_service.DeleteRequest{OrganizationID: orgID.String(), RoleID: roleID.String()}
	accessInfo := s.accessInfoWithRoleDelete(orgID, userID)
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.roleRepo.On("GetByID", orgID.String(), roleID.String()).Return(role, nil)
	s.memberRepo.On("ListUserEmailsForRole", orgID.String(), roleID.String()).Return([]string{}, nil)
	s.roleRepo.On("Delete", nil, orgID.String(), roleID.String()).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return((*models.AuditLog)(nil), errors.New("audit failed"))
	s.logger.On("Log", mock.MatchedBy(func(data yca_log.LogData) bool {
		return data.Level == "error" && data.Message == "Failed to create audit log"
	})).Return()

	err := s.svc.Delete(req, accessInfo)
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
	s.memberRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
	s.auditLogSvc.AssertExpectations(s.T())
	s.logger.AssertExpectations(s.T())
}
