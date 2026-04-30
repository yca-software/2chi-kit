package user_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type GetRequest struct {
	UserID string `json:"-" validate:"required,uuid"`
}

type GetResponse struct {
	User        *models.User                                        `json:"user"`
	AdminAccess *models.AdminAccess                                 `json:"adminAccess"`
	Roles       *[]models.OrganizationMemberWithOrganizationAndRole `json:"roles"`
}

func (s *service) Get(req *GetRequest, accessInfo *models.AccessInfo) (*GetResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	user, err := s.repos.User.GetByID(nil, req.UserID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizer.CheckOwnResource(accessInfo, user.ID.String()); err != nil {
		return nil, err
	}

	adminAccess, err := s.repos.AdminAccess.GetByUserID(req.UserID)
	if err != nil {
		if e, ok := err.(*yca_error.Error); ok {
			if e.ErrorCode != constants.NOT_FOUND_CODE {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	roles, err := s.repos.OrganizationMember.ListByUserIDWithRole(req.UserID)
	if err != nil {
		return nil, err
	}

	return &GetResponse{
		User:        user,
		AdminAccess: adminAccess,
		Roles:       roles,
	}, nil
}
