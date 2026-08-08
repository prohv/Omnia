package worker

import (
	"context"
	"os"
	"sync"
	"time"

	"omnia/internal/jobs"
	"omnia/internal/router"
)

// Pool manages a bounded goroutine worker pool executing processing jobs.
type Pool struct {
	numWorkers int
	router     *router.Router
	jobChan    chan jobs.Job
	resChan    chan jobs.JobResult
	wg         sync.WaitGroup
}

func NewPool(numWorkers int, r *router.Router) *Pool {
	if numWorkers < 1 {
		numWorkers = 1
	}
	if r == nil {
		r = router.NewRouter(nil)
	}

	return &Pool{
		numWorkers: numWorkers,
		router:     r,
		jobChan:    make(chan jobs.Job, 100),
		resChan:    make(chan jobs.JobResult, 100),
	}
}

// Start launches worker goroutines.
func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}
}

func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-p.jobChan:
			if !ok {
				return
			}
			p.processJob(ctx, job)
		}
	}
}

func (p *Pool) processJob(ctx context.Context, job jobs.Job) {
	startTime := time.Now()

	var inputSize int64
	if info, err := os.Stat(job.InputPath); err == nil {
		inputSize = info.Size()
	}

	engine, err := p.router.Route(&job)
	if err != nil {
		p.resChan <- jobs.JobResult{
			JobID:     job.ID,
			Success:   false,
			Error:     err,
			Duration:  time.Since(startTime),
			InputSize: inputSize,
		}
		return
	}

	execErr := engine.Execute(ctx, job)
	duration := time.Since(startTime)

	var outputSize int64
	if execErr == nil {
		if info, err := os.Stat(job.OutputPath); err == nil {
			outputSize = info.Size()

			// Post-conversion verify & cleanup for PDF conversions ONLY
			if job.Operation == jobs.OperationConvert && job.TargetFormat == "pdf" && info.Size() > 0 && job.InputPath != job.OutputPath {
				_ = os.Remove(job.InputPath)
			}
		}
	}

	p.resChan <- jobs.JobResult{
		JobID:      job.ID,
		Success:    execErr == nil,
		Error:      execErr,
		Duration:   duration,
		InputSize:  inputSize,
		OutputSize: outputSize,
	}
}

// Submit enqueues a job into the worker queue.
func (p *Pool) Submit(job jobs.Job) {
	p.jobChan <- job
}

// Close closes the input job channel and waits for workers to finish.
func (p *Pool) Close() {
	close(p.jobChan)
	p.wg.Wait()
	close(p.resChan)
}

// Results returns the channel receiving job execution results.
func (p *Pool) Results() <-chan jobs.JobResult {
	return p.resChan
}
