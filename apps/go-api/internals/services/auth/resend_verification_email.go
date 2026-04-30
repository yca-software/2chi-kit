package auth_service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type ResendVerificationEmailRequest struct {
	UserID   string `json:"-" validate:"required,uuid"`
	Language string `json:"-" validate:"omitempty,len=2"`
}

func (s *service) ResendVerificationEmail(req *ResendVerificationEmailRequest) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	user, err := s.repos.User.GetByID(nil, req.UserID)
	if err != nil {
		return err
	}

	if user.EmailVerifiedAt != nil {
		return yca_error.NewConflictError(nil, constants.EMAIL_ALREADY_VERIFIED_CODE, nil)
	}

	now := s.now()

	verificationToken, err := s.generateToken()
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.auth.ResendVerificationEmail",
			Message:  "Failed to generate email verification token",
			Data:     map[string]any{"user_id": user.ID},
			Error:    err,
		})
		return err
	}

	verificationTokenID, err := s.generateID()
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.auth.ResendVerificationEmail",
			Message:  "Failed to generate email verification token ID",
			Data:     map[string]any{"user_id": user.ID},
			Error:    err,
		})
		return err
	}

	tx, err := s.repos.User.BeginTx()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.auth.ResendVerificationEmail",
				Message:  "Transaction rollback failed",
				Data:     map[string]any{"user_id": user.ID},
				Error:    err,
			})
		}
	}()

	if err := s.repos.UserEmailVerificationToken.Create(tx, &models.UserEmailVerificationToken{
		ID:        verificationTokenID,
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(s.emailVerificationTTL) * time.Hour),
		TokenHash: s.hashToken(verificationToken),
	}); err != nil {
		return err
	}

	language := req.Language
	if language == "" {
		language = user.Language
	}

	emailBody, err := s.emailService.PrepareEmailBody("verification", map[string]any{
		"Lang":         language,
		"Title":        s.translator.Translate(language, "email.verification.title", nil),
		"Greeting":     s.translator.Translate(language, "email.verification.greeting", nil),
		"Content":      s.translator.Translate(language, "email.verification.content", nil),
		"Warning":      s.translator.Translate(language, "email.verification.warning", nil),
		"ButtonText":   s.translator.Translate(language, "email.verification.button", nil),
		"FooterIgnore": s.translator.Translate(language, "email.verification.footer.ignore", nil),
		"FooterLink":   s.translator.Translate(language, "email.verification.footer.link", nil),
		"C2ALink":      fmt.Sprintf("%s/verify-email?token=%s", s.appURL, verificationToken),
	})
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.auth.ResendVerificationEmail",
			Message:  "Failed to prepare email verification email body",
			Data:     map[string]any{"user_id": user.ID},
			Error:    err,
		})
		return yca_error.NewInternalServerError(err, "", nil)
	}

	subject := s.translator.Translate(language, "email.verification.subject", nil)
	if err := s.emailService.SendEmail(user.Email, subject, emailBody); err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.auth.ResendVerificationEmail",
			Message:  "Failed to send email verification email",
			Data:     map[string]any{"user_id": user.ID},
			Error:    err,
		})
		return yca_error.NewInternalServerError(err, "", nil)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
