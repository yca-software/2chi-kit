package organization_service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

// AdminUpdateSubscriptionSettingsRequest allows admins to update custom subscription fields. All fields optional; only non-nil are applied. Enterprise = subscription type 3.
type AdminUpdateSubscriptionSettingsRequest struct {
	OrganizationID        string  `json:"-" validate:"required,uuid"`
	CustomSubscription    *bool   `json:"customSubscription"`
	SubscriptionType      *int    `json:"subscriptionType" validate:"omitempty,min=1,max=3"` // SUBSCRIPTION_TYPE_FREE, BASIC, PRO, or ENTERPRISE
	SubscriptionSeats     *int    `json:"subscriptionSeats" validate:"omitempty,min=1"`
	SubscriptionExpiresAt *string `json:"subscriptionExpiresAt"` // RFC3339 or empty string to clear; validated by parsing in handler
}

func (s *service) AdminUpdateSubscriptionSettings(req *AdminUpdateSubscriptionSettingsRequest, accessInfo *models.AccessInfo) (*models.Organization, error) {
	if err := s.authorizer.CheckAdmin(accessInfo); err != nil {
		return nil, err
	}
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
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
				Level: "error", Location: "services.organizationService.AdminUpdateSubscriptionSettings",
				Error: err, Message: "Transaction rollback failed", Data: map[string]any{"organization_id": req.OrganizationID},
			})
		}
	}()

	updated := *org
	if req.CustomSubscription != nil {
		updated.CustomSubscription = *req.CustomSubscription
	}
	if req.SubscriptionType != nil {
		updated.SubscriptionType = *req.SubscriptionType
	}
	if req.SubscriptionSeats != nil {
		updated.SubscriptionSeats = *req.SubscriptionSeats
	}
	if req.SubscriptionExpiresAt != nil {
		if *req.SubscriptionExpiresAt == "" {
			updated.SubscriptionExpiresAt = nil
		} else {
			t, err := time.Parse(time.RFC3339, *req.SubscriptionExpiresAt)
			if err != nil {
				return nil, yca_error.NewUnprocessableEntityError(err, "", nil)
			}
			updated.SubscriptionExpiresAt = &t
		}
	}

	if err := s.repos.Organization.Update(tx, &updated); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	changes, _ := json.Marshal(map[string]any{
		"customSubscription":    updated.CustomSubscription,
		"subscriptionType":      updated.SubscriptionType,
		"subscriptionSeats":     updated.SubscriptionSeats,
		"subscriptionExpiresAt": updated.SubscriptionExpiresAt,
	})
	changesRaw := json.RawMessage(changes)
	nameCopy := updated.Name
	_, _ = s.auditLogService.Create(&audit_log_service.CreateRequest{
		OrganizationID: req.OrganizationID,
		Action:         constants.AUDIT_ACTION_TYPE_UPDATE,
		ResourceType:   constants.RESOURCE_TYPE_ORGANIZATION,
		ResourceID:     org.ID.String(),
		ResourceName:   &nameCopy,
		Data:           &changesRaw,
	}, accessInfo)

	return &updated, nil
}
