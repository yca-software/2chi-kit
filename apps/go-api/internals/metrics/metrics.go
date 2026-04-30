package metrics

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yca-software/2chi-kit/go-api/internals/jobs"
	"github.com/yca-software/go-common/observer"
)

const (
	NAMESPACE = "2chi-kit"
)

type Metrics struct {
	Base               *observer.BaseMetrics
	RateLimitHitsTotal *prometheus.CounterVec

	// Job consumer metrics (optional; use JobConsumerRecorder() to pass to jobs.Client).
	// Each job type has separate time series via the "job" label.
	jobOutcomeTotal    *prometheus.CounterVec
	jobDurationSeconds *prometheus.HistogramVec
	jobPublishedTotal  *prometheus.CounterVec
}

func New(appName string) (*Metrics, error) {
	base, err := observer.NewBaseMetrics(NAMESPACE, appName)
	if err != nil {
		return nil, err
	}

	jobOutcomeTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: NAMESPACE,
		Subsystem: "job_consumer",
		Name:      "outcome_total",
		Help:      "Job processing outcomes per job type (success, dead_letter, retry_republished)",
	}, []string{"job", "outcome"})
	jobDurationSeconds := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: NAMESPACE,
		Subsystem: "job_consumer",
		Name:      "duration_seconds",
		Help:      "Job processing duration in seconds per job type",
		Buckets:   prometheus.ExponentialBuckets(0.01, 2, 12),
	}, []string{"job"})
	jobPublishedTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: NAMESPACE,
		Subsystem: "job_consumer",
		Name:      "published_total",
		Help:      "Total number of job messages published per job type",
	}, []string{"job"})

	if err := prometheus.DefaultRegisterer.Register(jobOutcomeTotal); err != nil {
		return nil, err
	}
	if err := prometheus.DefaultRegisterer.Register(jobDurationSeconds); err != nil {
		return nil, err
	}
	if err := prometheus.DefaultRegisterer.Register(jobPublishedTotal); err != nil {
		return nil, err
	}

	rateLimitHitsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: NAMESPACE,
		Subsystem: "rate_limit",
		Name:      "hits_total",
		Help:      "Total number of rate limit hits",
	}, []string{"route"})
	if err := prometheus.DefaultRegisterer.Register(rateLimitHitsTotal); err != nil {
		return nil, err
	}

	return &Metrics{
		Base:               base,
		RateLimitHitsTotal: rateLimitHitsTotal,
		jobOutcomeTotal:    jobOutcomeTotal,
		jobDurationSeconds: jobDurationSeconds,
		jobPublishedTotal:  jobPublishedTotal,
	}, nil
}

// JobConsumerRecorder returns a recorder for job consumer metrics. Pass to jobs.NewClient so consumed and published jobs are measured.
func (m *Metrics) JobConsumerRecorder() jobs.ConsumerRecorder {
	if m == nil {
		return nil
	}
	return &jobConsumerRecorder{
		outcomes:  m.jobOutcomeTotal,
		duration:  m.jobDurationSeconds,
		published: m.jobPublishedTotal,
	}
}

type jobConsumerRecorder struct {
	outcomes  *prometheus.CounterVec
	duration  *prometheus.HistogramVec
	published *prometheus.CounterVec
}

func (r *jobConsumerRecorder) RecordOutcome(job string, outcome string) {
	if r == nil || r.outcomes == nil {
		return
	}
	r.outcomes.WithLabelValues(job, outcome).Inc()
}

func (r *jobConsumerRecorder) RecordDuration(job string, d time.Duration) {
	if r == nil || r.duration == nil {
		return
	}
	r.duration.WithLabelValues(job).Observe(d.Seconds())
}

func (r *jobConsumerRecorder) RecordPublished(job string) {
	if r == nil || r.published == nil {
		return
	}
	r.published.WithLabelValues(job).Inc()
}

func (m *Metrics) EchoMiddleware(skipper func(echo.Context) bool) echo.MiddlewareFunc {
	return m.Base.EchoMiddleware(skipper)
}

func (m *Metrics) GetQueryMetricsHook() observer.QueryMetricsHook {
	return m.Base.GetQueryMetricsHook()
}

func (m *Metrics) RecordRateLimitHit(route string) {
	if m.RateLimitHitsTotal != nil {
		m.RateLimitHitsTotal.WithLabelValues(route).Inc()
	}
}
