package core

import (
	"encoding/json"
	"time"
)

type JobState string

const (
	StatePending   JobState = "pending"
	StateRunning   JobState = "running"
	StateCompleted JobState = "completed"
	StateFailed    JobState = "failed"
	StateDead      JobState = "dead"
)

type Job struct {
	ID         int             `json:"id"`
	Payload    json.RawMessage `json:"payload"`
	QueueName  string          `json:"queue_name"`
	BatchID    string          `json:"batch_id"`
	State      JobState        `json:"state"`
	Retries    int             `json:"retries"`
	MaxRetries int             `json:"max_retries"`
	RunsAt     time.Time       `json:"runs_at"`
	WebhookURL string          `json:"webhook_url"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type JobRequest struct {
	Payload   json.RawMessage `json:"payload"`
	RunsAt    time.Time       `json:"runs_at"`
	QueueName string          `json:"queue_name"`
	BatchID   string          `json:"batch_id"`
}
