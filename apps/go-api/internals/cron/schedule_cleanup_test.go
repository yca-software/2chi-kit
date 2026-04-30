package cron

import (
	"context"
	"testing"

	"github.com/robfig/cron/v3"
	yca_logger "github.com/yca-software/go-common/logger"

	"github.com/stretchr/testify/suite"
)

type ScheduleCleanupSuite struct {
	suite.Suite
}

func TestScheduleCleanupSuite(t *testing.T) {
	suite.Run(t, new(ScheduleCleanupSuite))
}

func (s *ScheduleCleanupSuite) TestScheduleCleanup_DisabledInterval_DoesNothing() {
	s.T().Setenv(CRON_CLEANUP_ARCHIVED_INTERVAL_ENV, "0")

	pub := &mockPublisher{}
	scheduler := cron.New()

	scheduleCleanup(context.Background(), scheduler, pub, yca_logger.New())

	s.Equal(int64(0), pub.cleanupCalls.Load())
	s.Len(scheduler.Entries(), 0)
}

func (s *ScheduleCleanupSuite) TestScheduleCleanup_ImmediateAndPeriodicPublish() {
	s.T().Setenv(CRON_CLEANUP_ARCHIVED_INTERVAL_ENV, "50ms")

	pub := &mockPublisher{}
	scheduler := cron.New()
	scheduleCleanup(context.Background(), scheduler, pub, yca_logger.New())

	// Immediate publish should happen synchronously on scheduling.
	s.Equal(int64(1), pub.cleanupCalls.Load())
	s.Len(scheduler.Entries(), 1)

	// Simulate one scheduled tick deterministically.
	scheduler.Entries()[0].Job.Run()
	s.Equal(int64(2), pub.cleanupCalls.Load())
}
