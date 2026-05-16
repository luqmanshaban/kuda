package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/luqmanshaban/kuda/internal/api/handlers"
	"github.com/luqmanshaban/kuda/internal/api/middleware"
	"github.com/luqmanshaban/kuda/internal/config"
	"github.com/luqmanshaban/kuda/internal/store"
)

type Server struct {
	cfg *config.Config
	httpServer *http.Server
}

func NewServer(cfg *config.Config, js *store.JobStore, qs *store.QueueStore, dl *store.DeadLetterJobStore) *Server {
	mux := http.NewServeMux()

	// initializing the handlers
	jobHandler := &handlers.JobHandler{Store: js}
	queueHandler := &handlers.QueueHandler{Store: qs}
	dljHandler := &handlers.DeadLetterHandler{Store: dl}

	// api key initialization
	apiKey := middleware.InitAPIKey()
	auth := middleware.AuthMiddleware(apiKey)

	// job routes
	mux.Handle("POST /jobs", auth(http.HandlerFunc(jobHandler.CreateJH)))
	mux.Handle("GET /jobs", auth(http.HandlerFunc(jobHandler.GetJobsH)))
	mux.Handle("GET /jobs/{job_id}", auth(http.HandlerFunc(jobHandler.GetSingleJobH)))
	mux.Handle("GET /jobs/batch/{batch_id}", auth(http.HandlerFunc(jobHandler.GetJobsBatchH)))

	// dead letters 
	mux.Handle("GET /dead-letter-jobs", auth(http.HandlerFunc(dljHandler.GetDeadJobs)))

	// queues routes
	mux.Handle("POST /queues", auth(http.HandlerFunc(queueHandler.CreateQH)))
	mux.Handle("GET /queues/{name}", auth(http.HandlerFunc(queueHandler.GetSingleQH)))
	mux.Handle("GET /queues", auth(http.HandlerFunc(queueHandler.GetQH)))
	mux.Handle("PUT /queues", auth(http.HandlerFunc(queueHandler.UpdateQH)))

	return &Server{
		cfg: cfg,
		httpServer: &http.Server{
			Addr: cfg.Port,
			Handler: mux,
		},
	}
	
}

func (s *Server) Start() {
	slog.Info("Http server starting...", "Addr: ", s.cfg.Port)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Http server failed", "error", err)
	}
	
}

func (s *Server)  ShutDown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}