package models

type PaddleWebhookEvent struct {
	EventID    string         `json:"event_id"`
	EventType  string         `json:"event_type"`
	OccurredAt string         `json:"occurred_at"`
	Data       map[string]any `json:"data"`
}
