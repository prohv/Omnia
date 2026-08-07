package cmd

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"time"

	"omnia/internal/dependencies"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Automatically install and verify missing external dependencies",
	Long: `Cross-platform setup command that detects missing external dependencies (like LibreOffice)
and automatically runs the native package manager (winget on Windows, brew on macOS, apt on Linux)
to install and verify them for first-time use.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgMagenta)).WithWriter(os.Stdout).Println("Omnia Automated Dependency Setup")

		deps := dependencies.CheckDependencies()
		missingOffice := false

		for _, dep := range deps {
			if !dep.IsNative && !dep.Installed {
				missingOffice = true
				break
			}
		}

		if !missingOffice {
			pterm.Success.Println("All external dependencies and native engines are already installed and verified!")
			return nil
		}

		pterm.Info.Printf("Missing LibreOffice dependency detected on %s/%s.\n", runtime.GOOS, runtime.GOARCH)
		pterm.Info.Println("Starting automated package installation...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		var execCmd *exec.Cmd

		switch runtime.GOOS {
		case "windows":
			pterm.Info.Println("Executing: winget install --id TheDocumentFoundation.LibreOffice -e")
			execCmd = exec.CommandContext(ctx, "winget", "install", "--id", "TheDocumentFoundation.LibreOffice", "-e", "--accept-source-agreements", "--accept-package-agreements")

		case "darwin":
			pterm.Info.Println("Executing: brew install --cask libreoffice")
			execCmd = exec.CommandContext(ctx, "brew", "install", "--cask", "libreoffice")

		default:
			pterm.Info.Println("Executing: sudo apt update && sudo apt install -y libreoffice")
			execCmd = exec.CommandContext(ctx, "sudo", "apt", "install", "-y", "libreoffice")
		}

		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr

		if err := execCmd.Run(); err != nil {
			pterm.Error.Printf("Package manager installation encountered an issue: %v\n", err)
		}

		// Re-verify dependencies
		pterm.Println()
		pterm.Info.Println("Verifying installation...")

		recheckedDeps := dependencies.CheckDependencies()
		nowInstalled := false

		for _, dep := range recheckedDeps {
			if !dep.IsNative && dep.Installed {
				nowInstalled = true
				pterm.Success.Printf("Verified %s at: %s\n", dep.Name, dep.Path)
				break
			}
		}

		if nowInstalled {
			pterm.Success.Println("Omnia setup completed successfully! All processing engines are ready.")
		} else {
			pterm.Warning.Println("Installation completed, but soffice was not found on system PATH. You may need to restart your terminal.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
