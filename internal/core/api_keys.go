package core

import "time"

type APIKey struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Key string `json:"key"`
	CreatedAt time.Time `json:"created_at"`
}