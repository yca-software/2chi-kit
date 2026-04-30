package cron

import (
	"context"
	"sync/atomic"

	"github.com/robfig/cron/v3"
	"github.com/yca-software/go-common/logger"
)

func scheduleCleanup(
	ctx context.Context,
	scheduler *cron.Cron,
	publisher JobPublisher,
	appLogger logger.Logger,
) {
	if publisher == nil {
		return
	}

	interval := parseDurationEnv(CRON_CLEANUP_ARCHIVED_INTERVAL_ENV, defaultCronInterval, appLogger)
	if interval <= 0 {
		return
	}

	// Run once immediately so the schedule is effective after restarts.
	if err := publisher.PublishCleanup(ctx); err != nil {
		appLogger.Log(logger.LogData{
			Level:   "error",
			Message: "cron publish failed",
			Error:   err,
			Data:    map[string]any{"job": "cleanup_archived"},
		})
	}

	var running atomic.Bool
	spec := cronEverySpec(interval)
	if _, err := scheduler.AddFunc(spec, func() {
		if running.Swap(true) {
			return // avoid overlapping publishes
		}
		defer running.Store(false)
		if ctx.Err() != nil {
			return
		}

		if err := publisher.PublishCleanup(ctx); err != nil && ctx.Err() == nil {
			appLogger.Log(logger.LogData{
				Level:   "error",
				Message: "cron publish failed",
				Error:   err,
				Data:    map[string]any{"job": "cleanup_archived"},
			})
		}
	}); err != nil {
		appLogger.Log(logger.LogData{
			Level:   "error",
			Message: "failed to register cleanup cron job",
			Error:   err,
			Data: map[string]any{
				"interval": interval.String(),
			},
		})
	}
}
