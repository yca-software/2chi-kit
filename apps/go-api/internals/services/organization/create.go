package organization_service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type CreateRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=255"`
	PlaceID      string `json:"placeId" validate:"required,min=1,max=255"`
	BillingEmail string `json:"billingEmail" validate:"required,email"`
}

type CreateResponse struct {
	Organization *models.Organization       `json:"organization"`
	Roles        *[]models.Role             `json:"roles"`
	Members      *models.OrganizationMember `json:"members"`
}

func (s *service) Create(req *CreateRequest, accessInfo *models.AccessInfo) (*CreateResponse, error) {
	if accessInfo == nil {
		return nil, yca_error.NewForbiddenError(nil, constants.FORBIDDEN_CODE, nil)
	}
	if accessInfo.ApiKey != nil {
		return nil, yca_error.NewForbiddenError(nil, "API keys cannot create organizations", nil)
	}
	if accessInfo.User == nil {
		return nil, yca_error.NewForbiddenError(nil, constants.FORBIDDEN_CODE, nil)
	}

	ctx := context.Background()
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

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
				Level:    "error",
				Location: "services.organizationService.Create",
				Error:    err,
				Message:  "Transaction rollback failed",
				Data:     map[string]any{"organization_id": orgID},
			})
		}
	}()

	// Get location data from place ID (address, city, zip, country, geo, timezone)
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
		SubscriptionType:            constants.SUBSCRIPTION_TYPE_FREE,
		SubscriptionPaymentInterval: constants.PAYMENT_INTERVAL_MONTHLY,
		SubscriptionSeats:           constants.SUBSCRIPTION_TYPE_SEATS_INCLUDED_FREE,
	}

	paddleCustomer, err := s.paddleService.CreateCustomer(org)
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.organizationService.Create",
			Error:    err,
			Message:  "Failed to create or get paddle customer",
			Data:     map[string]any{"organization_id": orgID},
		})
		return nil, err
	}
	org.PaddleCustomerID = paddleCustomer.ID

	if err := s.repos.Organization.Create(tx, org); err != nil {
		return nil, err
	}

	roles := []models.Role{}
	var ownerRoleID uuid.UUID

	for i, role := range constants.DEFAULT_ROLES_TO_CREATE_FOR_ORGANIZATION {
		roleID, err := s.generateID()
		if err != nil {
			return nil, err
		}

		role.ID = roleID
		role.CreatedAt = now
		role.OrganizationID = orgID

		if i == 0 {
			ownerRoleID = roleID
		}

		roles = append(roles, role)
	}

	if err := s.repos.Role.CreateMany(tx, &roles); err != nil {
		return nil, err
	}

	membershipID, err := s.generateID()
	if err != nil {
		return nil, err
	}

	membership := models.OrganizationMember{
		ID:             membershipID,
		CreatedAt:      now,
		OrganizationID: orgID,
		UserID:         accessInfo.User.UserID,
		RoleID:         ownerRoleID,
	}

	if err := s.repos.OrganizationMember.Create(tx, &membership); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	changes, err := json.Marshal(map[string]any{
		"updated": map[string]any{
			"name":             org.Name,
			"address":          org.Address,
			"city":             org.City,
			"zip":              org.Zip,
			"country":          org.Country,
			"placeId":          org.PlaceID,
			"geo":              org.Geo,
			"timezone":         org.Timezone,
			"billingEmail":     org.BillingEmail,
			"paddleCustomerId": org.PaddleCustomerID,
		},
	})
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.organizationService.Create",
			Error:    err,
			Message:  "Failed to marshal changes for audit log",
			Data:     map[string]any{"organization_id": orgID},
		})
	} else {
		changesRaw := json.RawMessage(changes)
		orgNameCopy := org.Name
		if _, err := s.auditLogService.Create(&audit_log_service.CreateRequest{
			OrganizationID: orgID.String(),
			Action:         constants.AUDIT_ACTION_TYPE_CREATE,
			ResourceType:   constants.RESOURCE_TYPE_ORGANIZATION,
			ResourceID:     orgID.String(),
			ResourceName:   &orgNameCopy,
			Data:           &changesRaw,
		}, accessInfo); err != nil {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.organizationService.Create",
				Error:    err,
				Message:  "Failed to create audit log",
				Data:     map[string]any{"organization_id": orgID},
			})
		}
	}

	return &CreateResponse{
		Organization: org,
		Roles:        &roles,
		Members:      &membership,
	}, nil
}
