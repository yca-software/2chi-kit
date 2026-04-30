package auth_service

import (
	"database/sql"
	"errors"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

func (s *service) ResetPassword(req *ResetPasswordRequest) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	tokenHash := s.hashToken(req.Token)
	resetToken, err := s.repos.UserPasswordResetToken.GetByHash(nil, tokenHash)
	if err != nil {
		if e, ok := err.(*yca_error.Error); ok {
			if e.ErrorCode == constants.NOT_FOUND_CODE {
				return yca_error.NewUnauthorizedError(nil, constants.INVALID_PASSWORD_RESET_TOKEN_CODE, nil)
			}
		}
		return err
	}

	user, err := s.repos.User.GetByID(nil, resetToken.UserID.String())
	if err != nil {
		return err
	}

	now := s.now()

	if resetToken.ExpiresAt.Before(now) {
		return yca_error.NewUnauthorizedError(nil, constants.EXPIRED_PASSWORD_RESET_TOKEN_CODE, nil)
	}

	if resetToken.UsedAt != nil {
		return yca_error.NewUnauthorizedError(nil, constants.INVALID_PASSWORD_RESET_TOKEN_CODE, nil)
	}

	tx, err := s.repos.User.BeginTx()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.auth.ResetPassword",
				Message:  "Transaction rollback failed",
				Error:    err,
				Data:     map[string]any{"user_id": user.ID},
			})
		}
	}()

	if err := s.repos.UserPasswordResetToken.MarkAsUsed(tx, resetToken.ID.String()); err != nil {
		return err
	}

	hashedPassword, err := s.passwordHashFn(req.Password)
	if err != nil {
		return err
	}
	user.Password = &hashedPassword

	if err := s.repos.User.Update(tx, user); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
