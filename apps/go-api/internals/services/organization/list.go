package organization_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type ListRequest struct {
	SearchPhrase string `json:"-"`
	Limit        int    `json:"-" validate:"required,min=1,max=100"`
	Offset       int    `json:"-" validate:"gte=0"`
}

type PaginatedListResponse models.PaginatedListResponse[models.Organization]

func (s *service) List(req *ListRequest, accessInfo *models.AccessInfo) (*PaginatedListResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	if err := s.authorizer.CheckAdmin(accessInfo); err != nil {
		return nil, err
	}

	orgs, err := s.repos.Organization.Search(req.SearchPhrase, req.Limit+1, req.Offset)
	if err != nil {
		return nil, err
	}

	hasNext := len(*orgs) > req.Limit
	if hasNext {
		items := (*orgs)[:req.Limit]
		orgs = &items
	}

	return &PaginatedListResponse{Items: *orgs, HasNext: hasNext}, nil
}
