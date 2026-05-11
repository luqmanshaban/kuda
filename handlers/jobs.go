package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/luqmanshaban/kuda/metrics"
	"github.com/luqmanshaban/kuda/repository"
	"github.com/luqmanshaban/kuda/structs"
	"github.com/luqmanshaban/kuda/utils"
)

type JobHandler struct {
	Repo *repository.JobRepository
}

func (h *JobHandler) CreateJH(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("user").(structs.User).ID
	// read the row body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "unable to read body"})
		return
	}
	var incomingJobs []repository.JobRequest

	// check if input is array or object
	trimmedBody := bytes.TrimSpace(body)

	if len(trimmedBody) > 0 && trimmedBody[0] == '[' {
		err := json.Unmarshal(trimmedBody, &incomingJobs)
		if err != nil {
			utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
	} else {
		var singleRequest repository.JobRequest
		if err := json.Unmarshal(trimmedBody, &singleRequest); err != nil {
			utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
			return
		}
		incomingJobs = append(incomingJobs, singleRequest)
	}

	for i := range incomingJobs {
		if incomingJobs[i].RunsAt.IsZero() {
			incomingJobs[i].RunsAt = time.Now()
		}
	}

	jobs, err := h.Repo.CreateJob(incomingJobs, id)
	if err != nil {
		slog.Error("job creation failed", "component", "repository", "op", "create_job", "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to create job"})
		return
	}
	// promethues job enque
	metrics.JobsEnqueued.Add(float64(len(jobs)))

	utils.WriteJson(w, http.StatusCreated, jobs)
}

func (h *JobHandler) GetJH(w http.ResponseWriter, r *http.Request) {
	uId := r.Context().Value("user").(structs.User).ID
	id := r.PathValue("job_id")
	if id == "" {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "job id not provided"})
		return
	}

	job_id, err := strconv.Atoi(id)
	if err != nil {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "job id format is invalid"})
		return
	}

	j, err := h.Repo.GetJob(job_id, uId)
	if err != nil {
		slog.Error("job fetching failed", "component", "repository", "op", "fetch_job", "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to fetch for job"})
		return
	}

	utils.WriteJson(w, http.StatusOK, j)
}

func (h *JobHandler) GetUsersJobs(w http.ResponseWriter, r *http.Request) {

}
