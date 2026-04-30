package team_service_test

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
	team_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/team"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	team_service "github.com/yca-software/2chi-kit/go-api/internals/services/team"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
	yca_validate "github.com/yca-software/go-common/validator"
)

type TeamServiceTestSuite struct {
	suite.Suite
	svc          team_service.Service
	repos        *repositories.Repositories
	orgRepo      *organization_repository.MockRepository
	teamRepo     *team_repository.MockRepository
	auditLogRepo *audit_log_repository.MockRepository
	auditLogSvc  *audit_log_service.MockService
	logger       *yca_log.MockLogger
	authorizer   *helpers.Authorizer
	now          time.Time
}

func TestTeamServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TeamServiceTestSuite))
}

func (s *TeamServiceTestSuite) SetupTest() {
	s.now = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	s.orgRepo = organization_repository.NewMock()
	s.teamRepo = team_repository.NewMock()
	s.auditLogRepo = audit_log_repository.NewMock()
	s.auditLogSvc = audit_log_service.NewMockService()
	s.logger = &yca_log.MockLogger{}

	s.repos = &repositories.Repositories{
		Organization: s.orgRepo,
		Team:         s.teamRepo,
		AuditLog:     s.auditLogRepo,
	}

	s.authorizer = helpers.NewAuthorizer(func() time.Time { return s.now })

	s.svc = team_service.NewService(&team_service.Dependencies{
		GenerateID:      uuid.NewV7,
		Now:             func() time.Time { return s.now },
		Validator:       yca_validate.New(),
		Repositories:    s.repos,
		Authorizer:      s.authorizer,
		Logger:          s.logger,
		AuditLogService: s.auditLogSvc,
	})
}

func (s *TeamServiceTestSuite) subscribedProOrg(orgID uuid.UUID) *models.Organization {
	expiresAt := s.now.Add(24 * time.Hour)
	return &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_PRO,
		SubscriptionExpiresAt: &expiresAt,
	}
}

func (s *TeamServiceTestSuite) accessInfoWithTeamDelete(orgID uuid.UUID, userID uuid.UUID) *models.AccessInfo {
	return &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_TEAM_DELETE},
				},
			},
		},
	}
}

// --- Create: validations ---

