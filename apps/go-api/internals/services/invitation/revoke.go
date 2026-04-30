package invitation_service

import (
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

type RevokeRequest struct {
	OrganizationID string `json:"-"`
	InvitationID   string `json:"-" validate:"required,uuid"`
}

func (s *service) Revoke(req *RevokeRequest, accessInfo *models.AccessInfo) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return err
	}

	if err := s.authorizer.CheckOrganizationPermissionWithSubscription(accessInfo, org, constants.PERMISSION_MEMBERS_DELETE); err != nil {
		return err
	}

	invitation, err := s.repos.Invitation.GetByID(org.ID.String(), req.InvitationID)
	if err != nil {
		return err
	}

	if invitation.RevokedAt != nil {
		return yca_error.NewUnprocessableEntityError(nil, constants.INVITATION_REVOKED_CODE, nil)
	}

	if invitation.AcceptedAt != nil {
		return yca_error.NewUnprocessableEntityError(nil, constants.INVITATION_ALREADY_ACCEPTED_CODE, nil)
	}

	if invitation.ExpiresAt.Before(s.now()) {
		return yca_error.NewUnprocessableEntityError(nil, constants.INVITATION_EXPIRED_CODE, nil)
	}

	now := s.now()
	invitation.RevokedAt = &now
	if err := s.repos.Invitation.Update(nil, invitation); err != nil {
		return err
	}

	return nil
}
