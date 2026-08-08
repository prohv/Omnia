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
	processClip     bool
)

var processCmd = &cobra.Command{
	Use:   "process [path1 path2...]",
	Short: "Concurrently process multiple files or folders together",
	Long: `Batch scan multiple files or folders, detect file types, plan execution pipelines,
and process all files concurrently using a high-performance worker pool.

Use --clip flag to automatically read copied file paths from system clipboard.

Examples:
  omnia process my_folder/ --workers 4 --to pdf
  omnia process --clip --to pdf
  omnia process folder1/ folder2/ --recursive`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime := time.Now()

		if processClip {
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

		// 1. Scan Paths
		scn := scanner.NewScanner()
		var files []string
		for _, targetPath := range args {
			scanned, err := scn.ScanPath(targetPath, recursiveScan)
			if err == nil {
				files = append(files, scanned...)
			}
		}

		if len(files) == 0 {
			pterm.Warning.Println("No processable files found in specified paths")
			return nil
		}

		pterm.Info.Printf("Discovered %d file(s) across target path(s) (Workers: %d, Target: %s)\n", len(files), numWorkers, processToFormat)

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
	processCmd.Flags().StringVarP(&processOutDir, "out", "o", "", "output directory (default: same directory as input file)")
	processCmd.Flags().BoolVarP(&processClip, "clip", "c", false, "read copied file paths directly from clipboard")
	rootCmd.AddCommand(processCmd)
}
