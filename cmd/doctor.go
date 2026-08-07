package cmd

import (
	"fmt"
	"os"

	"omnia/internal/dependencies"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system dependencies and configuration",
	Long:  `Run diagnostics on native Go engines, external executables (soffice), and configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).WithWriter(os.Stdout).Println("Omnia Diagnostic Doctor")

		deps := dependencies.CheckDependencies()

		tableData := pterm.TableData{
			{"Status", "Engine / Dependency", "Type", "Location / Details"},
		}

		allOk := true
		for _, dep := range deps {
			statusStr := pterm.Green("✔ Found")
			typeStr := "Native Go"

			if !dep.IsNative {
				typeStr = "External Binary"
				if !dep.Installed {
					statusStr = pterm.Red("✗ Missing")
					allOk = false
				}
			}

			tableData = append(tableData, []string{
				statusStr,
				dep.Name,
				typeStr,
				dep.Path,
			})
		}

		_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

		fmt.Println()

		for _, dep := range deps {
			if !dep.Installed && !dep.IsNative {
				pterm.Warning.Printf("Missing Dependency: %s\n", dep.Name)
				pterm.Info.Printf("Purpose: %s\n", dep.Purpose)
				pterm.Info.Printf("Installation Guide: %s\n\n", dep.InstallGuide)
			}
		}

		if AppConfig != nil {
			pterm.Success.Printf("Configuration loaded successfully (Workers: %d, LogLevel: %s)\n", AppConfig.Workers, AppConfig.LogLevel)
		}

		if allOk {
			pterm.Success.Println("All core engines and external dependencies are ready!")
		} else {
			pterm.Warning.Println("Omnia can process PDFs and Images natively, but LibreOffice (soffice) is required for Office document conversion.")
		}
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
