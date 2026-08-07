package cmd

import (
	"fmt"
	"os"

	"omnia/internal/config"
	"omnia/internal/logger"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	quiet   bool
	jsonLog bool

	AppConfig *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "omnia",
	Short: "Omnia — Fast, intelligent, cross-platform file processing CLI",
	Long: `Omnia is a fast, intelligent, cross-platform file processing CLI written in Go.
It follows a "Native Go First" philosophy, orchestrating native PDF/Image processing engines
and external LibreOffice conversion via a unified interface.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load Configuration
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}
		AppConfig = cfg

		// Determine Logger Mode
		logMode := cfg.LogLevel
		if verbose {
			logMode = string(logger.ModeVerbose)
		} else if quiet {
			logMode = string(logger.ModeQuiet)
		} else if jsonLog {
			logMode = string(logger.ModeJSON)
		}

		logger.SetupLogger(logMode, os.Stdout)
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.omnia/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose (debug) logging")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "enable quiet (error only) logging")
	rootCmd.PersistentFlags().BoolVar(&jsonLog, "json", false, "output logs in JSON format")
}
