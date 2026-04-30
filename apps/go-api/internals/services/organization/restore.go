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

type RestoreRequest struct {
	OrganizationID string `json:"organizationId" validate:"required,uuid"`
}

func (s *service) Restore(req *RestoreRequest, accessInfo *models.AccessInfo) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	if err := s.authorizer.CheckAdmin(accessInfo); err != nil {
		return err
	}

	org, err := s.repos.Organization.GetByIDIncludeArchived(req.OrganizationID)
	if err != nil {
		return err
	}
	if org.DeletedAt == nil {
		return yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil)
	}

	tx, err := s.repos.Organization.BeginTx()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.organizationService.Restore",
				Error:    err,
				Message:  "Transaction rollback failed",
				Data:     map[string]any{"organization_id": req.OrganizationID},
			})
		}
	}()

	if err := s.repos.Organization.Restore(tx, org.ID.String()); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	emptyData := json.RawMessage(`{}`)
	nameCopy := org.Name
	if _, err := s.auditLogService.Create(&audit_log_service.CreateRequest{
		OrganizationID: org.ID.String(),
		Action:         constants.AUDIT_ACTION_TYPE_RESTORE,
		ResourceType:   constants.RESOURCE_TYPE_ORGANIZATION,
		ResourceID:     org.ID.String(),
		ResourceName:   &nameCopy,
		Data:           &emptyData,
	}, accessInfo); err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.organizationService.Restore",
			Error:    err,
			Message:  "Failed to create audit log",
			Data:     map[string]any{"organization_id": req.OrganizationID},
		})
	}

	return nil
}
