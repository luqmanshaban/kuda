package main

import (
	"fmt"
	"sync"

	"github.com/luqmanshaban/kuda/repository"
	"github.com/luqmanshaban/kuda/structs"
)

type JobWorker struct {
	Repo *repository.JobRepository
}

func (w JobWorker) Worker(jch <- chan structs.Job, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jch {
		fmt.Println("Worker Worked on job id: ", j.ID)
		_, err := w.Repo.UpdateJobState(j.ID, "running")
		if err != nil {
			fmt.Println(err)
		}
	}
	 
}