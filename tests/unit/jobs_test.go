package unit

import (
	"testing"

	"omnia/internal/jobs"
)

func TestJobValidate(t *testing.T) {
	validJob := jobs.Job{
		ID:        "job-1",
		InputPath: "test.docx",
		Operation: jobs.OperationConvert,
	}
	if err := validJob.Validate(); err != nil {
		t.Errorf("expected valid job to pass validation, got: %v", err)
	}

	invalidJob := jobs.Job{
		ID: "job-2",
	}
	if err := invalidJob.Validate(); err == nil {
		t.Errorf("expected invalid job with empty input path to fail validation")
	}
}

func TestJobGetOption(t *testing.T) {
	job := jobs.Job{
		ID:        "job-1",
		InputPath: "test.docx",
		Operation: jobs.OperationCompress,
		Options: map[string]string{
			"level": "high",
		},
	}

	if val := job.GetOption("level", "balanced"); val != "high" {
		t.Errorf("expected 'high', got %s", val)
	}
	if val := job.GetOption("nonexistent", "default_val"); val != "default_val" {
		t.Errorf("expected 'default_val', got %s", val)
	}
}
