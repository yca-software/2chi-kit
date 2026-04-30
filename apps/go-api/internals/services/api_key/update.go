package api_key_service

import (
	"encoding/json"
	"strings"

	"slices"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type UpdateRequest struct {
	OrganizationID string                 `json:"-" validate:"required,uuid"`
	ApiKeyID       string                 `json:"-" validate:"required,uuid"`
	Name           string                 `json:"name" validate:"required,min=1,max=255"`
	Permissions    models.RolePermissions `json:"permissions" validate:"required,min=1"`
}

func (s *service) Update(req *UpdateRequest, accessInfo *models.AccessInfo) (*models.APIKey, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	for _, permission := range req.Permissions {
		if !slices.Contains(assignablePermissions, permission) {
			return nil, yca_error.NewUnprocessableEntityError(nil, constants.INVALID_API_KEY_PERMISSION_CODE, nil)
		}
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, org, constants.PERMISSION_API_KEY_WRITE); err != nil {
		return nil, err
	}
	if err := s.authorizer.CheckOrganizationFeature(accessInfo, org, constants.FEATURE_API_ACCESS); err != nil {
		return nil, err
	}

	apiKey, err := s.repos.ApiKey.GetByID(req.OrganizationID, req.ApiKeyID)
	if err != nil {
		return nil, err
	}

	updated := *apiKey
	updated.Name = strings.TrimSpace(req.Name)
	updated.Permissions = req.Permissions

	if err := s.repos.ApiKey.Update(nil, &updated); err != nil {
		return nil, err
	}

	changes, err := json.Marshal(map[string]any{
		"previous": map[string]any{
			"name":        apiKey.Name,
			"permissions": apiKey.Permissions,
		},
		"updated": map[string]any{
			"name":        updated.Name,
			"permissions": updated.Permissions,
		},
	})
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.api_keyService.Update",
			Error:    err,
			Message:  "Failed to marshal changes",
			Data:     map[string]any{"organization_id": req.OrganizationID, "api_key_id": req.ApiKeyID},
		})
	} else {
		changesRaw := json.RawMessage(changes)
		nameCopy := updated.Name
		if _, err := s.auditLogService.Create(&audit_log_service.CreateRequest{
			OrganizationID: org.ID.String(),
			Action:         constants.AUDIT_ACTION_TYPE_UPDATE,
			ResourceType:   constants.RESOURCE_TYPE_API_KEY,
			ResourceID:     apiKey.ID.String(),
			ResourceName:   &nameCopy,
			Data:           &changesRaw,
		}, accessInfo); err != nil {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.api_keyService.Update",
				Error:    err,
				Message:  "Failed to create audit log",
				Data:     map[string]any{"organization_id": req.OrganizationID, "api_key_id": req.ApiKeyID},
			})
		}
	}

	return &updated, nil
}
