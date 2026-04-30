package auth_service

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type AuthenticateWithGoogleRequest struct {
	Code            string `json:"code" validate:"required"`
	TermsVersion    string `json:"termsVersion" validate:"required,semver"`
	InvitationToken string `json:"invitationToken"`

	IPAddress string `json:"ipAddress" validate:"required,ip"`
	UserAgent string `json:"userAgent" validate:"required"`
	Language  string `json:"language" validate:"required,len=2"`
}

func (s *service) AuthenticateWithGoogle(req *AuthenticateWithGoogleRequest) (*AuthenticateResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	ctx := context.Background()

	googleUser, err := s.googleService.GetUserInfo(ctx, req.Code)
	if err != nil {
		return nil, yca_error.NewUnauthorizedError(err, constants.INVALID_TOKEN_CODE, nil)
	}

	if !googleUser.VerifiedEmail {
		return nil, yca_error.NewUnauthorizedError(nil, constants.INVALID_TOKEN_CODE, nil)
	}

	emailLower := strings.ToLower(googleUser.Email)
	googleID := googleUser.ID

	// Check if user exists by Google ID
	user, err := s.repos.User.GetByGoogleID(nil, googleID)
	if err != nil {
		if e, ok := err.(*yca_error.Error); ok {
			if e.ErrorCode != constants.NOT_FOUND_CODE {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	now := s.now()

	tx, err := s.repos.User.BeginTx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.auth.AuthenticateWithGoogle",
				Message:  "Transaction rollback failed",
				Error:    err,
				Data:     map[string]any{"user_id": user.ID},
			})
		}
	}()

	if user == nil {
		var resolvedUser *models.User
		existingUser, err := s.repos.User.GetByEmail(nil, emailLower)
		if err != nil {
			if e, ok := err.(*yca_error.Error); ok {
				if e.ErrorCode != constants.NOT_FOUND_CODE {
					return nil, err
				}
			} else {
				return nil, err
			}
		}

		if existingUser != nil {
			// user exists, link Google and update
			existingUser.GoogleID = &googleID
			if googleUser.Picture != "" {
				existingUser.AvatarURL = googleUser.Picture
			}
			if existingUser.EmailVerifiedAt == nil {
				existingUser.EmailVerifiedAt = &now
			}
			if err := s.repos.User.Update(tx, existingUser); err != nil {
				return nil, err
			}
			resolvedUser = existingUser
		} else {
			// user does not exist, create user
			userID, err := s.generateID()
			if err != nil {
				return nil, err
			}

			language := req.Language
			if !slices.Contains(constants.SUPPORTED_LANGUAGES, language) {
				language = constants.DEFAULT_LANGUAGE
			}

			names := strings.Fields(googleUser.Name)
			firstName := googleUser.GivenName
			lastName := googleUser.FamilyName
			if firstName == "" && len(names) > 0 {
				firstName = names[0]
			}
			if lastName == "" && len(names) > 1 {
				lastName = strings.Join(names[1:], " ")
			}
			if lastName == "" {
				lastName = firstName
			}

			avatarURL := ""
			if googleUser.Picture != "" {
				avatarURL = googleUser.Picture
			}

			resolvedUser = &models.User{
				ID:              userID,
				CreatedAt:       now,
				FirstName:       firstName,
				LastName:        lastName,
				Language:        language,
				Email:           emailLower,
				Password:        nil, // OAuth users don't have passwords
				GoogleID:        &googleID,
				AvatarURL:       avatarURL,
				EmailVerifiedAt: &now,
				TermsAcceptedAt: now,
				TermsVersion:    req.TermsVersion,
			}

			if err := s.repos.User.Create(tx, resolvedUser); err != nil {
				return nil, err
			}

			if req.InvitationToken != "" {
				tokenHash := s.hashToken(req.InvitationToken)
				invitation, err := s.repos.Invitation.GetByTokenHash(tokenHash)
				if err != nil {
					if e, ok := err.(*yca_error.Error); ok {
						if e.ErrorCode == constants.NOT_FOUND_CODE {
							return nil, yca_error.NewUnprocessableEntityError(nil, constants.INVALID_INVITATION_TOKEN_CODE, nil)
						}
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
					UserID:         resolvedUser.ID,
					OrganizationID: invitation.OrganizationID,
					RoleID:         invitation.RoleID,
				}); err != nil {
					if e, ok := err.(*yca_error.Error); ok {
						if e.ErrorCode == constants.CONFLICT_CODE {
							return nil, yca_error.NewConflictError(nil, constants.USER_ALREADY_MEMBER_CODE, nil)
						}
					}
					return nil, err
				}
			}
		}
		user = resolvedUser
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

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &AuthenticateResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
