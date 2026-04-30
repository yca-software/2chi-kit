package organization_member_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type ListByOrganizationRequest struct {
	OrganizationID string `json:"-" validate:"required,uuid"`
}

func (s *service) ListByOrganization(req *ListByOrganizationRequest, accessInfo *models.AccessInfo) (*[]models.OrganizationMemberWithUser, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizer.CheckOrganizationPermission(accessInfo, org.ID.String(), constants.PERMISSION_MEMBERS_READ); err != nil {
		return nil, err
	}

	members, err := s.repos.OrganizationMember.ListByOrganizationID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	return members, nil
}
