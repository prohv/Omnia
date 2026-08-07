package engines

import (
	"fmt"
	"sync"

	"omnia/internal/jobs"
)

// Registry manages registered processing engines and resolves appropriate engines for jobs.
type Registry struct {
	mu      sync.RWMutex
	engines []Engine
}

// NewRegistry creates a new empty Engine Registry.
func NewRegistry() *Registry {
	return &Registry{
		engines: make([]Engine, 0),
	}
}

// Register appends a new engine to the registry.
func (r *Registry) Register(e Engine) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.engines = append(r.engines, e)
}

// GetEngineForJob finds the first registered engine capable of handling the job.
func (r *Registry) GetEngineForJob(job jobs.Job) (Engine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, e := range r.engines {
		if e.CanHandle(job) {
			return e, nil
		}
	}

	return nil, fmt.Errorf("no engine found capable of handling job %s (operation: %s, mime: %s, target: %s)", job.ID, job.Operation, job.MimeType, job.TargetFormat)
}

// ListEngines returns a slice of all registered engines.
func (r *Registry) ListEngines() []Engine {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cp := make([]Engine, len(r.engines))
	copy(cp, r.engines)
	return cp
}

// GlobalRegistry holds the default application engine registry instance.
var GlobalRegistry = NewRegistry()
