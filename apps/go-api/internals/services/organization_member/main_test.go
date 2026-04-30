package organization_member_service_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	"github.com/yca-software/2chi-kit/go-api/internals/helpers"
	"github.com/yca-software/2chi-kit/go-api/internals/repositories"
	audit_log_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/audit_log"
	organization_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization"
	organization_member_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization_member"
	role_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/role"
	user_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	organization_member_service "github.com/yca-software/2chi-kit/go-api/internals/services/organization_member"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type OrganizationMemberServiceTestSuite struct {
	suite.Suite
	svc           organization_member_service.Service
	repos         *repositories.Repositories
	orgRepo       *organization_repository.MockRepository
	orgMemberRepo *organization_member_repository.MockRepository
	roleRepo      *role_repository.MockRepository
	userRepo      *user_repository.MockRepository
	auditLogRepo  *audit_log_repository.MockRepository
	auditLogSvc   *audit_log_service.MockService
	logger        *yca_log.MockLogger
	authorizer    *helpers.Authorizer
	now           time.Time
}

func TestOrganizationMemberServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationMemberServiceTestSuite))
}

// orgWithActiveSubscription returns an org that passes CheckOrganizationPermissionWithSubscription (non-free, future expiry).
func (s *OrganizationMemberServiceTestSuite) orgWithActiveSubscription(orgID uuid.UUID, name string) *models.Organization {
	expiresAt := s.now.Add(30 * 24 * time.Hour)
	return &models.Organization{
		ID:                   orgID,
		Name:                 name,
		SubscriptionType:     constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &expiresAt,
	}
}

func (s *OrganizationMemberServiceTestSuite) SetupTest() {
	s.now = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	s.orgRepo = organization_repository.NewMock()
	s.orgMemberRepo = organization_member_repository.NewMock()
	s.roleRepo = role_repository.NewMock()
	s.userRepo = user_repository.NewMock()
	s.auditLogRepo = audit_log_repository.NewMock()
	s.auditLogSvc = audit_log_service.NewMockService()
	s.logger = &yca_log.MockLogger{}

	s.repos = &repositories.Repositories{
		Organization:       s.orgRepo,
		OrganizationMember: s.orgMemberRepo,
		Role:               s.roleRepo,
		User:               s.userRepo,
		AuditLog:           s.auditLogRepo,
	}

	s.authorizer = helpers.NewAuthorizer(func() time.Time { return s.now })

	s.svc = organization_member_service.NewService(&organization_member_service.Dependencies{
		Validator:       yca_validate.New(),
		Repositories:    s.repos,
		Authorizer:      s.authorizer,
		Logger:          s.logger,
		AuditLogService: s.auditLogSvc,
	})
}

// --- ListByOrganization: validations ---