func (s *TeamServiceTestSuite) TestCreate_Validation_InvalidOrganizationID() {
	req := &team_service.CreateRequest{
		OrganizationID: "not-a-uuid",
		Name:           "Test Team",
		Description:    "Test description",
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

func (s *TeamServiceTestSuite) TestCreate_Validation_MissingName() {
	req := &team_service.CreateRequest{
		OrganizationID: uuid.New().String(),
		Name:           "",
		Description:    "Test description",
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

func (s *TeamServiceTestSuite) TestCreate_OrganizationNotFound() {
	orgID := uuid.New()
	req := &team_service.CreateRequest{
		OrganizationID: orgID.String(),
		Name:           "Test Team",
		Description:    "Test description",
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

func (s *TeamServiceTestSuite) TestCreate_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_PRO,
		SubscriptionExpiresAt: &expiresAt,
	}
	req := &team_service.CreateRequest{
		OrganizationID: orgID.String(),
		Name:           "Test Team",
		Description:    "Test description",
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

func (s *TeamServiceTestSuite) TestCreate_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_PRO,
		SubscriptionExpiresAt: &expiresAt,
	}
	req := &team_service.CreateRequest{
		OrganizationID: orgID.String(),
		Name:           "  Test Team  ",
		Description:    "Test description",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_TEAM_WRITE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("Create", nil, mock.AnythingOfType("*models.Team")).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	resp, err := s.svc.Create(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal("Test Team", resp.Name) // Should be trimmed
	s.Equal(req.Description, resp.Description)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
}

// --- List: validations ---

func (s *TeamServiceTestSuite) TestList_Validation_InvalidUUID() {
	req := &team_service.ListRequest{
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

func (s *TeamServiceTestSuite) TestList_OrganizationNotFound() {
	orgID := uuid.New()
	req := &team_service.ListRequest{
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

func (s *TeamServiceTestSuite) TestList_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_PRO,
		SubscriptionExpiresAt: &expiresAt,
	}
	req := &team_service.ListRequest{
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

func (s *TeamServiceTestSuite) TestList_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_PRO,
		SubscriptionExpiresAt: &expiresAt,
	}
	req := &team_service.ListRequest{
		OrganizationID: orgID.String(),
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_TEAM_READ},
				},
			},
		},
	}
	teams := []models.Team{
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Name:           "Team 1",
			Description:    "Description 1",
		},
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Name:           "Team 2",
			Description:    "Description 2",
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("ListByOrganizationID", orgID.String()).Return(&teams, nil)

	resp, err := s.svc.List(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(*resp, 2)
	s.Equal(teams[0].Name, (*resp)[0].Name)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
}

// --- Update: validations ---

func (s *TeamServiceTestSuite) TestUpdate_Validation_InvalidOrganizationID() {
	req := &team_service.UpdateRequest{
		OrganizationID: "not-a-uuid",
		TeamID:         uuid.New().String(),
		Name:           "Updated Team",
		Description:    "Updated description",
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

func (s *TeamServiceTestSuite) TestUpdate_Validation_InvalidTeamID() {
	req := &team_service.UpdateRequest{
		OrganizationID: uuid.New().String(),
		TeamID:         "not-a-uuid",
		Name:           "Updated Team",
		Description:    "Updated description",
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

func (s *TeamServiceTestSuite) TestUpdate_Validation_MissingName() {
	req := &team_service.UpdateRequest{
		OrganizationID: uuid.New().String(),
		TeamID:         uuid.New().String(),
		Name:           "",
		Description:    "Updated description",
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

func (s *TeamServiceTestSuite) TestUpdate_OrganizationNotFound() {
	orgID := uuid.New()
	req := &team_service.UpdateRequest{
		OrganizationID: orgID.String(),
		TeamID:         uuid.New().String(),
		Name:           "Updated Team",
		Description:    "Updated description",
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

func (s *TeamServiceTestSuite) TestUpdate_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_PRO,
		SubscriptionExpiresAt: &expiresAt,
	}
	req := &team_service.UpdateRequest{
		OrganizationID: orgID.String(),
		TeamID:         uuid.New().String(),
		Name:           "Updated Team",
		Description:    "Updated description",
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

func (s *TeamServiceTestSuite) TestUpdate_TeamNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_PRO,
		SubscriptionExpiresAt: &expiresAt,
	}
	req := &team_service.UpdateRequest{
		OrganizationID: orgID.String(),
		TeamID:         uuid.New().String(),
		Name:           "Updated Team",
		Description:    "Updated description",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_TEAM_WRITE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), req.TeamID).Return((*models.Team)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	resp, err := s.svc.Update(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
}

func (s *TeamServiceTestSuite) TestUpdate_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	teamID := uuid.New()
	expiresAt := s.now.Add(24 * time.Hour)
	org := &models.Organization{
		ID:                    orgID,
		Name:                  "Test Org",
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_PRO,
		SubscriptionExpiresAt: &expiresAt,
	}
	team := &models.Team{
		ID:             teamID,
		OrganizationID: orgID,
		Name:           "Old Team",
		Description:    "Old description",
	}
	req := &team_service.UpdateRequest{
		OrganizationID: orgID.String(),
		TeamID:         teamID.String(),
		Name:           "  Updated Team  ",
		Description:    "Updated description",
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_TEAM_WRITE},
				},
			},
		},
	}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), teamID.String()).Return(team, nil)
	s.teamRepo.On("Update", nil, mock.AnythingOfType("*models.Team")).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	resp, err := s.svc.Update(req, accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Equal("Updated Team", resp.Name) // Should be trimmed
	s.Equal(req.Description, resp.Description)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
}

// --- Delete: validations ---

func (s *TeamServiceTestSuite) TestDelete_Validation_InvalidOrganizationID() {
	req := &team_service.DeleteRequest{
		OrganizationID: "not-a-uuid",
		TeamID:         uuid.New().String(),
	}
	accessInfo := &models.AccessInfo{User: &models.UserAccessInfo{UserID: uuid.New()}}
	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *TeamServiceTestSuite) TestDelete_Validation_InvalidTeamID() {
	req := &team_service.DeleteRequest{
		OrganizationID: uuid.New().String(),
		TeamID:         "not-a-uuid",
	}
	accessInfo := &models.AccessInfo{User: &models.UserAccessInfo{UserID: uuid.New()}}
	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Delete: business logic ---

func (s *TeamServiceTestSuite) TestDelete_OrganizationNotFound() {
	orgID := uuid.New()
	teamID := uuid.New().String()
	req := &team_service.DeleteRequest{OrganizationID: orgID.String(), TeamID: teamID}
	accessInfo := &models.AccessInfo{User: &models.UserAccessInfo{UserID: uuid.New()}}
	s.orgRepo.On("GetByID", orgID.String()).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *TeamServiceTestSuite) TestDelete_PermissionDenied() {
	orgID := uuid.New()
	userID := uuid.New()
	org := s.subscribedProOrg(orgID)
	req := &team_service.DeleteRequest{OrganizationID: orgID.String(), TeamID: uuid.New().String()}
	accessInfo := &models.AccessInfo{User: &models.UserAccessInfo{UserID: userID}}
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)

	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
}

func (s *TeamServiceTestSuite) TestDelete_TeamNotFound() {
	orgID := uuid.New()
	userID := uuid.New()
	teamIDStr := uuid.New().String()
	org := s.subscribedProOrg(orgID)
	req := &team_service.DeleteRequest{OrganizationID: orgID.String(), TeamID: teamIDStr}
	accessInfo := s.accessInfoWithTeamDelete(orgID, userID)
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), teamIDStr).Return((*models.Team)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
}

func (s *TeamServiceTestSuite) TestDelete_RepoDeleteFails() {
	orgID := uuid.New()
	userID := uuid.New()
	teamID := uuid.New()
	org := s.subscribedProOrg(orgID)
	team := &models.Team{
		ID:             teamID,
		OrganizationID: orgID,
		Name:           "Team",
		Description:    "desc",
	}
	req := &team_service.DeleteRequest{OrganizationID: orgID.String(), TeamID: teamID.String()}
	accessInfo := s.accessInfoWithTeamDelete(orgID, userID)
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), teamID.String()).Return(team, nil)
	s.teamRepo.On("Delete", nil, orgID.String(), teamID.String()).Return(errors.New("db error"))

	err := s.svc.Delete(req, accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
	s.auditLogSvc.AssertNotCalled(s.T(), "Create")
}

func (s *TeamServiceTestSuite) TestDelete_Success() {
	orgID := uuid.New()
	userID := uuid.New()
	teamID := uuid.New()
	org := s.subscribedProOrg(orgID)
	team := &models.Team{
		ID:             teamID,
		OrganizationID: orgID,
		Name:           "To Delete",
		Description:    "Team description",
	}
	req := &team_service.DeleteRequest{OrganizationID: orgID.String(), TeamID: teamID.String()}
	accessInfo := s.accessInfoWithTeamDelete(orgID, userID)
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), teamID.String()).Return(team, nil)
	s.teamRepo.On("Delete", nil, orgID.String(), teamID.String()).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return(&models.AuditLog{}, nil).Maybe()
	s.logger.On("Log", mock.Anything).Return().Maybe()

	err := s.svc.Delete(req, accessInfo)
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
}

func (s *TeamServiceTestSuite) TestDelete_AuditLogFailureStillSucceeds() {
	orgID := uuid.New()
	userID := uuid.New()
	teamID := uuid.New()
	org := s.subscribedProOrg(orgID)
	team := &models.Team{
		ID:             teamID,
		OrganizationID: orgID,
		Name:           "To Delete",
		Description:    "d",
	}
	req := &team_service.DeleteRequest{OrganizationID: orgID.String(), TeamID: teamID.String()}
	accessInfo := s.accessInfoWithTeamDelete(orgID, userID)
	s.orgRepo.On("GetByID", orgID.String()).Return(org, nil)
	s.teamRepo.On("GetByID", orgID.String(), teamID.String()).Return(team, nil)
	s.teamRepo.On("Delete", nil, orgID.String(), teamID.String()).Return(nil)
	s.auditLogSvc.On("Create", mock.Anything, accessInfo).Return((*models.AuditLog)(nil), errors.New("audit failed"))
	s.logger.On("Log", mock.MatchedBy(func(data yca_log.LogData) bool {
		return data.Level == "error" && data.Message == "Failed to create audit log"
	})).Return()

	err := s.svc.Delete(req, accessInfo)
	s.NoError(err)
	s.orgRepo.AssertExpectations(s.T())
	s.teamRepo.AssertExpectations(s.T())
	s.auditLogSvc.AssertExpectations(s.T())
	s.logger.AssertExpectations(s.T())
}
