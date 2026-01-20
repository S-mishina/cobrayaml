//go:generate go run ./internal/docgen/main.go

// Package cobrayaml provides a YAML-based command builder for creating cobra CLI applications.
//
// This package allows you to define CLI commands declaratively in a commands.yaml file,
// then build and execute them using the CommandBuilder.
//
// # Quick Start
//
// 1. Create a commands.yaml file defining your CLI structure
// 2. Use NewCommandBuilder to load the configuration
// 3. Register your handler functions with RegisterFunction
// 4. Build and execute with BuildRootCommand
//
// Example:
//
//	builder, _ := cobrayaml.NewCommandBuilder("commands.yaml")
//	builder.RegisterFunction("runList", runList)
//	rootCmd, _ := builder.BuildRootCommand()
//	rootCmd.Execute()
package cobrayaml

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)


// ArgsConfig represents argument validation configuration in commands.yaml.
//
// Fields:
//   - Type: Validation type (none, any, exact, min, max, range)
//   - Count: Required count for "exact" type
//   - Min: Minimum count for "min" or "range" type
//   - Max: Maximum count for "max" or "range" type
//
// Example YAML:
//
//	args:
//	  type: exact
//	  count: 2
//
//	args:
//	  type: range
//	  min: 1
//	  max: 3
type ArgsConfig struct {
	Type  string `yaml:"type"`            // none, any, exact, min, max, range
	Count int    `yaml:"count,omitempty"` // for exact
	Min   int    `yaml:"min,omitempty"`   // for min, range
	Max   int    `yaml:"max,omitempty"`   // for max, range
}

// Supported args types for commands.yaml.
const (
	ArgsTypeNone  = "none"
	ArgsTypeAny   = "any"
	ArgsTypeExact = "exact"
	ArgsTypeMin   = "min"
	ArgsTypeMax   = "max"
	ArgsTypeRange = "range"
)

// SupportedArgsTypes lists all supported argument validation types.
var SupportedArgsTypes = []string{
	ArgsTypeNone,
	ArgsTypeAny,
	ArgsTypeExact,
	ArgsTypeMin,
	ArgsTypeMax,
	ArgsTypeRange,
}

// Supported flag types for commands.yaml.
// Use these values in the "type" field of flag definitions.
const (
	// FlagTypeString represents a string flag.
	// Go type: string
	// Example: --config /path/to/file
	FlagTypeString = "string"

	// FlagTypeBool represents a boolean flag.
	// Go type: bool
	// Default value in YAML: "true" or "false"
	// Example: --debug
	FlagTypeBool = "bool"

	// FlagTypeInt represents an integer flag.
	// Go type: int
	// Example: --timeout 30
	FlagTypeInt = "int"

	// FlagTypeStringSlice represents a comma-separated string list flag.
	// Go type: []string
	// Example: --tags a,b,c
	FlagTypeStringSlice = "stringSlice"
)

// SupportedFlagTypes lists all supported flag types.
var SupportedFlagTypes = []string{
	FlagTypeString,
	FlagTypeBool,
	FlagTypeInt,
	FlagTypeStringSlice,
}

// CommandConfig represents a command configuration in commands.yaml.
//
// Fields:
//   - Use: Command name and argument pattern (e.g., "add <name> <value>")
//   - Aliases: Alternative command names
//   - Short: Brief description shown in help
//   - Long: Detailed description
//   - Args: Argument validation configuration (see ArgsConfig)
//   - RunFunc: Name of the handler function registered with RegisterFunction
//   - Flags: List of flag definitions
//   - Commands: Nested subcommands
//   - Hidden: Hide command from help output
type CommandConfig struct {
	Use      string                   `yaml:"use"`
	Aliases  []string                 `yaml:"aliases,omitempty"`
	Short    string                   `yaml:"short"`
	Long     string                   `yaml:"long,omitempty"`
	Args     *ArgsConfig              `yaml:"args,omitempty"`
	RunFunc  string                   `yaml:"run_func,omitempty"`
	Flags    []FlagConfig             `yaml:"flags,omitempty"`
	Commands map[string]CommandConfig `yaml:"commands,omitempty"`
	Hidden   bool                     `yaml:"hidden,omitempty"`
}

