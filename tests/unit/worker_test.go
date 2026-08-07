package unit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"omnia/internal/jobs"
	"omnia/internal/router"
	"omnia/internal/worker"
	"omnia/pkg/engines"
)

func TestWorkerPool(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "input.txt")
	dstFile := filepath.Join(tempDir, "output.txt")
	_ = os.WriteFile(srcFile, []byte("worker pool test content"), 0644)

	reg := engines.NewRegistry()
	reg.Register(engines.NewOpenXMLEngine())

	r := router.NewRouter(reg)
	pool := worker.NewPool(2, r)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool.Start(ctx)

	job := jobs.Job{
		ID:           "w-job-1",
		InputPath:    srcFile,
		OutputPath:   dstFile,
		Operation:    jobs.OperationExtractText,
		TargetFormat: "txt",
	}

	pool.Submit(job)
	pool.Close()

	var resList []jobs.JobResult
	for res := range pool.Results() {
		resList = append(resList, res)
	}

	if len(resList) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resList))
	}
}
