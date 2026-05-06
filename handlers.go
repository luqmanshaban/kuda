package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func CreateJH(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Payload json.RawMessage `json:"payload"`
		RunsAt time.Time `json:"runs_at"`
	} 


	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, map[string]string{ "message": "payload required" })
		return 
	}


	jId, err := CreateJob(payload.Payload, payload.RunsAt)
	if err != nil {
		fmt.Println(err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to create job"})
		return
	}

	WriteJson(w, http.StatusCreated, map[string]int{"job_id": jId})
}

func GetJH(w http.ResponseWriter, r*http.Request) {
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

	j, err := GetJob(job_id)
	if err != nil {
		fmt.Println(err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to fetch for job"})
		return
	}

	WriteJson(w, http.StatusOK, j)
}