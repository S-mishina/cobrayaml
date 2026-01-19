package cobrayaml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerator_GenerateDocs_Basic(t *testing.T) {
	yamlContent := `
name: my-tool
description: A simple CLI tool
version: "1.0.0"
root:
  use: my-tool
  short: My CLI tool
  long: A longer description of my CLI tool
commands:
  list:
    use: list
    short: List items
    run_func: runList
  add:
    use: "add <name>"
    short: Add an item
    run_func: runAdd
    flags:
      - name: force
        shorthand: f
        type: bool
        usage: Force the operation
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	docs, err := gen.GenerateDocs()
	if err != nil {
		t.Fatalf("GenerateDocs() error = %v", err)
	}

	// Check tool name
	if !strings.Contains(docs, "# my-tool") {
		t.Error("docs should contain tool name as heading")
	}

	// Check description
	if !strings.Contains(docs, "A simple CLI tool") {
		t.Error("docs should contain tool description")
	}

	// Check version
	if !strings.Contains(docs, "**Version:** 1.0.0") {
		t.Error("docs should contain version")
	}

	// Check commands section
	if !strings.Contains(docs, "## Commands") {
		t.Error("docs should contain Commands section")
	}

	// Check list command
	if !strings.Contains(docs, "### list") {
		t.Error("docs should contain list command")
	}

	// Check add command
	if !strings.Contains(docs, "### add") {
		t.Error("docs should contain add command")
	}

	// Check flag table
	if !strings.Contains(docs, "`--force`") {
		t.Error("docs should contain force flag")
	}

	if !strings.Contains(docs, "`-f`") {
		t.Error("docs should contain flag shorthand")
	}
}

func TestGenerator_GenerateDocs_NestedSubcommands(t *testing.T) {
	yamlContent := `
name: kubectl
description: Kubernetes CLI
version: "1.25.0"
root:
  use: kubectl
  short: Kubernetes command line tool
commands:
  get:
    use: get
    short: Display resources
    commands:
      pods:
        use: pods
        short: List pods
        run_func: runGetPods
      services:
        use: services
        short: List services
        run_func: runGetServices
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	docs, err := gen.GenerateDocs()
	if err != nil {
		t.Fatalf("GenerateDocs() error = %v", err)
	}

	// Check parent command
	if !strings.Contains(docs, "### get") {
		t.Error("docs should contain get command")
	}

	// Check nested commands with deeper heading
	if !strings.Contains(docs, "#### pods") {
		t.Error("docs should contain pods subcommand with deeper heading")
	}

	if !strings.Contains(docs, "#### services") {
		t.Error("docs should contain services subcommand with deeper heading")
	}

	// Check full path for nested command
	if !strings.Contains(docs, "kubectl get pods") {
		t.Error("docs should contain full command path for nested command")
	}
}

