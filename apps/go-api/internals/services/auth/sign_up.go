package auth_service

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type SignUpRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required"`
	FirstName       string `json:"firstName" validate:"required"`
	LastName        string `json:"lastName" validate:"required"`
	Language        string `json:"-" validate:"required,len=2"`
	IPAddress       string `json:"ipAddress" validate:"required,ip"`
	UserAgent       string `json:"userAgent" validate:"required"`
	TermsVersion    string `json:"termsVersion" validate:"required,semver"`
	InvitationToken string `json:"invitationToken"`
}

type SignUpResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (s *service) SignUp(req *SignUpRequest) (*SignUpResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	emailLower := strings.ToLower(req.Email)

	existingUser, err := s.repos.User.GetByEmail(nil, emailLower)
	if err == nil && existingUser != nil {
		return nil, yca_error.NewConflictError(nil, constants.EMAIL_ALREADY_IN_USE_CODE, nil)
	}
	if err != nil {
		if e, ok := yca_error.AsError(err); ok && e.ErrorCode != constants.NOT_FOUND_CODE {
			return nil, e
		}
		if _, ok := yca_error.AsError(err); !ok {
			return nil, err
		}
	}

	now := s.now()

	userID, err := s.generateID()
	if err != nil {
		return nil, err
	}

	hashedPassword, err := s.passwordHashFn(req.Password)
	if err != nil {
		return nil, err
	}
	hashedPasswordPtr := &hashedPassword

	language := strings.ToLower(req.Language)
	if !slices.Contains(constants.SUPPORTED_LANGUAGES, language) {
		language = constants.DEFAULT_LANGUAGE
	}

	tx, err := s.repos.User.BeginTx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.auth.SignUp",
				Message:  "Transaction rollback failed",
				Error:    err,
				Data:     map[string]any{"user_id": userID},
			})
		}
	}()

	user := &models.User{
		ID:              userID,
		CreatedAt:       now,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Language:        language,
		Email:           emailLower,
		Password:        hashedPasswordPtr,
		TermsAcceptedAt: now,
		TermsVersion:    req.TermsVersion,
	}

	if err := s.repos.User.Create(tx, user); err != nil {
		return nil, err
	}

	if req.InvitationToken != "" {
		tokenHash := s.hashToken(req.InvitationToken)
		invitation, err := s.repos.Invitation.GetByTokenHash(tokenHash)
		if err != nil {
			if e, ok := yca_error.AsError(err); ok && e.ErrorCode == constants.NOT_FOUND_CODE {
				return nil, yca_error.NewUnprocessableEntityError(nil, constants.INVALID_INVITATION_TOKEN_CODE, nil)
			}
			return nil, err
		}

		if invitation.RevokedAt != nil {
			return nil, yca_error.NewUnprocessableEntityError(nil, constants.INVITATION_REVOKED_CODE, nil)
		}

		if invitation.AcceptedAt != nil {
			return nil, yca_error.NewUnprocessableEntityError(nil, constants.INVITATION_ALREADY_ACCEPTED_CODE, nil)
		}

		if invitation.ExpiresAt.Before(s.now()) {
			return nil, yca_error.NewUnprocessableEntityError(nil, constants.INVITATION_EXPIRED_CODE, nil)
		}

		if strings.ToLower(invitation.Email) != emailLower {
			return nil, yca_error.NewForbiddenError(nil, constants.INVITATION_EMAIL_MISMATCH_CODE, nil)
		}

		organization, err := s.repos.Organization.GetByID(invitation.OrganizationID.String())
		if err != nil {
			return nil, err
		}

		members, err := s.repos.OrganizationMember.ListByOrganizationID(organization.ID.String())
		if err != nil {
			return nil, err
		}

		if len(*members) >= organization.SubscriptionSeats {
			return nil, yca_error.NewForbiddenError(nil, constants.ORGANIZATION_SEATS_LIMIT_CODE, nil)
		}

		now := time.Now()
		invitation.AcceptedAt = &now
		if err := s.repos.Invitation.Update(tx, invitation); err != nil {
			return nil, err
		}

		membershipID, err := s.generateID()
		if err != nil {
			return nil, err
		}

		if err := s.repos.OrganizationMember.Create(tx, &models.OrganizationMember{
			ID:             membershipID,
			CreatedAt:      now,
			UserID:         user.ID,
			OrganizationID: invitation.OrganizationID,
			RoleID:         invitation.RoleID,
		}); err != nil {
			if e, ok := yca_error.AsError(err); ok && e.ErrorCode == constants.CONFLICT_CODE {
				return nil, yca_error.NewConflictError(nil, constants.USER_ALREADY_MEMBER_CODE, nil)
			}
			return nil, err
		}
	}

	accessToken, err := s.generateAccessToken(user, "", "")
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken()
	if err != nil {
		return nil, err
	}

	refreshTokenID, err := s.generateID()
	if err != nil {
		return nil, err
	}

	if err = s.repos.UserRefreshToken.Create(tx, &models.UserRefreshToken{
		ID:        refreshTokenID,
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(s.refreshTTL) * time.Hour),
		IP:        req.IPAddress,
		UserAgent: req.UserAgent,
		TokenHash: s.hashToken(refreshToken),
	}); err != nil {
		return nil, err
	}

	verificationToken, err := s.generateToken()
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.auth.SignUp",
			Message:  "Failed to generate email verification token",
			Data:     map[string]any{"user_id": user.ID},
		})
		return nil, err
	}

	verificationTokenID, err := s.generateID()
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.auth.SignUp",
			Message:  "Failed to generate email verification token ID",
			Data:     map[string]any{"user_id": user.ID},
		})
		return nil, err
	}

	if err := s.repos.UserEmailVerificationToken.Create(tx, &models.UserEmailVerificationToken{
		ID:        verificationTokenID,
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(s.emailVerificationTTL) * time.Hour),
		TokenHash: s.hashToken(verificationToken),
	}); err != nil {
		return nil, err
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
			Location: "services.auth.SignUp",
			Message:  "Failed to prepare email verification email body",
			Data:     map[string]any{"user_id": user.ID},
		})
		return nil, yca_error.NewInternalServerError(err, "", nil)
	}

	subject := s.translator.Translate(language, "email.verification.subject", nil)
	if err := s.emailService.SendEmail(user.Email, subject, emailBody); err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.auth.SignUp",
			Message:  "Failed to send email verification email",
			Data:     map[string]any{"user_id": user.ID},
		})
		return nil, yca_error.NewInternalServerError(err, "", nil)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &SignUpResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
