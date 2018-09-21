package main

import "time"

type WorkerPool struct {
	rateLimit time.Duration
	queue     []Worker
	jobsChan  chan APIGetter
}

type Worker struct {
	apiRequest APIGetter
}

func (wp *WorkerPool) Run() {
	for _, w := range wp.queue {

		wp.jobsChan <- w.apiRequest

	}
}
