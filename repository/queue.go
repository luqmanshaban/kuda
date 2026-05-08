package repository

import (
	"database/sql"

	"github.com/luqmanshaban/kuda/structs"
)

type QueueRepo struct {
	DB *sql.DB
}

func (r *QueueRepo) CreateQueue(name string, user_id int, webhook_url string) (structs.Queue, error) {
	var q structs.Queue
	err := r.DB.QueryRow("INSERT INTO queues (name, user_id, webhook_url) VALUES ($1, $2, $3) RETURNING id, name, user_id, webhook_url", name, user_id, webhook_url).Scan(
		&q.ID,
		&q.Name,
		&q.UserID,
		&q.WebhookUrl,
	)

	if err != nil {
		return  structs.Queue{}, err
	}

	return  q, nil
}