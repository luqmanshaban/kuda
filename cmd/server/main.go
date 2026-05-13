package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/luqmanshaban/kuda/internal/api"
	"github.com/luqmanshaban/kuda/internal/config"
	"github.com/luqmanshaban/kuda/internal/core"
	"github.com/luqmanshaban/kuda/internal/store"
	"github.com/luqmanshaban/kuda/internal/worker"
)

func main() {
	// 1. Config
	cfg := config.Load()

	// 2. Infrastructure logging setup
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 3. Database
	db := store.Connect(cfg)
	defer db.Close()


	// 4. Stores 
	jobStore := store.NewJobStore(db)
	queueStore := store.NewQueueStore(db)

	// 5. Stale jobs
	err := jobStore.ResetStaleJobs()
	if err != nil {
		slog.Error("failed to reset stale jobs", "error", err)
	}

	// 6. worker pool
	jobCh := make(chan core.Job, 100)
	pool := worker.NewPool(jobStore, 100)
	wg := pool.Start(jobCh)

	// 7. Producer
	pollCtx, pollCancel := context.WithCancel(context.Background())
	producer := worker.NewProducer(jobStore, jobCh)
	go producer.Start(pollCtx)
	
	// 8. Http Server
	srv := api.NewServer(cfg, jobStore, queueStore)

	go srv.Start()

	// 9. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit,syscall.SIGTERM, syscall.SIGINT)
	<- quit

	pollCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	if err := srv.ShutDown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}
	close(jobCh)
	wg.Wait()
	slog.Info("server exited gracefully")
	
}