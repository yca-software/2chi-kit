package organization_service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	invitation_service "github.com/yca-software/2chi-kit/go-api/internals/services/invitation"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

// AdminCreateOrganizationWithCustomSubscriptionRequest is used by admins to create an org with custom subscription and invite the owner.
type AdminCreateOrganizationWithCustomSubscriptionRequest struct {
	Name                  string     `json:"name" validate:"required,min=1,max=255"`
	PlaceID               string     `json:"placeId" validate:"required,min=1,max=255"`
	BillingEmail          string     `json:"billingEmail" validate:"required,email"`
	OwnerEmail            string     `json:"ownerEmail" validate:"required,email"`
	SubscriptionType      int        `json:"subscriptionType" validate:"required,min=1,max=3"` // SUBSCRIPTION_TYPE_FREE, BASIC, PRO, or ENTERPRISE
	SubscriptionSeats     int        `json:"subscriptionSeats" validate:"required,min=1"`
	SubscriptionExpiresAt *time.Time `json:"subscriptionExpiresAt"` // optional; nil = no expiry (e.g. enterprise contract)
	Language              string     `json:"language" validate:"required"`
}

// AdminCreateOrganizationWithCustomSubscription creates an org with custom subscription (same as Create: Paddle customer, cleanups). Sets custom_subscription and subscription fields; no member is added—caller sends the owner an invite.
func (s *service) AdminCreateOrganizationWithCustomSubscription(req *AdminCreateOrganizationWithCustomSubscriptionRequest, accessInfo *models.AccessInfo) (*models.Organization, error) {
	if err := s.authorizer.CheckAdmin(accessInfo); err != nil {
		return nil, err
	}
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	ctx := context.Background()
	now := s.now()

	orgID, err := s.generateID()
	if err != nil {
		return nil, err
	}

	tx, err := s.repos.Organization.BeginTx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.logger.Log(yca_log.LogData{
				Level: "error", Location: "services.organizationService.AdminCreateOrganizationWithCustomSubscription",
				Error: err, Message: "Transaction rollback failed", Data: map[string]any{"organization_id": orgID},
			})
		}
	}()

	locationData, err := s.googleService.GetLocationData(ctx, req.PlaceID)
	if err != nil {
		return nil, err
	}

	org := &models.Organization{
		ID:                          orgID,
		CreatedAt:                   now,
		Name:                        strings.TrimSpace(req.Name),
		Address:                     locationData.Address,
		City:                        locationData.City,
		Zip:                         locationData.Zip,
		Country:                     locationData.Country,
		PlaceID:                     locationData.PlaceID,
		Geo:                         locationData.Geo,
		Timezone:                    locationData.Timezone,
		BillingEmail:                req.BillingEmail,
		CustomSubscription:          true,
		SubscriptionExpiresAt:       req.SubscriptionExpiresAt,
		SubscriptionPaymentInterval: constants.PAYMENT_INTERVAL_MONTHLY,
		SubscriptionType:            req.SubscriptionType,
		SubscriptionSeats:           req.SubscriptionSeats,
		SubscriptionInTrial:         false,
	}

	paddleCustomer, err := s.paddleService.CreateCustomer(org)
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level: "error", Location: "services.organizationService.AdminCreateOrganizationWithCustomSubscription",
			Error: err, Message: "Failed to create or get paddle customer", Data: map[string]any{"organization_id": orgID},
		})
		return nil, err
	}
	org.PaddleCustomerID = paddleCustomer.ID

	if err := s.repos.Organization.Create(tx, org); err != nil {
		return nil, err
	}

	roles := make([]models.Role, 0, len(constants.DEFAULT_ROLES_TO_CREATE_FOR_ORGANIZATION))
	for _, role := range constants.DEFAULT_ROLES_TO_CREATE_FOR_ORGANIZATION {
		roleID, err := s.generateID()
		if err != nil {
			return nil, err
		}
		r := role
		r.ID = roleID
		r.CreatedAt = now
		r.OrganizationID = orgID
		roles = append(roles, r)
	}

	if err := s.repos.Role.CreateMany(tx, &roles); err != nil {
		return nil, err
	}

	ownerRoleID := roles[0].ID
	emailLower := strings.ToLower(strings.TrimSpace(req.OwnerEmail))
	existingUser, err := s.repos.User.GetByEmail(tx, emailLower)
	if err != nil {
		e, ok := yca_error.AsError(err)
		if ok && e.ErrorCode == constants.NOT_FOUND_CODE {
			existingUser = nil
		} else {
			return nil, err
		}
	}

	if existingUser != nil {
		memberID, err := s.generateID()
		if err != nil {
			return nil, err
		}
		if err := s.repos.OrganizationMember.Create(tx, &models.OrganizationMember{
			ID:             memberID,
			CreatedAt:      now,
			UserID:         existingUser.ID,
			OrganizationID: orgID,
			RoleID:         ownerRoleID,
		}); err != nil {
			s.logger.Log(yca_log.LogData{
				Level: "error", Location: "services.organizationService.AdminCreateOrganizationWithCustomSubscription",
				Error: err, Message: "Failed to add existing user as owner", Data: map[string]any{"organization_id": orgID, "user_id": existingUser.ID},
			})
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if existingUser == nil {
		_, err = s.invitationService.Create(&invitation_service.CreateRequest{
			Email:          req.OwnerEmail,
			OrganizationID: org.ID.String(),
			RoleID:         ownerRoleID.String(),
			InvitedByID:    accessInfo.User.UserID.String(),
			InvitedByEmail: accessInfo.User.Email,
			Language:       req.Language,
		}, accessInfo)
		if err != nil {
			s.logger.Log(yca_log.LogData{
				Level: "error", Location: "services.organizationService.AdminCreateOrganizationWithCustomSubscription",
				Error: err, Message: "Failed to send owner invitation", Data: map[string]any{"organization_id": orgID, "owner_email": req.OwnerEmail},
			})
			return nil, err
		}
	}

	return org, nil
}
