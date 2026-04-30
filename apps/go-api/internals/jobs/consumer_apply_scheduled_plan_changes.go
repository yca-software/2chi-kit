package jobs

import (
	"context"
)

func (c *Consumers) applyScheduledPlanChangesHandler(ctx context.Context) func() error {
	return func() error {
		return c.srvs.Paddle.ApplyScheduledPlanChanges(ctx)
	}
}
