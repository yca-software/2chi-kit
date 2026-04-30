package paddle_service

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	yca_error "github.com/yca-software/go-common/error"
)

// webhookNotRelevant returns true when the error indicates the webhook is not for this API
// (e.g. customer/org not in our DB). Caller should return nil so we respond 200 and avoid logs.
func webhookNotRelevant(err error) bool {
	var ycaErr *yca_error.Error
	if errors.As(err, &ycaErr) && ycaErr.StatusCode == http.StatusNotFound {
		return true
	}
	return false
}

// ourPriceIDFromWebhookItems returns the first line item price id that belongs to this app, or "" if none.
func (s *service) ourPriceIDFromWebhookItems(items []any) string {
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		price, ok := item["price"].(map[string]any)
		if !ok {
			continue
		}
		priceID, _ := price["id"].(string)
		if s.isOurPriceID(priceID) {
			return priceID
		}
	}
	return ""
}

func (s *service) HandleWebhook(payload []byte, signature string) error {
	var webhookEvent models.PaddleWebhookEvent
	if err := json.Unmarshal(payload, &webhookEvent); err != nil {
		return yca_error.NewUnprocessableEntityError(err, "", nil)
	}

	switch webhookEvent.EventType {
	case "subscription.created", "subscription.trialing":
		return s.handleSubscriptionUpdate(webhookEvent.Data, true)
	case "subscription.updated", "subscription.activated":
		return s.handleSubscriptionUpdate(webhookEvent.Data, false)
	case "subscription.canceled":
		return s.handleSubscriptionCanceled(webhookEvent.Data)
	case "transaction.completed":
		return s.handleTransactionCompleted(webhookEvent.Data)
	default:
		return nil
	}
}

// handleSubscriptionUpdate updates org from subscription webhook only when the subscription
// is for this product's pricing. applyPlan: true for subscription.created / subscription.trialing.
func (s *service) handleSubscriptionUpdate(data map[string]any, applyPlan bool) error {
	subscriptionData, ok := data["subscription"].(map[string]interface{})
	if !ok {
		subscriptionData = data
	}

	subscriptionID, ok := subscriptionData["id"].(string)
	if !ok {
		return yca_error.NewUnprocessableEntityError(errors.New("missing subscription id in webhook data"), constants.MISSING_PADDLE_SUBSCRIPTION_ID_CODE, nil)
	}

	customerID, ok := subscriptionData["customer_id"].(string)
	if !ok {
		return yca_error.NewUnprocessableEntityError(errors.New("missing customer id in webhook data"), constants.MISSING_PADDLE_CUSTOMER_ID_CODE, nil)
	}

	var items []any
	if rawItems, ok := subscriptionData["items"].([]any); ok {
		items = rawItems
	}
	ourPriceID := s.ourPriceIDFromWebhookItems(items)
	if ourPriceID == "" {
		return nil // no line item uses this app's prices; do not update org, no log
	}

	organization, err := s.repos.Organization.GetByPaddleCustomerID(customerID)
	if err != nil {
		if webhookNotRelevant(err) {
			return nil // customer not in our DB (e.g. other product); respond 200, no log
		}
		return err
	}

	if !applyPlan && organization.PaddleSubscriptionID != nil && *organization.PaddleSubscriptionID != subscriptionID {
		return nil // update is for a different subscription on the same Paddle customer
	}

	organization.PaddleSubscriptionID = &subscriptionID
	if status, _ := subscriptionData["status"].(string); status == "trialing" {
		organization.SubscriptionInTrial = true
	} else {
		organization.SubscriptionInTrial = false
	}
	if applyPlan {
		organization.SubscriptionType = s.getTierFromPriceID(ourPriceID)
		organization.SubscriptionSeats = getSeatsByTier(organization.SubscriptionType)
		organization.SubscriptionPaymentInterval = s.getIntervalFromPriceID(ourPriceID)
	}

	if currentBillingPeriod, ok := subscriptionData["current_billing_period"].(map[string]any); ok {
		if endsAt, ok := currentBillingPeriod["ends_at"].(string); ok {
			if t, err := time.Parse(time.RFC3339, endsAt); err == nil {
				organization.SubscriptionExpiresAt = &t
			}
		}
	} else if nextBilledAt, ok := subscriptionData["next_billed_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, nextBilledAt); err == nil {
			organization.SubscriptionExpiresAt = &t
		}
	}

	if err := s.repos.Organization.Update(nil, organization); err != nil {
		return err
	}

	return nil
}

