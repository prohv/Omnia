package engines

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
		ext == ".xlsx" || ext == ".xls" || ext == ".odt" || ext == ".rtf" ||
		ext == ".txt" || ext == ".md" || ext == ".csv"

	if !isOfficeInput {
		return false
	}

	targetExt := strings.ToLower(job.TargetFormat)
	if targetExt == "" && job.OutputPath != "" {
		targetExt = strings.TrimPrefix(strings.ToLower(filepath.Ext(job.OutputPath)), ".")
	}

	if job.Operation == jobs.OperationConvert || job.Operation == jobs.OperationCompress {
		if targetExt == "pdf" || targetExt == "png" || targetExt == "jpg" || targetExt == "" {
			return true
		}
	}

	return false
}

// Execute converts a single job using technique A (Fast Launch Flags).
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

	// Technique A: High-Speed Launch Flags
	cmd := exec.CommandContext(ctx, e.executablePath,
		"--headless",
		"--nologo",
		"--nofirststartwizard",
		"--norestore",
		"--nolockcheck",
		"--convert-to", targetExt,
		job.InputPath,
		"--outdir", tempDir,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("libreoffice: execution failed: %w (output: %s)", err, string(out))
	}

	files, err := os.ReadDir(tempDir)
	if err != nil || len(files) == 0 {
		return fmt.Errorf("libreoffice: output file was not generated in temp directory")
	}

	convertedFile := filepath.Join(tempDir, files[0].Name())

	if err := os.MkdirAll(filepath.Dir(job.OutputPath), 0755); err != nil {
		return fmt.Errorf("libreoffice: failed to create output destination directory: %w", err)
	}

	_ = os.Remove(job.OutputPath)

	if err := os.Rename(convertedFile, job.OutputPath); err != nil {
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

// ExecuteBatchWithProgress converts multiple jobs in 1 single LibreOffice call with real-time per-file progress updates.
func (e *LibreOfficeEngine) ExecuteBatchWithProgress(ctx context.Context, jobGroup []jobs.Job, onProgress func(j jobs.Job)) error {
	if len(jobGroup) == 0 {
		return nil
	}
	if len(jobGroup) == 1 {
		err := e.Execute(ctx, jobGroup[0])
		if err == nil && onProgress != nil {
			onProgress(jobGroup[0])
		}
		return err
	}

	tempDir, err := os.MkdirTemp("", "omnia_soffice_batch_*")
	if err != nil {
		return fmt.Errorf("libreoffice batch: failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	targetExt := strings.ToLower(jobGroup[0].TargetFormat)
	if targetExt == "" {
		targetExt = "pdf"
	}

	args := []string{
		"--headless",
		"--nologo",
		"--nofirststartwizard",
		"--norestore",
		"--nolockcheck",
		"--convert-to", targetExt,
	}

	for _, j := range jobGroup {
		args = append(args, j.InputPath)
	}
	args = append(args, "--outdir", tempDir)

	cmd := exec.CommandContext(ctx, e.executablePath, args...)

	// Poller goroutine to trigger onProgress as each PDF appears in tempDir
	doneChan := make(chan struct{})
	go func() {
		seen := make(map[string]bool)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-doneChan:
				return
			case <-ticker.C:
				for _, j := range jobGroup {
					rawName := strings.TrimSuffix(filepath.Base(j.InputPath), filepath.Ext(j.InputPath))
					expectedTempFile := filepath.Join(tempDir, fmt.Sprintf("%s.%s", rawName, targetExt))
					if !seen[j.ID] {
						if info, err := os.Stat(expectedTempFile); err == nil && info.Size() > 0 {
							seen[j.ID] = true
							if onProgress != nil {
								onProgress(j)
							}
						}
					}
				}
			}
		}
	}()

	out, err := cmd.CombinedOutput()
	close(doneChan)

	if err != nil {
		return fmt.Errorf("libreoffice batch: execution failed: %w (output: %s)", err, string(out))
	}

	// Move converted files to target output paths
	for _, j := range jobGroup {
		rawName := strings.TrimSuffix(filepath.Base(j.InputPath), filepath.Ext(j.InputPath))
		expectedTempFile := filepath.Join(tempDir, fmt.Sprintf("%s.%s", rawName, targetExt))

		if _, err := os.Stat(expectedTempFile); err == nil {
			_ = os.MkdirAll(filepath.Dir(j.OutputPath), 0755)
			_ = os.Remove(j.OutputPath)
			_ = os.Rename(expectedTempFile, j.OutputPath)
		}
	}

	return nil
}
