package team_member_service

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

type AddRequest struct {
	OrganizationID string `json:"-" validate:"required,uuid"`
	TeamID         string `json:"-" validate:"required,uuid"`
	UserID         string `json:"userId" validate:"required,uuid"`
}

func (s *service) Add(req *AddRequest, accessInfo *models.AccessInfo) (*models.TeamMemberWithUser, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, org, constants.PERMISSION_TEAM_MEMBER_WRITE); err != nil {
		return nil, err
	}

	team, err := s.repos.Team.GetByID(org.ID.String(), req.TeamID)
	if err != nil {
		return nil, err
	}

	user, err := s.repos.User.GetByID(nil, req.UserID)
	if err != nil {
		return nil, err
	}

	if _, err := s.repos.OrganizationMember.GetByUserIDAndOrganizationID(user.ID.String(), org.ID.String()); err != nil {
		return nil, err
	}

	now := s.now()
	memberID, err := s.generateID()
	if err != nil {
		return nil, err
	}

	tx, err := s.repos.Team.BeginTx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.teamMemberService.Add",
				Message:  "Transaction rollback failed",
				Error:    err,
			})
		}
	}()

	member := &models.TeamMember{
		ID:             memberID,
		CreatedAt:      now,
		OrganizationID: org.ID,
		TeamID:         team.ID,
		UserID:         user.ID,
	}

	if err := s.repos.TeamMember.Create(tx, member); err != nil {
		return nil, err
	}

	memberWithUser, err := s.repos.TeamMember.GetByIDWithUser(tx, org.ID.String(), member.ID.String())
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
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
			Location: "services.teamMemberService.Add",
			Error:    err,
			Message:  "Failed to marshal metadata",
			Data:     map[string]any{"organization_id": req.OrganizationID, "team_id": req.TeamID, "user_id": req.UserID},
		})
	} else {
		metadataRaw := json.RawMessage(metadata)
		teamNameCopy := team.Name
		if _, err := s.auditLogService.Create(&audit_log_service.CreateRequest{
			OrganizationID: org.ID.String(),
			Action:         constants.AUDIT_ACTION_TYPE_CREATE,
			ResourceType:   constants.RESOURCE_TYPE_TEAM_MEMBER,
			ResourceID:     member.ID.String(),
			ResourceName:   &teamNameCopy,
			Data:           &metadataRaw,
		}, accessInfo); err != nil {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.teamMemberService.Add",
				Error:    err,
				Message:  "Failed to create audit log",
				Data:     map[string]any{"organization_id": req.OrganizationID, "team_id": req.TeamID, "user_id": req.UserID},
			})
		}
	}

	return memberWithUser, nil
}
