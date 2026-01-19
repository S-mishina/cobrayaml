package cobrayaml

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// BaseCommand represents the base structure for CLI commands
type BaseCommand struct {
	RootCmd    *cobra.Command
	ConfigFile string
	ConfigPath string
}

// NewBaseCommand creates a new base command with common setup
func NewBaseCommand(use, short, long string) *BaseCommand {
	base := &BaseCommand{
		RootCmd: &cobra.Command{
			Use:   use,
			Short: short,
			Long:  long,
		},
	}

	base.setupCommonFlags()
	return base
}

// setupCommonFlags sets up common flags for all CLI tools
func (b *BaseCommand) setupCommonFlags() {
	b.RootCmd.PersistentFlags().StringVar(&b.ConfigFile, "config", "", "config file (default is $HOME/.config/<tool-name>.yaml)")
	b.RootCmd.PersistentFlags().StringVar(&b.ConfigPath, "config-path", "", "config directory path")
}

// InitConfig initializes the configuration
func (b *BaseCommand) InitConfig(toolName string) {
	if b.ConfigFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(b.ConfigFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}

		// Set config search paths
		configDir := filepath.Join(home, ".config")
		if b.ConfigPath != "" {
			configDir = b.ConfigPath
		}

		viper.AddConfigPath(configDir)
		viper.AddConfigPath(".")
		viper.SetConfigName(toolName)
		viper.SetConfigType("yaml")
	}

	// Set environment variable prefix
	viper.SetEnvPrefix(toolName)
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}
}

// Execute executes the root command
func (b *BaseCommand) Execute() error {
	return b.RootCmd.Execute()
}

// AddCommand adds a subcommand to the root command
func (b *BaseCommand) AddCommand(cmd *cobra.Command) {
	b.RootCmd.AddCommand(cmd)
}
