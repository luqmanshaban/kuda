// dead letter job
package core

import (
	"encoding/json"
	"time"
)

type DeadLetterJob struct {
	ID          int             `json:"id"`
	Payload     json.RawMessage `json:"payload"`
	QueueName   string          `json:"queue_name"`
	BatchID     string          `json:"batch_id"`
	ErrorReason string          `json:"error_reason"`
	CreatedAt   time.Time       `json:"created_at"`
}
