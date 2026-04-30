package role_service

import (
	"encoding/json"
	"strings"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type UpdateRequest struct {
	OrganizationID string                 `json:"-" validate:"required,uuid"`
	RoleID         string                 `json:"-" validate:"required,uuid"`
	Name           string                 `json:"name" validate:"required,min=1,max=255"`
	Description    string                 `json:"description"`
	Permissions    models.RolePermissions `json:"permissions" validate:"required,min=1"`
}

func (s *service) Update(req *UpdateRequest, accessInfo *models.AccessInfo) (*models.Role, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, org, constants.PERMISSION_ROLE_WRITE); err != nil {
		return nil, err
	}
	if err := s.authorizer.CheckOrganizationFeature(accessInfo, org, constants.FEATURE_ROLES); err != nil {
		return nil, err
	}

	role, err := s.repos.Role.GetByID(req.OrganizationID, req.RoleID)
	if err != nil {
		return nil, err
	}

	if role.Locked {
		return nil, yca_error.NewForbiddenError(nil, constants.ROLE_LOCKED_CODE, nil)
	}

	updatedRole := *role
	updatedRole.Name = strings.TrimSpace(req.Name)
	updatedRole.Description = strings.TrimSpace(req.Description)
	updatedRole.Permissions = req.Permissions

	if err := s.repos.Role.Update(nil, &updatedRole); err != nil {
		return nil, err
	}

	changes, err := json.Marshal(map[string]any{
		"previous": map[string]any{
			"name":        role.Name,
			"description": role.Description,
			"permissions": role.Permissions,
		},
		"updated": map[string]any{
			"name":        req.Name,
			"description": req.Description,
			"permissions": req.Permissions,
		},
	})
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.roleService.Update",
			Error:    err,
			Message:  "Failed to marshal changes",
			Data:     map[string]any{"organization_id": req.OrganizationID, "role_id": req.RoleID},
		})
	} else {
		changesRaw := json.RawMessage(changes)
		nameCopy := updatedRole.Name
		if _, err := s.auditLogService.Create(&audit_log_service.CreateRequest{
			OrganizationID: org.ID.String(),
			Action:         constants.AUDIT_ACTION_TYPE_UPDATE,
			ResourceType:   constants.RESOURCE_TYPE_ROLE,
			ResourceID:     role.ID.String(),
			ResourceName:   &nameCopy,
			Data:           &changesRaw,
		}, accessInfo); err != nil {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.roleService.Update",
				Error:    err,
				Message:  "Failed to create audit log",
				Data:     map[string]any{"organization_id": req.OrganizationID, "role_id": req.RoleID},
			})
		}
	}

	return &updatedRole, nil
}
