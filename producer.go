package main

import (
	"sync"

	"github.com/luqmanshaban/kuda/structs"
)

type JobProducer struct {
	Worker JobWorker
}

func (p *JobProducer) StartPool(numW int, jobs <- chan structs.Job) {
	var wg sync.WaitGroup

	for i := 1; i <= numW; i ++ {
		wg.Add(1)
		go p.Worker.Worker(jobs, &wg)
	}

	go func() {
		wg.Wait()
	}()
}