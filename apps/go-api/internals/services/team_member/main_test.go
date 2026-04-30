package team_member_service_test

import (
	"database/sql"
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
	organization_member_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization_member"
	organization_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization"
	team_member_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/team_member"
	team_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/team"
	user_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	team_member_service "github.com/yca-software/2chi-kit/go-api/internals/services/team_member"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
	yca_repository "github.com/yca-software/go-common/repository"
	yca_validate "github.com/yca-software/go-common/validator"
)

type TeamMemberServiceTestSuite struct {
	suite.Suite
	svc          team_member_service.Service
	repos        *repositories.Repositories
	orgRepo      *organization_repository.MockRepository
	teamRepo     *team_repository.MockRepository
	teamMemberRepo *team_member_repository.MockRepository
	orgMemberRepo  *organization_member_repository.MockRepository
	userRepo       *user_repository.MockRepository
	auditLogRepo   *audit_log_repository.MockRepository
	auditLogSvc    *audit_log_service.MockService
	logger         *yca_log.MockLogger
	authorizer     *helpers.Authorizer
	now            time.Time
}

func TestTeamMemberServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TeamMemberServiceTestSuite))
}

func (s *TeamMemberServiceTestSuite) SetupTest() {
	s.now = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	s.orgRepo = organization_repository.NewMock()
	s.teamRepo = team_repository.NewMock()
	s.teamMemberRepo = team_member_repository.NewMock()
	s.orgMemberRepo = organization_member_repository.NewMock()
	s.userRepo = user_repository.NewMock()
	s.auditLogRepo = audit_log_repository.NewMock()
	s.auditLogSvc = audit_log_service.NewMockService()
	s.logger = &yca_log.MockLogger{}

	s.repos = &repositories.Repositories{
		Organization:       s.orgRepo,
		Team:               s.teamRepo,
		TeamMember:         s.teamMemberRepo,
		OrganizationMember: s.orgMemberRepo,
		User:               s.userRepo,
		AuditLog:           s.auditLogRepo,
	}

	s.authorizer = helpers.NewAuthorizer(func() time.Time { return s.now })

	s.svc = team_member_service.NewService(&team_member_service.Dependencies{
		GenerateID:      uuid.NewV7,
		Now:             func() time.Time { return s.now },
		Validator:       yca_validate.New(),
		Repositories:    s.repos,
		Authorizer:      s.authorizer,
		Logger:          s.logger,
		AuditLogService: s.auditLogSvc,
	})
}

// --- Add: validations ---

