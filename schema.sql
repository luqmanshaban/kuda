CREATE TABLE IF NOT EXISTS users(
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    api_key TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS jobs(
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    queue_name TEXT NOT NULL,
    payload JSONB NOT NULL,
    user_id BIGINT REFERENCES users(id),
    state TEXT NOT NULL default 'pending',
    retries INT NOT NULL default 0,
    max_retries INT NOT NULL default 3,
    runs_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL default NOW(),
    updated_at TIMESTAMP NOT NULL default NOW()
);

CREATE TABLE IF NOT EXISTS queues(
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, 
    name TEXT NOT NULL UNIQUE,
    user_id BIGINT REFERENCES users(id),
    webhook_url TEXT NOT NULL
);