// FlagConfig represents a flag configuration in commands.yaml.
//
// Fields:
//   - Name: Flag name (e.g., "namespace" for --namespace)
//   - Shorthand: Short flag (e.g., "n" for -n)
//   - Type: Flag type (see SupportedFlagTypes)
//   - DefaultValue: Default value as string
//   - Usage: Description shown in help
//   - Required: Mark flag as required
//   - Persistent: Inherit flag to all subcommands
//   - Hidden: Hide flag from help output
type FlagConfig struct {
	Name         string `yaml:"name"`
	Shorthand    string `yaml:"shorthand,omitempty"`
	Type         string `yaml:"type"`
	DefaultValue string `yaml:"default,omitempty"`
	Usage        string `yaml:"usage"`
	Required     bool   `yaml:"required,omitempty"`
	Persistent   bool   `yaml:"persistent,omitempty"`
	Hidden       bool   `yaml:"hidden,omitempty"`
}

// ToolConfig represents the entire tool configuration in commands.yaml.
//
// Example YAML structure:
//
//	name: "my-tool"
//	description: "My CLI tool"
//	version: "1.0.0"
//	root:
//	  use: "my-tool"
//	  short: "A CLI tool"
//	  flags:
//	    - name: "config"
//	      type: "string"
//	      persistent: true
//	commands:
//	  list:
//	    use: "list"
//	    short: "List items"
//	    args: "NoArgs"
//	    run_func: "runList"
type ToolConfig struct {
	Name        string                    `yaml:"name"`
	Description string                    `yaml:"description,omitempty"`
	Version     string                    `yaml:"version,omitempty"`
	Root        CommandConfig             `yaml:"root"`
	Commands    map[string]CommandConfig  `yaml:"commands,omitempty"`
	Functions   map[string]string         `yaml:"functions,omitempty"`
}

// CommandBuilder builds cobra commands from YAML configuration
type CommandBuilder struct {
	config    *ToolConfig
	funcMap   map[string]any
}

// NewCommandBuilder creates a new command builder
func NewCommandBuilder(configPath string) (*CommandBuilder, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config ToolConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %v", err)
	}

	if err := ValidateConfig(&config); err != nil {
		return nil, err
	}

	return &CommandBuilder{
		config:  &config,
		funcMap: make(map[string]any),
	}, nil
}

// NewCommandBuilderFromString creates a new command builder from YAML string
func NewCommandBuilderFromString(yamlContent string) (*CommandBuilder, error) {
	var config ToolConfig
	if err := yaml.Unmarshal([]byte(yamlContent), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %v", err)
	}

	if err := ValidateConfig(&config); err != nil {
		return nil, err
	}

	return &CommandBuilder{
		config:  &config,
		funcMap: make(map[string]any),
	}, nil
}

// RegisterFunction registers a function that can be called from YAML config
func (cb *CommandBuilder) RegisterFunction(name string, fn any) {
	cb.funcMap[name] = fn
}

// BuildRootCommand builds the root command from configuration
func (cb *CommandBuilder) BuildRootCommand() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:     cb.config.Root.Use,
		Short:   cb.config.Root.Short,
		Long:    cb.config.Root.Long,
		Version: cb.config.Version,
	}

	// Set run function for root command
	if cb.config.Root.RunFunc != "" {
		if fn, exists := cb.funcMap[cb.config.Root.RunFunc]; exists {
			if runE, ok := fn.(func(*cobra.Command, []string) error); ok {
				rootCmd.RunE = runE
			} else {
				return nil, fmt.Errorf("function %s is not of type func(*cobra.Command, []string) error", cb.config.Root.RunFunc)
			}
		} else {
			return nil, fmt.Errorf("function %s not registered", cb.config.Root.RunFunc)
		}
	}

	// Add flags to root command
	if err := cb.addFlags(rootCmd, cb.config.Root.Flags); err != nil {
		return nil, err
	}

	// Build and add subcommands
	for name, cmdConfig := range cb.config.Commands {
		subCmd, err := cb.buildCommand(name, cmdConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build command %s: %v", name, err)
		}
		rootCmd.AddCommand(subCmd)
	}

	return rootCmd, nil
}

