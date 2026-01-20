package cobrayaml

import (
	"fmt"
	"slices"
	"strings"
)

// ValidationError represents multiple validation errors collected during config validation.
type ValidationError struct {
	Errors []string
}

// Error returns the formatted error message with all validation errors.
func (e *ValidationError) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "validation failed with %d error(s):\n", len(e.Errors))
	for _, err := range e.Errors {
		sb.WriteString("  - ")
		sb.WriteString(err)
		sb.WriteString("\n")
	}
	return sb.String()
}

// addError adds a new error to the ValidationError.
func (e *ValidationError) addError(format string, args ...any) {
	e.Errors = append(e.Errors, fmt.Sprintf(format, args...))
}

// hasErrors returns true if there are any validation errors.
func (e *ValidationError) hasErrors() bool {
	return len(e.Errors) > 0
}

// ValidateConfig validates the entire ToolConfig and returns an error if validation fails.
// It collects all validation errors and returns them together.
func ValidateConfig(config *ToolConfig) error {
	ve := &ValidationError{}

	// Validate ToolConfig required fields
	validateToolConfig(config, ve)

	// Validate root command
	validateCommandConfig(&config.Root, "root", ve)

	// Validate root command flags
	validateFlags(config.Root.Flags, "root", ve)
	validateFlagDuplicates(config.Root.Flags, "root", ve)

	// Collect all command names at root level for duplicate check
	commandNames := make(map[string]bool)

	// Validate all top-level commands
	for name, cmdConfig := range config.Commands {
		cmdName := extractCommandName(cmdConfig.Use)
		if cmdName == "" {
			cmdName = name
		}

		// Check for duplicate command names
		if commandNames[cmdName] {
			ve.addError("duplicate command name %q at root level", cmdName)
		}
		commandNames[cmdName] = true

		// Validate this command and its subcommands recursively
		validateCommandRecursive(&cmdConfig, name, ve)
	}

	if ve.hasErrors() {
		return ve
	}
	return nil
}

// validateToolConfig validates the ToolConfig required fields.
func validateToolConfig(config *ToolConfig, ve *ValidationError) {
	if config.Name == "" {
		ve.addError("tool config: name is required")
	}
}

// validateCommandConfig validates a CommandConfig's required fields.
func validateCommandConfig(config *CommandConfig, path string, ve *ValidationError) {
	if config.Use == "" {
		ve.addError("command %q: use is required", path)
	}
	if config.Short == "" {
		ve.addError("command %q: short description is required", path)
	}

	// Validate args config
	validateArgsConfig(config.Args, path, ve)
}

// validateCommandRecursive validates a command and all its subcommands recursively.
func validateCommandRecursive(config *CommandConfig, path string, ve *ValidationError) {
	// Validate command required fields
	validateCommandConfig(config, path, ve)

	// Validate flags
	validateFlags(config.Flags, path, ve)

	// Validate flag duplicates within this command
	validateFlagDuplicates(config.Flags, path, ve)

	// Collect subcommand names for duplicate check
	subCommandNames := make(map[string]bool)

	// Validate subcommands recursively
	for name, subConfig := range config.Commands {
		subPath := path + "/" + name

		cmdName := extractCommandName(subConfig.Use)
		if cmdName == "" {
			cmdName = name
		}

		// Check for duplicate command names at this level
		if subCommandNames[cmdName] {
			ve.addError("command %q: duplicate subcommand name %q", path, cmdName)
		}
		subCommandNames[cmdName] = true

		validateCommandRecursive(&subConfig, subPath, ve)
	}
}

// validateFlags validates each flag's required fields.
func validateFlags(flags []FlagConfig, cmdPath string, ve *ValidationError) {
	for _, flag := range flags {
		if flag.Name == "" {
			ve.addError("command %q: flag name is required", cmdPath)
		}
		if flag.Type == "" {
			if flag.Name != "" {
				ve.addError("command %q, flag %q: type is required", cmdPath, flag.Name)
			} else {
				ve.addError("command %q: flag type is required", cmdPath)
			}
		}
		if flag.Usage == "" {
			if flag.Name != "" {
				ve.addError("command %q, flag %q: usage is required", cmdPath, flag.Name)
			} else {
				ve.addError("command %q: flag usage is required", cmdPath)
			}
		}
	}
}

// validateFlagDuplicates checks for duplicate flag names and shorthands within a command.
func validateFlagDuplicates(flags []FlagConfig, cmdPath string, ve *ValidationError) {
	names := make(map[string]bool)
	shorthands := make(map[string]bool)

	for _, flag := range flags {
		if flag.Name != "" {
			if names[flag.Name] {
				ve.addError("command %q: duplicate flag name %q", cmdPath, flag.Name)
			}
			names[flag.Name] = true
		}

		if flag.Shorthand != "" {
			if shorthands[flag.Shorthand] {
				ve.addError("command %q: duplicate flag shorthand %q", cmdPath, flag.Shorthand)
			}
			shorthands[flag.Shorthand] = true
		}
	}
}

// validateArgsConfig validates the ArgsConfig for consistency.
func validateArgsConfig(args *ArgsConfig, cmdPath string, ve *ValidationError) {
	if args == nil {
		return
	}

	// Validate type is supported
	if args.Type != "" && !slices.Contains(SupportedArgsTypes, args.Type) {
		ve.addError("command %q: invalid args type %q (must be one of: %s)",
			cmdPath, args.Type, strings.Join(SupportedArgsTypes, ", "))
	}

	// Validate type-specific constraints
	switch args.Type {
	case ArgsTypeExact:
		if args.Count < 1 {
			ve.addError("command %q: args type 'exact' requires count >= 1", cmdPath)
		}
	case ArgsTypeMin:
		if args.Min < 0 {
			ve.addError("command %q: args type 'min' requires min >= 0", cmdPath)
		}
	case ArgsTypeMax:
		if args.Max < 1 {
			ve.addError("command %q: args type 'max' requires max >= 1", cmdPath)
		}
	case ArgsTypeRange:
		if args.Min < 0 {
			ve.addError("command %q: args type 'range' requires min >= 0", cmdPath)
		}
		if args.Max < 1 {
			ve.addError("command %q: args type 'range' requires max >= 1", cmdPath)
		}
		if args.Min > args.Max {
			ve.addError("command %q: args type 'range' requires min <= max (got min=%d, max=%d)",
				cmdPath, args.Min, args.Max)
		}
	}
}

// extractCommandName extracts the command name from the "use" field.
// For example, "add <name>" returns "add".
func extractCommandName(use string) string {
	if use == "" {
		return ""
	}
	parts := strings.Fields(use)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
