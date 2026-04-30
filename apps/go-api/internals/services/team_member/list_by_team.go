package team_member_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type ListByTeamRequest struct {
	OrganizationID string `json:"-" validate:"required,uuid"`
	TeamID         string `json:"-" validate:"required,uuid"`
}

func (s *service) ListByTeam(req *ListByTeamRequest, accessInfo *models.AccessInfo) (*[]models.TeamMemberWithUser, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, org, constants.PERMISSION_TEAM_MEMBER_READ); err != nil {
		return nil, err
	}

	team, err := s.repos.Team.GetByID(req.OrganizationID, req.TeamID)
	if err != nil {
		return nil, err
	}

	members, err := s.repos.TeamMember.ListByTeamID(org.ID.String(), team.ID.String())
	if err != nil {
		return nil, err
	}

	return members, nil
}
