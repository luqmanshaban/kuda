package main

import (
	"database/sql"
	"fmt"
	"net/http"
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

	// call the producer
	producer.StartPool(numWorkers, jobCh)

	go func() {
		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {

			jobs, err := jobRepo.FetchPending()
			if err != nil {
				fmt.Println(err)
				continue
			}

			for _, j := range jobs {
				if len(jobs) == 0 {
					fmt.Println("NO pending jobs")
				}
				jobCh <- j
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

	fmt.Println("SERVER RUNNING ON localhost:8000")

	http.ListenAndServe(":8000", mux)
}
