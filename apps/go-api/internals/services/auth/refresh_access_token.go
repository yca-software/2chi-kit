package auth_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type RefreshAccessTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
	IPAddress    string `json:"ipAddress" validate:"required,ip"`
	UserAgent    string `json:"userAgent" validate:"required"`
}

type RefreshAccessTokenResponse struct {
	AccessToken string `json:"accessToken"`
}

func (s *service) RefreshAccessToken(req *RefreshAccessTokenRequest) (*RefreshAccessTokenResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	tokenHash := s.hashToken(req.RefreshToken)
	refreshToken, err := s.repos.UserRefreshToken.GetByHash(nil, tokenHash)
	if err != nil {
		if e, ok := yca_error.AsError(err); ok && e.ErrorCode == constants.NOT_FOUND_CODE {
			return nil, yca_error.NewUnauthorizedError(nil, constants.INVALID_TOKEN_CODE, nil)
		}
		return nil, err
	}

	user, err := s.repos.User.GetByID(nil, refreshToken.UserID.String())
	if err != nil {
		if e, ok := yca_error.AsError(err); ok && e.ErrorCode == constants.NOT_FOUND_CODE {
			return nil, yca_error.NewUnauthorizedError(nil, constants.INVALID_TOKEN_CODE, nil)
		}
		return nil, err
	}

	if refreshToken.RevokedAt != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.auth.RefreshAccessToken",
			Message:  "Refresh attempt with revoked token",
			Data:     map[string]any{"refresh_token_id": refreshToken.ID, "user_id": refreshToken.UserID},
		})
		return nil, yca_error.NewUnauthorizedError(nil, constants.INVALID_TOKEN_CODE, nil)
	}

	if refreshToken.ExpiresAt.Before(s.now()) {
		return nil, yca_error.NewUnauthorizedError(nil, constants.EXPIRED_TOKEN_CODE, nil)
	}

	var impersonatedBy, impersonatedByEmail string
	if refreshToken.ImpersonatedBy.Valid {
		impersonatedBy = refreshToken.ImpersonatedBy.UUID.String()

		impersonatedByUser, err := s.repos.User.GetByID(nil, impersonatedBy)
		if err != nil {
			if e, ok := yca_error.AsError(err); ok && e.ErrorCode == constants.NOT_FOUND_CODE {
				return nil, yca_error.NewUnauthorizedError(nil, constants.INVALID_TOKEN_CODE, nil)
			}
			return nil, err
		}

		impersonatedByEmail = impersonatedByUser.Email
	}
	accessToken, err := s.generateAccessToken(user, impersonatedBy, impersonatedByEmail)
	if err != nil {
		return nil, err
	}

	return &RefreshAccessTokenResponse{
		AccessToken: accessToken,
	}, nil
}
