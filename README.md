# Kuda

Kuda is a self-hosted job queue for teams that already run Postgres and don't want to operate another piece of infrastructure.

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

## Project Structure

```text
kuda/
├── cmd/
│   └── server/
│       └── main.go           # Entry point: Orchestrates infrastructure & lifecycle
│
├── internal/                 # Private server code (Protected by Go internal/ rule)
│   ├── api/                  # HTTP Layer
│   │   ├── server.go         # Route registration & HTTP server setup
│   │   ├── handlers/         # Request handlers for Jobs and Queues
│   │   └── middleware/       # Auth (API Key verification) & Logging
│   │
│   ├── core/                 # Domain Layer: Business logic & shared structs
│   │   ├── job.go            # Job entity & State constants (Pending/Running/Dead)
│   │   └── queue.go          # Queue entity definitions
│   │
│   ├── store/                # Data Access Layer: Database-specific logic
│   │   ├── postgres.go       # Connection pool & health checks
│   │   ├── jobs.go           # SQL implementation for Job operations
│   │   └── queues.go         # SQL implementation for Queue management
│   │
│   ├── worker/               # Background Processing Layer
│   │   ├── pool.go           # Concurrency management (Worker Pool)
│   │   ├── worker.go         # Job delivery logic (Webhook POST execution)
│   │   └── producer.go       # DB polling loop with backpressure logic
│   │
│   └── config/               # Configuration: Type-safe environment loading
│
├── sdk/                      # Public Go Client: (In Development)
│
├── Dockerfile                # Multi-stage Alpine-based build
├── docker-compose.yml        # Orchestration for App, Postgres, & Prometheus
├── schema.sql                # PostgreSQL table definitions & indexing
└── prometheus.yml            # Metrics scraping configuration

```

## API Reference

### Authentication

All endpoints require an API key in the Authorization header:

```
Authorization: kuda_your_api_key_here
```

## 1. Queues

### Create a queue

```
POST /queues
{ "name": "emails", "webhook_url": "https://yourapp.com/webhook" }

→ { "id": 42, "name": "emails, "webhook_url": "https://yourapp.com/webhook", "created_at:"0001-01-01T00:00:00Z"}
```

### Get a queue

```
GET /queues/emails

→ { "id": 42, "name": "emails, "webhook_url": "https://yourapp.com/webhook","created_at:"0001-01-01T00:00:00Z"}
```

### Get all queues

```
GET /queues

→ [
{ "id": 42, "name": "emails, "webhook_url": "https://yourapp.com/webhook","created_at:"0001-01-01T00:00:00Z"},
{ "id": 43, "name": "sms, "webhook_url": "https://yourapp.com/sms","created_at:"0001-01-01T00:00:00Z"}
]
```

## 2. Jobs

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

_*NOTE*: make sure to convert your time to UTC_

### Check job status

```
GET /jobs/:id
→ { "id": 42, "state": "completed", ... }
```

### Check jobs filtered by status

```
GET /jobs?status=completed
→ [{ "id": 42, "state": "completed", ... }, { "id": 43, "state": "completed", ... }]
```

### Check job status in batch

```
GET /jobs/batch/{batch_id}
→ [{ "id": 42, "state": "completed", ... }, { "id": 43, "state": "pending", ... }]
```
### Dead Letter Queue (DLQ)

```
GET /dead-letter-jobs
→ [
{
   "id": 1,
   "payload": {
     "task_id": 42624,
     ...
   },
   "queue_name": "ghost-queue-that-does-not-exist",
   "batch_id": "ae0d406c98ddc3...",
   "error_reason": "unregistered_queue_name",
   "created_at": "2026-05-16T19:02:36.568589+03:00"
 },
 ...
]
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

| Variable     | Description       |
| ------------ | ----------------- |
| DB_HOST      | Postgres host     |
| DB_NAME      | Database name     |
| DB_USER      | Database user     |
| DB_PASSWORD  | Database password |
| PORT         | Server port       |
| KUDA_API_KEY | Authorization     |

_**NOTE**: If `KUDA_API_KEY` is not found in your environment variables, one will be generated on startup and logged to the console. Copy this key and save it to your .env file for future use._


## --- _coming soon_ ---
- [ ] UI Dashboard for displaying quick metrics
- [ ] webhook signature verification
- [ ] proper logging
- [ ] cron / recurring schedules
- [ ] Advanced job filtering
- [ ] Queue-level stats
- [ ] Testing
- [ ] Idempotency
- [ ] Detailed Documentation