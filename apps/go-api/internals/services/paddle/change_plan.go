package paddle_service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	"github.com/yca-software/2chi-kit/go-api/internals/constants"
	"github.com/yca-software/2chi-kit/go-api/internals/models"
	audit_log_service "github.com/yca-software/2chi-kit/go-api/internals/services/audit_log"
	yca_error "github.com/yca-software/go-common/error"
	yca_log "github.com/yca-software/go-common/logger"
)

const (
	EffectiveAtImmediately       = "immediately"
	EffectiveAtNextBillingPeriod = "next_billing_period"
)

// ChangePlanRequest is used to change an existing subscription to a different plan (upgrade/downgrade).
type ChangePlanRequest struct {
	OrganizationID string `json:"organizationId" validate:"required,uuid"`
	PlanID         string `json:"planId" validate:"required"` // Paddle price ID for the new plan
}

// ChangePlanResult is returned by ChangePlan. EffectiveAt indicates when the new plan takes effect.
type ChangePlanResult struct {
	Organization *models.Organization `json:"organization"`
	EffectiveAt  string                 `json:"effectiveAt"` // "immediately" or "next_billing_period"
}

// ChangePlan updates the organization's Paddle subscription to the new plan. The organization must
// already have an active subscription (paddle_subscription_id). Upgrades and monthly→annual apply
// immediately; tier downgrades and annual→monthly are scheduled for end of billing period and applied by cron.
func (s *service) ChangePlan(ctx context.Context, req *ChangePlanRequest, accessInfo *models.AccessInfo) (*ChangePlanResult, error) {
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

	if org.PaddleSubscriptionID == nil || *org.PaddleSubscriptionID == "" {
		return nil, yca_error.NewUnprocessableEntityError(
			errors.New("organization has no Paddle subscription to change"),
			"",
			nil,
		)
	}

	currentTier := org.SubscriptionType
	currentInterval := org.SubscriptionPaymentInterval
	targetTier := s.getTierFromPriceID(req.PlanID)
	targetInterval := s.getIntervalFromPriceID(req.PlanID)

	// Same plan and interval: no-op.
	if targetTier == currentTier && targetInterval == currentInterval {
		return &ChangePlanResult{Organization: org, EffectiveAt: EffectiveAtImmediately}, nil
	}

	// Schedule for end of period: tier downgrade (e.g. Pro→Basic) or same-tier annual→monthly. Do not call Paddle yet.
	tierDowngrade := targetTier < currentTier
	sameTierAnnualToMonthly := targetTier == currentTier &&
		targetInterval == constants.PAYMENT_INTERVAL_MONTHLY && currentInterval == constants.PAYMENT_INTERVAL_ANNUAL
	if tierDowngrade || sameTierAnnualToMonthly {
		org.ScheduledPlanPriceID = &req.PlanID
		if err := s.repos.Organization.Update(nil, org); err != nil {
			return nil, err
		}
		s.auditLogSubscriptionChange(org, accessInfo, currentTier, currentInterval, targetTier, targetInterval, EffectiveAtNextBillingPeriod)
		return &ChangePlanResult{Organization: org, EffectiveAt: EffectiveAtNextBillingPeriod}, nil
	}

	// Upgrades (tier or monthly→annual): apply immediately in Paddle and DB.
	var prorationMode paddle.ProrationBillingMode
	if currentSub, getErr := s.paddleClient.GetSubscription(ctx, &paddle.GetSubscriptionRequest{
		SubscriptionID: *org.PaddleSubscriptionID,
	}); getErr == nil && currentSub != nil && currentSub.Status == paddle.SubscriptionStatusTrialing {
		prorationMode = paddle.ProrationBillingModeDoNotBill
	} else {
		prorationMode = paddle.ProrationBillingModeProratedImmediately
	}

	org.ScheduledPlanPriceID = nil
	sub, err := s.paddleClient.UpdateSubscription(ctx, &paddle.UpdateSubscriptionRequest{
		SubscriptionID: *org.PaddleSubscriptionID,
		Items: paddle.NewPatchField([]paddle.UpdateSubscriptionItems{
			*paddle.NewUpdateSubscriptionItemsSubscriptionUpdateItemFromCatalog(&paddle.SubscriptionUpdateItemFromCatalog{
				PriceID:  req.PlanID,
				Quantity: 1,
			}),
		}),
		ProrationBillingMode: paddle.NewPatchField(prorationMode),
	})
	if err != nil {
		return nil, yca_error.NewUnprocessableEntityError(err, constants.SUBSCRIPTION_CHANGE_FAILED_CODE, nil)
	}

	s.applySubscriptionFromPrice(org, req.PlanID)
	if sub != nil {
		if sub.CurrentBillingPeriod != nil && sub.CurrentBillingPeriod.EndsAt != "" {
			if t, err := time.Parse(time.RFC3339, sub.CurrentBillingPeriod.EndsAt); err == nil {
				org.SubscriptionExpiresAt = &t
			}
		}
		if org.SubscriptionExpiresAt == nil && sub.NextBilledAt != nil && *sub.NextBilledAt != "" {
			if t, err := time.Parse(time.RFC3339, *sub.NextBilledAt); err == nil {
				org.SubscriptionExpiresAt = &t
			}
		}
	}
	if err := s.repos.Organization.Update(nil, org); err != nil {
		return nil, err
	}
	s.auditLogSubscriptionChange(org, accessInfo, currentTier, currentInterval, targetTier, targetInterval, EffectiveAtImmediately)
	return &ChangePlanResult{Organization: org, EffectiveAt: EffectiveAtImmediately}, nil
}

func (s *service) auditLogSubscriptionChange(org *models.Organization, accessInfo *models.AccessInfo, fromTier, fromInterval, toTier, toInterval int, effectiveAt string) {
	data, _ := json.Marshal(map[string]any{
		"fromTier":     fromTier,
		"fromInterval": fromInterval,
		"toTier":       toTier,
		"toInterval":   toInterval,
		"effectiveAt":  effectiveAt,
	})
	dataRaw := json.RawMessage(data)
	orgID := org.ID.String()
	nameCopy := org.Name
	if _, err := s.auditLogService.Create(&audit_log_service.CreateRequest{
		OrganizationID: orgID,
		Action:         constants.AUDIT_ACTION_TYPE_UPDATE,
		ResourceType:   constants.RESOURCE_TYPE_SUBSCRIPTION,
		ResourceID:     orgID,
		ResourceName:   &nameCopy,
		Data:           &dataRaw,
	}, accessInfo); err != nil {
		s.logger.Log(yca_log.LogData{
			Level: "error", Message: "Failed to create subscription change audit log", Error: err,
			Data: map[string]any{"organization_id": org.ID.String()},
		})
	}
}
