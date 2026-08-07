package progress

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
)

// ProgressTracker manages terminal UI output for job execution updates.
type ProgressTracker struct {
	pb *pterm.ProgressbarPrinter
}

func NewProgressTracker(totalJobs int, title string) *ProgressTracker {
	if title == "" {
		title = "Processing jobs..."
	}
	pb, _ := pterm.DefaultProgressbar.
		WithTotal(totalJobs).
		WithTitle(title).
		WithWriter(os.Stdout).
		Start()

	return &ProgressTracker{pb: pb}
}

func (pt *ProgressTracker) Increment(msg string) {
	if pt.pb != nil {
		if msg != "" {
			pt.pb.UpdateTitle(msg)
		}
		pt.pb.Increment()
	}
}

func (pt *ProgressTracker) Finish() {
	if pt.pb != nil {
		_, _ = pt.pb.Stop()
	}
}

// PrintSummary outputs clean batch processing statistics.
func PrintSummary(completed, failed, skipped int, totalDuration string) {
	pterm.Println()
	pterm.DefaultSection.Println("Execution Summary")

	tableData := pterm.TableData{
		{"Metric", "Value"},
		{"Completed", pterm.Green(fmt.Sprintf("%d", completed))},
		{"Failed", pterm.Red(fmt.Sprintf("%d", failed))},
		{"Skipped", pterm.Yellow(fmt.Sprintf("%d", skipped))},
		{"Total Duration", totalDuration},
	}

	_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
}
