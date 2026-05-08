package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/luqmanshaban/kuda/structs"
)

type JobRepository struct {
	DB *sql.DB
}

type JobRequest struct {
	Payload   json.RawMessage `json:"payload"`
	RunsAt    time.Time       `json:"runs_at"`
	UserID    int             `json:"user_id"`
	QueueName string          `json:"queue_name"`
}

func (r *JobRepository) CreateJob(jobs []JobRequest) ([]structs.Job, error) {

	var nj []structs.Job

	// build one query with all values
	q := "INSERT INTO jobs (payload, runs_at, user_id, queue_name) VALUES "
	var args []any

	for i, job := range jobs {
		q += fmt.Sprintf("($%d, $%d, $%d, $%d),", i*4+1, i*4+2, i*4+3, i*4+4)
		args = append(args, job.Payload, job.RunsAt, job.UserID, job.QueueName)
	}

	q = strings.TrimSuffix(q, ",")

	q += " RETURNING id, payload, queue_name, user_id, state, retries, max_retries, runs_at, created_at, updated_at"
	

	rows, err := r.DB.Query(q, args...)
	if err != nil {
		return []structs.Job{}, err
	}
	defer rows.Close()


	for rows.Next() {
		var j structs.Job

		rows.Scan(
			&j.ID,
			&j.Payload,
			&j.QueueName,
			&j.UserID,
			&j.State,
			&j.Retries,
			&j.MaxRetries,
			&j.RunsAt,
			&j.CreatedAt,
			&j.UpdatedAt,
		)

		nj = append(nj, j)
	}
	
	return nj, nil
}

func (r *JobRepository) GetJob(id int) (structs.Job, error) {
	var j structs.Job

	err := r.DB.QueryRow(
		`
			SELECT id, payload, queue_name, user_id, state, retries, max_retries, runs_at, created_at, updated_at
			FROM jobs
			WHERE id = $1`, id).Scan(
		&j.ID,
		&j.Payload,
		&j.QueueName,
		&j.UserID,
		&j.State,
		&j.Retries,
		&j.MaxRetries,
		&j.RunsAt,
		&j.CreatedAt,
		&j.UpdatedAt,
	)

	if err != nil {
		return structs.Job{}, err
	}

	return j, nil
}

// check for pending jobs
func (r *JobRepository) FetchPending() ([]structs.Job, error) {
	var job []structs.Job

	rows, err := r.DB.Query("SELECT id, payload, queue_name, user_id, state, retries, max_retries, runs_at, created_at, updated_at FROM jobs WHERE state = 'pending' AND runs_at <= NOW() LIMIT 100 FOR UPDATE SKIP LOCKED")
	if err != nil {
		return []structs.Job{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var j structs.Job
		if err := rows.Scan(
			&j.ID,
			&j.Payload,
			&j.QueueName,
			&j.UserID,
			&j.State,
			&j.Retries,
			&j.MaxRetries,
			&j.RunsAt,
			&j.CreatedAt,
			&j.UpdatedAt,
		); err != nil {
			return []structs.Job{}, err
		}

		job = append(job, j)
	}

	return job, err
}

// update job status
func (r *JobRepository) UpdateJobState(id int, state string) (structs.Job, error) {
	var j structs.Job
	_, err := r.DB.Exec("UPDATE jobs SET state = $1 WHERE id = $2 RETURNING id, payload, queue_name, user_id, state, retries, max_retries, runs_at, created_at, updated_at", state, id)
	if err != nil {
		return structs.Job{}, err
	}

	err = r.DB.QueryRow("SELECT id, payload, queue_name, user_id, state, retries, max_retries, runs_at, created_at, updated_at FROM jobs WHERE id = $1", id).Scan(
		&j.ID,
		&j.Payload,
		&j.QueueName,
		&j.UserID,
		&j.State,
		&j.Retries,
		&j.MaxRetries,
		&j.RunsAt,
		&j.CreatedAt,
		&j.UpdatedAt,
	)

	return j, nil
}
