package cmd

import (
	"context"
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

		for _, inputFile := range args {
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
				successCount++
			}
		}

		if len(args) > 1 {
			pterm.Success.Printf("Batch convert completed: %d succeeded, %d failed in %v\n",
				successCount, failCount, time.Since(startTime).Round(time.Millisecond))
		}

		return nil
	},
}

func init() {
	convertCmd.Flags().StringVarP(&targetFormat, "to", "t", "pdf", "target output format (e.g. pdf, txt, jpg, png)")
	convertCmd.Flags().StringVarP(&outputDir, "out", "o", "", "output directory (default ./output)")
	rootCmd.AddCommand(convertCmd)
}
