package repository

import (
	"database/sql"

	"github.com/luqmanshaban/kuda/structs"
)

type UserRepository struct {
	DB *sql.DB
}

func (r *UserRepository) CreateUser(email string, password string) (structs.User, error)  {
	var nu structs.User 
	err := r.DB.QueryRow("INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id, email", email, password).Scan(&nu.ID, &nu.Email )
	if err != nil {
		return  structs.User{}, err 
	}

	return nu, nil
}

func (r *UserRepository) GetUserJobs(userId int) ([]structs.Job, error) {
	var j []structs.Job

	rows, err := r.DB.Query(
		`
			SELECT id, payload, queue_name, user_id, state, retries, max_retries, runs_at, created_at, updated_at
			FROM jobs
			WHERE user_id = $1
			`, userId)

	if err != nil {
		return []structs.Job{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var job structs.Job 
		rows.Scan(
			&job.ID,
			&job.Payload,
			&job.QueueName,
			&job.UserID,
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

func (r *UserRepository) GetFilteredUserJobs(userId int, filter string) ([]structs.Job, error) {
	var j []structs.Job

	rows, err := r.DB.Query(
		`
			SELECT id, payload, queue_name, user_id, state, retries, max_retries, runs_at, created_at, updated_at
			FROM jobs
			WHERE user_id = $1
			AND state = $2
			`, userId, filter)

	if err != nil {
		return []structs.Job{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var job structs.Job 
		rows.Scan(
			&job.ID,
			&job.Payload,
			&job.QueueName,
			&job.UserID,
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