func TestGenerator_GenerateDocs_AllFlagTypes(t *testing.T) {
	yamlContent := `
name: test-tool
root:
  use: test-tool
  short: Test tool
commands:
  run:
    use: run
    short: Run something
    run_func: runRun
    flags:
      - name: config
        type: string
        default: config.yaml
        usage: Path to config file
      - name: verbose
        shorthand: v
        type: bool
        default: "false"
        usage: Enable verbose output
      - name: count
        type: int
        default: "10"
        usage: Number of items
      - name: tags
        type: stringSlice
        usage: List of tags
      - name: required-flag
        type: string
        usage: A required flag
        required: true
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	docs, err := gen.GenerateDocs()
	if err != nil {
		t.Fatalf("GenerateDocs() error = %v", err)
	}

	// Check all flag types appear
	if !strings.Contains(docs, "`--config`") {
		t.Error("docs should contain config flag")
	}

	if !strings.Contains(docs, "`--verbose`") {
		t.Error("docs should contain verbose flag")
	}

	if !strings.Contains(docs, "`--count`") {
		t.Error("docs should contain count flag")
	}

	if !strings.Contains(docs, "`--tags`") {
		t.Error("docs should contain tags flag")
	}

	// Check default value
	if !strings.Contains(docs, "`config.yaml`") {
		t.Error("docs should contain default value for config flag")
	}

	// Check required flag marker
	if !strings.Contains(docs, "**(required)**") {
		t.Error("docs should mark required flags")
	}
}

func TestGenerator_GenerateDocs_ArgsTypes(t *testing.T) {
	tests := []struct {
		name     string
		argsYAML string
		expected string
	}{
		{
			name: "exact args",
			argsYAML: `args:
      type: exact
      count: 2`,
			expected: "Exactly 2 argument(s) required",
		},
		{
			name: "min args",
			argsYAML: `args:
      type: min
      min: 1`,
			expected: "At least 1 argument(s) required",
		},
		{
			name: "max args",
			argsYAML: `args:
      type: max
      max: 5`,
			expected: "At most 5 argument(s) allowed",
		},
		{
			name: "range args",
			argsYAML: `args:
      type: range
      min: 1
      max: 3`,
			expected: "1 to 3 argument(s)",
		},
		{
			name: "no args",
			argsYAML: `args:
      type: none`,
			expected: "No arguments allowed",
		},
		{
			name: "any args",
			argsYAML: `args:
      type: any`,
			expected: "Any number of arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yamlContent := `
name: test
root:
  use: test
commands:
  cmd:
    use: cmd
    short: Test command
    run_func: runCmd
    ` + tt.argsYAML

			gen, err := NewGeneratorFromString(yamlContent)
			if err != nil {
				t.Fatalf("NewGeneratorFromString() error = %v", err)
			}

			docs, err := gen.GenerateDocs()
			if err != nil {
				t.Fatalf("GenerateDocs() error = %v", err)
			}

			if !strings.Contains(docs, tt.expected) {
				t.Errorf("docs should contain %q for %s", tt.expected, tt.name)
			}
		})
	}
}

func TestGenerator_GenerateDocs_HiddenCommands(t *testing.T) {
	yamlContent := `
name: test-tool
root:
  use: test-tool
  short: Test tool
commands:
  visible:
    use: visible
    short: A visible command
    run_func: runVisible
  hidden:
    use: hidden
    short: A hidden command
    run_func: runHidden
    hidden: true
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	docs, err := gen.GenerateDocs()
	if err != nil {
		t.Fatalf("GenerateDocs() error = %v", err)
	}

	// Visible command should appear
	if !strings.Contains(docs, "visible") {
		t.Error("docs should contain visible command")
	}

	// Hidden command should not appear
	if strings.Contains(docs, "### hidden") {
		t.Error("docs should not contain hidden command")
	}
}

func TestGenerator_GenerateDocs_HiddenFlags(t *testing.T) {
	yamlContent := `
name: test-tool
root:
  use: test-tool
  short: Test tool
commands:
  run:
    use: run
    short: Run command
    run_func: runRun
    flags:
      - name: visible-flag
        type: string
        usage: A visible flag
      - name: hidden-flag
        type: string
        usage: A hidden flag
        hidden: true
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	docs, err := gen.GenerateDocs()
	if err != nil {
		t.Fatalf("GenerateDocs() error = %v", err)
	}

	// Visible flag should appear
	if !strings.Contains(docs, "`--visible-flag`") {
		t.Error("docs should contain visible flag")
	}

	// Hidden flag should not appear
	if strings.Contains(docs, "`--hidden-flag`") {
		t.Error("docs should not contain hidden flag")
	}
}

func TestGenerator_GenerateDocs_Aliases(t *testing.T) {
	yamlContent := `
name: test-tool
root:
  use: test-tool
  short: Test tool
commands:
  list:
    use: list
    short: List items
    run_func: runList
    aliases:
      - ls
      - l
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	docs, err := gen.GenerateDocs()
	if err != nil {
		t.Fatalf("GenerateDocs() error = %v", err)
	}

	// Check aliases
	if !strings.Contains(docs, "**Aliases:**") {
		t.Error("docs should contain Aliases section")
	}

	if !strings.Contains(docs, "ls") {
		t.Error("docs should contain ls alias")
	}

	if !strings.Contains(docs, "l") {
		t.Error("docs should contain l alias")
	}
}

