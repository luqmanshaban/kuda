package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/lib/pq"
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
    tx, err := r.DB.Begin()
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    rows, err := tx.Query(`
        SELECT 
            j.id, j.payload, j.queue_name, j.user_id,
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

    var jobs []structs.Job
    var ids []int

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

// exponential backoff with jitter 
func (r *JobRepository) RetryJob(id int, retries int) error {

	backoff := time.Duration(10 << retries) * time.Second
	jitter := time.Duration(rand.Intn(5)) * time.Second

	nextRunAt := time.Now().Add(backoff + jitter)

	_,err := r.DB.Exec(`
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
	return  nil
}

func (r *JobRepository) DeadJob(id int) error {
	_, err := r.DB.Exec(`
		UPDATE jobs 
		SET state = 'dead',
		updated_at = NOW()
		WHERE id = $1
		`, id)
	if err != nil {
		return err 
	}
	return  nil
}

func (r *JobRepository) ResetStaleJobs() error {

	_,err := r.DB.Exec(`
		UPDATE jobs
		SET state = 'pending',
		runs_at = NOW(),
		updated_at = NOW()
		WHERE state = 'running'
		`)
    return err
}
