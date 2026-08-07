package unit

import (
	"os"
	"path/filepath"
	"testing"

	"omnia/internal/jobs"
	"omnia/internal/router"
	"omnia/pkg/engines"
)

func TestRouter(t *testing.T) {
	tempDir := t.TempDir()
	txtFile := filepath.Join(tempDir, "test.txt")
	_ = os.WriteFile(txtFile, []byte("sample text content"), 0644)

	reg := engines.NewRegistry()
	reg.Register(engines.NewImageEngine())
	reg.Register(engines.NewPDFCPUEngine())
	reg.Register(engines.NewOpenXMLEngine())

	r := router.NewRouter(reg)

	job := jobs.Job{
		ID:           "route-1",
		InputPath:    txtFile,
		Operation:    jobs.OperationConvert,
		TargetFormat: "pdf",
	}

	// Should attempt routing
	_, err := r.Route(&job)
	if job.MimeType == "" {
		t.Errorf("router failed to populate MIME type on job")
	}
	if err == nil {
		t.Logf("routed engine for txt to pdf successfully")
	}
}
