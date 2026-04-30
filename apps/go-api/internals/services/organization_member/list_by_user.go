package organization_member_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type ListByUserRequest struct {
	UserID string `json:"-" validate:"required,uuid"`
}

func (s *service) ListByUser(req *ListByUserRequest, accessInfo *models.AccessInfo) (*[]models.OrganizationMemberWithOrganization, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	if err := s.authorizer.CheckOwnResource(accessInfo, req.UserID); err != nil {
		return nil, err
	}

	members, err := s.repos.OrganizationMember.ListByUserID(req.UserID)
	if err != nil {
		return nil, err
	}

	return members, nil
}
