package invitation_service_test

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
	invitation_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/invitation"
	organization_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization"
	organization_member_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/organization_member"
	user_repository "github.com/yca-software/2chi-kit/go-api/internals/repositories/user"
	invitation_service "github.com/yca-software/2chi-kit/go-api/internals/services/invitation"
	yca_email "github.com/yca-software/go-common/email"
	yca_error "github.com/yca-software/go-common/error"
	yca_translate "github.com/yca-software/go-common/localizer"
	yca_log "github.com/yca-software/go-common/logger"
	yca_repository "github.com/yca-software/go-common/repository"
	yca_validate "github.com/yca-software/go-common/validator"
)

type InvitationServiceTestSuite struct {
	suite.Suite
	svc            invitation_service.Service
	repos          *repositories.Repositories
	orgRepo        *organization_repository.MockRepository
	invitationRepo *invitation_repository.MockRepository
	orgMemberRepo  *organization_member_repository.MockRepository
	userRepo       *user_repository.MockRepository
	emailSvc       *yca_email.MockEmailService
	translator     *yca_translate.MockTranslator
	logger         *yca_log.MockLogger
	now            time.Time
	accessInfo     *models.AccessInfo
}

func TestInvitationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(InvitationServiceTestSuite))
}

func (s *InvitationServiceTestSuite) SetupTest() {
	s.now = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	s.orgRepo = organization_repository.NewMock()
	s.invitationRepo = invitation_repository.NewMock()
	s.orgMemberRepo = organization_member_repository.NewMock()
	s.userRepo = user_repository.NewMock()
	s.emailSvc = yca_email.NewMockEmailService()
	s.translator = yca_translate.NewMockTranslator()
	s.translator.On("Translate", mock.Anything, mock.Anything, mock.Anything).Return("translated").Maybe()
	s.translator.On("Translate", mock.Anything, mock.Anything, mock.AnythingOfType("map[string]interface {}")).Return("translated").Maybe()
	s.logger = &yca_log.MockLogger{}

	s.repos = &repositories.Repositories{
		Organization:       s.orgRepo,
		Invitation:         s.invitationRepo,
		OrganizationMember: s.orgMemberRepo,
		User:               s.userRepo,
	}

	s.accessInfo = &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			Email:   "admin@example.com",
			IsAdmin: true,
		},
	}

	s.svc = invitation_service.New(&invitation_service.Dependencies{
		Validator:     yca_validate.New(),
		Authorizer:    helpers.NewAuthorizer(func() time.Time { return s.now }),
		Repos:         s.repos,
		Logger:        s.logger,
		EmailService:  s.emailSvc,
		Now:           func() time.Time { return s.now },
		GenerateID:    uuid.NewV7,
		HashToken:     func(token string) string { return "hashed:" + token },
		GenerateToken: func() (string, error) { return "invite-token-123", nil },
		Translator:    s.translator,
	})
}

// --- List ---

func (s *InvitationServiceTestSuite) TestList_OrganizationNotFound() {
	orgID := uuid.New().String()
	s.orgRepo.On("GetByID", orgID).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &invitation_service.ListRequest{OrganizationID: orgID}
	resp, err := s.svc.List(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *InvitationServiceTestSuite) TestList_PermissionDenied() {
	orgID := uuid.New().String()
	org := &models.Organization{ID: uuid.MustParse(orgID), Name: "Test Org", SubscriptionSeats: 10}
	s.orgRepo.On("GetByID", orgID).Return(org, nil)

	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			IsAdmin: false,
			Roles:   []models.JWTAccessTokenPermissionData{}, // no permission
		},
	}

	req := &invitation_service.ListRequest{OrganizationID: orgID}
	resp, err := s.svc.List(req, accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
	s.orgRepo.AssertExpectations(s.T())
}

