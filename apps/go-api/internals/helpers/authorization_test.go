package helpers

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type AuthorizationTestSuite struct {
	suite.Suite
	authorizer *Authorizer
	now        time.Time
}

func TestAuthorizationTestSuite(t *testing.T) {
	suite.Run(t, new(AuthorizationTestSuite))
}

func (s *AuthorizationTestSuite) SetupTest() {
	s.now = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	s.authorizer = NewAuthorizer(func() time.Time { return s.now })
}

func (s *AuthorizationTestSuite) TestCheckAdmin_WithAdminAccess() {
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			IsAdmin: true,
		},
	}

	err := s.authorizer.CheckAdmin(accessInfo)
	s.NoError(err)
}

func (s *AuthorizationTestSuite) TestCheckAdmin_WithoutAdminAccess() {
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			IsAdmin: false,
		},
	}

	err := s.authorizer.CheckAdmin(accessInfo)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
}

func (s *AuthorizationTestSuite) TestCheckAdmin_NilAccessInfo() {
	err := s.authorizer.CheckAdmin(nil)
	s.Error(err)
}
func (s *AuthorizationTestSuite) TestCheckOwnResource_OwnResource() {
	userID := uuid.New()
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}

	err := s.authorizer.CheckOwnResource(accessInfo, userID.String())
	s.NoError(err)
}

func (s *AuthorizationTestSuite) TestCheckOwnResource_AdminAccess() {
	userID := uuid.New()
	otherUserID := uuid.New()
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  userID,
			IsAdmin: true,
		},
	}

	err := s.authorizer.CheckOwnResource(accessInfo, otherUserID.String())
	s.NoError(err) // Admin can access any resource
}

func (s *AuthorizationTestSuite) TestCheckOwnResource_OtherUserResource() {
	userID := uuid.New()
	otherUserID := uuid.New()
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: userID,
		},
	}

	err := s.authorizer.CheckOwnResource(accessInfo, otherUserID.String())
	s.Error(err)
}

func (s *AuthorizationTestSuite) TestCheckOrganizationPermission_WithPermission() {
	orgID := uuid.New()
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
				},
			},
		},
	}

	err := s.authorizer.CheckOrganizationPermission(accessInfo, orgID.String(), constants.PERMISSION_ORG_READ)
	s.NoError(err)
}

func (s *AuthorizationTestSuite) TestCheckOrganizationPermission_WithoutPermission() {
	orgID := uuid.New()
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
				},
			},
		},
	}

	err := s.authorizer.CheckOrganizationPermission(accessInfo, orgID.String(), constants.PERMISSION_ORG_WRITE)
	s.Error(err)
}

func (s *AuthorizationTestSuite) TestCheckOrganizationPermission_AdminAccess() {
	orgID := uuid.New()
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			IsAdmin: true,
		},
	}

	err := s.authorizer.CheckOrganizationPermission(accessInfo, orgID.String(), constants.PERMISSION_ORG_READ)
	s.NoError(err) // Admin has all permissions
}

func (s *AuthorizationTestSuite) TestCheckOrganizationPermission_ApiKeyAccess_WithPermission() {
	orgID := uuid.New()
	accessInfo := &models.AccessInfo{
		ApiKey: &models.APIKey{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
		},
	}

	err := s.authorizer.CheckOrganizationPermission(accessInfo, orgID.String(), constants.PERMISSION_ORG_READ)
	s.NoError(err)
}

func (s *AuthorizationTestSuite) TestCheckOrganizationPermission_ApiKeyAccess_WithoutPermission() {
	orgID := uuid.New()
	accessInfo := &models.AccessInfo{
		ApiKey: &models.APIKey{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
		},
	}

	err := s.authorizer.CheckOrganizationPermission(accessInfo, orgID.String(), constants.PERMISSION_ORG_WRITE)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FORBIDDEN_CODE, e.ErrorCode)
	}
}

