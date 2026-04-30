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

type CreateCheckoutSessionRequest struct {
	OrganizationID string `json:"organizationId" validate:"required,uuid"`
	PlanID         string `json:"planId" validate:"required"` // Paddle plan ID for the subscription tier
}

type CheckoutSessionResponse struct {
	TransactionID string `json:"transactionId"`
}

func (s *service) CreateCheckoutSession(req CreateCheckoutSessionRequest, accessInfo *models.AccessInfo) (*CheckoutSessionResponse, error) {
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
			Location: "paddle_service.CreateCheckoutSession",
			Error:    errors.New("organization without Paddle customer ID"),
			Message:  "Organization not found",
			Data:     map[string]any{"organization_id": req.OrganizationID},
		})
		return nil, yca_error.NewInternalServerError(errors.New("organization without Paddle customer ID"), "", nil)
	}

	ctx := context.Background()
	transaction, err := s.paddleClient.CreateTransaction(ctx, &paddle.CreateTransactionRequest{
		CustomerID: &organization.PaddleCustomerID,
		Items: []paddle.CreateTransactionItems{
			*paddle.NewCreateTransactionItemsTransactionItemFromCatalog(&paddle.TransactionItemFromCatalog{
				PriceID:  req.PlanID,
				Quantity: 1,
			}),
		},
	})
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "paddle_service.CreateCheckoutSession",
			Error:    err,
			Message:  "Failed to create checkout session",
			Data:     map[string]any{"organization_id": req.OrganizationID},
		})
		return nil, err
	}

	if transaction.ID == "" {
		return nil, yca_error.NewInternalServerError(errors.New("transaction ID not found in response"), "", nil)
	}

	return &CheckoutSessionResponse{
		TransactionID: transaction.ID,
	}, nil
}
