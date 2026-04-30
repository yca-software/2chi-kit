package cron

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/yca-software/go-common/logger"
)

// Start starts the in-process cron publisher.
// It returns a cancel func that the caller should invoke on shutdown.
func Start(publisher JobPublisher, appLogger logger.Logger) (context.CancelFunc, error) {
	if publisher == nil {
		appLogger.Log(logger.LogData{
			Level:   "warn",
			Message: "RabbitMQ job publisher is not configured; cron dispatcher disabled",
		})
		return nil, nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	scheduler := cron.New()

	scheduleCleanup(ctx, scheduler, publisher, appLogger)
	scheduleApplyScheduledPlanChanges(ctx, scheduler, publisher, appLogger)

	scheduler.Start()
	go func() {
		<-ctx.Done()
		stopCtx := scheduler.Stop()
		<-stopCtx.Done()
	}()

	return cancel, nil
}
