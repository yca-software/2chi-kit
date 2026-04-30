package auth_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	yca_error "github.com/yca-software/go-common/error"
)

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

func (s *service) VerifyEmail(req *VerifyEmailRequest) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	tokenHash := s.hashToken(req.Token)

	verificationToken, err := s.repos.UserEmailVerificationToken.GetByHash(nil, tokenHash)
	if err != nil {
		if e, ok := err.(*yca_error.Error); ok {
			if e.ErrorCode == constants.NOT_FOUND_CODE {
				return yca_error.NewUnauthorizedError(nil, constants.INVALID_VERIFICATION_TOKEN_CODE, nil)
			}
		}
		return err
	}

	if _, err := s.repos.User.GetByID(nil, verificationToken.UserID.String()); err != nil {
		return err
	}

	if verificationToken.ExpiresAt.Before(s.now()) {
		return yca_error.NewUnauthorizedError(nil, constants.EXPIRED_VERIFICATION_TOKEN_CODE, nil)
	}

	if verificationToken.UsedAt != nil {
		return yca_error.NewUnauthorizedError(nil, constants.INVALID_VERIFICATION_TOKEN_CODE, nil)
	}

	if err := s.repos.UserEmailVerificationToken.MarkAsUsed(nil, verificationToken.ID.String()); err != nil {
		return err
	}

	return nil
}
