package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"omnia/internal/jobs"
	"omnia/internal/planner"
	"omnia/internal/router"
	"omnia/pkg/engines"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	compressLevel string
	compressOut   string
)

var compressCmd = &cobra.Command{
	Use:   "compress <file>",
	Short: "Compress PDF or Image file size",
	Long: `Compress PDF or image file using native optimization algorithms.

Compression Levels: low, balanced, high, extreme (default: balanced)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		startTime := time.Now()

		if compressOut == "" && AppConfig != nil {
			compressOut = AppConfig.OutputDirectory
		}

		p := planner.NewPlanner()
		opts := map[string]string{"compression": compressLevel}

		job, err := p.CreateJob(inputFile, jobs.OperationCompress, "", compressOut, opts)
		if err != nil {
			return fmt.Errorf("failed to plan compression job: %w", err)
		}

		r := router.NewRouter(engines.GlobalRegistry)
		engine, err := r.Route(&job)
		if err != nil {
			return fmt.Errorf("failed to route compression job: %w", err)
		}

		pterm.Info.Printf("Compressing %s (%s, Level: %s)\n", inputFile, engine.Name(), compressLevel)

		ctx := context.Background()
		if err := engine.Execute(ctx, job); err != nil {
			pterm.Error.Printf("Compression failed: %v\n", err)
			return err
		}

		origStat, _ := os.Stat(inputFile)
		newStat, _ := os.Stat(job.OutputPath)

		var origSize, newSize int64
		if origStat != nil {
			origSize = origStat.Size()
		}
		if newStat != nil {
			newSize = newStat.Size()
		}

		pterm.Success.Printf("Successfully compressed %s to %s in %v (Size: %d ➔ %d bytes)\n",
			inputFile, job.OutputPath, time.Since(startTime).Round(time.Millisecond), origSize, newSize)

		return nil
	},
}

func init() {
	compressCmd.Flags().StringVarP(&compressLevel, "level", "l", "balanced", "compression level (low, balanced, high, extreme)")
	compressCmd.Flags().StringVarP(&compressOut, "out", "o", "", "output directory (default ./output)")
	rootCmd.AddCommand(compressCmd)
}
