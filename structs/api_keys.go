package structs

import "time"

type APIKEY struct {
	ID int `json:"id"`
	Key string `json:"key"`
	UserID string `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}