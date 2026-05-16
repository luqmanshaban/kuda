// dead letter
package store

import (
	"database/sql"

	"github.com/luqmanshaban/kuda/internal/core"
)

type DeadLetterJobStore struct {
	DB *sql.DB
}

func NewDeadLetter(db *sql.DB) *DeadLetterJobStore {
	return &DeadLetterJobStore{DB: db}
}

func (r *DeadLetterJobStore) CreateDeadJob(job core.Job) (int, error) {
	var dj core.DeadLetterJob

	err := r.DB.QueryRow("INSERT INTO dead_letter_jobs (payload, queue_name) VALUES ($1, $2) RETURNING id", job.Payload, job.QueueName).Scan(
		&dj.ID,
	)
	if err != nil {
		return 0, err
	}
	return job.ID, nil
}

func (r *DeadLetterJobStore) CreateDeadJobWithBatchID(job core.Job) (int, error) {
	var dj core.DeadLetterJob

	err := r.DB.QueryRow("INSERT INTO dead_letter_jobs (payload, queue_name, batch_id) VALUES ($1, $2) RETURNING id", job.Payload, job.QueueName, job.BatchID).Scan(
		&dj.ID,
	)
	if err != nil {
		return 0, err
	}
	return job.ID, nil
}

func (r *DeadLetterJobStore) GetDeadJobs() ([]core.DeadLetterJob, error) {
	var djs []core.DeadLetterJob

	rows, err := r.DB.Query("SELECT id, payload, queue_name, batch_id, error_reason, created_at from dead_letter_jobs")
	if err != nil {
		return []core.DeadLetterJob{}, err
	}

	for rows.Next() {
		var dj core.DeadLetterJob
		if err := rows.Scan(
			&dj.ID,
			&dj.Payload,
			&dj.QueueName,
			&dj.BatchID,
			&dj.ErrorReason,
			&dj.CreatedAt,
		); err != nil {
			return []core.DeadLetterJob{}, err
		}
		djs = append(djs, dj)
	}
	rows.Close()

	return djs, nil
}

func (r *DeadLetterJobStore) GetDeadJob(jobId int) (core.DeadLetterJob, error) {
	var dj core.DeadLetterJob

	err := r.DB.QueryRow("SELECT id, payload, queue_name, batch_id, error_reason, created_at from dead_letter_jobs WHERE id = $1", jobId).Scan(
		&dj.ID,
		&dj.Payload,
		&dj.QueueName,
		&dj.BatchID,
		&dj.ErrorReason,
		&dj.CreatedAt,
	)
	if err != nil {
		return core.DeadLetterJob{}, err
	}

	return dj, nil
}