func (s *AuthorizationTestSuite) TestCheckOrganizationPermission_ApiKeyAccess_WrongOrg() {
	orgID := uuid.New()
	otherOrgID := uuid.New()
	accessInfo := &models.AccessInfo{
		ApiKey: &models.APIKey{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
		},
	}

	err := s.authorizer.CheckOrganizationPermission(accessInfo, otherOrgID.String(), constants.PERMISSION_ORG_READ)
	s.Error(err)
}

func (s *AuthorizationTestSuite) TestCheckOrganizationPermissionWithSubscription_ActiveSubscription() {
	orgID := uuid.New()
	futureTime := s.now.Add(24 * time.Hour)
	organization := &models.Organization{
		ID:                    orgID,
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &futureTime,
	}

	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
				},
			},
		},
	}

	err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, organization, constants.PERMISSION_ORG_READ)
	s.NoError(err)
}

func (s *AuthorizationTestSuite) TestCheckOrganizationPermissionWithSubscription_PastDueWithinGrace() {
	orgID := uuid.New()
	pastTime := s.now.Add(-24 * time.Hour) // 1 day ago, within 7-day grace
	organization := &models.Organization{
		ID:                    orgID,
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &pastTime,
	}

	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
				},
			},
		},
	}

	err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, organization, constants.PERMISSION_ORG_READ)
	s.NoError(err)
}

func (s *AuthorizationTestSuite) TestCheckOrganizationPermissionWithSubscription_ExpiredSubscription() {
	orgID := uuid.New()
	pastTime := s.now.Add(-8 * 24 * time.Hour) // 8 days ago, beyond 7-day grace
	organization := &models.Organization{
		ID:                    orgID,
		SubscriptionType:      constants.SUBSCRIPTION_TYPE_BASIC,
		SubscriptionExpiresAt: &pastTime,
	}

	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
			Roles: []models.JWTAccessTokenPermissionData{
				{
					OrganizationID: orgID,
					RoleID:         uuid.New(),
					Permissions:    models.RolePermissions{constants.PERMISSION_ORG_READ},
				},
			},
		},
	}

	err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, organization, constants.PERMISSION_ORG_READ)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.PAYMENT_REQUIRED_CODE, e.ErrorCode)
	}
}

func (s *AuthorizationTestSuite) TestCheckOrganizationFeature_Admin_SkipsCheck() {
	org := &models.Organization{
		ID:               uuid.New(),
		SubscriptionType: constants.SUBSCRIPTION_TYPE_FREE, // audit_log not in FREE plan
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID:  uuid.New(),
			IsAdmin: true,
		},
	}
	err := s.authorizer.CheckOrganizationFeature(accessInfo, org, constants.FEATURE_AUDIT_LOG)
	s.NoError(err)
}

func (s *AuthorizationTestSuite) TestCheckOrganizationFeature_PlanIncludesFeature() {
	org := &models.Organization{
		ID:               uuid.New(),
		SubscriptionType: constants.SUBSCRIPTION_TYPE_BASIC,
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	err := s.authorizer.CheckOrganizationFeature(accessInfo, org, constants.FEATURE_AUDIT_LOG)
	s.NoError(err)
}

func (s *AuthorizationTestSuite) TestCheckOrganizationFeature_PlanExcludesFeature() {
	org := &models.Organization{
		ID:               uuid.New(),
		SubscriptionType: constants.SUBSCRIPTION_TYPE_FREE,
	}
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{
			UserID: uuid.New(),
		},
	}
	err := s.authorizer.CheckOrganizationFeature(accessInfo, org, constants.FEATURE_AUDIT_LOG)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FEATURE_NOT_INCLUDED_CODE, e.ErrorCode)
	}
}

func (s *AuthorizationTestSuite) TestCheckOrganizationFeature_NilOrganization() {
	accessInfo := &models.AccessInfo{
		User: &models.UserAccessInfo{UserID: uuid.New()},
	}
	err := s.authorizer.CheckOrganizationFeature(accessInfo, nil, constants.FEATURE_AUDIT_LOG)
	s.Error(err)
	if e, ok := err.(*yca_error.Error); ok {
		s.Equal(constants.FEATURE_NOT_AVAILABLE_CODE, e.ErrorCode)
	}
}
