package router

import (
	"fmt"

	"omnia/internal/jobs"
	"omnia/internal/mime"
	"omnia/pkg/engines"
)

// Router resolves processing engines based on file MIME types and job parameters.
type Router struct {
	registry *engines.Registry
}

func NewRouter(reg *engines.Registry) *Router {
	if reg == nil {
		reg = engines.GlobalRegistry
	}
	return &Router{registry: reg}
}

// Route inspects the job's file MIME type and resolves the appropriate processing Engine.
func (r *Router) Route(job *jobs.Job) (engines.Engine, error) {
	if job.MimeType == "" {
		detectedMime, err := mime.DetectMime(job.InputPath)
		if err == nil {
			job.MimeType = detectedMime
		}
	}

	engine, err := r.registry.GetEngineForJob(*job)
	if err != nil {
		return nil, fmt.Errorf("router: failed to route job %s: %w", job.ID, err)
	}

	return engine, nil
}
