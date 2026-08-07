package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"omnia/internal/jobs"
	"omnia/internal/planner"
	"omnia/internal/progress"
	"omnia/internal/router"
	"omnia/internal/scanner"
	"omnia/internal/worker"
	"omnia/pkg/engines"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	numWorkers      int
	recursiveScan   bool
	processToFormat string
	processOutDir   string
)

var processCmd = &cobra.Command{
	Use:   "process <directory>",
	Short: "Concurrently process all files in a folder",
	Long: `Batch scan a folder, detect file types, plan execution pipelines,
and process all files concurrently using a high-performance worker pool.

Example:
  omnia process my_folder/ --workers 4 --to pdf --recursive`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := args[0]
		startTime := time.Now()

		if numWorkers < 1 {
			if AppConfig != nil {
				numWorkers = AppConfig.Workers
			} else {
				numWorkers = 4
			}
		}
		if processOutDir == "" && AppConfig != nil {
			processOutDir = AppConfig.OutputDirectory
		}
		if processToFormat == "" {
			processToFormat = "pdf"
		}

		// 1. Scan Path
		scn := scanner.NewScanner()
		files, err := scn.ScanPath(targetPath, recursiveScan)
		if err != nil {
			return fmt.Errorf("failed to scan path %s: %w", targetPath, err)
		}

		if len(files) == 0 {
			pterm.Warning.Printf("No processable files found in %s\n", targetPath)
			return nil
		}

		pterm.Info.Printf("Discovered %d file(s) in %s (Workers: %d, Target: %s)\n", len(files), targetPath, numWorkers, processToFormat)

		// 2. Setup Signal Context for Ctrl+C
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			pterm.Warning.Println("\nReceived cancellation signal. Gracefully shutting down workers...")
			cancel()
		}()

		// 3. Plan Jobs
		pln := planner.NewPlanner()
		var jobList []jobs.Job
		for _, f := range files {
			job, err := pln.CreateJob(f, jobs.OperationConvert, processToFormat, processOutDir, nil)
			if err == nil {
				jobList = append(jobList, job)
			}
		}

		// 4. Start Worker Pool
		rtr := router.NewRouter(engines.GlobalRegistry)
		pool := worker.NewPool(numWorkers, rtr)
		pool.Start(ctx)

		// Submit Jobs
		go func() {
			for _, job := range jobList {
				pool.Submit(job)
			}
			pool.Close()
		}()

		// 5. Track Progress
		tracker := progress.NewProgressTracker(len(jobList), "Processing batch files...")

		completedCount := 0
		failedCount := 0

		for res := range pool.Results() {
			if res.Success {
				completedCount++
				tracker.Increment(fmt.Sprintf("Completed %s", res.JobID))
			} else {
				failedCount++
				tracker.Increment(fmt.Sprintf("Failed %s", res.JobID))
			}
		}

		tracker.Finish()

		// 6. Summary Report
		durationStr := time.Since(startTime).Round(time.Millisecond).String()
		progress.PrintSummary(completedCount, failedCount, 0, durationStr)

		return nil
	},
}

func init() {
	processCmd.Flags().IntVarP(&numWorkers, "workers", "w", 0, "number of worker goroutines (default min(CPU, 6))")
	processCmd.Flags().BoolVarP(&recursiveScan, "recursive", "r", false, "scan directory recursively")
	processCmd.Flags().StringVarP(&processToFormat, "to", "t", "pdf", "target output format for converted files")
	processCmd.Flags().StringVarP(&processOutDir, "out", "o", "", "output directory (default ./output)")
	rootCmd.AddCommand(processCmd)
}
