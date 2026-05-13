package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/luqmanshaban/kuda/internal/core"
	"github.com/luqmanshaban/kuda/internal/store"
)

type Producer struct {
	Store *store.JobStore
	ch    chan<- core.Job
}

func NewProducer(s *store.JobStore, ch chan<- core.Job) *Producer {
	return &Producer{Store: s, ch: ch}
}

func (p *Producer) Start(context context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-context.Done():
			slog.Info("Shutting down")
			return
		case <- ticker.C:
		    // Backpressure (if channel is more than half...wait)
			if len(p.ch) >= cap(p.ch) / 2 {
				continue
			}

			jobs, err := p.Store.FetchPending()
			if err != nil {
				slog.Error("failed to fetch jobs", "error", err)
				continue
			}

			for _, j := range jobs {
				p.ch <- j
			}
		}
	}
}
