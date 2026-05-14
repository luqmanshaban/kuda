package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	// "github.com/luqmanshaban/kuda/metrics"
	"github.com/luqmanshaban/kuda/internal/core"
	"github.com/luqmanshaban/kuda/internal/store"
)

type JobHandler struct {
	Store *store.JobStore
}

func (h *JobHandler) CreateJH(w http.ResponseWriter, r *http.Request) {

	// read the row body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"message": "unable to read body"})
		return
	}
	var incomingJobs []core.JobRequest

	// check if input is array or object
	trimmedBody := bytes.TrimSpace(body)

	if len(trimmedBody) > 0 && trimmedBody[0] == '[' {
		err := json.Unmarshal(trimmedBody, &incomingJobs)
		if err != nil {
			WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
	} else {
		var singleRequest core.JobRequest
		if err := json.Unmarshal(body, &singleRequest); err != nil {
			fmt.Printf("JSON Decode Error: %v\n", err)
			WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		if singleRequest.RunsAt.IsZero() {
			singleRequest.RunsAt = time.Now().UTC()
		}

		job_id, err := h.Store.CreateSingleJob(singleRequest)
		if err != nil {
			slog.Error("job creation failed", "error", err)
			WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to create job"})
			return
		}
		WriteJson(w, http.StatusCreated, map[string]int{"job_id": job_id})
		return
	}

	batch_id := GenerateBatchId()
	for i := range incomingJobs {
		if incomingJobs[i].RunsAt.IsZero() {
			incomingJobs[i].RunsAt = time.Now().UTC()
		}
		
		incomingJobs[i].BatchID = batch_id
	}

	count, err := h.Store.CreateJobs(incomingJobs)
	if err != nil {
		slog.Error("job creation failed", "component", "repository", "op", "create_job", "error", err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to create job"})
		return
	}
	// promethues job enque
	// metrics.JobsEnqueued.Add(float64(len(jobs)))

	WriteJson(w, http.StatusCreated, map[string]any{"batch_id": batch_id, "count": count})
}

func (h *JobHandler) GetSingleJobH(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("job_id")
	if id == "" {
		WriteJson(w, http.StatusBadRequest, map[string]string{"message": "job id not provided"})
		return
	}

	job_id, err := strconv.Atoi(id)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{"message": "job id format is invalid"})
		return
	}

	j, err := h.Store.GetJob(job_id)
	if err != nil {
		slog.Error("job fetching failed", "component", "repository", "op", "fetch_job", "error", err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to fetch for job"})
		return
	}

	WriteJson(w, http.StatusOK, j)
}


func (h *JobHandler) GetJobsH(w http.ResponseWriter, r *http.Request) {
	// CHECK if query filters are passed
	statuses := []string{"pending", "running", "completed", "failed", "dead"}
	param := r.URL.Query().Get("status")
	for _, status := range statuses {
		if param == status {
			j, err := h.Store.GetFilteredJobs(status)

			if err != nil {
				slog.Error("job filtration failed", "component", "repository", "op", "filter_user_jobs", "error", err)
				WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to fetch for job"})
				return
			}
			WriteJson(w, http.StatusOK, j)
			return
		}
	}

	js, err := h.Store.GetJobs()
	if err != nil {
		slog.Error("user fetching failed", "component", "repository", "op", "fetch_user", "error", err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to fetch for job"})
		return
	}

	WriteJson(w, http.StatusOK, js)
}

func (h *JobHandler) GetJobsBatchH(w http.ResponseWriter, r *http.Request) {
	batchId := r.PathValue("batch_id")
	if batchId == "" {
		WriteJson(w, http.StatusBadRequest, map[string]string{"message":"batch id required"})
		return
	}
	
	js, err := h.Store.GetJobsBatchId(batchId)
	if err != nil {
		slog.Error("user fetching failed", "component", "repository", "op", "fetch_user", "error", err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to fetch for job"})
		return
	}

	WriteJson(w, http.StatusOK, js)
}
