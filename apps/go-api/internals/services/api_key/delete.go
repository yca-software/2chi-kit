package api_key_service

import (
	"encoding/json"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type DeleteRequest struct {
	OrganizationID string `json:"organizationId" validate:"required,uuid"`
	ApiKeyID       string `json:"apiKeyId" validate:"required,uuid"`
}

func (s *service) Delete(req *DeleteRequest, accessInfo *models.AccessInfo) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return err
	}

	if err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, org, constants.PERMISSION_API_KEY_DELETE); err != nil {
		return err
	}
	if err := s.authorizer.CheckOrganizationFeature(accessInfo, org, constants.FEATURE_API_ACCESS); err != nil {
		return err
	}

	apiKey, err := s.repos.ApiKey.GetByID(req.OrganizationID, req.ApiKeyID)
	if err != nil {
		return err
	}

	if err := s.repos.ApiKey.Delete(nil, req.OrganizationID, req.ApiKeyID); err != nil {
		return err
	}

	emptyData := json.RawMessage(`{}`)
	nameCopy := apiKey.Name
	if _, err := s.auditLogService.Create(&audit_log_service.CreateRequest{
		OrganizationID: req.OrganizationID,
		Action:         constants.AUDIT_ACTION_TYPE_DELETE,
		ResourceType:   constants.RESOURCE_TYPE_API_KEY,
		ResourceID:     req.ApiKeyID,
		ResourceName:   &nameCopy,
		Data:           &emptyData,
	}, accessInfo); err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.api_keyService.Delete",
			Error:    err,
			Message:  "Failed to create audit log",
			Data:     map[string]any{"organization_id": req.OrganizationID, "api_key_id": req.ApiKeyID},
		})
	}

	return nil
}
