package paddle_service

import (
	"context"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_log "github.com/yca-software/go-common/logger"
)

func (s *service) UpdateCustomer(organization *models.Organization) (*paddle.Customer, error) {
	ctx := context.Background()

	customer, err := s.paddleClient.UpdateCustomer(ctx, &paddle.UpdateCustomerRequest{
		CustomerID: organization.PaddleCustomerID,
		Email:      paddle.NewPatchField(organization.BillingEmail),
		Name:       paddle.NewPatchField(&organization.Name),
		CustomData: paddle.NewPatchField(paddle.CustomData{
			"organization_id": organization.ID.String(),
			"address":         organization.Address,
			"city":            organization.City,
			"zip":             organization.Zip,
			"country":         organization.Country,
			"timezone":        organization.Timezone,
		}),
	})
	if err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "paddle_service.UpdateCustomer",
			Error:    err,
			Message:  "Failed to update paddle customer: " + err.Error(),
			Data: map[string]any{
				"organization_id":    organization.ID.String(),
				"paddle_customer_id": organization.PaddleCustomerID,
				"billing_email":      organization.BillingEmail,
			},
		})
		return nil, err
	}

	return customer, nil
}
