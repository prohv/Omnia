package cmd

import (
	"context"
	"fmt"
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
	Use:   "convert <file>",
	Short: "Convert file format to target (default PDF)",
	Long: `Convert input file to specified target format.
If no target format is specified, Omnia automatically converts the file to PDF.

Examples:
  omnia convert document.docx --to pdf
  omnia convert presentation.pptx --to pdf
  omnia convert photo.png --to pdf
  omnia convert document.docx --to txt`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		startTime := time.Now()

		if targetFormat == "" {
			targetFormat = "pdf"
		}
		if outputDir == "" && AppConfig != nil {
			outputDir = AppConfig.OutputDirectory
		}

		p := planner.NewPlanner()
		job, err := p.CreateJob(inputFile, jobs.OperationConvert, targetFormat, outputDir, nil)
		if err != nil {
			return fmt.Errorf("failed to plan conversion job: %w", err)
		}

		r := router.NewRouter(engines.GlobalRegistry)
		engine, err := r.Route(&job)
		if err != nil {
			return fmt.Errorf("failed to route conversion job: %w", err)
		}

		pterm.Info.Printf("Converting %s ➔ %s (%s)\n", inputFile, job.OutputPath, engine.Name())

		ctx := context.Background()
		if err := engine.Execute(ctx, job); err != nil {
			pterm.Error.Printf("Conversion failed: %v\n", err)
			return err
		}

		pterm.Success.Printf("Successfully converted %s to %s in %v\n", inputFile, job.OutputPath, time.Since(startTime).Round(time.Millisecond))
		return nil
	},
}

func init() {
	convertCmd.Flags().StringVarP(&targetFormat, "to", "t", "pdf", "target output format (e.g. pdf, txt, jpg, png)")
	convertCmd.Flags().StringVarP(&outputDir, "out", "o", "", "output directory (default ./output)")
	rootCmd.AddCommand(convertCmd)
}
