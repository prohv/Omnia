package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version values populated during build or defaults
var (
	Version   = "0.1.2"
	GitCommit = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Omnia",
	Long:  `Display version, git commit, build date, and system runtime info for Omnia.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Omnia v%s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Go Version : %s\n", runtime.Version())
		fmt.Printf("Git Commit : %s\n", GitCommit)
		fmt.Printf("Build Date : %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
