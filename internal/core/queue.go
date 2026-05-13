package core

import "time"

type Queue struct {
	ID int `json:"id"`
	Name string `json:"name"`
	WebhookUrl string `json:"webhook_url"`
	CreatedAt time.Time `json:"created_at"`
}