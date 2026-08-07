package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"omnia/internal/jobs"
	"omnia/internal/mime"
	"omnia/internal/router"
	"omnia/pkg/engines"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <file>",
	Short: "Display metadata and MIME type of a file",
	Long: `Inspect file metadata, magic-number MIME signature, file size,
and determine the target processing engines available for the file.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		info, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("info: failed to inspect file %s: %w", filePath, err)
		}

		absPath, _ := filepath.Abs(filePath)
		mimeType, err := mime.DetectMime(filePath)
		if err != nil {
			mimeType = "unknown"
		}

		pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgBlue)).WithWriter(os.Stdout).Println("Omnia File Information Inspector")

		tableData := pterm.TableData{
			{"Property", "Value"},
			{"Filename", filepath.Base(filePath)},
			{"Absolute Path", absPath},
			{"File Size", fmt.Sprintf("%d bytes (%.2f KB)", info.Size(), float64(info.Size())/1024.0)},
			{"MIME Type", pterm.Cyan(mimeType)},
			{"Last Modified", info.ModTime().Format("2006-01-02 15:04:05")},
		}

		// Query available engine for convert
		r := router.NewRouter(engines.GlobalRegistry)
		dummyJob := jobs.Job{
			ID:           "info-check",
			InputPath:    absPath,
			Operation:    jobs.OperationConvert,
			TargetFormat: "pdf",
			MimeType:     mimeType,
		}
		if eng, err := r.Route(&dummyJob); err == nil {
			tableData = append(tableData, []string{"Primary Engine", pterm.Green(eng.Name())})
		} else {
			tableData = append(tableData, []string{"Primary Engine", pterm.Yellow("No direct engine (or text fallback)")})
		}

		_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
