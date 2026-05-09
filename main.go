package main

import (
	"context"
	"database/sql"
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
	"github.com/luqmanshaban/kuda/repository"
	"github.com/luqmanshaban/kuda/structs"
)

var db *sql.DB

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
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

	// jobs
	mux.HandleFunc("POST /jobs", jobHandler.CreateJH)
	mux.HandleFunc("GET /jobs/{job_id}", jobHandler.GetJH)

	// users
	mux.HandleFunc("POST /users", userHandler.CreateUH)
	mux.HandleFunc("GET /users/jobs/{user_id}", userHandler.GetUserJH)

	// queues
	mux.HandleFunc("POST /queues", queueHandler.CreateQH)


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
