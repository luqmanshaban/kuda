package main

import (
	"time"
)

func CreateJob(payload any, runs_at time.Time) (int, error) {
	var nj Job

	err := db.QueryRow("INSERT INTO jobs (payload, runs_at) VALUES ($1, $2) RETURNING id", payload, runs_at).Scan(&nj.ID)
	if err != nil {
		return -1, err
	}

	return nj.ID, nil
}

func GetJob(id int) (Job, error) {
	var j Job

	err := db.QueryRow(
		`
			SELECT id, payload, state, retries, max_retries, runs_at, created_at, updated_at
			FROM jobs
			WHERE id = $1`, id).Scan(
		&j.ID,
		&j.Payload,
		&j.State,
		&j.Retries,
		&j.MaxRetries,
		&j.RunsAt,
		&j.CreatedAt,
		&j.UpdatedAt,
	)

	if err != nil {
		return Job{}, err
	}

	return j, nil
}