func TestGenerator_GenerateDocs_GlobalFlags(t *testing.T) {
	yamlContent := `
name: test-tool
root:
  use: test-tool
  short: Test tool
  flags:
    - name: config
      shorthand: c
      type: string
      default: ~/.config
      usage: Path to config file
      persistent: true
    - name: debug
      type: bool
      usage: Enable debug mode
commands:
  run:
    use: run
    short: Run something
    run_func: runRun
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	docs, err := gen.GenerateDocs()
	if err != nil {
		t.Fatalf("GenerateDocs() error = %v", err)
	}

	// Check global flags section
	if !strings.Contains(docs, "### Global Flags") {
		t.Error("docs should contain Global Flags section")
	}

	if !strings.Contains(docs, "`--config`") {
		t.Error("docs should contain config flag in global flags")
	}

	if !strings.Contains(docs, "`-c`") {
		t.Error("docs should contain shorthand for config flag")
	}
}

func TestGenerator_GenerateDocsToFile(t *testing.T) {
	yamlContent := `
name: test-tool
description: Test tool
root:
  use: test-tool
  short: Test tool
commands:
  hello:
    use: hello
    short: Say hello
    run_func: runHello
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "README.md")

	err = gen.GenerateDocsToFile(outputPath)
	if err != nil {
		t.Fatalf("GenerateDocsToFile() error = %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	if !strings.Contains(string(content), "# test-tool") {
		t.Error("generated file should contain tool name")
	}

	if !strings.Contains(string(content), "### hello") {
		t.Error("generated file should contain hello command")
	}
}

func TestGenerator_GenerateDocs_EmptyCommands(t *testing.T) {
	yamlContent := `
name: minimal-tool
root:
  use: minimal-tool
  short: A minimal tool with no commands
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	docs, err := gen.GenerateDocs()
	if err != nil {
		t.Fatalf("GenerateDocs() error = %v", err)
	}

	// Should still generate valid markdown
	if !strings.Contains(docs, "# minimal-tool") {
		t.Error("docs should contain tool name even with no commands")
	}

	if !strings.Contains(docs, "## Commands") {
		t.Error("docs should contain Commands section even if empty")
	}
}

func TestCollectDocsConfig(t *testing.T) {
	yamlContent := `
name: test-tool
description: Test description
version: "2.0.0"
root:
  use: test-tool
  short: Test tool
  long: A longer description
  flags:
    - name: global
      type: string
commands:
  cmd1:
    use: cmd1
    short: Command 1
  cmd2:
    use: cmd2
    short: Command 2
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	config := gen.collectDocsConfig()

	if config.ToolName != "test-tool" {
		t.Errorf("ToolName = %q, want %q", config.ToolName, "test-tool")
	}

	if config.ToolDescription != "Test description" {
		t.Errorf("ToolDescription = %q, want %q", config.ToolDescription, "Test description")
	}

	if config.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", config.Version, "2.0.0")
	}

	if len(config.Commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(config.Commands))
	}

	if len(config.RootCommand.Flags) != 1 {
		t.Errorf("expected 1 root flag, got %d", len(config.RootCommand.Flags))
	}
}

func TestFilterVisibleFlags(t *testing.T) {
	flags := []FlagConfig{
		{Name: "visible1", Type: "string", Hidden: false},
		{Name: "hidden1", Type: "string", Hidden: true},
		{Name: "visible2", Type: "bool", Hidden: false},
		{Name: "hidden2", Type: "bool", Hidden: true},
	}

	visible := filterVisibleFlags(flags)

	if len(visible) != 2 {
		t.Errorf("expected 2 visible flags, got %d", len(visible))
	}

	for _, f := range visible {
		if f.Hidden {
			t.Errorf("flag %q should not be hidden", f.Name)
		}
	}
}
