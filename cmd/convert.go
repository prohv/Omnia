package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"omnia/internal/jobs"
	"omnia/internal/planner"
	"omnia/internal/router"
	"omnia/pkg/engines"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	targetFormat string
	outputDir    string
)

var convertCmd = &cobra.Command{
	Use:   "convert <file1> [file2...]",
	Short: "Convert single or multiple files to target format (default PDF)",
	Long: `Convert single or multiple input files to specified target format.
If no target format is specified, Omnia automatically converts files to PDF.
By default, newly created files are saved in the same directory as the input files.
For PDF conversions, the original input file is safely removed after verifying the new PDF is valid.
Files that are already PDFs are automatically skipped.

Examples:
  omnia convert document.docx --to pdf
  omnia convert doc1.docx pres2.pptx photo3.png --to pdf
  omnia convert report.docx --to txt`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime := time.Now()

		if targetFormat == "" {
			targetFormat = "pdf"
		}
		if outputDir == "" && AppConfig != nil {
			outputDir = AppConfig.OutputDirectory
		}

		p := planner.NewPlanner()
		r := router.NewRouter(engines.GlobalRegistry)
		ctx := context.Background()

		successCount := 0
		failCount := 0
		skipCount := 0

		for _, inputFile := range args {
			ext := strings.ToLower(filepath.Ext(inputFile))
			if ext == ".pdf" && strings.ToLower(targetFormat) == "pdf" {
				pterm.Info.Printf("Skipped %s (file is already a PDF)\n", inputFile)
				skipCount++
				continue
			}

			job, err := p.CreateJob(inputFile, jobs.OperationConvert, targetFormat, outputDir, nil)
			if err != nil {
				pterm.Error.Printf("Failed to plan job for %s: %v\n", inputFile, err)
				failCount++
				continue
			}

			engine, err := r.Route(&job)
			if err != nil {
				pterm.Error.Printf("Failed to route job for %s: %v\n", inputFile, err)
				failCount++
				continue
			}

			pterm.Info.Printf("Converting %s ➔ %s (%s)\n", inputFile, job.OutputPath, engine.Name())

			if err := engine.Execute(ctx, job); err != nil {
				pterm.Error.Printf("Conversion failed for %s: %v\n", inputFile, err)
				failCount++
			} else {
				pterm.Success.Printf("Successfully converted %s to %s\n", inputFile, job.OutputPath)
				
				// Post-conversion verify & cleanup for PDF conversions ONLY
				if job.Operation == jobs.OperationConvert && job.TargetFormat == "pdf" {
					if stat, err := os.Stat(job.OutputPath); err == nil && stat.Size() > 0 && job.InputPath != job.OutputPath {
						if err := os.Remove(job.InputPath); err == nil {
							pterm.Info.Printf("Verified & removed original file %s\n", job.InputPath)
						}
					}
				}

				successCount++
			}
		}

		if len(args) > 1 {
			pterm.Success.Printf("Batch convert completed: %d succeeded, %d skipped, %d failed in %v\n",
				successCount, skipCount, failCount, time.Since(startTime).Round(time.Millisecond))
		}

		return nil
	},
}

func init() {
	convertCmd.Flags().StringVarP(&targetFormat, "to", "t", "pdf", "target output format (e.g. pdf, txt, jpg, png)")
	convertCmd.Flags().StringVarP(&outputDir, "out", "o", "", "output directory (default: same directory as input file)")
	rootCmd.AddCommand(convertCmd)
}
