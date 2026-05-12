# Kuda

A lightweight, language-agnostic job queue service built in Go.
Kuda lets any application schedule and deliver background jobs
reliably over HTTP webhooks, without running any additional
infrastructure beyond a Postgres database.

## Why Kuda

Background jobs are simple. Making them reliable isn't.
Kuda handles delivery guarantees, automatic retries with
exponential backoff, dead-letter queuing, and real-time
observability, so your application doesn't have to.

## Performance

Tested on a single Docker container (local machine):

| Test                      | Jobs  | Time  | Req/sec | p99   |
| ------------------------- | ----- | ----- | ------- | ----- |
| Single job submissions    | 5,000 | 3.9s  | 1,280   | 408ms |
| Batch submissions (5/req) | 5,000 | 0.79s | 1,272   | 102ms |

Zero failures across both tests.

## How It Works

1. Your app registers a queue with a webhook URL
2. Your app submits jobs (_immediately_ or _scheduled_)
3. Kuda delivers each job to your webhook via HTTP POST
4. On failure, Kuda retries with exponential backoff
5. After max retries, jobs move to dead-letter state

```
Your App → POST /jobs → Kuda → POST your-webhook.com/handler
                                    ↓ on failure
                               retry with backoff
                                    ↓ max retries exceeded
                               dead letter
```

## Quick Start

```bash
git clone https://github.com/luqmanshaban/kuda
cd kuda
docker compose up --build
```

## API Reference

### Authentication

All endpoints require an API key in the Authorization header:

```
Authorization: kuda_your_api_key_here
```

### Create a user

```
POST /users
{ "email": "you@example.com" }

→ { "api_key": "kuda_..." }
```

### Create a queue

```
POST /queues
{ "name": "emails", "webhook_url": "https://yourapp.com/webhook" }
```

### Submit a job

```
POST /jobs
{ "queue_name": "emails", "payload": { "any": "data" } }

→ { "job_id": 42 }
```

### Submit jobs in batch

```
POST /jobs
[
  { "queue_name": "emails", "payload": { "to": "a@example.com" } },
  { "queue_name": "emails", "payload": { "to": "b@example.com" } }
]

→ { "batch_id": "uuid", "count": 2 }
```

### Schedule a job

```
POST /jobs
{
  "queue_name": "emails",
  "payload": { "to": "a@example.com" },
  "runs_at": "2026-06-01T09:00:00Z"
}
```

### Check job status

```
GET /jobs/:id
→ { "id": 42, "state": "completed", ... }
```

### Health check

```
GET /health
→ { "status": "healthy" }
```

## Architecture

- **HTTP API** — job submission and status queries
- **Postgres** — durable job store, source of truth
- **Worker pool** — 100 concurrent goroutines
- **Producer** — polls DB every 500ms using SELECT FOR UPDATE SKIP LOCKED
- **Webhook delivery** — HTTP POST with 10s timeout per job
- **Retry** — exponential backoff (10s → 20s → 40s) with jitter
- **Graceful shutdown** — in-flight jobs complete before process exits

## Key Technical Decisions

**Postgres over Redis** — durability is the core value proposition.
Jobs survive crashes, restarts, and deployments.

**SELECT FOR UPDATE SKIP LOCKED** — atomic job claiming prevents
duplicate processing across concurrent workers without any
application-level locking.

**TIMESTAMPTZ over TIMESTAMP** — timezone-aware storage ensures
correct scheduling across environments.

## Environment Variables

| Variable | Description       |
| -------- | ----------------- |
| DB_HOST  | Postgres host     |
| DB_NAME  | Database name     |
| DB_USER  | Database user     |
| DB_PASS  | Database password |


