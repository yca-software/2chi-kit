package team_service

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
	OrganizationID string `json:"-" validate:"required,uuid"`
	TeamID         string `json:"-" validate:"required,uuid"`
	Name           string `json:"name" validate:"required,min=1,max=255"`
	Description    string `json:"description"`
}

func (s *service) Update(req *UpdateRequest, accessInfo *models.AccessInfo) (*models.Team, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, org, constants.PERMISSION_TEAM_WRITE); err != nil {
		return nil, err
	}
	if err := s.authorizer.CheckOrganizationFeature(accessInfo, org, constants.FEATURE_TEAMS); err != nil {
		return nil, err
	}

	team, err := s.repos.Team.GetByID(org.ID.String(), req.TeamID)
	if err != nil {
		return nil, err
	}

	updatedTeam := *team
	updatedTeam.Name = strings.TrimSpace(req.Name)
	updatedTeam.Description = req.Description

	if err := s.repos.Team.Update(nil, &updatedTeam); err != nil {
		return nil, err
	}

	changes, err := json.Marshal(map[string]any{
		"previous": map[string]any{
			"name":        team.Name,
			"description": team.Description,
		},
		"updated": map[string]any{
			"name":        req.Name,
			"description": req.Description,
		},
	})
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.teamService.Update",
			Error:    err,
			Message:  "Failed to marshal changes",
			Data:     map[string]any{"organization_id": req.OrganizationID, "team_id": req.TeamID},
		})
	} else {
		changesRaw := json.RawMessage(changes)
		nameCopy := updatedTeam.Name
		if _, err := s.auditLogService.Create(&audit_log_service.CreateRequest{
			OrganizationID: org.ID.String(),
			Action:         constants.AUDIT_ACTION_TYPE_UPDATE,
			ResourceType:   constants.RESOURCE_TYPE_TEAM,
			ResourceID:     team.ID.String(),
			ResourceName:   &nameCopy,
			Data:           &changesRaw,
		}, accessInfo); err != nil {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.teamService.Update",
				Error:    err,
				Message:  "Failed to create audit log",
				Data:     map[string]any{"organization_id": req.OrganizationID, "team_id": req.TeamID},
			})
		}
	}

	return &updatedTeam, nil
}
