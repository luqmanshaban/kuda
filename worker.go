package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/luqmanshaban/kuda/repository"
	"github.com/luqmanshaban/kuda/structs"
)

type JobWorker struct {
	Repo *repository.JobRepository
}

func (w JobWorker) Worker(worker int, jch <-chan structs.Job, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jch {
		err := w.Deliver(j)
		if err != nil {
			fmt.Printf("Failed to send job to webhook, error: %v\n", err)
			_, err := w.Repo.UpdateJobState(j.ID, "failed")
			if err != nil {
				fmt.Printf("Failed to update job status, error: %v\n", err)
			}
		} else {
			log.Println("Job sent to webhook successfully ---- attempting to update database...")
			_, err := w.Repo.UpdateJobState(j.ID, "completed")
			if err != nil {
				fmt.Printf("Failed to update job status: %v\n", err)
			}
		    log.Println("Worker completed successfully")
		}
	}

}

func (w JobWorker) Deliver(job structs.Job) error {
	body, err := json.Marshal(map[string]any{
		"job_id":  job.ID,
		"payload": job.Payload,
	})
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(job.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("webhook returned %d", resp.StatusCode)
}
