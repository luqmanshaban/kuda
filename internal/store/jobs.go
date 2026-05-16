package store

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/luqmanshaban/kuda/internal/core"
)

type JobStore struct {
	DB *sql.DB
}

func NewJobStore(db *sql.DB) *JobStore {
	return &JobStore{DB: db}
}

func (r *JobStore) CreateJobs(jobs []core.JobRequest) (int64, error) {
	if len(jobs) == 0 {
		return 0, nil
	}

	// fetch for existing queues
	rows, err := r.DB.Query("SELECT name FROM queues")
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var validQueues = make(map[string]bool)

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return 0, err
		}
		validQueues[name] = true
	}

	// seperating healthy jobs from dead jobs
	var healthyJobs []core.JobRequest
	var deadJobs []core.JobRequest

	for _, job := range jobs {
		if validQueues[job.QueueName] {
			healthyJobs = append(healthyJobs, job)
		} else {
			deadJobs = append(deadJobs, job)
		}
	}

	var totalInserted int64

	// save the dead letter jobs
	if len(deadJobs) > 0 {
		dq := "INSERT INTO dead_letter_jobs (payload, queue_name, batch_id, error_reason) VALUES "
		var dArgs []any
		for i, job := range deadJobs {
			dq += fmt.Sprintf("($%d, $%d, $%d, $%d),", i*4+1, i*4+2, i*4+3, i*4+4)
			dArgs = append(dArgs, job.Payload, job.QueueName, job.BatchID, "unregistered_queue_name")
		}
		dq = strings.TrimSuffix(dq, ",")

		_, err := r.DB.Exec(dq, dArgs...)
		if err != nil {
			return 0, fmt.Errorf("dlq batch routing failed: %w", err)
		}
		// Note: We still count these as processed by the ingestion layer
		totalInserted += int64(len(deadJobs))
	}

	if len(healthyJobs) > 0 {
		// save healthy jobs
		q := "INSERT INTO jobs (payload, runs_at, queue_name, batch_id) VALUES "
		var args []any

		for i, job := range healthyJobs {
			q += fmt.Sprintf("($%d, $%d, $%d, $%d),", i*4+1, i*4+2, i*4+3, i*4+4)
			args = append(args, job.Payload, job.RunsAt, job.QueueName, job.BatchID)
		}

		q = strings.TrimSuffix(q, ",")

		_, err := r.DB.Exec(q, args...)
		if err != nil {
			return 0, err
		}

		totalInserted += int64(len(healthyJobs))
	}

	return totalInserted, nil

}

func (r *JobStore) CreateSingleJob(job core.JobRequest) (int, error) {
	var j core.Job

	rows, err := r.DB.Query("SELECT name from queues")
	if err != nil {
		return 0, err
	}

	var validQueues = make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return 0, err
		}
	}

	if validQueues[job.QueueName] {
		err := r.DB.QueryRow(
			`
        INSERT INTO jobs (payload, runs_at, queue_name)
        VALUES ($1, $2, $3)
        RETURNING id
		`, job.Payload, job.RunsAt, job.QueueName).Scan(
			&j.ID,
		)

		if err != nil {
			return j.ID, err
		}

		return j.ID, nil
	}

	return 0, errors.New("queue name does not exist in the database")
}

func (r *JobStore) GetJob(id int) (core.Job, error) {
	var j core.Job

	err := r.DB.QueryRow(
		`
			SELECT id, payload, queue_name, state, retries, max_retries, runs_at, created_at, updated_at
			FROM jobs
			WHERE id = $1`, id).Scan(
		&j.ID,
		&j.Payload,
		&j.QueueName,
		&j.State,
		&j.Retries,
		&j.MaxRetries,
		&j.RunsAt,
		&j.CreatedAt,
		&j.UpdatedAt,
	)

	if err != nil {
		return core.Job{}, err
	}

	return j, nil
}

