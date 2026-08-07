package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration settings for Omnia.
type Config struct {
	Workers         int    `mapstructure:"workers"`
	Compression     string `mapstructure:"compression"`
	OutputDirectory string `mapstructure:"output_directory"`
	KeepOriginal    bool   `mapstructure:"keep_original"`
	Overwrite       bool   `mapstructure:"overwrite"`
	LogLevel        string `mapstructure:"log_level"`
}

// DefaultWorkers returns min(runtime.NumCPU(), 6).
func DefaultWorkers() int {
	cpus := runtime.NumCPU()
	if cpus > 6 {
		return 6
	}
	if cpus < 1 {
		return 1
	}
	return cpus
}

// DefaultConfig returns default configuration values.
func DefaultConfig() *Config {
	return &Config{
		Workers:         DefaultWorkers(),
		Compression:     "balanced",
		OutputDirectory: "./output",
		KeepOriginal:    true,
		Overwrite:       false,
		LogLevel:        "info",
	}
}

// LoadConfig initializes Viper, sets defaults, checks config file path or ~/.omnia/config.yaml, and unmarshals into Config.
func LoadConfig(cfgFile string) (*Config, error) {
	v := viper.New()

	v.SetDefault("workers", DefaultWorkers())
	v.SetDefault("compression", "balanced")
	v.SetDefault("output_directory", "./output")
	v.SetDefault("keep_original", true)
	v.SetDefault("overwrite", false)
	v.SetDefault("log_level", "info")

	v.SetEnvPrefix("OMNIA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			configDir := filepath.Join(home, ".omnia")
			v.AddConfigPath(configDir)
			v.SetConfigName("config")
			v.SetConfigType("yaml")
		}
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && !os.IsNotExist(err) && cfgFile != "" {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return cfg, nil
}