func (s *InvitationServiceTestSuite) TestList_Success() {
	orgID := uuid.New().String()
	org := &models.Organization{ID: uuid.MustParse(orgID), Name: "Test Org", SubscriptionSeats: 10}
	invitations := &[]models.Invitation{
		{ID: uuid.New(), OrganizationID: uuid.MustParse(orgID), Email: "a@example.com"},
	}
	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	s.invitationRepo.On("ListByOrganizationID", orgID).Return(invitations, nil)

	req := &invitation_service.ListRequest{OrganizationID: orgID}
	resp, err := s.svc.List(req, s.accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Len(*resp, 1)
	s.Equal("a@example.com", (*resp)[0].Email)
	s.orgRepo.AssertExpectations(s.T())
	s.invitationRepo.AssertExpectations(s.T())
}

// --- Revoke: validations ---

func (s *InvitationServiceTestSuite) TestRevoke_Validation_MissingInvitationID() {
	req := &invitation_service.RevokeRequest{
		OrganizationID: uuid.New().String(),
		InvitationID:   "",
	}
	err := s.svc.Revoke(req, s.accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *InvitationServiceTestSuite) TestRevoke_Validation_InvalidInvitationID() {
	req := &invitation_service.RevokeRequest{
		OrganizationID: uuid.New().String(),
		InvitationID:   "not-a-uuid",
	}
	err := s.svc.Revoke(req, s.accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Revoke: business logic ---

func (s *InvitationServiceTestSuite) TestRevoke_OrganizationNotFound() {
	orgID := uuid.New().String()
	invID := uuid.New().String()
	s.orgRepo.On("GetByID", orgID).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &invitation_service.RevokeRequest{OrganizationID: orgID, InvitationID: invID}
	err := s.svc.Revoke(req, s.accessInfo)
	s.Error(err)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *InvitationServiceTestSuite) TestRevoke_InvitationNotFound() {
	orgID := uuid.New().String()
	invID := uuid.New().String()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: uuid.MustParse(orgID), SubscriptionType: constants.SUBSCRIPTION_TYPE_BASIC, SubscriptionExpiresAt: &futureExpiry}
	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	s.invitationRepo.On("GetByID", orgID, invID).Return((*models.Invitation)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &invitation_service.RevokeRequest{OrganizationID: orgID, InvitationID: invID}
	err := s.svc.Revoke(req, s.accessInfo)
	s.Error(err)
	s.invitationRepo.AssertExpectations(s.T())
}

func (s *InvitationServiceTestSuite) TestRevoke_AlreadyRevoked() {
	orgID := uuid.New().String()
	invID := uuid.New().String()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: uuid.MustParse(orgID), SubscriptionType: constants.SUBSCRIPTION_TYPE_BASIC, SubscriptionExpiresAt: &futureExpiry}
	revokedAt := s.now.Add(-1 * time.Hour)
	inv := &models.Invitation{
		ID:        uuid.MustParse(invID),
		RevokedAt: &revokedAt,
	}
	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	s.invitationRepo.On("GetByID", orgID, invID).Return(inv, nil)

	req := &invitation_service.RevokeRequest{OrganizationID: orgID, InvitationID: invID}
	err := s.svc.Revoke(req, s.accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVITATION_REVOKED_CODE, e.ErrorCode)
	}
	s.invitationRepo.AssertExpectations(s.T())
}

func (s *InvitationServiceTestSuite) TestRevoke_AlreadyAccepted() {
	orgID := uuid.New().String()
	invID := uuid.New().String()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: uuid.MustParse(orgID), SubscriptionType: constants.SUBSCRIPTION_TYPE_BASIC, SubscriptionExpiresAt: &futureExpiry}
	acceptedAt := s.now.Add(-1 * time.Hour)
	inv := &models.Invitation{
		ID:         uuid.MustParse(invID),
		RevokedAt:  nil,
		AcceptedAt: &acceptedAt,
		ExpiresAt:  s.now.Add(24 * time.Hour),
	}
	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	s.invitationRepo.On("GetByID", orgID, invID).Return(inv, nil)

	req := &invitation_service.RevokeRequest{OrganizationID: orgID, InvitationID: invID}
	err := s.svc.Revoke(req, s.accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVITATION_ALREADY_ACCEPTED_CODE, e.ErrorCode)
	}
	s.invitationRepo.AssertExpectations(s.T())
}

func (s *InvitationServiceTestSuite) TestRevoke_Expired() {
	orgID := uuid.New().String()
	invID := uuid.New().String()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: uuid.MustParse(orgID), SubscriptionType: constants.SUBSCRIPTION_TYPE_BASIC, SubscriptionExpiresAt: &futureExpiry}
	inv := &models.Invitation{
		ID:         uuid.MustParse(invID),
		RevokedAt:  nil,
		AcceptedAt: nil,
		ExpiresAt:  s.now.Add(-1 * time.Hour),
	}
	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	s.invitationRepo.On("GetByID", orgID, invID).Return(inv, nil)

	req := &invitation_service.RevokeRequest{OrganizationID: orgID, InvitationID: invID}
	err := s.svc.Revoke(req, s.accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.INVITATION_EXPIRED_CODE, e.ErrorCode)
	}
	s.invitationRepo.AssertExpectations(s.T())
}

