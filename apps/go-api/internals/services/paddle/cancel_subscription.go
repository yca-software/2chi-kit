package paddle_service

import (
	"context"

	"github.com/PaddleHQ/paddle-go-sdk/v4"
	yca_log "github.com/yca-software/go-common/logger"
)

func (s *service) CancelSubscription(paddleSubscriptionID string) error {
	ctx := context.Background()

	if _, err := s.paddleClient.CancelSubscription(ctx, &paddle.CancelSubscriptionRequest{
		SubscriptionID: paddleSubscriptionID,
	}); err != nil {
		s.logger.Log(yca_log.LogData{
			Level:    "error",
			Location: "paddle_service.CancelSubscription",
			Error:    err,
			Message:  "Failed to cancel paddle subscription: " + err.Error(),
			Data: map[string]any{
				"paddle_subscription_id": paddleSubscriptionID,
			},
		})
		return err
	}

	return nil
}
