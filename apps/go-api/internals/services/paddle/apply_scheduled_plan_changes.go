package paddle_service

import (
	"context"
	"time"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	yca_log "github.com/yca-software/go-common/logger"
)

// ApplyScheduledPlanChanges finds orgs whose billing period has ended and have a scheduled plan change
// (e.g. annual→monthly), updates their Paddle subscription to the scheduled plan (do_not_bill), and syncs the org.
// Called by the apply_scheduled_plan_changes job consumer. Usually the job is published by cmd/cron at least daily.
func (s *service) ApplyScheduledPlanChanges(ctx context.Context) error {
	orgs, err := s.repos.Organization.GetOrganizationsWithScheduledPlanChangeDue()
	if err != nil {
		return err
	}
	if orgs == nil || len(*orgs) == 0 {
		return nil
	}
	for _, org := range *orgs {
		if org.ScheduledPlanPriceID == nil || *org.ScheduledPlanPriceID == "" ||
			org.PaddleSubscriptionID == nil || *org.PaddleSubscriptionID == "" {
			continue
		}
		planID := *org.ScheduledPlanPriceID
		sub, err := s.paddleClient.UpdateSubscription(ctx, &paddle.UpdateSubscriptionRequest{
			SubscriptionID: *org.PaddleSubscriptionID,
			Items: paddle.NewPatchField([]paddle.UpdateSubscriptionItems{
				*paddle.NewUpdateSubscriptionItemsSubscriptionUpdateItemFromCatalog(&paddle.SubscriptionUpdateItemFromCatalog{
					PriceID:  planID,
					Quantity: 1,
				}),
			}),
			ProrationBillingMode: paddle.NewPatchField(paddle.ProrationBillingModeDoNotBill),
		})
		if err != nil {
			s.logger.Log(yca_log.LogData{
				Level:   "error",
				Message: "ApplyScheduledPlanChanges: Paddle UpdateSubscription failed",
				Error:   err,
				Data: map[string]any{
					"organization_id": org.ID.String(),
					"plan_id":         planID,
				},
			})
			continue
		}
		org.ScheduledPlanPriceID = nil
		s.applySubscriptionFromPrice(&org, planID)
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
		if err := s.repos.Organization.Update(nil, &org); err != nil {
			s.logger.Log(yca_log.LogData{
				Level:   "error",
				Message: "ApplyScheduledPlanChanges: Organization.Update failed",
				Error:   err,
				Data:    map[string]any{"organization_id": org.ID.String()},
			})
			continue
		}
	}
	return nil
}
