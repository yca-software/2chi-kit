package helpers

import (
	"slices"
	"time"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

// PastDueGraceDuration is the duration after subscription expiry during which access is still allowed (show warning in UI; after this, return PAYMENT_REQUIRED).
var PastDueGraceDuration = time.Duration(constants.SUBSCRIPTION_PAST_DUE_GRACE_DAYS) * 24 * time.Hour

type Authorizer struct {
	now func() time.Time
}

func NewAuthorizer(now func() time.Time) *Authorizer {
	return &Authorizer{
		now: now,
	}
}

func (a *Authorizer) CheckAdmin(accessInfo *models.AccessInfo) error {
	if accessInfo == nil || accessInfo.User == nil || !accessInfo.User.IsAdmin {
		return yca_error.NewForbiddenError(nil, constants.FORBIDDEN_CODE, nil)
	}
	return nil
}

func (a *Authorizer) CheckOwnResource(accessInfo *models.AccessInfo, resourceUserID string) error {
	if accessInfo == nil || accessInfo.User == nil {
		return yca_error.NewForbiddenError(nil, constants.FORBIDDEN_CODE, nil)
	}

	if accessInfo.User.IsAdmin {
		return nil
	}

	if accessInfo.User.UserID.String() == resourceUserID {
		return nil
	}

	return yca_error.NewForbiddenError(nil, constants.FORBIDDEN_CODE, nil)
}

func (a *Authorizer) CheckOrganizationPermission(accessInfo *models.AccessInfo, organizationID string, permission string) error {
	if accessInfo == nil {
		return yca_error.NewForbiddenError(nil, constants.FORBIDDEN_CODE, nil)
	}

	if accessInfo.User != nil {
		if accessInfo.User.IsAdmin {
			return nil
		}

		for _, role := range accessInfo.User.Roles {
			if role.OrganizationID.String() == organizationID {
				if slices.Contains([]string(role.Permissions), permission) {
					return nil
				}
			}
		}
	}

	if accessInfo.ApiKey != nil {
		if accessInfo.ApiKey.OrganizationID.String() == organizationID {
			if slices.Contains([]string(accessInfo.ApiKey.Permissions), permission) {
				return nil
			}
		}
		return yca_error.NewForbiddenError(nil, constants.FORBIDDEN_CODE, nil)
	}

	return yca_error.NewForbiddenError(nil, constants.FORBIDDEN_CODE, nil)
}

func (a *Authorizer) CheckOrganizationPermissionWithSubscription(accessInfo *models.AccessInfo, organization *models.Organization, permission string) error {
	// Subscription type 0 (FREE) does not include paid-gated capabilities.
	if organization.SubscriptionType == constants.SUBSCRIPTION_TYPE_FREE {
		return yca_error.NewForbiddenError(nil, constants.FEATURE_NOT_INCLUDED_CODE, nil)
	}

	// Subscription expiry must be set (by pricing step trial or Paddle). Enterprise/custom may have nil.
	// Past due: allow access for SUBSCRIPTION_PAST_DUE_GRACE_DAYS after expiry; after that return PAYMENT_REQUIRED.
	if organization.SubscriptionType != constants.SUBSCRIPTION_TYPE_ENTERPRISE && !organization.CustomSubscription {
		if organization.SubscriptionExpiresAt == nil {
			return yca_error.NewPaymentRequiredError(nil, constants.PAYMENT_REQUIRED_CODE, nil)
		}
		now := a.now()
		if organization.SubscriptionExpiresAt.Before(now) {
			if now.Sub(*organization.SubscriptionExpiresAt) > PastDueGraceDuration {
				return yca_error.NewPaymentRequiredError(nil, constants.PAYMENT_REQUIRED_CODE, nil)
			}
		}
	}

	return a.CheckOrganizationPermission(accessInfo, organization.ID.String(), permission)
}

func (a *Authorizer) CheckOrganizationFeature(accessInfo *models.AccessInfo, organization *models.Organization, featureKey string) error {
	if accessInfo != nil && accessInfo.User != nil && accessInfo.User.IsAdmin {
		return nil
	}
	if organization == nil {
		return yca_error.NewForbiddenError(nil, constants.FEATURE_NOT_AVAILABLE_CODE, nil)
	}
	allowedTypes, ok := constants.FEATURES_FOR_PLANS[featureKey]
	if !ok {
		return yca_error.NewForbiddenError(nil, constants.FEATURE_NOT_AVAILABLE_CODE, nil)
	}
	if !slices.Contains(allowedTypes, organization.SubscriptionType) {
		return yca_error.NewForbiddenError(nil, constants.FEATURE_NOT_INCLUDED_CODE, nil)
	}
	return nil
}
