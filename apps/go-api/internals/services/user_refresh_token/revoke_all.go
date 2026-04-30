package user_refresh_token_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type RevokeAllRequest struct {
	UserID           string `json:"userId" validate:"required,uuid"`
	KeepRefreshToken string `json:"keepRefreshToken,omitempty"`
}

func (s *service) RevokeAll(req *RevokeAllRequest, accessInfo *models.AccessInfo) error {
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

	if req.KeepRefreshToken != "" {
		tokenHash := s.hashToken(req.KeepRefreshToken)
		current, err := s.repos.UserRefreshToken.GetByHash(nil, tokenHash)
		if err != nil {
			if e, ok := yca_error.AsError(err); ok && e.ErrorCode == constants.NOT_FOUND_CODE {
				return yca_error.NewUnprocessableEntityError(nil, constants.INVALID_TOKEN_CODE, nil)
			}
			return err
		}
		if current.UserID.String() != user.ID.String() {
			return yca_error.NewForbiddenError(nil, constants.FORBIDDEN_CODE, nil)
		}
		if current.RevokedAt != nil || current.ExpiresAt.Before(s.now()) {
			return yca_error.NewUnprocessableEntityError(nil, constants.INVALID_TOKEN_CODE, nil)
		}
		if err := s.repos.UserRefreshToken.RevokeAllExcept(nil, user.ID.String(), current.ID.String()); err != nil {
			return err
		}
		return nil
	}

	if err := s.repos.UserRefreshToken.RevokeAll(nil, user.ID.String()); err != nil {
		return err
	}

	return nil
}
