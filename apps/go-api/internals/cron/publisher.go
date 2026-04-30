package cron

import "context"

// JobPublisher is the minimal interface cron needs to enqueue job triggers.
// `*jobs.Client` satisfies this automatically.
type JobPublisher interface {
	PublishCleanup(context.Context) error
	PublishApplyScheduledPlanChanges(context.Context) error
}
