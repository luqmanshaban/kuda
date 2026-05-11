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

func (r *QueueRepo) GetQueues(user_id int) ([]structs.Queue, error) {
	var qs []structs.Queue

	rows, err := r.DB.Query("SELECT id, name, webhook_url FROM queues WHERE user_id = $1 ", user_id)
	if err != nil {
		return  nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var q structs.Queue
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

func (r *QueueRepo) UpdateQueue(url string, id, userId int) (structs.Queue, error) {
	var q structs.Queue

	err := r.DB.QueryRow("UPDATE queues SET webhook_url = $1 WHERE id = $2 AND user_id = $3 RETURNING id, name, webhook_url", url, id, userId).Scan(
		&q.ID,
		&q.Name,
		&q.WebhookUrl,
	)
	if err != nil {
		return  structs.Queue{}, err
	}


	return q, nil
}
