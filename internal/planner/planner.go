package planner

import (
	"fmt"
	"path/filepath"
	"strings"

	"omnia/internal/jobs"
)

// Planner constructs execution jobs and output path layouts.
type Planner struct{}

func NewPlanner() *Planner {
	return &Planner{}
}

// CreateJob constructs a single processing job for an input file.
func (p *Planner) CreateJob(inputPath string, op jobs.Operation, targetFormat string, outputDir string, opts map[string]string) (jobs.Job, error) {
	absInput, err := filepath.Abs(inputPath)
	if err != nil {
		return jobs.Job{}, fmt.Errorf("planner: invalid input path %s: %w", inputPath, err)
	}

	baseName := filepath.Base(absInput)
	ext := filepath.Ext(baseName)
	rawName := strings.TrimSuffix(baseName, ext)

	if targetFormat == "" {
		targetFormat = "pdf"
	}
	targetFormat = strings.TrimPrefix(strings.ToLower(targetFormat), ".")

	if outputDir == "" {
		outputDir = "./output"
	}

	outFileName := fmt.Sprintf("%s.%s", rawName, targetFormat)
	outputPath := filepath.Join(outputDir, outFileName)

	jobID := fmt.Sprintf("job-%s-%s", op, rawName)

	return jobs.Job{
		ID:           jobID,
		InputPath:    absInput,
		OutputPath:   outputPath,
		Operation:    op,
		TargetFormat: targetFormat,
		Options:      opts,
	}, nil
}
