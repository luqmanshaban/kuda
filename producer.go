package main

import (
	"sync"

	"github.com/luqmanshaban/kuda/structs"
)

type JobProducer struct {
	Worker JobWorker
}

func (p *JobProducer) StartPool(numW int, jobs <- chan structs.Job) *sync.WaitGroup {
	var wg sync.WaitGroup

	for i := 1; i <= numW; i ++ {
		wg.Add(1)
		go p.Worker.Worker(i,jobs, &wg)
	}

	// go func() {
	// 	wg.Wait()
	// }()
	return &wg
}