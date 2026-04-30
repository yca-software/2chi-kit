package invitation_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type ListRequest struct {
	OrganizationID string `json:"-"`
}

func (s *service) List(req *ListRequest, accessInfo *models.AccessInfo) (*[]models.Invitation, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	organization, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizer.CheckOrganizationPermission(accessInfo, organization.ID.String(), constants.PERMISSION_MEMBERS_READ); err != nil {
		return nil, err
	}

	invitations, err := s.repos.Invitation.ListByOrganizationID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	return invitations, nil
}
