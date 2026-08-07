package jobs

import (
	"fmt"
	"time"
)

// Operation represents the processing action required for a job.
type Operation string

const (
	OperationConvert     Operation = "convert"
	OperationCompress    Operation = "compress"
	OperationInfo        Operation = "info"
	OperationExtractText Operation = "extract_text"
)

// Job defines the complete execution unit for file processing.
type Job struct {
	ID           string
	InputPath    string
	OutputPath   string
	Operation    Operation
	TargetFormat string
	MimeType     string
	Options      map[string]string
}

// JobResult captures the execution outcome of a job.
type JobResult struct {
	JobID      string
	Success    bool
	Error      error
	Duration   time.Duration
	InputSize  int64
	OutputSize int64
}

// Validate checks that required job fields are properly initialized.
func (j *Job) Validate() error {
	if j.InputPath == "" {
		return fmt.Errorf("job %s: input path cannot be empty", j.ID)
	}
	if j.Operation == "" {
		return fmt.Errorf("job %s: operation cannot be empty", j.ID)
	}
	return nil
}

// GetOption retrieves a job option value with a fallback default.
func (j *Job) GetOption(key, defaultValue string) string {
	if j.Options == nil {
		return defaultValue
	}
	if val, ok := j.Options[key]; ok && val != "" {
		return val
	}
	return defaultValue
}
