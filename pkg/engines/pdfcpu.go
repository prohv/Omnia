package engines

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"omnia/internal/jobs"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// PDFCPUEngine handles PDF compression, optimization, split/merge natively in Go using pdfcpu.
type PDFCPUEngine struct{}

func NewPDFCPUEngine() *PDFCPUEngine {
	return &PDFCPUEngine{}
}

func (e *PDFCPUEngine) Name() string {
	return "PDFCPUEngine (Native PDF)"
}

func (e *PDFCPUEngine) CanHandle(job jobs.Job) bool {
	ext := strings.ToLower(filepath.Ext(job.InputPath))
	mime := strings.ToLower(job.MimeType)

	isPDFInput := ext == ".pdf" || mime == "application/pdf"
	isPDFTarget := strings.ToLower(job.TargetFormat) == "pdf" || strings.ToLower(filepath.Ext(job.OutputPath)) == ".pdf"

	if job.Operation == jobs.OperationCompress && isPDFInput {
		return true
	}
	if job.Operation == jobs.OperationConvert && isPDFInput && isPDFTarget {
		return true
	}
	if job.Operation == jobs.OperationInfo && isPDFInput {
		return true
	}
	return false
}

func (e *PDFCPUEngine) Execute(ctx context.Context, job jobs.Job) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	conf := model.NewDefaultConfiguration()

	switch job.Operation {
	case jobs.OperationCompress, jobs.OperationConvert:
		// Optimize/Compress PDF file
		if job.OutputPath == "" {
			return fmt.Errorf("pdfcpu: output path is required")
		}

		// Ensure destination directory exists
		if err := os.MkdirAll(filepath.Dir(job.OutputPath), 0755); err != nil {
			return fmt.Errorf("pdfcpu: failed to create output directory: %w", err)
		}

		err := api.OptimizeFile(job.InputPath, job.OutputPath, conf)
		if err != nil {
			return fmt.Errorf("pdfcpu: failed to optimize PDF: %w", err)
		}
		return nil

	case jobs.OperationInfo:
		// Read PDF metadata context
		_, err := api.ReadContextFile(job.InputPath)
		if err != nil {
			return fmt.Errorf("pdfcpu: failed to read PDF info: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("pdfcpu: unsupported operation %s", job.Operation)
	}
}