func (s *TeamMemberServiceTestSuite) TestAdd_Validation_InvalidOrganizationID() {
	req := &team_member_service.AddRequest{
		OrganizationID: "not-a-uuid",
		TeamID:         uuid.New().String(),
		UserID:         uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	resp, err := s.svc.Add(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *TeamMemberServiceTestSuite) TestAdd_Validation_InvalidTeamID() {
	req := &team_member_service.AddRequest{
		OrganizationID: uuid.New().String(),
		TeamID:         "not-a-uuid",
		UserID:         uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	resp, err := s.svc.Add(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *TeamMemberServiceTestSuite) TestAdd_Validation_InvalidUserID() {
	req := &team_member_service.AddRequest{
		OrganizationID: uuid.New().String(),
		TeamID:         uuid.New().String(),
		UserID:         "not-a-uuid",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	resp, err := s.svc.Add(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Add: business logic ---

func (s *TeamMemberServiceTestSuite) TestAdd_OrganizationNotFound() {
	orgID := uuid.New()
	req := &team_member_service.AddRequest{
		OrganizationID: orgID.String(),
		TeamID:         uuid.New().String(),
		UserID:         uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.Add(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *TeamMemberServiceTestSuite) TestAdd_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &expiresAt,
	}
	req := &team_member_service.AddRequest{
		OrganizationID: orgID.String(),
		TeamID:         uuid.New().String(),
		UserID:         uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)

	resp, err := s.svc.Add(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
}

func (s *TeamMemberServiceTestSuite) TestAdd_TeamNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &expiresAt,
	}
	req := &team_member_service.AddRequest{
		OrganizationID: orgID.String(),
		TeamID:         uuid.New().String(),
		UserID:         uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_TEAM_MEMBER_WRITE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), req.TeamID).Return((*models.Team)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.Add(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
}

func (s *TeamMemberServiceTestSuite) TestAdd_UserNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	teamID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &expiresAt,
	}
	team := &models.Team{
		ID:             teamID,
		OrganizationID: orgID,
		Name:           "Test Team",
	}
	req := &team_member_service.AddRequest{
		OrganizationID: orgID.String(),
		TeamID:         teamID.String(),
		UserID:         uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_TEAM_MEMBER_WRITE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), teamID.String()).Return(team, nil)
	s.userRepo.On("GetByID", nil, req.UserID).Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.Add(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
	s.userRepo.AssertExpectations(s.T())
}

func (s *TeamMemberServiceTestSuite) TestAdd_OrganizationMemberNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	teamID := uuid.New()
	memberUserID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &expiresAt,
	}
	team := &models.Team{
		ID:             teamID,
		OrganizationID: orgID,
		Name:           "Test Team",
	}
	user := &models.User{
		ID:    memberUserID,
		Email: "member@example.com",
	}
	req := &team_member_service.AddRequest{
		OrganizationID: orgID.String(),
		TeamID:         teamID.String(),
		UserID:         memberUserID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_TEAM_MEMBER_WRITE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), teamID.String()).Return(team, nil)
	s.userRepo.On("GetByID", nil, memberUserID.String()).Return(user, nil)
	s.orgMemberRepo.On("GetByUserIDAndOrganizationID", memberUserID.String(), orgID.String()).Return((*models.OrganizationMember)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.Add(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
	s.userRepo.AssertExpectations(s.T())
	s.orgMemberRepo.AssertExpectations(s.T())
}

func (s *TeamMemberServiceTestSuite) TestAdd_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	teamID := uuid.New()
	memberUserID := uuid.New()
	memberID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &expiresAt,
	}
	team := &models.Team{
		ID:             teamID,
		OrganizationID: orgID,
		Name:           "Test Team",
	}
	user := &models.User{
		ID:    memberUserID,
		Email: "member@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}
	orgMember := &models.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         memberUserID,
	}
	memberWithUser := &models.TeamMemberWithUser{
		TeamMember: models.TeamMember{
			ID:             memberID,
			CreatedAt:      s.now,
			OrganizationID: orgID,
			TeamID:         teamID,
			UserID:         memberUserID,
		},
		UserEmail:     user.Email,
		UserFirstName: user.FirstName,
		UserLastName:  user.LastName,
	}
	req := &team_member_service.AddRequest{
		OrganizationID: orgID.String(),
		TeamID:         teamID.String(),
		UserID:         memberUserID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_TEAM_MEMBER_WRITE},
				},
			},
		},
	}
	tx := yca_repository.NewMockTx()
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), teamID.String()).Return(team, nil)
	s.userRepo.On("GetByID", nil, memberUserID.String()).Return(user, nil)
	s.orgMemberRepo.On("GetByUserIDAndOrganizationID", memberUserID.String(), orgID.String()).Return(orgMember, nil)
	s.teamRepo.On("BeginTx").Return(tx, nil)
	s.teamMemberRepo.On("Create", tx, mock.AnythingOfType("*models.TeamMember")).Return(nil)
	s.teamMemberRepo.On("GetByIDWithUser", tx, orgID.String(), mock.AnythingOfType("string")).Return(memberWithUser, nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(sql.ErrTxDone).Maybe()
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	resp, err := s.svc.Add(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal(memberID, resp.ID)
	s.Equal(memberUserID, resp.UserID)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
	s.userRepo.AssertExpectations(s.T())
	s.orgMemberRepo.AssertExpectations(s.T())
	s.teamMemberRepo.AssertExpectations(s.T())
}

// --- ListByTeam: validations ---

func (s *TeamMemberServiceTestSuite) TestListByTeam_Validation_InvalidOrganizationID() {
	req := &team_member_service.ListByTeamRequest{
		OrganizationID: "not-a-uuid",
		TeamID:         uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	resp, err := s.svc.ListByTeam(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *TeamMemberServiceTestSuite) TestListByTeam_Validation_InvalidTeamID() {
	req := &team_member_service.ListByTeamRequest{
		OrganizationID: uuid.New().String(),
		TeamID:         "not-a-uuid",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	resp, err := s.svc.ListByTeam(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- ListByTeam: business logic ---

func (s *TeamMemberServiceTestSuite) TestListByTeam_OrganizationNotFound() {
	orgID := uuid.New()
	req := &team_member_service.ListByTeamRequest{
		OrganizationID: orgID.String(),
		TeamID:         uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.ListByTeam(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *TeamMemberServiceTestSuite) TestListByTeam_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &expiresAt,
	}
	req := &team_member_service.ListByTeamRequest{
		OrganizationID: orgID.String(),
		TeamID:         uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)

	resp, err := s.svc.ListByTeam(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
}

func (s *TeamMemberServiceTestSuite) TestListByTeam_TeamNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &expiresAt,
	}
	req := &team_member_service.ListByTeamRequest{
		OrganizationID: orgID.String(),
		TeamID:         uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_TEAM_MEMBER_READ},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), req.TeamID).Return((*models.Team)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.ListByTeam(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
}

func (s *TeamMemberServiceTestSuite) TestListByTeam_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	teamID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &expiresAt,
	}
	team := &models.Team{
		ID:             teamID,
		OrganizationID: orgID,
		Name:           "Test Team",
	}
	members := []models.TeamMemberWithUser{
		{
			TeamMember: models.TeamMember{
				ID:             uuid.New(),
				OrganizationID: orgID,
				TeamID:         teamID,
				UserID:         uuid.New(),
			},
			UserEmail:     "user1@example.com",
			UserFirstName: "John",
			UserLastName:  "Doe",
		},
		{
			TeamMember: models.TeamMember{
				ID:             uuid.New(),
				OrganizationID: orgID,
				TeamID:         teamID,
				UserID:         uuid.New(),
			},
			UserEmail:     "user2@example.com",
			UserFirstName: "Jane",
			UserLastName:  "Smith",
		},
	}
	req := &team_member_service.ListByTeamRequest{
		OrganizationID: orgID.String(),
		TeamID:         teamID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_TEAM_MEMBER_READ},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), teamID.String()).Return(team, nil)
	s.teamMemberRepo.On("ListByTeamID", orgID.String(), teamID.String()).Return(&members, nil)

	resp, err := s.svc.ListByTeam(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(*resp, 2)
	s.Equal(members[0].UserEmail, (*resp)[0].UserEmail)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
	s.teamMemberRepo.AssertExpectations(s.T())
}

// --- Remove: validations ---

func (s *TeamMemberServiceTestSuite) TestRemove_Validation_InvalidOrganizationID() {
	req := &team_member_service.RemoveRequest{
		OrganizationID: "not-a-uuid",
		TeamID:         uuid.New().String(),
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

func (s *TeamMemberServiceTestSuite) TestRemove_Validation_InvalidTeamID() {
	req := &team_member_service.RemoveRequest{
		OrganizationID: uuid.New().String(),
		TeamID:         "not-a-uuid",
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

func (s *TeamMemberServiceTestSuite) TestRemove_Validation_InvalidMemberID() {
	req := &team_member_service.RemoveRequest{
		OrganizationID: uuid.New().String(),
		TeamID:         uuid.New().String(),
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

func (s *TeamMemberServiceTestSuite) TestRemove_OrganizationNotFound() {
	orgID := uuid.New()
	req := &team_member_service.RemoveRequest{
		OrganizationID: orgID.String(),
		TeamID:         uuid.New().String(),
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

func (s *TeamMemberServiceTestSuite) TestRemove_TeamNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	org := &models.Organization{
		ID:   orgID,
		Name: "Test Org",
	}
	req := &team_member_service.RemoveRequest{
		OrganizationID: orgID.String(),
		TeamID:         uuid.New().String(),
		MemberID:       uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_TEAM_MEMBER_DELETE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), req.TeamID).Return((*models.Team)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	err := s.svc.Remove(req, accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
}

func (s *TeamMemberServiceTestSuite) TestRemove_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &expiresAt,
	}
	team := &models.Team{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           "Test Team",
	}
	req := &team_member_service.RemoveRequest{
		OrganizationID: orgID.String(),
		TeamID:         team.ID.String(),
		MemberID:       uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), team.ID.String()).Return(team, nil)

	err := s.svc.Remove(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
}

func (s *TeamMemberServiceTestSuite) TestRemove_MemberNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &expiresAt,
	}
	team := &models.Team{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           "Test Team",
	}
	req := &team_member_service.RemoveRequest{
		OrganizationID: orgID.String(),
		TeamID:         team.ID.String(),
		MemberID:       uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_TEAM_MEMBER_DELETE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), team.ID.String()).Return(team, nil)
	s.teamMemberRepo.On("GetByID", orgID.String(), req.MemberID).Return((*models.TeamMember)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	err := s.svc.Remove(req, accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
	s.teamMemberRepo.AssertExpectations(s.T())
}

func (s *TeamMemberServiceTestSuite) TestRemove_UserNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	teamID := uuid.New()
	memberID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &expiresAt,
	}
	team := &models.Team{
		ID:             teamID,
		OrganizationID: orgID,
		Name:           "Test Team",
	}
	member := &models.TeamMember{
		ID:             memberID,
		OrganizationID: orgID,
		TeamID:         teamID,
		UserID:         uuid.New(),
	}
	req := &team_member_service.RemoveRequest{
		OrganizationID: orgID.String(),
		TeamID:         teamID.String(),
		MemberID:       memberID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_TEAM_MEMBER_DELETE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), teamID.String()).Return(team, nil)
	s.teamMemberRepo.On("GetByID", orgID.String(), memberID.String()).Return(member, nil)
	s.userRepo.On("GetByID", nil, member.UserID.String()).Return((*models.User)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	err := s.svc.Remove(req, accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
	s.teamMemberRepo.AssertExpectations(s.T())
	s.userRepo.AssertExpectations(s.T())
}

func (s *TeamMemberServiceTestSuite) TestRemove_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	teamID := uuid.New()
	memberID := uuid.New()
	memberUserID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &expiresAt,
	}
	team := &models.Team{
		ID:             teamID,
		OrganizationID: orgID,
		Name:           "Test Team",
	}
	member := &models.TeamMember{
		ID:             memberID,
		OrganizationID: orgID,
		TeamID:         teamID,
		UserID:         memberUserID,
	}
	user := &models.User{
		ID:    memberUserID,
		Email: "member@example.com",
	}
	req := &team_member_service.RemoveRequest{
		OrganizationID: orgID.String(),
		TeamID:         teamID.String(),
		MemberID:       memberID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_TEAM_MEMBER_DELETE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), teamID.String()).Return(team, nil)
	s.teamMemberRepo.On("GetByID", orgID.String(), memberID.String()).Return(member, nil)
	s.userRepo.On("GetByID", nil, memberUserID.String()).Return(user, nil)
	s.teamMemberRepo.On("Delete", nil, orgID.String(), memberID.String()).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	err := s.svc.Remove(req, accessInfo)
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
	s.teamMemberRepo.AssertExpectations(s.T())
	s.userRepo.AssertExpectations(s.T())
}
