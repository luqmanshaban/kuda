package worker

import (
	"log/slog"
	"sync"

	"github.com/luqmanshaban/kuda/internal/core"
	"github.com/luqmanshaban/kuda/internal/store"
)

type Pool struct {
	store      *store.JobStore
	numWorkers int
}

func NewPool(s *store.JobStore, n int) *Pool {
	return &Pool{store: s, numWorkers: n}
}

func (p *Pool) Start(jobCh <-chan core.Job) *sync.WaitGroup {
	var wg sync.WaitGroup
	w := &Worker{Store: p.store}

	for i := 1; i <= p.numWorkers; i++{
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for job := range jobCh {
				p.process(w, i, job)
			}
		}(i)
	}
	return  &wg
}

func (p *Pool) process(w *Worker, workerId int, j core.Job) {
	err := w.Deliver(j)
	if err != nil {
		slog.Error("job delivery failed", "job_id", j.ID, "error", err)
		if j.Retries >= j.MaxRetries {
			p.store.DeadJob(j.ID)
		} else {
			p.store.RetryJob(j.ID, j.Retries)
		}
		return
	}
	p.store.UpdateJobState(j.ID, "completed")
	slog.Info("job completed", "job_id", j.ID, "worker_id", workerId)
}
