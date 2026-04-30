package user_refresh_token_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type ListActiveRequest struct {
	UserID string `json:"-" validate:"required,uuid"`
}

func (s *service) ListActive(req *ListActiveRequest, accessInfo *models.AccessInfo) (*[]models.UserRefreshToken, error) {
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

	tokens, err := s.repos.UserRefreshToken.GetActiveByUserID(nil, user.ID.String())
	if err != nil {
		return nil, err
	}

	return tokens, nil
}
