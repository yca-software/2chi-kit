package api_key_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type ListRequest struct {
	OrganizationID string `json:"organizationId" validate:"required,uuid"`
}

func (s *service) List(req *ListRequest, accessInfo *models.AccessInfo) (*[]models.APIKey, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizer.CheckOrganizationPermission(accessInfo, org.ID.String(), constants.PERMISSION_API_KEY_READ); err != nil {
		return nil, err
	}
	if err := s.authorizer.CheckOrganizationFeature(accessInfo, org, constants.FEATURE_API_ACCESS); err != nil {
		return nil, err
	}

	return s.repos.ApiKey.ListByOrganizationID(req.OrganizationID)
}
