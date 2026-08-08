package progress

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

const FixedTitleWidth = 35

// ProgressTracker manages terminal UI output for job execution updates.
type ProgressTracker struct {
	total     int
	current   int
	lastMsg    string
	flashMsg   string
	flashUntil time.Time
	startTime  time.Time
	mu         sync.Mutex
	done      chan struct{}
	wg        sync.WaitGroup
}

func NewProgressTracker(totalJobs int, title string) *ProgressTracker {
	if title == "" {
		title = "Processing jobs"
	}

	pt := &ProgressTracker{
		total:     totalJobs,
		startTime: time.Now(),
		done:      make(chan struct{}),
	}

	pt.wg.Add(1)
	go pt.renderLoop()
	return pt
}

// Remove SetActive since we no longer use it

func (pt *ProgressTracker) Increment(msg string) {
	pt.mu.Lock()
	pt.current++
	if msg != "" {
		pt.lastMsg = msg
		pt.flashMsg = msg
		pt.flashUntil = time.Now().Add(1 * time.Second)
	}
	pt.mu.Unlock()
}

func (pt *ProgressTracker) Finish() {
	pt.done <- struct{}{}
	pt.wg.Wait()
}

func (pt *ProgressTracker) renderLoop() {
	defer pt.wg.Done()
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()
	dotCount := 0

	for {
		select {
		case <-pt.done:
			pt.renderLine(0, true)
			fmt.Println() // Move to next line on completion
			return
		case <-ticker.C:
			dotCount = (dotCount + 1) % 4
			pt.renderLine(dotCount, false)
		}
	}
}

func (pt *ProgressTracker) renderLine(dots int, isFinished bool) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	// 1. Build Title
	title := ""
	if isFinished {
		title = pterm.Green("Done   ")
	} else {
		if time.Now().Before(pt.flashUntil) && pt.flashMsg != "" {
			title = pterm.Green(pt.flashMsg)
		} else {
			dotsStr := strings.Repeat(".", dots) + strings.Repeat(" ", 3-dots)
			title = pterm.Yellow("Processing" + dotsStr)
		}
	}

	// Ensure fixed width
	rawTitle := pterm.RemoveColorFromString(title)
	if len(rawTitle) < FixedTitleWidth {
		title += strings.Repeat(" ", FixedTitleWidth-len(rawTitle))
	} else if len(rawTitle) > FixedTitleWidth {
		// Truncate cleanly while trying to preserve color formatting
		// Since color codes make blind truncation messy, we can rebuild the string truncated.
		if isFinished {
			title = pterm.Green("Done" + strings.Repeat(" ", FixedTitleWidth-4))
		} else if time.Now().Before(pt.flashUntil) && pt.flashMsg != "" {
			truncated := pt.flashMsg
			if len(truncated) > FixedTitleWidth {
				truncated = truncated[:FixedTitleWidth-3] + "..."
			}
			title = pterm.Green(truncated)
		}
	}

	// 2. Build Bar (matching pterm default style: no brackets, ThemeDefault color)
	barWidth := 25
	current := pt.current
	if current > pt.total {
		current = pt.total
	}
	filled := 0
	if pt.total > 0 {
		filled = (current * barWidth) / pt.total
	}
	unfilled := barWidth - filled
	
	barStyle := pterm.ThemeDefault.ProgressbarBarStyle
	if len(barStyle) == 0 {
		barStyle = pterm.Style{pterm.FgLightCyan}
	}
	
	barChar := pterm.DefaultProgressbar.BarCharacter
	if barChar == "" {
		barChar = "█"
	}
	barFiller := pterm.DefaultProgressbar.BarFiller
	if barFiller == "" {
		barFiller = " "
	}

	barStr := barStyle.Sprint(strings.Repeat(barChar, filled)) + strings.Repeat(barFiller, unfilled)

	// 3. Build Percentage (Yellow -> Green gradient)
	percentage := 0
	if pt.total > 0 {
		percentage = int((float64(current) / float64(pt.total)) * 100)
	}
	startColor := pterm.NewRGB(255, 255, 0)
	endColor := pterm.NewRGB(0, 255, 0)
	currentRGB := startColor.Fade(0, float32(pt.total), float32(current), endColor)
	percStr := currentRGB.Sprintf("%3d%%", percentage)

	// 4. Build Time
	elapsed := time.Since(pt.startTime)
	timeStr := fmt.Sprintf("%dms", elapsed.Milliseconds())
	if elapsed.Seconds() >= 1 {
		timeStr = fmt.Sprintf("%.1fs", elapsed.Seconds())
	}
	timeStr = pterm.Gray(timeStr)

	// 5. Output line
	pterm.Printo(fmt.Sprintf("%s %s %s %s %s", title, barStr, percStr, pterm.Gray("|"), timeStr))
}

// PrintSummary outputs clean batch processing statistics.
func PrintSummary(completed, failed, skipped int, totalDuration string) {
	pterm.Println()
	pterm.DefaultSection.Println("Execution Summary")

	tableData := pterm.TableData{
		{"Metric", "Value"},
		{"Completed", pterm.Green(fmt.Sprintf("%d", completed))},
	}

	if failed > 0 {
		tableData = append(tableData, []string{"Failed", pterm.Red(fmt.Sprintf("%d", failed))})
	}
	if skipped > 0 {
		tableData = append(tableData, []string{"Skipped", pterm.Yellow(fmt.Sprintf("%d", skipped))})
	}

	tableData = append(tableData, []string{"Total Duration", totalDuration})

	_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
}
