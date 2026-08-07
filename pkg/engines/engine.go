package engines

import (
	"context"

	"omnia/internal/jobs"
)

// Engine defines the uniform interface implemented by all Omnia processing backends.
type Engine interface {
	// Name returns the descriptive unique identifier for the engine.
	Name() string

	// CanHandle returns true if this engine is capable of processing the given job.
	CanHandle(job jobs.Job) bool

	// Execute performs the processing operation specified in the job.
	Execute(ctx context.Context, job jobs.Job) error
}
