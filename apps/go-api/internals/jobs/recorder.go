package jobs

import "time"

// ConsumerRecorder records job consumer metrics (duration, published count, outcome) per job type.
// Implementations are provided by the metrics package; pass from app when creating the job client.
type ConsumerRecorder interface {
	RecordDuration(job string, d time.Duration)
	RecordPublished(job string)
	RecordOutcome(job string, outcome string)
}
