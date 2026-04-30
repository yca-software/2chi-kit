package auth_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

func (s *service) Logout(req *LogoutRequest, accessInfo *models.AccessInfo) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	tokenHash := s.hashToken(req.RefreshToken)
	refreshToken, err := s.repos.UserRefreshToken.GetByHash(nil, tokenHash)
	if err != nil {
		if e, ok := yca_error.AsError(err); ok && e.ErrorCode == constants.NOT_FOUND_CODE {
			return yca_error.NewUnauthorizedError(nil, constants.INVALID_TOKEN_CODE, nil)
		}
		return err
	}

	if accessInfo == nil || accessInfo.User == nil || accessInfo.User.UserID != refreshToken.UserID {
		return yca_error.NewForbiddenError(nil, constants.FORBIDDEN_CODE, nil)
	}

	if refreshToken.RevokedAt != nil {
		return yca_error.NewUnauthorizedError(nil, constants.INVALID_TOKEN_CODE, nil)
	}

	if refreshToken.ExpiresAt.Before(s.now()) {
		return yca_error.NewUnauthorizedError(nil, constants.EXPIRED_TOKEN_CODE, nil)
	}

	if err := s.repos.UserRefreshToken.RevokeByHash(nil, tokenHash); err != nil {
		return err
	}

	return nil
}
