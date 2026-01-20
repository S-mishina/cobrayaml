package cobrayaml

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCommandBuilderFromString(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantName    string
		wantDesc    string
		wantErr     bool
	}{
		{
			name: "valid yaml",
			yamlContent: `
name: test-tool
description: A test tool
root:
  use: test
  short: Test command
  long: A test command for testing
commands:
  sub:
    use: sub
    short: Sub command
`,
			wantName: "test-tool",
			wantDesc: "A test tool",
			wantErr:  false,
		},
		{
			name: "minimal yaml",
			yamlContent: `
name: minimal
description: Minimal config
root:
  use: minimal
  short: Minimal command
`,
			wantName: "minimal",
			wantDesc: "Minimal config",
			wantErr:  false,
		},
		{
			name:        "invalid yaml",
			yamlContent: `invalid: yaml: content: [`,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb, err := NewCommandBuilderFromString(tt.yamlContent)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCommandBuilderFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			config := cb.GetConfig()
			if config.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", config.Name, tt.wantName)
			}
			if config.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", config.Description, tt.wantDesc)
			}
		})
	}
}

func TestNewCommandBuilder(t *testing.T) {
	// Create a temporary YAML file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "commands.yaml")

	yamlContent := `
name: file-test
description: Test from file
root:
  use: file-test
  short: File test command
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	cb, err := NewCommandBuilder(configPath)
	if err != nil {
		t.Fatalf("NewCommandBuilder() error = %v", err)
	}

	config := cb.GetConfig()
	if config.Name != "file-test" {
		t.Errorf("Name = %q, want %q", config.Name, "file-test")
	}
}

func TestNewCommandBuilder_FileNotFound(t *testing.T) {
	_, err := NewCommandBuilder("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestCommandBuilder_RegisterFunction(t *testing.T) {
	yamlContent := `
name: test
description: test
root:
  use: test
  short: Test command
commands:
  run:
    use: run
    short: Run command
    run_func: myFunc
`
	cb, err := NewCommandBuilderFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewCommandBuilderFromString() error = %v", err)
	}

	called := false
	cb.RegisterFunction("myFunc", func(cmd *cobra.Command, args []string) error {
		called = true
		return nil
	})

	rootCmd, err := cb.BuildRootCommand()
	if err != nil {
		t.Fatalf("BuildRootCommand() error = %v", err)
	}

	// Execute the run subcommand via root
	rootCmd.SetArgs([]string{"run"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !called {
		t.Error("registered function was not called")
	}
}

func TestCommandBuilder_BuildRootCommand(t *testing.T) {
	yamlContent := `
name: build-test
description: Build test
root:
  use: build-test
  short: Build test short
  long: Build test long description
commands:
  add:
    use: add <name>
    short: Add something
    args:
      type: exact
      count: 1
    flags:
      - name: force
        shorthand: f
        type: bool
        default: "false"
        usage: Force the operation
  list:
    use: list
    short: List all items
    flags:
      - name: output
        shorthand: o
        type: string
        default: "table"
        usage: Output format
      - name: limit
        type: int
        default: "10"
        usage: Limit results
`
	cb, err := NewCommandBuilderFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewCommandBuilderFromString() error = %v", err)
	}

	rootCmd, err := cb.BuildRootCommand()
	if err != nil {
		t.Fatalf("BuildRootCommand() error = %v", err)
	}

	if rootCmd.Use != "build-test" {
		t.Errorf("Use = %q, want %q", rootCmd.Use, "build-test")
	}

	commands := rootCmd.Commands()
	if len(commands) != 2 {
		t.Errorf("expected 2 subcommands, got %d", len(commands))
	}

	// Check add command
	var addCmd, listCmd *cobra.Command
	for _, cmd := range commands {
		switch cmd.Use {
		case "add <name>":
			addCmd = cmd
		case "list":
			listCmd = cmd
		}
	}

	if addCmd == nil {
		t.Error("add command not found")
	} else {
		forceFlag := addCmd.Flags().Lookup("force")
		if forceFlag == nil {
			t.Error("force flag not found on add command")
		}
	}

	if listCmd == nil {
		t.Error("list command not found")
	} else {
		outputFlag := listCmd.Flags().Lookup("output")
		if outputFlag == nil {
			t.Error("output flag not found on list command")
		}
		limitFlag := listCmd.Flags().Lookup("limit")
		if limitFlag == nil {
			t.Error("limit flag not found on list command")
		}
	}
}

func TestCommandBuilder_ArgsValidation(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		testArgs    []string
		wantErr     bool
	}{
		{
			name: "none with no args",
			yamlContent: `
name: args-test
description: Args test
root:
  use: args-test
  short: Args test command
commands:
  test:
    use: test
    short: Test command
    args:
      type: none
`,
			testArgs: []string{"test"},
			wantErr:  false,
		},
		{
			name: "none with args should fail",
			yamlContent: `
name: args-test
description: Args test
root:
  use: args-test
  short: Args test command
commands:
  test:
    use: test
    short: Test command
    args:
      type: none
`,
			testArgs: []string{"test", "arg1"},
			wantErr:  true,
		},
		{
			name: "exact count 1 with 1 arg",
			yamlContent: `
name: args-test
description: Args test
root:
  use: args-test
  short: Args test command
commands:
  test:
    use: test
    short: Test command
    args:
      type: exact
      count: 1
`,
			testArgs: []string{"test", "arg1"},
			wantErr:  false,
		},
		{
			name: "exact count 1 with 0 args should fail",
			yamlContent: `
name: args-test
description: Args test
root:
  use: args-test
  short: Args test command
commands:
  test:
    use: test
    short: Test command
    args:
      type: exact
      count: 1
`,
			testArgs: []string{"test"},
			wantErr:  true,
		},
		{
			name: "min 1 with 1 arg",
			yamlContent: `
name: args-test
description: Args test
root:
  use: args-test
  short: Args test command
commands:
  test:
    use: test
    short: Test command
    args:
      type: min
      min: 1
`,
			testArgs: []string{"test", "arg1"},
			wantErr:  false,
		},
		{
			name: "min 1 with 2 args",
			yamlContent: `
name: args-test
description: Args test
root:
  use: args-test
  short: Args test command
commands:
  test:
    use: test
    short: Test command
    args:
      type: min
      min: 1
`,
			testArgs: []string{"test", "arg1", "arg2"},
			wantErr:  false,
		},
		{
			name: "max 2 with 2 args",
			yamlContent: `
name: args-test
description: Args test
root:
  use: args-test
  short: Args test command
commands:
  test:
    use: test
    short: Test command
    args:
      type: max
      max: 2
`,
			testArgs: []string{"test", "arg1", "arg2"},
			wantErr:  false,
		},
		{
			name: "max 2 with 3 args should fail",
			yamlContent: `
name: args-test
description: Args test
root:
  use: args-test
  short: Args test command
commands:
  test:
    use: test
    short: Test command
    args:
      type: max
      max: 2
`,
			testArgs: []string{"test", "arg1", "arg2", "arg3"},
			wantErr:  true,
		},
		{
			name: "range 1-3 with 2 args",
			yamlContent: `
name: args-test
description: Args test
root:
  use: args-test
  short: Args test command
commands:
  test:
    use: test
    short: Test command
    args:
      type: range
      min: 1
      max: 3
`,
			testArgs: []string{"test", "arg1", "arg2"},
			wantErr:  false,
		},
		{
			name: "range 1-3 with 0 args should fail",
			yamlContent: `
name: args-test
description: Args test
root:
  use: args-test
  short: Args test command
commands:
  test:
    use: test
    short: Test command
    args:
      type: range
      min: 1
      max: 3
`,
			testArgs: []string{"test"},
			wantErr:  true,
		},
		{
			name: "range 1-3 with 4 args should fail",
			yamlContent: `
name: args-test
description: Args test
root:
  use: args-test
  short: Args test command
commands:
  test:
    use: test
    short: Test command
    args:
      type: range
      min: 1
      max: 3
`,
			testArgs: []string{"test", "arg1", "arg2", "arg3", "arg4"},
			wantErr:  true,
		},
		{
			name: "any with multiple args",
			yamlContent: `
name: args-test
description: Args test
root:
  use: args-test
  short: Args test command
commands:
  test:
    use: test
    short: Test command
    args:
      type: any
`,
			testArgs: []string{"test", "arg1", "arg2", "arg3"},
			wantErr:  false,
		},
		{
			name: "no args config allows any",
			yamlContent: `
name: args-test
description: Args test
root:
  use: args-test
  short: Args test command
commands:
  test:
    use: test
    short: Test command
`,
			testArgs: []string{"test", "arg1", "arg2"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb, err := NewCommandBuilderFromString(tt.yamlContent)
			if err != nil {
				t.Fatalf("NewCommandBuilderFromString() error = %v", err)
			}

			rootCmd, err := cb.BuildRootCommand()
			if err != nil {
				t.Fatalf("BuildRootCommand() error = %v", err)
			}

			// Find and set run function for test command
			for _, cmd := range rootCmd.Commands() {
				if cmd.Use == "test" {
					cmd.Run = func(cmd *cobra.Command, args []string) {}
					break
				}
			}

			rootCmd.SetArgs(tt.testArgs)
			err = rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandBuilder_FlagTypes(t *testing.T) {
	yamlContent := `
name: flag-test
description: Flag test
root:
  use: flag-test
  short: Flag test command
commands:
  test:
    use: test
    short: Test command
    flags:
      - name: str
        type: string
        default: "default-value"
        usage: String flag
      - name: num
        type: int
        default: "42"
        usage: Int flag
      - name: flag
        type: bool
        default: "true"
        usage: Bool flag
      - name: short-flag
        shorthand: s
        type: string
        usage: Flag with shorthand
`
	cb, err := NewCommandBuilderFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewCommandBuilderFromString() error = %v", err)
	}

	rootCmd, err := cb.BuildRootCommand()
	if err != nil {
		t.Fatalf("BuildRootCommand() error = %v", err)
	}

	var testCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "test" {
			testCmd = cmd
			break
		}
	}

	if testCmd == nil {
		t.Fatal("test command not found")
	}

	// Check string flag
	strVal, err := testCmd.Flags().GetString("str")
	if err != nil {
		t.Errorf("GetString(str) error = %v", err)
	}
	if strVal != "default-value" {
		t.Errorf("str default = %q, want %q", strVal, "default-value")
	}

	// Check int flag
	intVal, err := testCmd.Flags().GetInt("num")
	if err != nil {
		t.Errorf("GetInt(num) error = %v", err)
	}
	if intVal != 42 {
		t.Errorf("num default = %d, want %d", intVal, 42)
	}

	// Check bool flag
	boolVal, err := testCmd.Flags().GetBool("flag")
	if err != nil {
		t.Errorf("GetBool(flag) error = %v", err)
	}
	if boolVal != true {
		t.Errorf("flag default = %v, want %v", boolVal, true)
	}

	// Check shorthand flag
	shortFlag := testCmd.Flags().Lookup("short-flag")
	if shortFlag == nil {
		t.Error("short-flag not found")
	} else if shortFlag.Shorthand != "s" {
		t.Errorf("shorthand = %q, want %q", shortFlag.Shorthand, "s")
	}
}

func TestCommandBuilder_UnsupportedFlagType(t *testing.T) {
	yamlContent := `
name: unsupported-test
description: Test
root:
  use: test
  short: Test command
commands:
  test:
    use: test
    short: Test
    flags:
      - name: bad
        type: unsupported_type
        usage: Bad flag
`
	cb, err := NewCommandBuilderFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewCommandBuilderFromString() error = %v", err)
	}

	_, err = cb.BuildRootCommand()
	if err == nil {
		t.Error("expected error for unsupported flag type")
	}
}

func TestCommandBuilder_UnregisteredFunction(t *testing.T) {
	yamlContent := `
name: unregistered-test
description: Test
root:
  use: test
  short: Test command
commands:
  test:
    use: test
    short: Test
    run_func: nonexistent
`
	cb, err := NewCommandBuilderFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewCommandBuilderFromString() error = %v", err)
	}

	_, err = cb.BuildRootCommand()
	if err == nil {
		t.Error("expected error for unregistered function")
	}
}

func TestCommandBuilder_PersistentFlags(t *testing.T) {
	yamlContent := `
name: persistent-test
description: Test
root:
  use: test
  short: Test command
  flags:
    - name: global
      type: string
      persistent: true
      usage: Global flag
commands:
  sub:
    use: sub
    short: Sub command
`
	cb, err := NewCommandBuilderFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewCommandBuilderFromString() error = %v", err)
	}

	rootCmd, err := cb.BuildRootCommand()
	if err != nil {
		t.Fatalf("BuildRootCommand() error = %v", err)
	}

	// Check persistent flag exists on root
	globalFlag := rootCmd.PersistentFlags().Lookup("global")
	if globalFlag == nil {
		t.Error("global persistent flag not found")
	}
}

func TestCommandBuilder_Aliases(t *testing.T) {
	yamlContent := `
name: alias-test
description: Test aliases
root:
  use: alias-test
  short: Alias test command
commands:
  get:
    use: get
    short: Get resources
    aliases:
      - g
      - list
`
	cb, err := NewCommandBuilderFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewCommandBuilderFromString() error = %v", err)
	}

	rootCmd, err := cb.BuildRootCommand()
	if err != nil {
		t.Fatalf("BuildRootCommand() error = %v", err)
	}

	// Find get command
	var getCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "get" {
			getCmd = cmd
			break
		}
	}

	if getCmd == nil {
		t.Fatal("get command not found")
	}

	// Check aliases
	expectedAliases := []string{"g", "list"}
	if len(getCmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases count = %d, want %d", len(getCmd.Aliases), len(expectedAliases))
	}

	for i, alias := range expectedAliases {
		if i < len(getCmd.Aliases) && getCmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, getCmd.Aliases[i], alias)
		}
	}
}

func TestCommandBuilder_AliasesExecution(t *testing.T) {
	yamlContent := `
name: alias-exec-test
description: Test alias execution
root:
  use: alias-exec-test
  short: Alias execution test command
commands:
  get:
    use: get
    short: Get resources
    aliases:
      - g
    run_func: runGet
`
	cb, err := NewCommandBuilderFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewCommandBuilderFromString() error = %v", err)
	}

	called := false
	cb.RegisterFunction("runGet", func(cmd *cobra.Command, args []string) error {
		called = true
		return nil
	})

	rootCmd, err := cb.BuildRootCommand()
	if err != nil {
		t.Fatalf("BuildRootCommand() error = %v", err)
	}

	// Execute using alias
	rootCmd.SetArgs([]string{"g"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() with alias error = %v", err)
	}

	if !called {
		t.Error("function was not called when using alias")
	}
}

func TestCommandBuilder_Version(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantVersion string
	}{
		{
			name: "with version",
			yamlContent: `
name: version-test
description: Test version
version: "1.2.3"
root:
  use: version-test
  short: Version test
`,
			wantVersion: "1.2.3",
		},
		{
			name: "without version",
			yamlContent: `
name: no-version-test
description: Test without version
root:
  use: no-version-test
  short: No version test
`,
			wantVersion: "",
		},
		{
			name: "semantic version",
			yamlContent: `
name: semver-test
description: Test semantic version
version: "v2.0.0-beta.1"
root:
  use: semver-test
  short: Semver test
`,
			wantVersion: "v2.0.0-beta.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb, err := NewCommandBuilderFromString(tt.yamlContent)
			if err != nil {
				t.Fatalf("NewCommandBuilderFromString() error = %v", err)
			}

			rootCmd, err := cb.BuildRootCommand()
			if err != nil {
				t.Fatalf("BuildRootCommand() error = %v", err)
			}

			if rootCmd.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", rootCmd.Version, tt.wantVersion)
			}
		})
	}
}

func TestCommandBuilder_HiddenCommand(t *testing.T) {
	yamlContent := `
name: hidden-cmd-test
description: Test hidden command
root:
  use: hidden-cmd-test
  short: Hidden command test
commands:
  visible:
    use: visible
    short: Visible command
  internal:
    use: internal
    short: Internal command (hidden)
    hidden: true
`
	cb, err := NewCommandBuilderFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewCommandBuilderFromString() error = %v", err)
	}

	rootCmd, err := cb.BuildRootCommand()
	if err != nil {
		t.Fatalf("BuildRootCommand() error = %v", err)
	}

	var visibleCmd, internalCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		switch cmd.Use {
		case "visible":
			visibleCmd = cmd
		case "internal":
			internalCmd = cmd
		}
	}

	if visibleCmd == nil {
		t.Fatal("visible command not found")
	}
	if internalCmd == nil {
		t.Fatal("internal command not found")
	}

	if visibleCmd.Hidden {
		t.Error("visible command should not be hidden")
	}
	if !internalCmd.Hidden {
		t.Error("internal command should be hidden")
	}
}

func TestCommandBuilder_HiddenFlag(t *testing.T) {
	yamlContent := `
name: hidden-flag-test
description: Test hidden flag
root:
  use: hidden-flag-test
  short: Hidden flag test
  flags:
    - name: "visible-flag"
      type: "string"
      usage: "A visible flag"
    - name: "hidden-flag"
      type: "string"
      usage: "A hidden flag"
      hidden: true
`
	cb, err := NewCommandBuilderFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewCommandBuilderFromString() error = %v", err)
	}

	rootCmd, err := cb.BuildRootCommand()
	if err != nil {
		t.Fatalf("BuildRootCommand() error = %v", err)
	}

	visibleFlag := rootCmd.Flags().Lookup("visible-flag")
	hiddenFlag := rootCmd.Flags().Lookup("hidden-flag")

	if visibleFlag == nil {
		t.Fatal("visible-flag not found")
	}
	if hiddenFlag == nil {
		t.Fatal("hidden-flag not found")
	}

	if visibleFlag.Hidden {
		t.Error("visible-flag should not be hidden")
	}
	if !hiddenFlag.Hidden {
		t.Error("hidden-flag should be hidden")
	}
}

func TestCommandBuilder_HiddenPersistentFlag(t *testing.T) {
	yamlContent := `
name: hidden-persistent-flag-test
description: Test hidden persistent flag
root:
  use: hidden-persistent-flag-test
  short: Hidden persistent flag test
  flags:
    - name: "debug"
      type: "bool"
      usage: "Enable debug mode"
      persistent: true
      hidden: true
`
	cb, err := NewCommandBuilderFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewCommandBuilderFromString() error = %v", err)
	}

	rootCmd, err := cb.BuildRootCommand()
	if err != nil {
		t.Fatalf("BuildRootCommand() error = %v", err)
	}

	debugFlag := rootCmd.PersistentFlags().Lookup("debug")

	if debugFlag == nil {
		t.Fatal("debug flag not found")
	}

	if !debugFlag.Hidden {
		t.Error("debug flag should be hidden")
	}
}

// TestExampleCommandsYAML ensures the example YAML used in documentation is valid.
func TestExampleCommandsYAML(t *testing.T) {
	cb, err := NewCommandBuilderFromString(ExampleCommandsYAML)
	if err != nil {
		t.Fatalf("ExampleCommandsYAML is invalid: %v", err)
	}

	// Register the functions referenced in the example
	cb.RegisterFunction("runList", func(cmd *cobra.Command, args []string) error {
		return nil
	})
	cb.RegisterFunction("runAdd", func(cmd *cobra.Command, args []string) error {
		return nil
	})
	cb.RegisterFunction("runDelete", func(cmd *cobra.Command, args []string) error {
		return nil
	})

	rootCmd, err := cb.BuildRootCommand()
	if err != nil {
		t.Fatalf("Failed to build root command from ExampleCommandsYAML: %v", err)
	}

	// Verify basic structure
	if rootCmd.Use != "my-tool" {
		t.Errorf("Use = %q, want %q", rootCmd.Use, "my-tool")
	}
	if rootCmd.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", rootCmd.Version, "1.0.0")
	}

	// Verify subcommands
	commands := rootCmd.Commands()
	if len(commands) != 3 {
		t.Errorf("expected 3 subcommands, got %d", len(commands))
	}

	// Verify persistent flag
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Error("config persistent flag not found")
	}

	// Test list command execution
	rootCmd.SetArgs([]string{"list"})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("list command execution failed: %v", err)
	}

	// Test add command execution
	rootCmd.SetArgs([]string{"add", "test-item"})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("add command execution failed: %v", err)
	}

	// Test delete command execution
	rootCmd.SetArgs([]string{"delete", "test-item"})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("delete command execution failed: %v", err)
	}

	// Test delete command with alias
	rootCmd.SetArgs([]string{"rm", "test-item"})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("delete command (via alias 'rm') execution failed: %v", err)
	}
}
