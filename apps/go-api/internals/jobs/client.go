package jobs

import (
	"context"
	"fmt"
	"maps"
	"os"
	"strconv"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	yca_log "github.com/yca-software/go-common/logger"
	"golang.org/x/sync/errgroup"
)

const (
	HeaderRetryCount                   = "x-retry-count"
	QUEUE_CLEANUP                      = "cleanup"
	QUEUE_APPLY_SCHEDULED_PLAN_CHANGES = "apply_scheduled_plan_changes"
	defaultInfraMaxRetries             = 3
)

const (
	OutcomeSuccess          = "success"
	OutcomeDeadLetter       = "dead_letter"
	OutcomeRetryRepublished = "retry_republished"
)

type Client struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	log      yca_log.Logger
	mu       sync.Mutex
	recorder ConsumerRecorder

	infraMaxRetries int
}

func NewClient(ctx context.Context, url string, log yca_log.Logger, recorder ConsumerRecorder) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("jobs: RabbitMQ URL is required")
	}
	retries := envInt("JOB_INFRA_MAX_RETRIES", defaultInfraMaxRetries)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("jobs: connect: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("jobs: channel: %w", err)
	}
	log.Log(yca_log.LogData{
		Level:   "info",
		Message: "RabbitMQ job client connected",
		Data:    map[string]any{"infra_max_retries": retries},
	})
	return &Client{
		conn:            conn,
		ch:              ch,
		log:             log,
		recorder:        recorder,
		infraMaxRetries: retries,
	}, nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ch != nil {
		_ = c.ch.Close()
		c.ch = nil
	}
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// Logger returns the client's logger for handlers that run outside Client methods.
func (c *Client) Logger() yca_log.Logger {
	if c == nil {
		return nil
	}
	return c.log
}

func (c *Client) publish(ctx context.Context, queue string, pub amqp.Publishing) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ch == nil {
		return fmt.Errorf("jobs: client closed")
	}
	return c.ch.PublishWithContext(ctx, "", queue, false, false, pub)
}

func (c *Client) publishTrigger(ctx context.Context, queue string) error {
	err := c.publish(ctx, queue, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/octet-stream",
		Body:         []byte("1"),
	})
	if err != nil {
		return err
	}
	if c.recorder != nil {
		c.recorder.RecordPublished(queue)
	}
	return nil
}

func (c *Client) PublishCleanup(ctx context.Context) error {
	return c.publishTrigger(ctx, QUEUE_CLEANUP)
}

func (c *Client) PublishApplyScheduledPlanChanges(ctx context.Context) error {
	return c.publishTrigger(ctx, QUEUE_APPLY_SCHEDULED_PLAN_CHANGES)
}

func (c *Client) ConsumeCleanup(ctx context.Context, handler func() error) error {
	return c.consume(ctx, QUEUE_CLEANUP, "JOB_CLEANUP_CONCURRENCY", 4, handler, "Cleanup job failed")
}

func (c *Client) ConsumeApplyScheduledPlanChanges(ctx context.Context, handler func() error) error {
	return c.consume(ctx, QUEUE_APPLY_SCHEDULED_PLAN_CHANGES, "JOB_APPLY_SCHEDULED_PLAN_CHANGES_CONCURRENCY", 4, handler, "Apply scheduled plan changes job failed")
}

func (c *Client) RunCleanupConsumer(ctx context.Context, handler func() error) (stop func()) {
	return c.startConsumer(ctx, "Cleanup", func(ctx context.Context) error { return c.ConsumeCleanup(ctx, handler) })
}

func (c *Client) RunApplyScheduledPlanChangesConsumer(ctx context.Context, handler func() error) (stop func()) {
	return c.startConsumer(ctx, "Apply scheduled plan changes", func(ctx context.Context) error {
		return c.ConsumeApplyScheduledPlanChanges(ctx, handler)
	})
}

func (c *Client) startConsumer(ctx context.Context, label string, run func(context.Context) error) (stop func()) {
	subCtx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		if err := run(subCtx); err != nil && subCtx.Err() == nil {
			c.log.Log(yca_log.LogData{Level: "error", Message: label + " job consumer stopped", Error: err})
		}
	}()
	return func() {
		cancel()
		time.Sleep(100 * time.Millisecond)
	}
}

