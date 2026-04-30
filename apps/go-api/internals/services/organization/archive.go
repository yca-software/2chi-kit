package organization_service

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type ArchiveRequest struct {
	OrganizationID string `json:"organizationId" validate:"required,uuid"`
}

func (s *service) Archive(req *ArchiveRequest, accessInfo *models.AccessInfo) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return err
	}

	// Permission only; do not require active subscription — user may be archiving due to non-payment.
	if err := s.authorizer.CheckOrganizationPermission(accessInfo, org.ID.String(), constants.PERMISSION_ORG_DELETE); err != nil {
		return err
	}

	tx, err := s.repos.Organization.BeginTx()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.organizationService.Archive",
				Error:    err,
				Message:  "Transaction rollback failed",
				Data:     map[string]any{"organization_id": req.OrganizationID},
			})
		}
	}()

	if err := s.repos.Organization.Archive(tx, org.ID.String()); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Do not archive the Paddle customer; the same email may be used by other apps or for resubscription.

	if org.PaddleSubscriptionID != nil && *org.PaddleSubscriptionID != "" {
		if err := s.paddleService.CancelSubscription(*org.PaddleSubscriptionID); err != nil {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.organizationService.Archive",
				Error:    err,
				Message:  "Failed to cancel paddle subscription",
				Data:     map[string]any{"organization_id": req.OrganizationID},
			})
		}
	}

	emptyData := json.RawMessage(`{}`)
	nameCopy := org.Name
	if _, err := s.auditLogService.Create(&audit_log_service.CreateRequest{
		OrganizationID: org.ID.String(),
		Action:         constants.AUDIT_ACTION_TYPE_ARCHIVE,
		ResourceType:   constants.RESOURCE_TYPE_ORGANIZATION,
		ResourceID:     org.ID.String(),
		ResourceName:   &nameCopy,
		Data:           &emptyData,
	}, accessInfo); err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.organizationService.Archive",
			Error:    err,
			Message:  "Failed to create audit log",
			Data:     map[string]any{"organization_id": req.OrganizationID},
		})
	}

	return nil
}
