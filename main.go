package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/luqmanshaban/kuda/handlers"
	"github.com/luqmanshaban/kuda/metrics"
	"github.com/luqmanshaban/kuda/middleware"
	"github.com/luqmanshaban/kuda/repository"
	"github.com/luqmanshaban/kuda/structs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var db *sql.DB

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}
	db = ConnectToDB()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// slog definitions
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// initialize repository
	jobRepo := &repository.JobRepository{DB: db}
	userRepo := &repository.UserRepository{DB: db}
	queueRepo := &repository.QueueRepo{DB: db}
	authMid := middleware.NewAuthMiddleware(userRepo)
	// initialize handlers
	jobHandler := &handlers.JobHandler{Repo: jobRepo}
	userHandler := &handlers.UserHandler{Repo: userRepo}
	queueHandler := &handlers.QueueHandler{Repo: queueRepo}
	// initialize worker
	worker := JobWorker{Repo: jobRepo}
	// initialize producer
	producer := JobProducer{Worker: worker}

	// define the jobs and num of workers
	const numWorkers = 100
	jobCh := make(chan structs.Job, 100)

	// call the producer and pass out wg so main can wait on it
	wg := producer.StartPool(numWorkers, jobCh)

	// create a context for the poll
	pollContext, CancelPoll := context.WithCancel(context.Background())

	// check for staled jobs and reset them
	err := jobRepo.ResetStaleJobs()
	if err != nil {
		log.Fatal("failed to reset stale jobs:", err)
	}
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-pollContext.Done():
				return
			case <-ticker.C:
				if len(jobCh) > cap(jobCh)/2 {
					continue
				}
				jobs, err := jobRepo.FetchPending()
				if err != nil {
					slog.Error("fetch pending jobs failed", "component", "repository", "op", "fetch_pending", "error", err)
					continue
				}

				for _, j := range jobs {
					jobCh <- j
				}
			}
		}
	}()

	mux := http.NewServeMux()

	// Initialize prometheus
	metrics.Init()
	mux.Handle("GET /metrics", promhttp.Handler())
	// jobs
	mux.Handle("POST /jobs", authMid.Authenticate(http.HandlerFunc(jobHandler.CreateJH)))
	mux.Handle("GET /jobs/{job_id}",  authMid.Authenticate(http.HandlerFunc(jobHandler.GetJH)))

	// users
	mux.HandleFunc("POST /users", userHandler.CreateUH)
	mux.Handle("GET /users/jobs", authMid.Authenticate(http.HandlerFunc(userHandler.GetUserJH)))
	mux.Handle("GET /users/me", authMid.Authenticate(http.HandlerFunc(userHandler.GetUser)))

	// queues
	mux.Handle("POST /queues", authMid.Authenticate(http.HandlerFunc(queueHandler.CreateQH)))
	mux.Handle("GET /queues", authMid.Authenticate(http.HandlerFunc(queueHandler.GetQH)))
	mux.Handle("PUT /queues/{queue_id}", authMid.Authenticate(http.HandlerFunc(queueHandler.UpdateQH)))

	// health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string {"status": "Unhealthy"})
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string {"status": "healthy"})
	})


	// http server in a gouroutine for graceful shutdown
	srv := &http.Server{Addr: ":8000", Handler: mux}
	go func() {
		fmt.Println("SERVER RUNNING ON PORT 8000")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	CancelPoll()

	fmt.Println("shutting down...")

	// stop accepting http requests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	// close chanels - stop workers - wait for in-flight jobs to finish
	close(jobCh)
	wg.Wait()

	fmt.Println("All workers done exiting")
}
