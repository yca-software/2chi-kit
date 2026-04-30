package paddle_service

import (
	"context"
	"errors"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

type CreateCustomerPortalSessionRequest struct {
	OrganizationID string `json:"organizationId" validate:"required,uuid"`
}

type CustomerPortalSessionResponse struct {
	PortalURL string `json:"portalUrl"`
}

func (s *service) CreateCustomerPortalSession(req CreateCustomerPortalSessionRequest, accessInfo *models.AccessInfo) (*CustomerPortalSessionResponse, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	if err := s.authorizer.CheckOrganizationPermission(accessInfo, req.OrganizationID, constants.PERMISSION_SUBSCRIPTION_WRITE); err != nil {
		return nil, err
	}

	organization, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if organization.PaddleCustomerID == "" {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "paddle_service.CreateCustomerPortalSession",
			Error:    errors.New("organization not found"),
			Message:  "Organization not found",
			Data:     map[string]any{"organization_id": req.OrganizationID},
		})
		return nil, yca_error.NewInternalServerError(errors.New("organization without Paddle customer ID"), "", nil)
	}

	ctx := context.Background()
	portalSession, err := s.paddleClient.CreateCustomerPortalSession(ctx, &paddle.CreateCustomerPortalSessionRequest{
		CustomerID: organization.PaddleCustomerID,
	})
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "paddle_service.CreateCustomerPortalSession",
			Error:    err,
			Message:  "Failed to create customer portal session",
			Data:     map[string]any{"organization_id": req.OrganizationID},
		})
		return nil, err
	}

	if portalSession.URLs.General.Overview == "" {
		return nil, yca_error.NewInternalServerError(errors.New("portal URL not found in session response"), "", nil)
	}

	return &CustomerPortalSessionResponse{
		PortalURL: portalSession.URLs.General.Overview,
	}, nil
}
