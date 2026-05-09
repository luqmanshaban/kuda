package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/luqmanshaban/kuda/repository"
	"github.com/luqmanshaban/kuda/utils"
)

type JobHandler struct {
	Repo *repository.JobRepository
}

func (h *JobHandler) CreateJH(w http.ResponseWriter, r *http.Request) {
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
			incomingJobs[i].RunsAt = time.Now().Add(5 * time.Minute)
		}
	}

	jobs, err := h.Repo.CreateJob(incomingJobs)
	if err != nil {
		fmt.Println(err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to create job"})
		return
	}

	utils.WriteJson(w, http.StatusCreated, jobs)
}

func (h *JobHandler) GetJH(w http.ResponseWriter, r *http.Request) {
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

	j, err := h.Repo.GetJob(job_id)
	if err != nil {
		fmt.Println(err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to fetch for job"})
		return
	}

	utils.WriteJson(w, http.StatusOK, j)
}



func (h *JobHandler) GetUsersJobs(w http.ResponseWriter, r *http.Request) {
	
}
