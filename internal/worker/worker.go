package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/luqmanshaban/kuda/internal/core"
	"github.com/luqmanshaban/kuda/internal/store"
	// "github.com/luqmanshaban/kuda/metrics"
)

type Worker struct {
	Store *store.JobStore
}

func (w Worker) Worker(worker int, jch <-chan core.Job, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jch {
		// start := time.Now().UTC()
		err := w.Deliver(j)
		// prometheus data insertion
		// duration := time.Since(start).Seconds()
		// metrics.JobDeliveryDuration.Observe(duration)

		
		if err != nil {
			slog.Error("job delivery failed", "component", "worker", "op", "deliver_job", "job_id", j.ID, "error", err)
			// metrics.JobsFailed.Inc()

			if j.Retries >= j.MaxRetries {
				if err := w.Store.DeadJob(j.ID); err != nil {
					slog.Error("failed to mark job dead", "component", "worker", "op", "dead_job", "job_id", j.ID, "error", err)
				} else {
					slog.Warn("job dead", "component", "worker", "job_id", j.ID, "attempts", j.Retries+1)
				}
			} else {
				if err := w.Store.RetryJob(j.ID, j.Retries); err != nil {
					slog.Error("failed to schedule retry", "component", "worker", "op", "retry_job", "job_id", j.ID, "error", err)
				} else {
					slog.Info("job scheduled for retry", "component", "worker", "job_id", j.ID, "attempt", j.Retries+1, "max", j.MaxRetries)
				}
			}
		} else {
			_, err := w.Store.UpdateJobState(j.ID, "completed")
			if err != nil {
				slog.Error("failed to mark job completed", "component", "worker", "op", "update_state", "job_id", j.ID, "error", err)
			} else {
				// metrics.JobsCompleted.Inc()
				slog.Info("job completed", "component", "worker", "job_id", j.ID, "worker_id", worker)
			}
		}
	}
}

func (w Worker) Deliver(job core.Job) error {
	body, err := json.Marshal(map[string]any{
		"job_id":  job.ID,
		"payload": job.Payload,
	})
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(job.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("webhook returned %d", resp.StatusCode)
}