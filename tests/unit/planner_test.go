package unit

import (
	"path/filepath"
	"testing"

	"omnia/internal/jobs"
	"omnia/internal/planner"
)

func TestPlannerCreateJob(t *testing.T) {
	p := planner.NewPlanner()

	job, err := p.CreateJob("report.docx", jobs.OperationConvert, "pdf", "./out", nil)
	if err != nil {
		t.Fatalf("unexpected planner error: %v", err)
	}

	if job.TargetFormat != "pdf" {
		t.Errorf("expected target format 'pdf', got %s", job.TargetFormat)
	}
	if filepath.Base(job.OutputPath) != "report.pdf" {
		t.Errorf("expected output filename 'report.pdf', got %s", filepath.Base(job.OutputPath))
	}
}
