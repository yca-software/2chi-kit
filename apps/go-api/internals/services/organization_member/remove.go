package organization_member_service

import (
	"encoding/json"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type RemoveRequest struct {
	OrganizationID string `json:"-" validate:"required,uuid"`
	MemberID       string `json:"-" validate:"required,uuid"`
}

func (s *service) Remove(req *RemoveRequest, accessInfo *models.AccessInfo) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return err
	}

	if err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, org, constants.PERMISSION_MEMBERS_DELETE); err != nil {
		return err
	}

	member, err := s.repos.OrganizationMember.GetByID(org.ID.String(), req.MemberID)
	if err != nil {
		return err
	}

	if accessInfo.User != nil && member.UserID == accessInfo.User.UserID {
		return yca_error.NewForbiddenError(nil, constants.USER_CANNOT_REMOVE_OWN_MEMBER_CODE, nil)
	}

	user, err := s.repos.User.GetByID(nil, member.UserID.String())
	if err != nil {
		return err
	}

	role, err := s.repos.Role.GetByID(org.ID.String(), member.RoleID.String())
	if err != nil {
		return err
	}

	if err := s.repos.OrganizationMember.Delete(nil, org.ID.String(), member.ID.String()); err != nil {
		return err
	}

	data, err := json.Marshal(map[string]any{
		"metadata": map[string]any{
			"userId":    user.ID,
			"userEmail": user.Email,
			"roleId":    role.ID,
			"roleName":  role.Name,
		},
	})
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.organizationMemberService.Remove",
			Error:    err,
			Message:  "Failed to marshal metadata",
			Data:     map[string]any{"organization_id": req.OrganizationID, "member_id": req.MemberID},
		})
	} else {
		dataRaw := json.RawMessage(data)
		rn := "Organization member"
		if _, err := s.auditLogService.Create(&audit_log_service.CreateRequest{
			OrganizationID: org.ID.String(),
			Action:         constants.AUDIT_ACTION_TYPE_DELETE,
			ResourceType:   constants.RESOURCE_TYPE_MEMBER,
			ResourceID:     member.ID.String(),
			ResourceName:   &rn,
			Data:           &dataRaw,
		}, accessInfo); err != nil {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.organizationMemberService.Remove",
				Error:    err,
				Message:  "Failed to create audit log",
				Data:     map[string]any{"organization_id": req.OrganizationID, "member_id": req.MemberID},
			})
		}
	}

	return nil
}
