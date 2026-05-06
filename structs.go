package main

import (
	"encoding/json"
	"time"
)

type JobState string 

const (
	StatePending JobState = "pending"
	StateRunning JobState = "running"
	StateCompleted JobState = "completed"
	StateFailed JobState = "failed"
)


type Job struct { 
   ID int `json:"id"`
   Payload json.RawMessage `json:"payload"`
   State JobState `json:"state"`
   Retries int `json:"retries"`
   MaxRetries int `json:"max_retries"`
   RunsAt time.Time `json:"runs_at"`
   CreatedAt time.Time `json:"created_at"`
   UpdatedAt time.Time `json:"updated_at"`
}