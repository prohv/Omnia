package engines

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"omnia/internal/jobs"
)

// LibreOfficeEngine invokes soffice in headless mode to convert Office documents to PDF/Images.
type LibreOfficeEngine struct {
	executablePath string
}

func NewLibreOfficeEngine() *LibreOfficeEngine {
	e := &LibreOfficeEngine{}
	e.executablePath = e.findExecutable()
	return e
}

func (e *LibreOfficeEngine) findExecutable() string {
	candidates := []string{"soffice", "libreoffice"}
	for _, cand := range candidates {
		if path, err := exec.LookPath(cand); err == nil {
			return path
		}
	}

	if runtime.GOOS == "windows" {
		winPaths := []string{
			`C:\Program Files\LibreOffice\program\soffice.exe`,
			`C:\Program Files (x86)\LibreOffice\program\soffice.exe`,
		}
		for _, p := range winPaths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	return ""
}

func (e *LibreOfficeEngine) Name() string {
	return "LibreOfficeEngine (soffice --headless)"
}

func (e *LibreOfficeEngine) CanHandle(job jobs.Job) bool {
	if e.executablePath == "" {
		return false
	}

	ext := strings.ToLower(filepath.Ext(job.InputPath))
	isOfficeInput := ext == ".docx" || ext == ".doc" || ext == ".pptx" || ext == ".ppt" ||
		ext == ".xlsx" || ext == ".xls" || ext == ".odt" || ext == ".rtf"

	if !isOfficeInput {
		return false
	}

	targetExt := strings.ToLower(job.TargetFormat)
	if targetExt == "" && job.OutputPath != "" {
		targetExt = strings.TrimPrefix(strings.ToLower(filepath.Ext(job.OutputPath)), ".")
	}

	// Office files to PDF or Image rendering via LibreOffice
	if job.Operation == jobs.OperationConvert || job.Operation == jobs.OperationCompress {
		if targetExt == "pdf" || targetExt == "png" || targetExt == "jpg" || targetExt == "" {
			return true
		}
	}

	return false
}

func (e *LibreOfficeEngine) Execute(ctx context.Context, job jobs.Job) error {
	if e.executablePath == "" {
		return fmt.Errorf("libreoffice: soffice executable not found on system PATH")
	}

	if job.OutputPath == "" {
		return fmt.Errorf("libreoffice: output path is required")
	}

	tempDir, err := os.MkdirTemp("", "omnia_soffice_*")
	if err != nil {
		return fmt.Errorf("libreoffice: failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	targetExt := strings.ToLower(job.TargetFormat)
	if targetExt == "" {
		targetExt = strings.TrimPrefix(strings.ToLower(filepath.Ext(job.OutputPath)), ".")
	}
	if targetExt == "" {
		targetExt = "pdf"
	}

	cmd := exec.CommandContext(ctx, e.executablePath,
		"--headless",
		"--convert-to", targetExt,
		job.InputPath,
		"--outdir", tempDir,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("libreoffice: execution failed: %w (output: %s)", err, string(out))
	}

	// Find the converted file in tempDir
	files, err := os.ReadDir(tempDir)
	if err != nil || len(files) == 0 {
		return fmt.Errorf("libreoffice: output file was not generated in temp directory")
	}

	convertedFile := filepath.Join(tempDir, files[0].Name())

	if err := os.MkdirAll(filepath.Dir(job.OutputPath), 0755); err != nil {
		return fmt.Errorf("libreoffice: failed to create output destination directory: %w", err)
	}

	// Remove target if it already exists
	_ = os.Remove(job.OutputPath)

	if err := os.Rename(convertedFile, job.OutputPath); err != nil {
		// Fallback copy if rename fails across drives
		inputBytes, err := os.ReadFile(convertedFile)
		if err != nil {
			return fmt.Errorf("libreoffice: failed to read converted file: %w", err)
		}
		if err := os.WriteFile(job.OutputPath, inputBytes, 0644); err != nil {
			return fmt.Errorf("libreoffice: failed to copy converted file: %w", err)
		}
	}

	return nil
}