func (s *InvitationServiceTestSuite) TestRevoke_Success() {
	orgID := uuid.New().String()
	invID := uuid.New().String()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: uuid.MustParse(orgID), SubscriptionType: constants.SUBSCRIPTION_TYPE_BASIC, SubscriptionExpiresAt: &futureExpiry}
	inv := &models.Invitation{
		ID:         uuid.MustParse(invID),
		RevokedAt:  nil,
		AcceptedAt: nil,
		ExpiresAt:  s.now.Add(24 * time.Hour),
	}
	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	s.invitationRepo.On("GetByID", orgID, invID).Return(inv, nil)
	s.invitationRepo.On("Update", nil, mock.AnythingOfType("*models.Invitation")).Return(nil)

	req := &invitation_service.RevokeRequest{OrganizationID: orgID, InvitationID: invID}
	err := s.svc.Revoke(req, s.accessInfo)
	s.NoError(err)
	s.invitationRepo.AssertExpectations(s.T())
}

// --- Create: validations ---

func (s *InvitationServiceTestSuite) TestCreate_Validation_MissingEmail() {
	req := &invitation_service.CreateRequest{
		Email:          "",
		OrganizationID: uuid.New().String(),
		RoleID:         uuid.New().String(),
		InvitedByID:    uuid.New().String(),
		InvitedByEmail: "admin@example.com",
		Language:       "en",
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *InvitationServiceTestSuite) TestCreate_Validation_InvalidEmail() {
	req := &invitation_service.CreateRequest{
		Email:          "not-an-email",
		OrganizationID: uuid.New().String(),
		RoleID:         uuid.New().String(),
		InvitedByID:    uuid.New().String(),
		InvitedByEmail: "admin@example.com",
		Language:       "en",
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *InvitationServiceTestSuite) TestCreate_Validation_MissingOrganizationID() {
	req := &invitation_service.CreateRequest{
		Email:          "user@example.com",
		OrganizationID: "",
		RoleID:         uuid.New().String(),
		InvitedByID:    uuid.New().String(),
		InvitedByEmail: "admin@example.com",
		Language:       "en",
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

func (s *InvitationServiceTestSuite) TestCreate_Validation_MissingRoleID() {
	req := &invitation_service.CreateRequest{
		Email:          "user@example.com",
		OrganizationID: uuid.New().String(),
		RoleID:         "",
		InvitedByID:    uuid.New().String(),
		InvitedByEmail: "admin@example.com",
		Language:       "en",
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.UNPROCESSABLE_ENTITY_CODE, e.ErrorCode)
	}
}

// --- Create: business logic ---

func (s *InvitationServiceTestSuite) TestCreate_OrganizationNotFound() {
	orgID := uuid.New().String()
	s.orgRepo.On("GetByID", orgID).Return((*models.Organization)(nil), yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))

	req := &invitation_service.CreateRequest{
		Email:          "user@example.com",
		OrganizationID: orgID,
		RoleID:         uuid.New().String(),
		InvitedByID:    uuid.New().String(),
		InvitedByEmail: "admin@example.com",
		Language:       "en",
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	s.orgRepo.AssertExpectations(s.T())
}

func (s *InvitationServiceTestSuite) TestCreate_ExistingUser_AlreadyMember() {
	orgID := uuid.New().String()
	userID := uuid.New()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: uuid.MustParse(orgID), Name: "Org", SubscriptionSeats: 10, SubscriptionType: constants.SUBSCRIPTION_TYPE_BASIC, SubscriptionExpiresAt: &futureExpiry}
	existingUser := &models.User{ID: userID, Email: "user@example.com"}
	members := &[]models.OrganizationMemberWithUser{
		{OrganizationMember: models.OrganizationMember{UserID: userID}},
	}

	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	s.userRepo.On("GetByEmail", nil, "user@example.com").Return(existingUser, nil)
	tx := yca_repository.NewMockTx()
	s.orgMemberRepo.On("BeginTx").Return(tx, nil)
	s.orgMemberRepo.On("ListByOrganizationID", orgID).Return(members, nil)
	tx.On("Rollback").Return(nil).Maybe()

	req := &invitation_service.CreateRequest{
		Email:          "user@example.com",
		OrganizationID: orgID,
		RoleID:         uuid.New().String(),
		InvitedByID:    uuid.New().String(),
		InvitedByEmail: "admin@example.com",
		Language:       "en",
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.Error(err)
	s.Nil(resp)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.USER_ALREADY_MEMBER_CODE, e.ErrorCode)
	}
	s.userRepo.AssertExpectations(s.T())
	s.orgMemberRepo.AssertExpectations(s.T())
}

func (s *InvitationServiceTestSuite) TestCreate_ExistingUser_AddMember_Success() {
	orgID := uuid.New().String()
	roleID := uuid.New().String()
	userID := uuid.New()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: uuid.MustParse(orgID), Name: "Org", SubscriptionSeats: 10, SubscriptionType: constants.SUBSCRIPTION_TYPE_BASIC, SubscriptionExpiresAt: &futureExpiry}
	existingUser := &models.User{ID: userID, Email: "user@example.com"}
	members := &[]models.OrganizationMemberWithUser{} // not a member yet
	memberWithUser := &models.OrganizationMemberWithUser{
		OrganizationMember: models.OrganizationMember{
			ID:             uuid.New(),
			UserID:         userID,
			OrganizationID: uuid.MustParse(orgID),
			RoleID:         uuid.MustParse(roleID),
		},
		UserEmail:     "user@example.com",
		UserFirstName: "User",
		UserLastName:  "Name",
	}

	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	s.userRepo.On("GetByEmail", nil, "user@example.com").Return(existingUser, nil)
	tx := yca_repository.NewMockTx()
	s.orgMemberRepo.On("BeginTx").Return(tx, nil)
	s.orgMemberRepo.On("ListByOrganizationID", orgID).Return(members, nil)
	s.orgMemberRepo.On("Create", tx, mock.AnythingOfType("*models.OrganizationMember")).Return(nil)
	s.orgMemberRepo.On("GetByIDWithUser", orgID, mock.AnythingOfType("string")).Return(memberWithUser, nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()

	req := &invitation_service.CreateRequest{
		Email:          "user@example.com",
		OrganizationID: orgID,
		RoleID:         roleID,
		InvitedByID:    uuid.New().String(),
		InvitedByEmail: "admin@example.com",
		Language:       "en",
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Nil(resp.Invitation)
	s.Require().NotNil(resp.Member)
	s.Equal("user@example.com", resp.Member.UserEmail)
	s.orgMemberRepo.AssertExpectations(s.T())
}

func (s *InvitationServiceTestSuite) TestCreate_NewUser_Invitation_Success() {
	orgID := uuid.New().String()
	roleID := uuid.New().String()
	futureExpiry := s.now.Add(24 * time.Hour)
	org := &models.Organization{ID: uuid.MustParse(orgID), Name: "Org", SubscriptionSeats: 10, SubscriptionType: constants.SUBSCRIPTION_TYPE_BASIC, SubscriptionExpiresAt: &futureExpiry}

	var nilUser *models.User
	s.orgRepo.On("GetByID", orgID).Return(org, nil)
	s.userRepo.On("GetByEmail", nil, "newuser@example.com").Return(nilUser, yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil))
	tx := yca_repository.NewMockTx()
	s.orgMemberRepo.On("BeginTx").Return(tx, nil)
	s.invitationRepo.On("Create", tx, mock.AnythingOfType("*models.Invitation")).Return(nil)
	s.emailSvc.On("PrepareEmailBody", "invitation", mock.Anything).Return("body", nil)
	s.emailSvc.On("SendEmail", "newuser@example.com", mock.AnythingOfType("string"), "body").Return(nil)
	tx.On("Commit").Return(nil)
	tx.On("Rollback").Return(nil).Maybe()

	req := &invitation_service.CreateRequest{
		Email:          "newuser@example.com",
		OrganizationID: orgID,
		RoleID:         roleID,
		InvitedByID:    uuid.New().String(),
		InvitedByEmail: "admin@example.com",
		Language:       "en",
	}
	resp, err := s.svc.Create(req, s.accessInfo)
	s.NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Invitation)
	s.Nil(resp.Member)
	s.Equal("newuser@example.com", resp.Invitation.Email)
	s.invitationRepo.AssertExpectations(s.T())
	s.emailSvc.AssertExpectations(s.T())
}

func (s *InvitationServiceTestSuite) TestCleanupStale() {
	s.invitationRepo.On("CleanupStale").Return(nil).Once()
	err := s.svc.CleanupStale()
	s.NoError(err)
	s.invitationRepo.AssertExpectations(s.T())
}

func (s *InvitationServiceTestSuite) TestCleanupStale_RepoError() {
	repoErr := errors.New("cleanup failed")
	s.invitationRepo.On("CleanupStale").Return(repoErr).Once()
	err := s.svc.CleanupStale()
	s.ErrorIs(err, repoErr)
	s.invitationRepo.AssertExpectations(s.T())
}
