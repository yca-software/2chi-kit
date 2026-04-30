package user_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type DeleteRequest struct {
	UserID string `json:"-" validate:"required,uuid"`
}

func (s *service) Delete(req *DeleteRequest, accessInfo *models.AccessInfo) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	user, err := s.repos.User.GetByID(nil, req.UserID)
	if err != nil {
		return err
	}

	if err := s.authorizer.CheckAdmin(accessInfo); err != nil {
		return err
	}

	if err := s.repos.User.Delete(nil, user); err != nil {
		return err
	}

	return nil
}
