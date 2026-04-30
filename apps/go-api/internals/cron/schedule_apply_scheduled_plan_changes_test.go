package cron

import (
	"context"
	"testing"

	"github.com/robfig/cron/v3"
	yca_logger "github.com/yca-software/go-common/logger"

	"github.com/stretchr/testify/suite"
)

type ScheduleApplyScheduledPlanChangesSuite struct {
	suite.Suite
}

func TestScheduleApplyScheduledPlanChangesSuite(t *testing.T) {
	suite.Run(t, new(ScheduleApplyScheduledPlanChangesSuite))
}

func (s *ScheduleApplyScheduledPlanChangesSuite) TestScheduleApplyScheduledPlanChanges_DisabledInterval_DoesNothing() {
	s.T().Setenv(CRON_APPLY_SCHEDULED_PLAN_CHANGES_INTERVAL_ENV, "0")

	pub := &mockPublisher{}
	scheduler := cron.New()

	scheduleApplyScheduledPlanChanges(context.Background(), scheduler, pub, yca_logger.New())

	s.Equal(int64(0), pub.applyCalls.Load())
	s.Len(scheduler.Entries(), 0)
}

func (s *ScheduleApplyScheduledPlanChangesSuite) TestScheduleApplyScheduledPlanChanges_ImmediateAndPeriodicPublish() {
	s.T().Setenv(CRON_APPLY_SCHEDULED_PLAN_CHANGES_INTERVAL_ENV, "50ms")

	pub := &mockPublisher{}
	scheduler := cron.New()
	scheduleApplyScheduledPlanChanges(context.Background(), scheduler, pub, yca_logger.New())

	// Immediate publish should happen synchronously on scheduling.
	s.Equal(int64(1), pub.applyCalls.Load())
	s.Len(scheduler.Entries(), 1)

	// Simulate one scheduled tick deterministically.
	scheduler.Entries()[0].Job.Run()
	s.Equal(int64(2), pub.applyCalls.Load())
}
