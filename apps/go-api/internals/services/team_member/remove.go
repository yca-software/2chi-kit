package team_member_service

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
	TeamID         string `json:"-" validate:"required,uuid"`
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

	team, err := s.repos.Team.GetByID(org.ID.String(), req.TeamID)
	if err != nil {
		return err
	}

	if err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, org, constants.PERMISSION_TEAM_MEMBER_DELETE); err != nil {
		return err
	}

	member, err := s.repos.TeamMember.GetByID(org.ID.String(), req.MemberID)
	if err != nil {
		return err
	}

	user, err := s.repos.User.GetByID(nil, member.UserID.String())
	if err != nil {
		return err
	}

	if err := s.repos.TeamMember.Delete(nil, org.ID.String(), member.ID.String()); err != nil {
		return err
	}

	metadata, err := json.Marshal(map[string]any{
		"userId":    user.ID,
		"userEmail": user.Email,
		"teamId":    team.ID,
		"teamName":  team.Name,
	})
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.teamMemberService.Remove",
			Error:    err,
			Message:  "Failed to marshal metadata",
			Data:     map[string]any{"organization_id": req.OrganizationID, "team_id": req.TeamID, "member_id": req.MemberID},
		})
	} else {
		metadataRaw := json.RawMessage(metadata)
		teamNameCopy := team.Name
		if _, err := s.auditLogService.Create(&audit_log_service.CreateRequest{
			OrganizationID: org.ID.String(),
			Action:         constants.AUDIT_ACTION_TYPE_DELETE,
			ResourceType:   constants.RESOURCE_TYPE_TEAM_MEMBER,
			ResourceID:     member.ID.String(),
			ResourceName:   &teamNameCopy,
			Data:           &metadataRaw,
		}, accessInfo); err != nil {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.teamMemberService.Remove",
				Error:    err,
				Message:  "Failed to create audit log",
				Data:     map[string]any{"organization_id": req.OrganizationID, "team_id": req.TeamID, "member_id": req.MemberID},
			})
		}
	}

	return nil
}
