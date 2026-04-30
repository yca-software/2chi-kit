package jobs

import (
	"context"

	"github.com/yca-software/2chi-kit/go-api/internals/services"
	yca_log "github.com/yca-software/go-common/logger"
)

type Consumers struct {
	srvs   *services.Services
	client *Client
	logger yca_log.Logger
}

func NewConsumers(srvs *services.Services, client *Client, logger yca_log.Logger) *Consumers {
	return &Consumers{srvs: srvs, client: client, logger: logger}
}

// Start starts all job consumers. Returns a stop function that stops every consumer.
func (c *Consumers) Start(ctx context.Context) (stop func()) {
	stops := []func(){
		c.client.RunCleanupConsumer(ctx, c.cleanupHandler()),
		c.client.RunApplyScheduledPlanChangesConsumer(ctx, c.applyScheduledPlanChangesHandler(ctx)),
	}
	return func() {
		for _, fn := range stops {
			fn()
		}
	}
}
