package user_refresh_token_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type RevokeRequest struct {
	UserID         string `json:"userId" validate:"required,uuid"`
	RefreshTokenID string `json:"refreshTokenId" validate:"required,uuid"`
}

func (s *service) Revoke(req *RevokeRequest, accessInfo *models.AccessInfo) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	user, err := s.repos.User.GetByID(nil, req.UserID)
	if err != nil {
		return err
	}

	if err := s.authorizer.CheckOwnResource(accessInfo, user.ID.String()); err != nil {
		return err
	}

	if err := s.repos.UserRefreshToken.Revoke(nil, user.ID.String(), req.RefreshTokenID); err != nil {
		return err
	}

	return nil
}
