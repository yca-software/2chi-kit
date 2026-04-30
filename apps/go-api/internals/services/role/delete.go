package role_service

import (
	"encoding/json"
	"strings"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_error "github.com/yca-software/go-common/error"
	logger "github.com/yca-software/go-common/logger"
)

type DeleteRequest struct {
	OrganizationID string `json:"-" validate:"required,uuid"`
	RoleID         string `json:"-" validate:"required,uuid"`
}

func (s *service) Delete(req *DeleteRequest, accessInfo *models.AccessInfo) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return err
	}

	if err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, org, constants.PERMISSION_ROLE_DELETE); err != nil {
		return err
	}
	if err := s.authorizer.CheckOrganizationFeature(accessInfo, org, constants.FEATURE_ROLES); err != nil {
		return err
	}

	role, err := s.repos.Role.GetByID(req.OrganizationID, req.RoleID)
	if err != nil {
		return err
	}

	if role.Locked {
		return yca_error.NewForbiddenError(nil, constants.ROLE_LOCKED_CODE, nil)
	}

	emails, err := s.repos.OrganizationMember.ListUserEmailsForRole(req.OrganizationID, req.RoleID)
	if err != nil {
		return err
	}
	if len(emails) > 0 {
		return yca_error.NewConflictError(nil, constants.ROLE_HAS_MEMBERS_CODE, map[string]any{
			"memberEmails":     emails,
			"memberEmailsText": strings.Join(emails, ", "),
		})
	}

	if err := s.repos.Role.Delete(nil, req.OrganizationID, req.RoleID); err != nil {
		return err
	}

	changes, err := json.Marshal(map[string]any{
		"previous": map[string]any{
			"name":        role.Name,
			"description": role.Description,
			"permissions": role.Permissions,
		},
	})
	if err != nil {
		s.logger.Log(logger.LogData{
			Level:    "error",
			Location: "services.roleService.Delete",
			Error:    err,
			Message:  "Failed to marshal changes",
			Data:     map[string]any{"organization_id": req.OrganizationID, "role_id": req.RoleID},
		})
	} else {
		changesRaw := json.RawMessage(changes)
		nameCopy := role.Name
		if _, err := s.auditLogService.Create(&audit_log_service.CreateRequest{
			OrganizationID: org.ID.String(),
			Action:         constants.AUDIT_ACTION_TYPE_DELETE,
			ResourceType:   constants.RESOURCE_TYPE_ROLE,
			ResourceID:     role.ID.String(),
			ResourceName:   &nameCopy,
			Data:           &changesRaw,
		}, accessInfo); err != nil {
			s.logger.Log(logger.LogData{
				Level:    "error",
				Location: "services.roleService.Delete",
				Error:    err,
				Message:  "Failed to create audit log",
				Data:     map[string]any{"organization_id": req.OrganizationID, "role_id": req.RoleID},
			})
		}
	}

	return nil
}
