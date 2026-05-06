CREATE TABLE IF NOT EXISTS jobs(
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    payload JSONB NOT NULL,
    state TEXT NOT NULL default 'pending',
    retries INT NOT NULL default 0,
    max_retries INT NOT NULL default 3,
    runs_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL default NOW(),
    updated_at TIMESTAMP NOT NULL default NOW()
);
