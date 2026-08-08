package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"omnia/internal/jobs"
	"omnia/internal/planner"
	"omnia/internal/progress"
	"omnia/internal/router"
	"omnia/internal/scanner"
	"omnia/pkg/engines"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	targetFormat string
	outputDir    string
	useClipboard bool
)

var convertCmd = &cobra.Command{
	Use:   "convert [file1 file2...]",
	Short: "Convert single or multiple files to target format (default PDF)",
	Long: `Convert single or multiple input files to specified target format.
If no target format is specified, Omnia automatically converts files to PDF.
By default, newly created files are saved in the same directory as the input files.
For PDF conversions, the original input file is safely removed after verifying the new PDF is valid.
Files that are already PDFs are automatically skipped.

Use --clip flag to automatically read copied file paths from system clipboard.

Examples:
  omnia convert document.docx --to pdf
  omnia convert --clip --to pdf
  omnia convert doc1.docx pres2.pptx photo3.png --to pdf`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime := time.Now()

		if useClipboard {
			clipPaths, err := scanner.ReadClipboardPaths()
			if err != nil || len(clipPaths) == 0 {
				pterm.Error.Println("Clipboard is empty or does not contain valid file paths")
				return nil
			}
			pterm.Info.Printf("Read %d file path(s) from clipboard\n", len(clipPaths))
			args = append(args, clipPaths...)
		}

		if len(args) == 0 {
			return cmd.Help()
		}

		if targetFormat == "" {
			targetFormat = "pdf"
		}
		if outputDir == "" && AppConfig != nil {
			outputDir = AppConfig.OutputDirectory
		}

		p := planner.NewPlanner()
		r := router.NewRouter(engines.GlobalRegistry)
		ctx := context.Background()

		// Filter files and skip existing PDFs
		var inputFiles []string
		skipCount := 0

		for _, f := range args {
			ext := strings.ToLower(filepath.Ext(f))
			if ext == ".pdf" && strings.ToLower(targetFormat) == "pdf" {
				pterm.Info.Printf("Skipped %s (file is already a PDF)\n", f)
				skipCount++
				continue
			}
			inputFiles = append(inputFiles, f)
		}

		if len(inputFiles) == 0 {
			if skipCount > 0 {
				pterm.Success.Printf("All %d file(s) are already PDFs. Nothing to convert.\n", skipCount)
			}
			return nil
		}

		// Initialize Live Progress Bar
		tracker := progress.NewProgressTracker(len(inputFiles), "Preparing engines")

		// Check if we can batch dispatch Office files to LibreOffice
		officeEngine := engines.NewLibreOfficeEngine()
		var officeJobs []jobs.Job
		var otherJobs []jobs.Job

		for _, inputFile := range inputFiles {
			job, err := p.CreateJob(inputFile, jobs.OperationConvert, targetFormat, outputDir, nil)
			if err != nil {
				tracker.Increment(fmt.Sprintf("Failed %s", inputFile))
				continue
			}

			if officeEngine.CanHandle(job) {
				officeJobs = append(officeJobs, job)
			} else {
				otherJobs = append(otherJobs, job)
			}
		}

		successCount := 0
		failCount := 0

		// Execute Batch LibreOffice Dispatch with Real-Time Per-File Progress Updates
		if len(officeJobs) > 0 {
			err := officeEngine.ExecuteBatchWithProgress(ctx, officeJobs, func(j jobs.Job) {
				tracker.Increment(fmt.Sprintf("Converted %s", filepath.Base(j.InputPath)))
			})

			if err == nil {
				for _, j := range officeJobs {
					successCount++
					if stat, err := os.Stat(j.OutputPath); err == nil && stat.Size() > 0 && j.InputPath != j.OutputPath {
						_ = os.Remove(j.InputPath)
					}
				}
			} else {
				// Fallback to individual executions
				for _, j := range officeJobs {
					if err := officeEngine.Execute(ctx, j); err == nil {
						successCount++
						tracker.Increment(fmt.Sprintf("Converted %s", filepath.Base(j.InputPath)))
						if stat, err := os.Stat(j.OutputPath); err == nil && stat.Size() > 0 && j.InputPath != j.OutputPath {
							_ = os.Remove(j.InputPath)
						}
					} else {
						failCount++
						tracker.Increment(fmt.Sprintf("Failed %s", filepath.Base(j.InputPath)))
					}
				}
			}
		}

		// Execute remaining native jobs
		for _, job := range otherJobs {
			engine, err := r.Route(&job)
			if err != nil {
				failCount++
				tracker.Increment(fmt.Sprintf("Failed %s", filepath.Base(job.InputPath)))
				continue
			}

			if err := engine.Execute(ctx, job); err != nil {
				failCount++
				tracker.Increment(fmt.Sprintf("Failed %s", filepath.Base(job.InputPath)))
			} else {
				successCount++
				tracker.Increment(fmt.Sprintf("Converted %s", filepath.Base(job.InputPath)))
				
				if job.Operation == jobs.OperationConvert && job.TargetFormat == "pdf" {
					if stat, err := os.Stat(job.OutputPath); err == nil && stat.Size() > 0 && job.InputPath != job.OutputPath {
						_ = os.Remove(job.InputPath)
					}
				}
			}
		}

		tracker.Finish()

		// Print Summary
		durationStr := time.Since(startTime).Round(time.Millisecond).String()
		progress.PrintSummary(successCount, failCount, skipCount, durationStr)

		return nil
	},
}

func init() {
	convertCmd.Flags().StringVarP(&targetFormat, "to", "t", "pdf", "target output format (e.g. pdf, txt, jpg, png)")
	convertCmd.Flags().StringVarP(&outputDir, "out", "o", "", "output directory (default: same directory as input file)")
	convertCmd.Flags().BoolVarP(&useClipboard, "clip", "c", false, "read copied file paths directly from clipboard")
	rootCmd.AddCommand(convertCmd)
}