func (c *Client) consume(ctx context.Context, queue, workersEnv string, defWorkers int, handler func() error, errLog string) error {
	n := envParallelWorkers(workersEnv, defWorkers)
	if n < 1 {
		n = 1
	}
	g, ctx := errgroup.WithContext(ctx)
	for w := range n {
		w := w
		g.Go(func() error { return c.consumeOneChannel(ctx, queue, handler, errLog, w) })
	}
	return g.Wait()
}

func (c *Client) consumeOneChannel(ctx context.Context, queue string, handler func() error, errLog string, workerID int) error {
	wch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("jobs: worker %d channel: %w", workerID, err)
	}
	defer wch.Close()
	if err := wch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("jobs: worker %d qos: %w", workerID, err)
	}
	deliveries, err := wch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("jobs: worker %d consume: %w", workerID, err)
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("jobs: worker %d deliveries closed", workerID)
			}
			msg := d
			c.ProcessJobDelivery(ctx, queue, &msg, handler, errLog)
		}
	}
}

func (c *Client) ProcessJobDelivery(ctx context.Context, queueName string, d *amqp.Delivery, handler func() error, errorMessage string) {
	start := time.Now()
	retryCount := parseRetryCount(d.Headers)
	err := handler()
	dur := time.Since(start)
	if c.recorder != nil {
		c.recorder.RecordDuration(queueName, dur)
	}

	ack := func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		if e := d.Ack(false); e != nil {
			c.log.Log(yca_log.LogData{Level: "error", Message: "jobs: ack", Error: e})
		}
	}
	nackDLQ := func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		if e := d.Nack(false, false); e != nil {
			c.log.Log(yca_log.LogData{Level: "error", Message: "jobs: nack dead letter", Error: e})
		}
	}

	if err == nil {
		if c.recorder != nil {
			c.recorder.RecordOutcome(queueName, OutcomeSuccess)
		}
		ack()
		return
	}
	if errorMessage != "" {
		c.log.Log(yca_log.LogData{
			Level:   "error",
			Message: errorMessage,
			Error:   err,
			Data: map[string]any{
				"queue":            queueName,
				"retry_count":      retryCount,
				"duration_seconds": dur.Seconds(),
			},
		})
	}

	if dl, _ := classifyJobError(err, retryCount, c.infraMaxRetries); dl {
		if c.recorder != nil {
			c.recorder.RecordOutcome(queueName, OutcomeDeadLetter)
		}
		nackDLQ()
		return
	}
	headers := cloneHeaders(d.Headers)
	headers[HeaderRetryCount] = int32(retryCount + 1)
	pubErr := c.publish(ctx, queueName, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/octet-stream",
		Body:         d.Body,
		Headers:      headers,
	})
	if pubErr != nil {
		c.log.Log(yca_log.LogData{Level: "error", Message: "jobs: republish failed, dead-lettering", Error: pubErr})
		if c.recorder != nil {
			c.recorder.RecordOutcome(queueName, OutcomeDeadLetter)
		}
		nackDLQ()
		return
	}
	if c.recorder != nil {
		c.recorder.RecordOutcome(queueName, OutcomeRetryRepublished)
	}
	ack()
}

func parseRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	v, ok := headers[HeaderRetryCount]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int32:
		return int(n)
	case int64:
		return int(n)
	case uint16:
		return int(n)
	case uint32:
		return int(n)
	case uint64:
		return int(n)
	default:
		return 0
	}
}

func cloneHeaders(h amqp.Table) amqp.Table {
	if len(h) == 0 {
		return amqp.Table{}
	}
	out := make(amqp.Table, len(h)+1)
	maps.Copy(out, h)
	return out
}

func classifyJobError(err error, retryCount, infraMax int) (deadLetter, republish bool) {
	if !IsRetryable(err) {
		return true, false
	}
	if retryCount >= infraMax {
		return true, false
	}
	return false, true
}

func envInt(key string, def int) int {
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func envParallelWorkers(key string, def int) int {
	n := envInt(key, def)
	if n < 1 {
		return def
	}
	return n
}