func (s *OrganizationMemberServiceTestSuite) TestListByOrganization_Validation_InvalidUUID() {
	req := &organization_member_service.ListByOrganizationRequest{
		OrganizationID: "not-a-uuid",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	resp, err := s.svc.ListByOrganization(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- ListByOrganization: business logic ---

func (s *OrganizationMemberServiceTestSuite) TestListByOrganization_NotFound() {
	orgID := uuid.New()
	req := &organization_member_service.ListByOrganizationRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.ListByOrganization(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *OrganizationMemberServiceTestSuite) TestListByOrganization_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	org := &models.Organization{
		ID:   orgID,
		Name: "Test Org",
	}
	req := &organization_member_service.ListByOrganizationRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)

	resp, err := s.svc.ListByOrganization(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
}

func (s *OrganizationMemberServiceTestSuite) TestListByOrganization_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	org := &models.Organization{
		ID:   orgID,
		Name: "Test Org",
	}
	req := &organization_member_service.ListByOrganizationRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_MEMBERS_READ},
				},
			},
		},
	}
	members := []models.OrganizationMemberWithUser{
		{
			OrganizationMember: models.OrganizationMember{
				ID:             uuid.New(),
				OrganizationID: orgID,
				UserID:         uuid.New(),
				RoleID:         uuid.New(),
			},
			UserEmail:     "member@example.com",
			UserFirstName: "John",
			UserLastName:  "Doe",
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.orgMemberRepo.On("ListByOrganizationID", orgID.String()).Return(&members, nil)

	resp, err := s.svc.ListByOrganization(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(*resp, 1)
	s.Equal(members[0].UserEmail, (*resp)[0].UserEmail)
	s.orgRepo.AssertExpectations(s.T())
	s.orgMemberRepo.AssertExpectations(s.T())
}

// --- ListByUser: validations ---

func (s *OrganizationMemberServiceTestSuite) TestListByUser_Validation_InvalidUUID() {
	req := &organization_member_service.ListByUserRequest{
		UserID: "not-a-uuid",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	resp, err := s.svc.ListByUser(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- ListByUser: business logic ---

func (s *OrganizationMemberServiceTestSuite) TestListByUser_NotOwnResource() {
	userID := uuid.New()
	otherUserID := uuid.New()
	req := &organization_member_service.ListByUserRequest{
		UserID: otherUserID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}

	resp, err := s.svc.ListByUser(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
}

func (s *OrganizationMemberServiceTestSuite) TestListByUser_Success() {
	userID := uuid.New()
	req := &organization_member_service.ListByUserRequest{
		UserID: userID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}
	members := []models.OrganizationMemberWithOrganization{
		{
			OrganizationMember: models.OrganizationMember{
				ID:             uuid.New(),
				OrganizationID: uuid.New(),
				UserID:         userID,
				RoleID:         uuid.New(),
			},
			OrganizationName: "Test Org",
		},
	}
	s.orgMemberRepo.On("ListByUserID", userID.String()).Return(&members, nil)

	resp, err := s.svc.ListByUser(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(*resp, 1)
	s.Equal(members[0].OrganizationName, (*resp)[0].OrganizationName)
	s.orgMemberRepo.AssertExpectations(s.T())
}

// --- Update: validations ---

func (s *OrganizationMemberServiceTestSuite) TestUpdate_Validation_InvalidOrganizationID() {
	req := &organization_member_service.UpdateRequest{
		OrganizationID: "not-a-uuid",
		MemberID:       uuid.New().String(),
		RoleID:         uuid.New().String(),
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

func (s *OrganizationMemberServiceTestSuite) TestUpdate_Validation_InvalidMemberID() {
	req := &organization_member_service.UpdateRequest{
		OrganizationID: uuid.New().String(),
		MemberID:       "not-a-uuid",
		RoleID:         uuid.New().String(),
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

func (s *OrganizationMemberServiceTestSuite) TestUpdate_Validation_InvalidRoleID() {
	req := &organization_member_service.UpdateRequest{
		OrganizationID: uuid.New().String(),
		MemberID:       uuid.New().String(),
		RoleID:         "not-a-uuid",
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

func (s *OrganizationMemberServiceTestSuite) TestUpdate_OrganizationNotFound() {
	orgID := uuid.New()
	req := &organization_member_service.UpdateRequest{
		OrganizationID: orgID.String(),
		MemberID:       uuid.New().String(),
		RoleID:         uuid.New().String(),
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

func (s *OrganizationMemberServiceTestSuite) TestUpdate_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	org := s.orgWithActiveSubscription(orgID, "Test Org")
	req := &organization_member_service.UpdateRequest{
		OrganizationID: orgID.String(),
		MemberID:       uuid.New().String(),
		RoleID:         uuid.New().String(),
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

func (s *OrganizationMemberServiceTestSuite) TestUpdate_MemberNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	org := s.orgWithActiveSubscription(orgID, "Test Org")
	req := &organization_member_service.UpdateRequest{
		OrganizationID: orgID.String(),
		MemberID:       uuid.New().String(),
		RoleID:         uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_MEMBERS_WRITE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.orgMemberRepo.On("GetByID", orgID.String(), req.MemberID).Return((*models.OrganizationMember)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.Update(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
	s.orgMemberRepo.AssertExpectations(s.T())
}

func (s *OrganizationMemberServiceTestSuite) TestUpdate_CannotUpdateOwnMember() {
	orgID := uuid.New()
	userID := uuid.New()
	memberID := uuid.New()
	org := s.orgWithActiveSubscription(orgID, "Test Org")
	member := &models.OrganizationMember{
		ID:             memberID,
		OrganizationID: orgID,
		UserID:         userID,
		RoleID:         uuid.New(),
	}
	req := &organization_member_service.UpdateRequest{
		OrganizationID: orgID.String(),
		MemberID:       memberID.String(),
		RoleID:         uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_MEMBERS_WRITE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.orgMemberRepo.On("GetByID", orgID.String(), memberID.String()).Return(member, nil)

	resp, err := s.svc.Update(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.USER_CANNOT_UPDATE_OWN_MEMBER_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
	s.orgMemberRepo.AssertExpectations(s.T())
}

func (s *OrganizationMemberServiceTestSuite) TestUpdate_RoleNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	memberID := uuid.New()
	org := s.orgWithActiveSubscription(orgID, "Test Org")
	member := &models.OrganizationMember{
		ID:             memberID,
		OrganizationID: orgID,
		UserID:         uuid.New(),
		RoleID:         uuid.New(),
	}
	req := &organization_member_service.UpdateRequest{
		OrganizationID: orgID.String(),
		MemberID:       memberID.String(),
		RoleID:         uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_MEMBERS_WRITE},
				},
			},
		},
	}
	currentRole := &models.Role{
		ID:   member.RoleID,
		Name: "Current Role",
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.orgMemberRepo.On("GetByID", orgID.String(), memberID.String()).Return(member, nil)
	s.roleRepo.On("GetByID", orgID.String(), member.RoleID.String()).Return(currentRole, nil)
	s.roleRepo.On("GetByID", orgID.String(), req.RoleID).Return((*models.Role)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.Update(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
	s.orgMemberRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
}

func (s *OrganizationMemberServiceTestSuite) TestUpdate_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	memberID := uuid.New()
	org := s.orgWithActiveSubscription(orgID, "Test Org")
	member := &models.OrganizationMember{
		ID:             memberID,
		OrganizationID: orgID,
		UserID:         uuid.New(),
		RoleID:         uuid.New(),
	}
	newRoleID := uuid.New()
	req := &organization_member_service.UpdateRequest{
		OrganizationID: orgID.String(),
		MemberID:       memberID.String(),
		RoleID:         newRoleID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_MEMBERS_WRITE},
				},
			},
		},
	}
	currentRole := &models.Role{
		ID:   member.RoleID,
		Name: "Current Role",
	}
	newRole := &models.Role{
		ID:   newRoleID,
		Name: "New Role",
	}
	memberWithUser := &models.OrganizationMemberWithUser{
		OrganizationMember: *member,
		UserEmail:          "member@example.com",
		UserFirstName:      "John",
		UserLastName:       "Doe",
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.orgMemberRepo.On("GetByID", orgID.String(), memberID.String()).Return(member, nil)
	s.roleRepo.On("GetByID", orgID.String(), member.RoleID.String()).Return(currentRole, nil)
	s.roleRepo.On("GetByID", orgID.String(), newRoleID.String()).Return(newRole, nil)
	s.orgMemberRepo.On("Update", nil, mock.AnythingOfType("*models.OrganizationMember")).Return(nil)
	s.orgMemberRepo.On("GetByIDWithUser", orgID.String(), memberID.String()).Return(memberWithUser, nil)
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	resp, err := s.svc.Update(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal(memberWithUser.UserEmail, resp.UserEmail)
	s.orgRepo.AssertExpectations(s.T())
	s.orgMemberRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
}

// --- Remove: validations ---

func (s *OrganizationMemberServiceTestSuite) TestRemove_Validation_InvalidOrganizationID() {
	req := &organization_member_service.RemoveRequest{
		OrganizationID: "not-a-uuid",
		MemberID:       uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	err := s.svc.Remove(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *OrganizationMemberServiceTestSuite) TestRemove_Validation_InvalidMemberID() {
	req := &organization_member_service.RemoveRequest{
		OrganizationID: uuid.New().String(),
		MemberID:       "not-a-uuid",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	err := s.svc.Remove(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Remove: business logic ---

func (s *OrganizationMemberServiceTestSuite) TestRemove_OrganizationNotFound() {
	orgID := uuid.New()
	req := &organization_member_service.RemoveRequest{
		OrganizationID: orgID.String(),
		MemberID:       uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	err := s.svc.Remove(req, accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *OrganizationMemberServiceTestSuite) TestRemove_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	org := s.orgWithActiveSubscription(orgID, "Test Org")
	req := &organization_member_service.RemoveRequest{
		OrganizationID: orgID.String(),
		MemberID:       uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)

	err := s.svc.Remove(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
}

func (s *OrganizationMemberServiceTestSuite) TestRemove_MemberNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	org := s.orgWithActiveSubscription(orgID, "Test Org")
	req := &organization_member_service.RemoveRequest{
		OrganizationID: orgID.String(),
		MemberID:       uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_MEMBERS_DELETE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.orgMemberRepo.On("GetByID", orgID.String(), req.MemberID).Return((*models.OrganizationMember)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	err := s.svc.Remove(req, accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
	s.orgMemberRepo.AssertExpectations(s.T())
}

func (s *OrganizationMemberServiceTestSuite) TestRemove_CannotRemoveOwnMember() {
	orgID := uuid.New()
	userID := uuid.New()
	memberID := uuid.New()
	org := s.orgWithActiveSubscription(orgID, "Test Org")
	member := &models.OrganizationMember{
		ID:             memberID,
		OrganizationID: orgID,
		UserID:         userID,
		RoleID:         uuid.New(),
	}
	req := &organization_member_service.RemoveRequest{
		OrganizationID: orgID.String(),
		MemberID:       memberID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_MEMBERS_DELETE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.orgMemberRepo.On("GetByID", orgID.String(), memberID.String()).Return(member, nil)

	err := s.svc.Remove(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.USER_CANNOT_REMOVE_OWN_MEMBER_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
	s.orgMemberRepo.AssertExpectations(s.T())
}

func (s *OrganizationMemberServiceTestSuite) TestRemove_UserNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	memberID := uuid.New()
	org := s.orgWithActiveSubscription(orgID, "Test Org")
	member := &models.OrganizationMember{
		ID:             memberID,
		OrganizationID: orgID,
		UserID:         uuid.New(),
		RoleID:         uuid.New(),
	}
	req := &organization_member_service.RemoveRequest{
		OrganizationID: orgID.String(),
		MemberID:       memberID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_MEMBERS_DELETE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.orgMemberRepo.On("GetByID", orgID.String(), memberID.String()).Return(member, nil)
	s.userRepo.On("GetByID", nil, member.UserID.String()).Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	err := s.svc.Remove(req, accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
	s.orgMemberRepo.AssertExpectations(s.T())
	s.userRepo.AssertExpectations(s.T())
}

func (s *OrganizationMemberServiceTestSuite) TestRemove_RoleNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	memberID := uuid.New()
	org := s.orgWithActiveSubscription(orgID, "Test Org")
	member := &models.OrganizationMember{
		ID:             memberID,
		OrganizationID: orgID,
		UserID:         uuid.New(),
		RoleID:         uuid.New(),
	}
	req := &organization_member_service.RemoveRequest{
		OrganizationID: orgID.String(),
		MemberID:       memberID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_MEMBERS_DELETE},
				},
			},
		},
	}
	user := &models.User{
		ID:    member.UserID,
		Email: "user@example.com",
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.orgMemberRepo.On("GetByID", orgID.String(), memberID.String()).Return(member, nil)
	s.userRepo.On("GetByID", nil, member.UserID.String()).Return(user, nil)
	s.roleRepo.On("GetByID", orgID.String(), member.RoleID.String()).Return((*models.Role)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	err := s.svc.Remove(req, accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
	s.orgMemberRepo.AssertExpectations(s.T())
	s.userRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
}

func (s *OrganizationMemberServiceTestSuite) TestRemove_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	memberID := uuid.New()
	org := s.orgWithActiveSubscription(orgID, "Test Org")
	member := &models.OrganizationMember{
		ID:             memberID,
		OrganizationID: orgID,
		UserID:         uuid.New(),
		RoleID:         uuid.New(),
	}
	req := &organization_member_service.RemoveRequest{
		OrganizationID: orgID.String(),
		MemberID:       memberID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_MEMBERS_DELETE},
				},
			},
		},
	}
	user := &models.User{
		ID:    member.UserID,
		Email: "user@example.com",
	}
	role := &models.Role{
		ID:   member.RoleID,
		Name: "Test Role",
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.orgMemberRepo.On("GetByID", orgID.String(), memberID.String()).Return(member, nil)
	s.userRepo.On("GetByID", nil, member.UserID.String()).Return(user, nil)
	s.roleRepo.On("GetByID", orgID.String(), member.RoleID.String()).Return(role, nil)
	s.orgMemberRepo.On("Delete", nil, orgID.String(), memberID.String()).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	err := s.svc.Remove(req, accessInfo)
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
	s.orgMemberRepo.AssertExpectations(s.T())
	s.userRepo.AssertExpectations(s.T())
	s.roleRepo.AssertExpectations(s.T())
}
