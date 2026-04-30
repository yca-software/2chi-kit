package cron

import (
	"os"
	"strings"
	"time"

	"github.com/yca-software/go-common/logger"
)

const (
	CRON_CLEANUP_ARCHIVED_INTERVAL_ENV             = "CRON_CLEANUP_ARCHIVED_INTERVAL"
	CRON_APPLY_SCHEDULED_PLAN_CHANGES_INTERVAL_ENV = "CRON_APPLY_SCHEDULED_PLAN_CHANGES_INTERVAL"
	defaultCronInterval                            = 24 * time.Hour
)

func parseDurationEnv(key string, defaultValue time.Duration, appLogger logger.Logger) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue
	}

	// Support "0" / "disable" style values.
	if raw == "0" || strings.EqualFold(raw, "disable") || strings.EqualFold(raw, "disabled") {
		return 0
	}

	d, err := time.ParseDuration(raw)
	if err != nil {
		appLogger.Log(logger.LogData{
			Level:   "warn",
			Message: "Invalid cron duration env value; using default",
			Data: map[string]any{
				"key":     key,
				"raw":     raw,
				"default": defaultValue.String(),
			},
			Error: err,
		})
		return defaultValue
	}

	return d
}

func cronEverySpec(d time.Duration) string {
	// Robfig cron supports an "@every <duration>" syntax.
	return "@every " + d.String()
}