func (r *JobStore) GetJobs() ([]core.Job, error) {
	var j []core.Job

	rows, err := r.DB.Query(
		`
			SELECT id, payload, queue_name, state, retries, max_retries, runs_at, created_at, updated_at
			FROM jobs
			`)

	if err != nil {
		return []core.Job{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var job core.Job
		rows.Scan(
			&job.ID,
			&job.Payload,
			&job.QueueName,
			&job.State,
			&job.Retries,
			&job.MaxRetries,
			&job.RunsAt,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		j = append(j, job)
	}

	return j, nil
}

func (r *JobStore) GetJobsBatchId(batch_id string) ([]core.Job, error) {
	var j []core.Job

	rows, err := r.DB.Query(
		`
			SELECT id, payload, queue_name, state, retries, max_retries, runs_at, created_at, updated_at
			FROM jobs
			WHERE batch_id = $1`, batch_id)

	if err != nil {
		return []core.Job{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var job core.Job
		rows.Scan(
			&job.ID,
			&job.Payload,
			&job.QueueName,
			&job.State,
			&job.Retries,
			&job.MaxRetries,
			&job.RunsAt,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		j = append(j, job)
	}

	return j, nil
}

func (r *JobStore) GetFilteredJobs(filter string) ([]core.Job, error) {
	var j []core.Job

	rows, err := r.DB.Query(
		`
			SELECT id, payload, queue_name, state, retries, max_retries, runs_at, created_at, updated_at
			FROM jobs
			WHERE state = $1
			`, filter)

	if err != nil {
		return []core.Job{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var job core.Job
		rows.Scan(
			&job.ID,
			&job.Payload,
			&job.QueueName,
			&job.State,
			&job.Retries,
			&job.MaxRetries,
			&job.RunsAt,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		j = append(j, job)
	}

	return j, nil
}

// check for pending jobs
func (r *JobStore) FetchPending() ([]core.Job, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
        SELECT
            j.id, j.payload, j.queue_name,
            j.state, j.retries, j.max_retries, j.runs_at,
            j.created_at, j.updated_at,
            q.webhook_url
        FROM jobs j
        JOIN queues q ON q.name = j.queue_name
        WHERE j.state = 'pending' AND j.runs_at <= NOW()
        LIMIT 50
        FOR UPDATE OF j SKIP LOCKED
    `)
	if err != nil {
		return nil, err
	}

	var jobs []core.Job
	var ids []int

	for rows.Next() {
		var j core.Job
		if err := rows.Scan(
			&j.ID,
			&j.Payload,
			&j.QueueName,
			&j.State,
			&j.Retries,
			&j.MaxRetries,
			&j.RunsAt,
			&j.CreatedAt,
			&j.UpdatedAt,
			&j.WebhookURL,
		); err != nil {
			rows.Close()
			return nil, err
		}
		jobs = append(jobs, j)
		ids = append(ids, j.ID)
	}
	rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return jobs, tx.Commit()
	}

	_, err = tx.Exec(`
        UPDATE jobs SET state = 'running', updated_at = NOW()
        WHERE id = ANY($1)
    `, pq.Array(ids))
	if err != nil {
		return nil, err
	}

	return jobs, tx.Commit()
}

// update job status
func (r *JobStore) UpdateJobState(id int, state string) (core.Job, error) {
	var j core.Job
	err := r.DB.QueryRow("UPDATE jobs SET state = $1 WHERE id = $2 RETURNING id, payload, queue_name, state, retries, max_retries, runs_at, created_at, updated_at", state, id).Scan(
		&j.ID,
		&j.Payload,
		&j.QueueName,
		&j.State,
		&j.Retries,
		&j.MaxRetries,
		&j.RunsAt,
		&j.CreatedAt,
		&j.UpdatedAt,
	)
	if err != nil {
		return core.Job{}, err
	}

	return j, nil
}

// exponential backoff with jitter
func (r *JobStore) RetryJob(id int, retries int) error {

	backoff := time.Duration(10<<retries) * time.Second
	jitter := time.Duration(rand.Intn(5)) * time.Second

	nextRunAt := time.Now().UTC().Add(backoff + jitter)

	_, err := r.DB.Exec(`
		UPDATE jobs
		SET state='pending',
		retries = retries + 1,
		runs_at = $1,
		updated_at = NOW()
		where id = $2
		`, nextRunAt, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *JobStore) DeadJob(job core.Job) error {
	_, err := r.DB.Exec(`
		UPDATE jobs
		SET state = 'dead',
		updated_at = NOW()
		WHERE id = $1
		`, job.ID)
	if err != nil {
		return err
	}
	_, err = r.DB.Exec(`
		INSERT INTO dead_letter_jobs (payload, queue_name, batch_id, error_reason) VALUES ($1,$2, $3, $4)
		`,job.Payload,job.QueueName, job.BatchID, "job `retry` exceeded maximum retries" )
	if err != nil {
		return err
	}
	return nil
}

func (r *JobStore) ResetStaleJobs() error {

	_, err := r.DB.Exec(`
		UPDATE jobs
		SET state = 'pending',
		runs_at = NOW(),
		updated_at = NOW()
		WHERE state = 'running'
		`)
	return err
}
