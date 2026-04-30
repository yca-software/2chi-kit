package auth_service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type ForgotPasswordRequest struct {
	Language string `json:"language" validate:"required,len=2"`
	Email    string `json:"email" validate:"required,email"`
}

func (s *service) ForgotPassword(req *ForgotPasswordRequest) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	emailLower := strings.ToLower(req.Email)

	user, err := s.repos.User.GetByEmail(nil, emailLower)
	if err != nil {
		if e, ok := err.(*yca_error.Error); ok {
			if e.ErrorCode == constants.NOT_FOUND_CODE {
				return nil
			}
		}
		return err
	}

	now := s.now()

	resetToken, err := s.generateToken()
	if err != nil {
		return err
	}

	resetTokenID, err := s.generateID()
	if err != nil {
		return err
	}

	tx, err := s.repos.UserPasswordResetToken.BeginTx()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.auth.ForgotPassword",
				Message:  "Transaction rollback failed",
				Error:    err,
			})
		}
	}()

	if err := s.repos.UserPasswordResetToken.Create(tx, &models.UserPasswordResetToken{
		ID:        resetTokenID,
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(s.passwordResetTTL) * time.Hour),
		TokenHash: s.hashToken(resetToken),
	}); err != nil {
		return err
	}

	emailBody, err := s.emailService.PrepareEmailBody("reset", map[string]any{
		"Lang":         req.Language,
		"Title":        s.translator.Translate(req.Language, "email.reset.title", nil),
		"Greeting":     s.translator.Translate(req.Language, "email.reset.greeting", nil),
		"Content":      s.translator.Translate(req.Language, "email.reset.content", nil),
		"Warning":      s.translator.Translate(req.Language, "email.reset.warning", nil),
		"ButtonText":   s.translator.Translate(req.Language, "email.reset.button", nil),
		"FooterIgnore": s.translator.Translate(req.Language, "email.reset.footer.ignore", nil),
		"FooterLink":   s.translator.Translate(req.Language, "email.reset.footer.link", nil),
		"C2ALink":      fmt.Sprintf("%s/reset-password?token=%s", s.appURL, resetToken),
	})
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.auth.ForgotPassword",
			Message:  "Failed to prepare email reset email body",
			Data:     map[string]any{"user_id": user.ID},
			Error:    err,
		})
		return err
	}

	subject := s.translator.Translate(req.Language, "email.reset.subject", nil)
	if err := s.emailService.SendEmail(user.Email, subject, emailBody); err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.auth.ForgotPassword",
			Message:  "Failed to send email reset email",
			Data:     map[string]any{"user_id": user.ID},
			Error:    err,
		})
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
