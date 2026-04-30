package organization_service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type UpdateRequest struct {
	OrganizationID string `json:"-" validate:"required,uuid"`
	Name           string `json:"name" validate:"required,min=1,max=255"`
	PlaceID        string `json:"placeId" validate:"required,min=1,max=255"`
}

func (s *service) Update(req *UpdateRequest, accessInfo *models.AccessInfo) (*models.Organization, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, org, constants.PERMISSION_ORG_WRITE); err != nil {
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
				Location: "services.organizationService.Update",
				Error:    err,
				Message:  "Transaction rollback failed",
				Data:     map[string]any{"organization_id": req.OrganizationID},
			})
		}
	}()

	updatedOrg := *org
	updatedOrg.Name = strings.TrimSpace(req.Name)

	// Only fetch location data if placeID has changed
	if req.PlaceID != org.PlaceID {
		ctx := context.Background()
		locationData, err := s.googleService.GetLocationData(ctx, req.PlaceID)
		if err != nil {
			return nil, err
		}

		updatedOrg.Address = locationData.Address
		updatedOrg.City = locationData.City
		updatedOrg.Zip = locationData.Zip
		updatedOrg.Country = locationData.Country
		updatedOrg.PlaceID = locationData.PlaceID
		updatedOrg.Geo = locationData.Geo
		updatedOrg.Timezone = locationData.Timezone
	}

	if _, err := s.paddleService.UpdateCustomer(&updatedOrg); err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.organizationService.Update",
			Error:    err,
			Message:  "Failed to update paddle customer",
			Data:     map[string]any{"organization_id": req.OrganizationID},
		})
		return nil, err
	}

	if err := s.repos.Organization.Update(tx, &updatedOrg); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		if _, err := s.paddleService.UpdateCustomer(org); err != nil {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.organizationService.Update",
				Error:    err,
				Message:  "Failed to update paddle customer",
				Data:     map[string]any{"organization_id": req.OrganizationID},
			})
		}
		return nil, err
	}

	changes, err := json.Marshal(map[string]any{
		"previous": map[string]any{
			"name":     org.Name,
			"address":  org.Address,
			"city":     org.City,
			"zip":      org.Zip,
			"country":  org.Country,
			"placeId":  org.PlaceID,
			"geo":      org.Geo,
			"timezone": org.Timezone,
		},
		"updated": map[string]any{
			"name":     updatedOrg.Name,
			"address":  updatedOrg.Address,
			"city":     updatedOrg.City,
			"zip":      updatedOrg.Zip,
			"country":  updatedOrg.Country,
			"placeId":  updatedOrg.PlaceID,
			"geo":      updatedOrg.Geo,
			"timezone": updatedOrg.Timezone,
		},
	})
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.organizationService.Update",
			Error:    err,
			Message:  "Failed to marshal changes",
			Data:     map[string]any{"organization_id": req.OrganizationID},
		})
	} else {
		changesRaw := json.RawMessage(changes)
		nameCopy := updatedOrg.Name
		if _, err := s.auditLogService.Create(&audit_log_service.CreateRequest{
			OrganizationID: org.ID.String(),
			Action:         constants.AUDIT_ACTION_TYPE_UPDATE,
			ResourceType:   constants.RESOURCE_TYPE_ORGANIZATION,
			ResourceID:     org.ID.String(),
			ResourceName:   &nameCopy,
			Data:           &changesRaw,
		}, accessInfo); err != nil {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.organizationService.Update",
				Error:    err,
				Message:  "Failed to create audit log",
				Data:     map[string]any{"organization_id": req.OrganizationID},
			})
		}
	}

	return &updatedOrg, nil
}
