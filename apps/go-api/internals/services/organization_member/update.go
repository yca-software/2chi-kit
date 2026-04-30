package organization_member_service

import (
	"encoding/json"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type UpdateRequest struct {
	OrganizationID string `json:"-" validate:"required,uuid"`
	MemberID       string `json:"-" validate:"required,uuid"`
	RoleID         string `json:"roleId" validate:"required,uuid"`
}

func (s *service) Update(req *UpdateRequest, accessInfo *models.AccessInfo) (*models.OrganizationMemberWithUser, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, org, constants.PERMISSION_MEMBERS_WRITE); err != nil {
		return nil, err
	}

	member, err := s.repos.OrganizationMember.GetByID(org.ID.String(), req.MemberID)
	if err != nil {
		return nil, err
	}

	if accessInfo.User != nil && member.UserID == accessInfo.User.UserID {
		return nil, yca_error.NewForbiddenError(nil, constants.USER_CANNOT_UPDATE_OWN_MEMBER_CODE, nil)
	}

	currentRole, err := s.repos.Role.GetByID(org.ID.String(), member.RoleID.String())
	if err != nil {
		return nil, err
	}

	newRole, err := s.repos.Role.GetByID(org.ID.String(), req.RoleID)
	if err != nil {
		return nil, err
	}

	member.RoleID = newRole.ID
	if err := s.repos.OrganizationMember.Update(nil, member); err != nil {
		return nil, err
	}

	memberWithUser, err := s.repos.OrganizationMember.GetByIDWithUser(org.ID.String(), member.ID.String())
	if err != nil {
		return nil, err
	}

	changes, err := json.Marshal(map[string]any{
		"previous": map[string]any{
			"roleId":   currentRole.ID,
			"roleName": currentRole.Name,
		},
		"updated": map[string]any{
			"roleId":   newRole.ID,
			"roleName": newRole.Name,
		},
	})
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.organizationMemberService.UpdateRole",
			Error:    err,
			Message:  "Failed to marshal changes",
			Data:     map[string]any{"organization_id": req.OrganizationID, "member_id": req.MemberID},
		})
	} else {
		changesRaw := json.RawMessage(changes)
		rn := "Organization member"
		if _, err := s.auditLogService.Create(&audit_log_service.CreateRequest{
			OrganizationID: org.ID.String(),
			Action:         constants.AUDIT_ACTION_TYPE_UPDATE,
			ResourceType:   constants.RESOURCE_TYPE_MEMBER,
			ResourceID:     member.ID.String(),
			ResourceName:   &rn,
			Data:           &changesRaw,
		}, accessInfo); err != nil {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.organizationMemberService.UpdateRole",
				Error:    err,
				Message:  "Failed to create audit log",
				Data:     map[string]any{"organization_id": req.OrganizationID, "member_id": req.MemberID},
			})
		}
	}

	return memberWithUser, nil
}