func (s *service) handleSubscriptionCanceled(data map[string]any) error {
	subscriptionData, ok := data["subscription"].(map[string]any)
	if !ok {
		subscriptionData = data
	}

	customerID, ok := subscriptionData["customer_id"].(string)
	if !ok {
		return yca_error.NewUnprocessableEntityError(errors.New("missing customer id in webhook data"), constants.MISSING_PADDLE_CUSTOMER_ID_CODE, nil)
	}

	organization, err := s.repos.Organization.GetByPaddleCustomerID(customerID)
	if err != nil {
		if webhookNotRelevant(err) {
			return nil // customer not in our DB (other product's webhook); respond 200, no log
		}
		return err
	}

	// Only clear subscription when it was our org's subscription; the customer may have other products' subscriptions.
	webhookSubID, _ := subscriptionData["id"].(string)
	if organization.PaddleSubscriptionID == nil || *organization.PaddleSubscriptionID != webhookSubID {
		return nil
	}

	organization.SubscriptionType = constants.SUBSCRIPTION_TYPE_FREE
	organization.SubscriptionSeats = constants.SUBSCRIPTION_TYPE_SEATS_INCLUDED_FREE
	organization.SubscriptionPaymentInterval = constants.PAYMENT_INTERVAL_MONTHLY
	organization.SubscriptionInTrial = false
	organization.PaddleSubscriptionID = nil

	if canceledAt, ok := subscriptionData["canceled_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, canceledAt); err == nil {
			organization.SubscriptionExpiresAt = &t
		}
	} else if currentBillingPeriod, ok := subscriptionData["current_billing_period"].(map[string]any); ok {
		if endsAt, ok := currentBillingPeriod["ends_at"].(string); ok {
			if t, err := time.Parse(time.RFC3339, endsAt); err == nil {
				organization.SubscriptionExpiresAt = &t
			}
		}
	}

	if err := s.repos.Organization.Update(nil, organization); err != nil {
		return err
	}

	return nil
}

func (s *service) handleTransactionCompleted(data map[string]any) error {
	transactionData, ok := data["transaction"].(map[string]any)
	if !ok {
		transactionData = data
	}

	customerID, _ := transactionData["customer_id"].(string)
	if customerID == "" {
		return yca_error.NewUnprocessableEntityError(errors.New("missing customer id in transaction webhook data"), constants.MISSING_PADDLE_CUSTOMER_ID_CODE, nil)
	}

	var items []any
	if rawItems, ok := transactionData["items"].([]any); ok {
		items = rawItems
	}
	ourPriceID := s.ourPriceIDFromWebhookItems(items)
	if ourPriceID == "" {
		return nil // transaction has no line item for this app's prices; do not update org, no log
	}

	organization, err := s.repos.Organization.GetByPaddleCustomerID(customerID)
	if err != nil {
		if webhookNotRelevant(err) {
			return nil // customer not in our DB (e.g. other product); respond 200, no log
		}
		return err
	}

	if subscriptionID, ok := transactionData["subscription_id"].(string); ok && subscriptionID != "" {
		organization.PaddleSubscriptionID = &subscriptionID
	}
	organization.SubscriptionInTrial = false

	organization.SubscriptionType = s.getTierFromPriceID(ourPriceID)
	organization.SubscriptionSeats = getSeatsByTier(organization.SubscriptionType)
	organization.SubscriptionPaymentInterval = s.getIntervalFromPriceID(ourPriceID)

	if billingPeriod, ok := transactionData["billing_period"].(map[string]any); ok {
		if endsAt, ok := billingPeriod["ends_at"].(string); ok {
			if t, err := time.Parse(time.RFC3339, endsAt); err == nil {
				organization.SubscriptionExpiresAt = &t
			}
		}
	}

	if err := s.repos.Organization.Update(nil, organization); err != nil {
		return err
	}

	return nil
}
