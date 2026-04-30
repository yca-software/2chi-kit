package invitation_service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type CreateRequest struct {
	Email          string `json:"email" validate:"required,email"`
	OrganizationID string `json:"organizationId" validate:"required,uuid"`
	RoleID         string `json:"roleId" validate:"required,uuid"`
	InvitedByID    string `json:"invitedById" validate:"required,uuid"`
	InvitedByEmail string `json:"invitedByEmail" validate:"required,email"`
	Language       string `json:"language" validate:"required"`
}

type CreateResponse struct {
	Invitation *models.Invitation                 `json:"invitation"`
	Member     *models.OrganizationMemberWithUser `json:"member"`
}

func (s *service) Create(req *CreateRequest, accessInfo *models.AccessInfo) (*CreateResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := s.checkCreateInvitationPermission(accessInfo, org); err != nil {
		return nil, err
	}

	emailLower := strings.ToLower(req.Email)
	existingUser, err := s.repos.User.GetByEmail(nil, emailLower)
	if err != nil {
		e, ok := yca_error.AsError(err)
		if ok && e.ErrorCode != constants.NOT_FOUND_CODE {
			return nil, e
		}
		if !ok {
			return nil, err
		}
	}

	now := s.now()
	tx, err := s.repos.OrganizationMember.BeginTx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			data := map[string]any{"organization_id": org.ID}
			if existingUser != nil {
				data["user_id"] = existingUser.ID
			}
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "services.invitation.Create",
				Message:  "Transaction rollback failed",
				Error:    err,
				Data:     data,
			})
		}
	}()

	if existingUser != nil {
		if members, err := s.repos.OrganizationMember.ListByOrganizationID(req.OrganizationID); err == nil {
			for _, member := range *members {
				if member.UserID == existingUser.ID {
					return nil, yca_error.NewConflictError(nil, constants.USER_ALREADY_MEMBER_CODE, nil)
				}
			}
		}

		memberID, err := s.generateID()
		if err != nil {
			return nil, err
		}

		member := &models.OrganizationMember{
			ID:             memberID,
			CreatedAt:      now,
			OrganizationID: org.ID,
			UserID:         existingUser.ID,
			RoleID:         uuid.MustParse(req.RoleID),
		}

		if err := s.repos.OrganizationMember.Create(tx, member); err != nil {
			return nil, err
		}

		if err := tx.Commit(); err != nil {
			return nil, err
		}

		memberWithUser, err := s.repos.OrganizationMember.GetByIDWithUser(org.ID.String(), member.ID.String())
		if err != nil {
			return nil, err
		}

		return &CreateResponse{
			Member: memberWithUser,
		}, nil
	}

	invitationID, err := s.generateID()
	if err != nil {
		return nil, err
	}

	inviteToken, err := s.generateToken()
	if err != nil {
		return nil, err
	}

	parsedInvitedByID, err := uuid.Parse(req.InvitedByID)
	if err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	invitation := &models.Invitation{
		ID:             invitationID,
		CreatedAt:      now,
		ExpiresAt:      now.Add(time.Duration(s.invitationTTL) * time.Hour),
		OrganizationID: org.ID,
		Email:          emailLower,
		RoleID:         uuid.MustParse(req.RoleID),
		InvitedByID:    uuid.NullUUID{UUID: parsedInvitedByID, Valid: true},
		InvitedByEmail: req.InvitedByEmail,
		TokenHash:      s.hashToken(inviteToken),
	}

	if err := s.repos.Invitation.Create(tx, invitation); err != nil {
		return nil, err
	}

	tokenTTLDays := s.invitationTTL / 24
	emailBody, err := s.emailService.PrepareEmailBody("invitation", map[string]any{
		"Lang":         req.Language,
		"Title":        s.translator.Translate(req.Language, "email.invitation.title", nil),
		"Greeting":     s.translator.Translate(req.Language, "email.invitation.greeting", nil),
		"Content":      s.translator.Translate(req.Language, "email.invitation.content", nil),
		"ButtonText":   s.translator.Translate(req.Language, "email.invitation.button", nil),
		"FooterIgnore": s.translator.Translate(req.Language, "email.invitation.footer.ignore", nil),
		"FooterExpiry": s.translator.Translate(req.Language, "email.invitation.footer.expiry", map[string]any{
			"TokenTTLDays": tokenTTLDays,
		}),
		"C2ALink": fmt.Sprintf("%s/signup?invitationToken=%s", s.appURL, inviteToken),
	})
	if err != nil {
		data := map[string]any{"organization_id": org.ID}
		if existingUser != nil {
			data["user_id"] = existingUser.ID
		}
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.invitation.Create",
			Message:  "Failed to prepare email invitation email body",
			Error:    err,
			Data:     data,
		})
		return nil, err
	}

	subject := s.translator.Translate(req.Language, "email.invitation.subject", map[string]any{
		"OrganizationName": org.Name,
	})

	if err := s.emailService.SendEmail(emailLower, subject, emailBody); err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "services.invitation.Create",
			Message:  "Failed to send email invitation email",
			Error:    err,
			Data:     map[string]any{"organization_id": org.ID, "email": emailLower},
		})
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &CreateResponse{
		Invitation: invitation,
	}, nil
}

func (s *service) checkCreateInvitationPermission(accessInfo *models.AccessInfo, organization *models.Organization) error {
	if err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, organization, constants.PERMISSION_MEMBERS_WRITE); err != nil {
		return err
	}
	// Non-admin: organization must have available seats
	if accessInfo != nil && accessInfo.User != nil && accessInfo.User.IsAdmin == false {
		organizationMembers, err := s.repos.OrganizationMember.ListByOrganizationID(organization.ID.String())
		if err != nil {
			return err
		}
		if len(*organizationMembers) >= organization.SubscriptionSeats {
			return yca_error.NewForbiddenError(nil, constants.ORGANIZATION_SEATS_LIMIT_CODE, nil)
		}
	}

	return nil
}
