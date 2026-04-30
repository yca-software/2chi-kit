package organization_service

import (
	"database/sql"
	"errors"

	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type DeleteRequest struct {
	OrganizationID string `json:"-" validate:"required,uuid"`
}

func (s *service) Delete(req *DeleteRequest, accessInfo *models.AccessInfo) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	if err := s.authorizer.CheckAdmin(accessInfo); err != nil {
		return err
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
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
				Location: "services.organizationService.Delete",
				Error:    err,
				Message:  "Transaction rollback failed",
				Data:     map[string]any{"organization_id": req.OrganizationID},
			})
		}
	}()

	if err := s.repos.Organization.Delete(tx, org.ID.String()); err != nil {
		return err
	}

	if org.PaddleSubscriptionID != nil && *org.PaddleSubscriptionID != "" {
		if err := s.paddleService.CancelSubscription(*org.PaddleSubscriptionID); err != nil {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.organizationService.Delete",
				Error:    err,
				Message:  "Failed to cancel paddle subscription",
				Data:     map[string]any{"organization_id": req.OrganizationID},
			})
		}
	}

	// Do not archive the Paddle customer; the same email may be used by other apps or for resubscription.

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
