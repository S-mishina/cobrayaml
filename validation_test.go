package cobrayaml

import (
	"strings"
	"testing"
)

func TestValidateConfig_ValidConfig(t *testing.T) {
	config := &ToolConfig{
		Name: "test-tool",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"add": {
				Use:   "add <name>",
				Short: "Add something",
				Flags: []FlagConfig{
					{Name: "force", Type: "bool", Usage: "Force the operation"},
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err != nil {
		t.Errorf("ValidateConfig() error = %v, want nil", err)
	}
}

func TestValidateConfig_MissingToolName(t *testing.T) {
	config := &ToolConfig{
		Name: "",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for missing tool name")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "name is required") {
		t.Errorf("error should contain 'name is required', got: %s", ve.Error())
	}
}

func TestValidateConfig_MissingCommandUse(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "",
			Short: "Test command",
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for missing use")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "use is required") {
		t.Errorf("error should contain 'use is required', got: %s", ve.Error())
	}
}

func TestValidateConfig_MissingCommandShort(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "",
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for missing short")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "short description is required") {
		t.Errorf("error should contain 'short description is required', got: %s", ve.Error())
	}
}

func TestValidateConfig_MissingFlagName(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
			Flags: []FlagConfig{
				{Name: "", Type: "string", Usage: "Some flag"},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for missing flag name")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "flag name is required") {
		t.Errorf("error should contain 'flag name is required', got: %s", ve.Error())
	}
}

func TestValidateConfig_MissingFlagType(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
			Flags: []FlagConfig{
				{Name: "output", Type: "", Usage: "Output format"},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for missing flag type")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "type is required") {
		t.Errorf("error should contain 'type is required', got: %s", ve.Error())
	}
}

func TestValidateConfig_MissingFlagUsage(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
			Flags: []FlagConfig{
				{Name: "output", Type: "string", Usage: ""},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for missing flag usage")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "usage is required") {
		t.Errorf("error should contain 'usage is required', got: %s", ve.Error())
	}
}

func TestValidateConfig_DuplicateFlagNameInRootCommand(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
			Flags: []FlagConfig{
				{Name: "config", Type: "string", Usage: "Config file"},
				{Name: "config", Type: "string", Usage: "Config again"},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for duplicate flag name in root command")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "duplicate flag name") {
		t.Errorf("error should contain 'duplicate flag name', got: %s", ve.Error())
	}
}

func TestValidateConfig_DuplicateFlagShorthandInRootCommand(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
			Flags: []FlagConfig{
				{Name: "config", Shorthand: "c", Type: "string", Usage: "Config file"},
				{Name: "count", Shorthand: "c", Type: "int", Usage: "Count"},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for duplicate flag shorthand in root command")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "duplicate flag shorthand") {
		t.Errorf("error should contain 'duplicate flag shorthand', got: %s", ve.Error())
	}
}

func TestValidateConfig_DuplicateFlagName(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"add": {
				Use:   "add",
				Short: "Add something",
				Flags: []FlagConfig{
					{Name: "force", Type: "bool", Usage: "Force operation"},
					{Name: "force", Type: "bool", Usage: "Force again"},
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for duplicate flag name")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "duplicate flag name") {
		t.Errorf("error should contain 'duplicate flag name', got: %s", ve.Error())
	}
}

func TestValidateConfig_DuplicateFlagShorthand(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"add": {
				Use:   "add",
				Short: "Add something",
				Flags: []FlagConfig{
					{Name: "force", Shorthand: "f", Type: "bool", Usage: "Force operation"},
					{Name: "fast", Shorthand: "f", Type: "bool", Usage: "Fast mode"},
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for duplicate flag shorthand")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "duplicate flag shorthand") {
		t.Errorf("error should contain 'duplicate flag shorthand', got: %s", ve.Error())
	}
}

func TestValidateConfig_ArgsExactValid(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"add": {
				Use:   "add <name>",
				Short: "Add something",
				Args: &ArgsConfig{
					Type:  "exact",
					Count: 1,
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err != nil {
		t.Errorf("ValidateConfig() error = %v, want nil", err)
	}
}

func TestValidateConfig_ArgsExactInvalidCount(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"add": {
				Use:   "add",
				Short: "Add something",
				Args: &ArgsConfig{
					Type:  "exact",
					Count: 0,
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for invalid exact count")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "count >= 1") {
		t.Errorf("error should contain 'count >= 1', got: %s", ve.Error())
	}
}

func TestValidateConfig_ArgsMinValid(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"add": {
				Use:   "add",
				Short: "Add something",
				Args: &ArgsConfig{
					Type: "min",
					Min:  0,
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err != nil {
		t.Errorf("ValidateConfig() error = %v, want nil", err)
	}
}

func TestValidateConfig_ArgsMinInvalid(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"add": {
				Use:   "add",
				Short: "Add something",
				Args: &ArgsConfig{
					Type: "min",
					Min:  -1,
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for invalid min value")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "min >= 0") {
		t.Errorf("error should contain 'min >= 0', got: %s", ve.Error())
	}
}

func TestValidateConfig_ArgsMaxValid(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"add": {
				Use:   "add",
				Short: "Add something",
				Args: &ArgsConfig{
					Type: "max",
					Max:  5,
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err != nil {
		t.Errorf("ValidateConfig() error = %v, want nil", err)
	}
}

func TestValidateConfig_ArgsMaxInvalid(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"add": {
				Use:   "add",
				Short: "Add something",
				Args: &ArgsConfig{
					Type: "max",
					Max:  0,
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for invalid max value")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "max >= 1") {
		t.Errorf("error should contain 'max >= 1', got: %s", ve.Error())
	}
}

func TestValidateConfig_ArgsRangeValid(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"add": {
				Use:   "add",
				Short: "Add something",
				Args: &ArgsConfig{
					Type: "range",
					Min:  1,
					Max:  5,
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err != nil {
		t.Errorf("ValidateConfig() error = %v, want nil", err)
	}
}

func TestValidateConfig_ArgsRangeInvalidMin(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"add": {
				Use:   "add",
				Short: "Add something",
				Args: &ArgsConfig{
					Type: "range",
					Min:  -1,
					Max:  5,
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for invalid range min")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "min >= 0") {
		t.Errorf("error should contain 'min >= 0', got: %s", ve.Error())
	}
}

func TestValidateConfig_ArgsRangeInvalidMax(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"add": {
				Use:   "add",
				Short: "Add something",
				Args: &ArgsConfig{
					Type: "range",
					Min:  0,
					Max:  0,
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for invalid range max")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "max >= 1") {
		t.Errorf("error should contain 'max >= 1', got: %s", ve.Error())
	}
}

func TestValidateConfig_ArgsRangeMinGreaterThanMax(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"add": {
				Use:   "add",
				Short: "Add something",
				Args: &ArgsConfig{
					Type: "range",
					Min:  5,
					Max:  3,
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for min > max")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "min <= max") {
		t.Errorf("error should contain 'min <= max', got: %s", ve.Error())
	}
}

func TestValidateConfig_InvalidArgsType(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"add": {
				Use:   "add",
				Short: "Add something",
				Args: &ArgsConfig{
					Type: "invalid_type",
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for invalid args type")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "invalid args type") {
		t.Errorf("error should contain 'invalid args type', got: %s", ve.Error())
	}
}

func TestValidateConfig_MultipleErrors(t *testing.T) {
	config := &ToolConfig{
		Name: "",
		Root: CommandConfig{
			Use:   "",
			Short: "",
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected errors")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if len(ve.Errors) < 3 {
		t.Errorf("expected at least 3 errors, got %d: %s", len(ve.Errors), ve.Error())
	}
}

func TestValidateConfig_DuplicateCommandNameAtRootLevel(t *testing.T) {
	// Test duplicate command names at root level
	// Map keys are different but Use field extracts to same command name
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"add-item": {
				Use:   "add",
				Short: "Add item",
			},
			"add-user": {
				Use:   "add", // Same command name extracted from Use
				Short: "Add user",
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for duplicate command name at root level")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "duplicate command name") {
		t.Errorf("error should contain 'duplicate command name', got: %s", ve.Error())
	}
}

func TestValidateConfig_DuplicateSubcommandName(t *testing.T) {
	// Test duplicate command names in subcommands
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"parent": {
				Use:   "parent",
				Short: "Parent command",
				Commands: map[string]CommandConfig{
					"child1": {
						Use:   "child",
						Short: "Child 1",
					},
					"child2": {
						Use:   "child", // Same command name extracted from Use
						Short: "Child 2",
					},
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for duplicate subcommand name")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "duplicate subcommand name") {
		t.Errorf("error should contain 'duplicate subcommand name', got: %s", ve.Error())
	}
}

func TestValidateConfig_NestedCommands(t *testing.T) {
	config := &ToolConfig{
		Name: "test",
		Root: CommandConfig{
			Use:   "test",
			Short: "Test command",
		},
		Commands: map[string]CommandConfig{
			"parent": {
				Use:   "parent",
				Short: "Parent command",
				Commands: map[string]CommandConfig{
					"child": {
						Use:   "child",
						Short: "", // Missing short
					},
				},
			},
		},
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("ValidateConfig() expected error for nested command missing short")
		return
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}

	if !strings.Contains(ve.Error(), "short description is required") {
		t.Errorf("error should contain 'short description is required', got: %s", ve.Error())
	}
}

func TestValidateConfig_ArgsNoneAndAny(t *testing.T) {
	tests := []struct {
		name     string
		argsType string
	}{
		{"none", ArgsTypeNone},
		{"any", ArgsTypeAny},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ToolConfig{
				Name: "test",
				Root: CommandConfig{
					Use:   "test",
					Short: "Test command",
				},
				Commands: map[string]CommandConfig{
					"cmd": {
						Use:   "cmd",
						Short: "Command",
						Args: &ArgsConfig{
							Type: tt.argsType,
						},
					},
				},
			}

			err := ValidateConfig(config)
			if err != nil {
				t.Errorf("ValidateConfig() error = %v, want nil", err)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{
		Errors: []string{
			"error 1",
			"error 2",
		},
	}

	errStr := ve.Error()
	if !strings.Contains(errStr, "2 error(s)") {
		t.Errorf("error string should contain '2 error(s)', got: %s", errStr)
	}
	if !strings.Contains(errStr, "error 1") {
		t.Errorf("error string should contain 'error 1', got: %s", errStr)
	}
	if !strings.Contains(errStr, "error 2") {
		t.Errorf("error string should contain 'error 2', got: %s", errStr)
	}
}

func TestValidationError_EmptyErrors(t *testing.T) {
	ve := &ValidationError{}
	if ve.Error() != "" {
		t.Errorf("empty ValidationError should return empty string, got: %s", ve.Error())
	}
}

func TestExtractCommandName(t *testing.T) {
	tests := []struct {
		use  string
		want string
	}{
		{"add", "add"},
		{"add <name>", "add"},
		{"delete <id> <reason>", "delete"},
		{"", ""},
		{"  spaced  ", "spaced"},
	}

	for _, tt := range tests {
		t.Run(tt.use, func(t *testing.T) {
			got := extractCommandName(tt.use)
			if got != tt.want {
				t.Errorf("extractCommandName(%q) = %q, want %q", tt.use, got, tt.want)
			}
		})
	}
}
