package organization_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

func (s *service) GetArchived(req *GetRequest, accessInfo *models.AccessInfo) (*models.Organization, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	if err := s.authorizer.CheckAdmin(accessInfo); err != nil {
		return nil, err
	}

	org, err := s.repos.Organization.GetByIDIncludeArchived(req.OrganizationID)
	if err != nil {
		return nil, err
	}
	if org.DeletedAt == nil {
		return nil, yca_error.NewNotFoundError(nil, constants.NOT_FOUND_CODE, nil)
	}

	return org, nil
}
