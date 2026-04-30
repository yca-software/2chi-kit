package paddle_service

import (
	"context"
	"errors"
	"time"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

// ProcessTransactionRequest is used to update an organization's subscription
// immediately after a successful Paddle checkout. It validates that the
// transaction belongs to the same customer as the organization before
// applying subscription changes derived from the price ID.
type ProcessTransactionRequest struct {
	OrganizationID string `json:"organizationId" validate:"required,uuid"`
	TransactionID  string `json:"transactionId" validate:"required"`
	PriceID        string `json:"priceId" validate:"required"`
}

func (s *service) ProcessTransaction(ctx context.Context, req *ProcessTransactionRequest, accessInfo *models.AccessInfo) (*models.Organization, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, yca_error.NewUnprocessableEntityError(nil, "", &err)
	}

	if err := s.authorizer.CheckOrganizationPermission(accessInfo, req.OrganizationID, constants.PERMISSION_SUBSCRIPTION_WRITE); err != nil {
		return nil, err
	}

	org, err := s.repos.Organization.GetByID(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if org.PaddleCustomerID == "" {
		return nil, yca_error.NewInternalServerError(errors.New("organization does not have a Paddle customer ID"), "", nil)
	}

	tx, err := s.paddleClient.GetTransaction(ctx, &paddle.GetTransactionRequest{
		TransactionID: req.TransactionID,
	})
	if err != nil {
		return nil, err
	}

	if tx.CustomerID == nil || *tx.CustomerID != org.PaddleCustomerID {
		return nil, yca_error.NewForbiddenError(nil, constants.FORBIDDEN_CODE, nil)
	}

	switch tx.Status {
	case paddle.TransactionStatusCompleted,
		paddle.TransactionStatusBilled,
		paddle.TransactionStatus("paid"):
	default:
		return nil, yca_error.NewUnprocessableEntityError(errors.New("transaction is not completed"), "", nil)
	}

	// Transaction must be for this product's pricing; a customer may have other products' transactions.
	var ourPriceID string
	for _, item := range tx.Items {
		if s.isOurPriceID(item.Price.ID) {
			ourPriceID = item.Price.ID
			break
		}
	}
	if ourPriceID == "" {
		return nil, yca_error.NewUnprocessableEntityError(errors.New("transaction is not for this product's pricing"), "", nil)
	}
	if req.PriceID != "" && req.PriceID != ourPriceID {
		return nil, yca_error.NewUnprocessableEntityError(errors.New("transaction price does not match request"), "", nil)
	}

	if tx.SubscriptionID != nil && *tx.SubscriptionID != "" {
		org.PaddleSubscriptionID = tx.SubscriptionID
	}
	if tx.BillingPeriod != nil && tx.BillingPeriod.EndsAt != "" {
		if t, err := time.Parse(time.RFC3339, tx.BillingPeriod.EndsAt); err == nil {
			org.SubscriptionExpiresAt = &t
		}
	}

	s.applySubscriptionFromPrice(org, ourPriceID)

	if err := s.repos.Organization.Update(nil, org); err != nil {
		return nil, err
	}

	return org, nil
}
