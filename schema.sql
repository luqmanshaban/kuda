CREATE TABLE IF NOT EXISTS api_keys(
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    key TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL default NOW()
);

CREATE TABLE IF NOT EXISTS queues(
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    webhook_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL default NOW()
);

CREATE TABLE IF NOT EXISTS jobs(
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    queue_name TEXT NOT NULL REFERENCES queues(name),
    batch_id TEXT,
    payload JSONB NOT NULL,
    state TEXT NOT NULL default 'pending',
    retries INT NOT NULL default 0,
    max_retries INT NOT NULL default 3,
    runs_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL default NOW(),
    updated_at TIMESTAMPTZ NOT NULL default NOW()
);

CREATE TABLE IF NOT EXISTS dead_letter_jobs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    queue_name TEXT NOT NULL,
    batch_id TEXT,
    payload JSONB NOT NULL,
    error_reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);