package store

import (
	"database/sql"

	"github.com/luqmanshaban/kuda/internal/core"
)

type QueueStore struct {
	DB *sql.DB 
}

func NewQueueStore(db *sql.DB) *QueueStore {
	return &QueueStore{DB: db}
}



func (r *QueueStore) CreateQueue(name, webhook_url string) (core.Queue, error) {
	var q core.Queue
	err := r.DB.QueryRow("INSERT INTO queues (name, webhook_url) VALUES ($1, $2) RETURNING id, name, webhook_url", name, webhook_url).Scan(
		&q.ID,
		&q.Name,
		&q.WebhookUrl,
	)

	if err != nil {
		return  core.Queue{}, err
	}

	return  q, nil
}

func (r *QueueStore) GetQueues() ([]core.Queue, error) {
	var qs []core.Queue

	rows, err := r.DB.Query("SELECT id, name, webhook_url FROM queues",)
	if err != nil {
		return  nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var q core.Queue
		err := rows.Scan(
			&q.ID, 
			&q.Name,
			&q.WebhookUrl,
		)
		if err != nil {
			return  nil, err 
		}
		qs = append(qs, q)
	}

	return qs, nil
}

func (r *QueueStore) UpdateQueue(url string, id int) (core.Queue, error) {
	var q core.Queue

	err := r.DB.QueryRow("UPDATE queues SET webhook_url = $1 WHERE id = $2 RETURNING id, name, webhook_url", url, id).Scan(
		&q.ID,
		&q.Name,
		&q.WebhookUrl,
	)
	if err != nil {
		return  core.Queue{}, err
	}


	return q, nil
}
