package user_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type UpdateLanguageRequest struct {
	UserID   string `json:"-" validate:"required,uuid"`
	Language string `json:"language" validate:"required,len=2"`
}

func (s *service) UpdateLanguage(req *UpdateLanguageRequest, accessInfo *models.AccessInfo) (*models.User, error) {
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

	user.Language = req.Language

	if err := s.repos.User.Update(nil, user); err != nil {
		return nil, err
	}

	return user, nil
}
