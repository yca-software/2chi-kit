package paddle_service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	paddleerr "github.com/PaddleHQ/paddle-go-sdk/v4/pkg/paddleerr"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

// Paddle customer IDs returned in API errors (e.g. "customer of id ctm_...").
var paddleCustomerIDFromMessageRE = regexp.MustCompile(`\b(ctm_[0-9a-z]+)\b`)

// CreateCustomer creates a Paddle customer for the organization, or returns the existing
// customer when the email is already registered (e.g. same user across multiple SaaS apps
// sharing one Paddle account). We do not archive Paddle customers so emails can be reused.
func (s *service) CreateCustomer(organization *models.Organization) (*paddle.Customer, error) {
	ctx := context.Background()

	customer, err := s.paddleClient.CreateCustomer(ctx, &paddle.CreateCustomerRequest{
		Email: organization.BillingEmail,
		Name:  &organization.Name,
		CustomData: paddle.CustomData{
			"organization_id": organization.ID.String(),
			"address":         organization.Address,
			"city":            organization.City,
			"zip":             organization.Zip,
			"country":         organization.Country,
			"timezone":        organization.Timezone,
		},
	})
	if err == nil {
		return customer, nil
	}

	// Email already exists in Paddle (e.g. customer from another app); reuse existing customer.
	if isCustomerAlreadyExistsError(err) {
		existing, resolveErr := s.resolveCustomerForDuplicateEmail(ctx, organization.BillingEmail, err)
		if resolveErr != nil {
			s.logger.Log(yca_log.LogData{
				Level:    "error",
				Location: "paddle_service.CreateCustomer",
				Error:    resolveErr,
				Message:  "Failed to resolve existing paddle customer: " + resolveErr.Error(),
				Data: map[string]any{
					"organization_id": organization.ID.String(),
					"billing_email":   organization.BillingEmail,
				},
			})
			return nil, yca_error.NewConflictError(err, constants.PADDLE_CUSTOMER_ALREADY_EXISTS_CODE, map[string]any{
				"organization_id": organization.ID.String(),
				"billing_email":   organization.BillingEmail,
			})
		}
		if existing != nil {
			if existing.Status != paddle.StatusActive {
				activated, activateErr := s.activateCustomer(ctx, existing.ID)
				if activateErr != nil {
					s.logger.Log(yca_log.LogData{
						Level:    "error",
						Location: "paddle_service.CreateCustomer",
						Error:    activateErr,
						Message:  "Failed to activate existing paddle customer: " + activateErr.Error(),
						Data: map[string]any{
							"organization_id": organization.ID.String(),
							"billing_email":   organization.BillingEmail,
							"customer_id":     existing.ID,
						},
					})
					return nil, activateErr
				}
				return activated, nil
			}
			return existing, nil
		}
	}

	data := map[string]any{
		"organization_id": organization.ID.String(),
		"billing_email":   organization.BillingEmail,
	}
	s.logger.Log(yca_log.LogData{
		Level:    "error",
		Location: "paddle_service.CreateCustomer",
		Error:    err,
		Message:  "Failed to create paddle customer: " + err.Error(),
		Data:     data,
	})
	return nil, yca_error.NewConflictError(err, constants.PADDLE_CUSTOMER_ALREADY_EXISTS_CODE, data)
}

// getCustomerByEmail returns the first Paddle customer for the given email, or nil if none.
func (s *service) getCustomerByEmail(ctx context.Context, email string) (*paddle.Customer, error) {
	coll, err := s.paddleClient.ListCustomers(ctx, &paddle.ListCustomersRequest{
		Email: []string{email},
	})
	if err != nil {
		return nil, err
	}
	var first *paddle.Customer
	_ = coll.Iter(ctx, func(c *paddle.Customer) (bool, error) {
		first = c
		return false, nil // stop after first
	})
	return first, nil
}

// resolveCustomerForDuplicateEmail loads the Paddle customer for a billing email.
// ListCustomers can return empty even when create failed with customer_already_exists;
// in that case Paddle includes the conflicting customer id in the error detail — fetch via GetCustomer.
func (s *service) resolveCustomerForDuplicateEmail(ctx context.Context, billingEmail string, createErr error) (*paddle.Customer, error) {
	existing, listErr := s.getCustomerByEmail(ctx, billingEmail)
	if listErr != nil {
		return nil, listErr
	}
	if existing != nil {
		return existing, nil
	}
	if id := paddleCustomerIDFromAlreadyExistsError(createErr); id != "" {
		return s.paddleClient.GetCustomer(ctx, &paddle.GetCustomerRequest{CustomerID: id})
	}
	return nil, nil
}

func paddleCustomerIDFromAlreadyExistsError(err error) string {
	if err == nil {
		return ""
	}
	var pe *paddleerr.Error
	if errors.As(err, &pe) && pe != nil {
		if id := extractPaddleCustomerIDFromText(pe.Detail); id != "" {
			return id
		}
	}
	return extractPaddleCustomerIDFromText(err.Error())
}

func extractPaddleCustomerIDFromText(s string) string {
	if s == "" {
		return ""
	}
	m := paddleCustomerIDFromMessageRE.FindStringSubmatch(s)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

// isCustomerAlreadyExistsError handles both SDK sentinel and raw API error shapes.
// Some Paddle responses are not wrapped with paddle.ErrCustomerAlreadyExists.
func isCustomerAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, paddle.ErrCustomerAlreadyExists) {
		return true
	}
	var pe *paddleerr.Error
	if errors.As(err, &pe) && pe != nil && pe.Code == "customer_already_exists" {
		return true
	}
	errText := strings.ToLower(err.Error())
	if strings.Contains(errText, "customer_already_exists") {
		return true
	}
	return strings.Contains(errText, "already exists") &&
		strings.Contains(errText, "customer")
}

func (s *service) activateCustomer(ctx context.Context, customerID string) (*paddle.Customer, error) {
	return s.paddleClient.UpdateCustomer(ctx, &paddle.UpdateCustomerRequest{
		CustomerID: customerID,
		Status:     paddle.NewPatchField(paddle.StatusActive),
	})
}
