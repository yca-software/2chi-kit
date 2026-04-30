package cron

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	yca_logger "github.com/yca-software/go-common/logger"

	"github.com/stretchr/testify/suite"
)

type mockPublisher struct {
	cleanupCalls atomic.Int64
	applyCalls   atomic.Int64
}

func (m *mockPublisher) PublishCleanup(context.Context) error {
	m.cleanupCalls.Add(1)
	return nil
}

func (m *mockPublisher) PublishApplyScheduledPlanChanges(context.Context) error {
	m.applyCalls.Add(1)
	return nil
}

type StartSuite struct {
	suite.Suite
}

func TestStartSuite(t *testing.T) {
	suite.Run(t, new(StartSuite))
}

func (s *StartSuite) TestStart_NilPublisherReturnsNilCancel() {
	cancel, err := Start(nil, yca_logger.New())
	s.Require().NoError(err)
	s.Require().Nil(cancel)
}

func (s *StartSuite) TestStart_InitialPublishes() {
	s.T().Setenv(CRON_CLEANUP_ARCHIVED_INTERVAL_ENV, "50ms")
	s.T().Setenv(CRON_APPLY_SCHEDULED_PLAN_CHANGES_INTERVAL_ENV, "50ms")
	pub := &mockPublisher{}
	cancel, err := Start(pub, yca_logger.New())
	s.Require().NoError(err)
	s.Require().NotNil(cancel)

	s.Equal(int64(1), pub.cleanupCalls.Load())
	s.Equal(int64(1), pub.applyCalls.Load())

	cancel()
}

func (s *StartSuite) TestStart_PeriodicPublishes() {
	s.T().Setenv(CRON_CLEANUP_ARCHIVED_INTERVAL_ENV, "50ms")
	s.T().Setenv(CRON_APPLY_SCHEDULED_PLAN_CHANGES_INTERVAL_ENV, "0")

	pub := &mockPublisher{}
	cancel, err := Start(pub, yca_logger.New())
	s.Require().NoError(err)
	s.Require().NotNil(cancel)

	// Initial publish + at least one periodic tick.
	deadline := time.Now().Add(2200 * time.Millisecond)
	for time.Now().Before(deadline) && pub.cleanupCalls.Load() < 2 {
		time.Sleep(10 * time.Millisecond)
	}

	s.GreaterOrEqual(pub.cleanupCalls.Load(), int64(2))
	s.Equal(int64(0), pub.applyCalls.Load())

	cancel()
}

func (s *StartSuite) TestCronEverySpec_SubSecondDurationAccepted() {
	scheduler := cron.New()
	_, err := scheduler.AddFunc(cronEverySpec(50*time.Millisecond), func() {})
	s.Require().NoError(err)
}