// buildCommand builds a single command from configuration
func (cb *CommandBuilder) buildCommand(_ string, config CommandConfig) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     config.Use,
		Aliases: config.Aliases,
		Short:   config.Short,
		Long:    config.Long,
		Hidden:  config.Hidden,
	}

	// Set args validation
	cb.setArgs(cmd, config.Args)

	// Set run function
	if config.RunFunc != "" {
		if fn, exists := cb.funcMap[config.RunFunc]; exists {
			if runE, ok := fn.(func(*cobra.Command, []string) error); ok {
				cmd.RunE = runE
			} else {
				return nil, fmt.Errorf("function %s is not of type func(*cobra.Command, []string) error", config.RunFunc)
			}
		} else {
			return nil, fmt.Errorf("function %s not registered", config.RunFunc)
		}
	}

	// Add flags
	if err := cb.addFlags(cmd, config.Flags); err != nil {
		return nil, err
	}

	// Build and add subcommands
	for subName, subConfig := range config.Commands {
		subCmd, err := cb.buildCommand(subName, subConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build subcommand %s: %v", subName, err)
		}
		cmd.AddCommand(subCmd)
	}

	return cmd, nil
}

// setArgs sets argument validation on a command based on ArgsConfig
func (cb *CommandBuilder) setArgs(cmd *cobra.Command, args *ArgsConfig) {
	if args == nil {
		return // default: no validation (any args allowed)
	}

	switch args.Type {
	case ArgsTypeNone:
		cmd.Args = cobra.NoArgs
	case ArgsTypeAny:
		cmd.Args = cobra.ArbitraryArgs
	case ArgsTypeExact:
		cmd.Args = cobra.ExactArgs(args.Count)
	case ArgsTypeMin:
		cmd.Args = cobra.MinimumNArgs(args.Min)
	case ArgsTypeMax:
		cmd.Args = cobra.MaximumNArgs(args.Max)
	case ArgsTypeRange:
		cmd.Args = cobra.RangeArgs(args.Min, args.Max)
	}
}

// addFlags adds flags to a command based on flag configuration
func (cb *CommandBuilder) addFlags(cmd *cobra.Command, flags []FlagConfig) error {
	for _, flag := range flags {
		var flagSet *pflag.FlagSet
		if flag.Persistent {
			flagSet = cmd.PersistentFlags()
		} else {
			flagSet = cmd.Flags()
		}

		switch flag.Type {
		case "string":
			if flag.Shorthand != "" {
				flagSet.StringP(flag.Name, flag.Shorthand, flag.DefaultValue, flag.Usage)
			} else {
				flagSet.String(flag.Name, flag.DefaultValue, flag.Usage)
			}
		case "bool":
			defaultBool := flag.DefaultValue == "true"
			if flag.Shorthand != "" {
				flagSet.BoolP(flag.Name, flag.Shorthand, defaultBool, flag.Usage)
			} else {
				flagSet.Bool(flag.Name, defaultBool, flag.Usage)
			}
		case "int":
			defaultInt := 0
			if flag.DefaultValue != "" {
				if _, err := fmt.Sscanf(flag.DefaultValue, "%d", &defaultInt); err != nil {
					return fmt.Errorf("invalid int default value %q for flag %s: %w", flag.DefaultValue, flag.Name, err)
				}
			}
			if flag.Shorthand != "" {
				flagSet.IntP(flag.Name, flag.Shorthand, defaultInt, flag.Usage)
			} else {
				flagSet.Int(flag.Name, defaultInt, flag.Usage)
			}
		case "stringSlice":
			var defaultSlice []string
			if flag.Shorthand != "" {
				flagSet.StringSliceP(flag.Name, flag.Shorthand, defaultSlice, flag.Usage)
			} else {
				flagSet.StringSlice(flag.Name, defaultSlice, flag.Usage)
			}
		default:
			return fmt.Errorf("unsupported flag type: %s", flag.Type)
		}

		if flag.Required {
			if err := cmd.MarkFlagRequired(flag.Name); err != nil {
				return fmt.Errorf("failed to mark flag %s as required: %w", flag.Name, err)
			}
		}

		if flag.Hidden {
			if err := flagSet.MarkHidden(flag.Name); err != nil {
				return fmt.Errorf("failed to mark flag %s as hidden: %w", flag.Name, err)
			}
		}
	}

	return nil
}

// GetConfig returns the tool configuration
func (cb *CommandBuilder) GetConfig() *ToolConfig {
	return cb.config
}
