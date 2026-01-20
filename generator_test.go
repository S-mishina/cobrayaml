package cobrayaml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGeneratorFromString(t *testing.T) {
	yamlContent := `
name: test-tool
description: A test tool
root:
  use: test
  short: Test command
commands:
  add:
    use: add
    short: Add something
    run_func: runAdd
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	if gen.config.Name != "test-tool" {
		t.Errorf("Name = %q, want %q", gen.config.Name, "test-tool")
	}
}

func TestNewGenerator(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "commands.yaml")

	yamlContent := `
name: file-test
description: Test from file
root:
  use: file-test
commands:
  list:
    use: list
    short: List items
    run_func: runList
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	gen, err := NewGenerator(configPath)
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	if gen.config.Name != "file-test" {
		t.Errorf("Name = %q, want %q", gen.config.Name, "file-test")
	}
}

func TestGenerator_CollectFunctions(t *testing.T) {
	yamlContent := `
name: test
description: test
root:
  use: test
  run_func: runRoot
commands:
  add:
    use: add <name>
    short: Add item
    run_func: runAdd
    flags:
      - name: force
        type: bool
      - name: output
        type: string
    args:
      type: exact
      count: 1
  list:
    use: list
    short: List items
    run_func: runList
  nested:
    use: nested
    short: Nested command
    commands:
      deep:
        use: deep
        short: Deep command
        run_func: runDeep
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	funcs := gen.CollectFunctions()

	// Should collect: runRoot, runAdd, runList, runDeep
	if len(funcs) != 4 {
		t.Errorf("expected 4 functions, got %d", len(funcs))
	}

	// Check that runAdd has the right flags
	var addFunc *FuncInfo
	for i := range funcs {
		if funcs[i].Name == "runAdd" {
			addFunc = &funcs[i]
			break
		}
	}

	if addFunc == nil {
		t.Fatal("runAdd function not found")
	}

	if len(addFunc.Flags) != 2 {
		t.Errorf("runAdd should have 2 flags, got %d", len(addFunc.Flags))
	}

	if addFunc.Args == nil || addFunc.Args.Type != "exact" || addFunc.Args.Count != 1 {
		t.Errorf("runAdd args not correctly captured")
	}
}

func TestGenerator_GenerateHandlers(t *testing.T) {
	yamlContent := `
name: test
description: test
root:
  use: test
commands:
  add:
    use: add <name>
    short: Add item
    run_func: runAdd
    flags:
      - name: force
        shorthand: f
        type: bool
      - name: output-format
        type: string
      - name: count
        type: int
      - name: tags
        type: stringSlice
    args:
      type: exact
      count: 2
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	code, err := gen.GenerateHandlers("main")
	if err != nil {
		t.Fatalf("GenerateHandlers() error = %v", err)
	}

	// Check package declaration
	if !strings.Contains(code, "package main") {
		t.Error("generated code should contain 'package main'")
	}

	// Check function signature
	if !strings.Contains(code, "func runAdd(cmd *cobra.Command, args []string) error") {
		t.Error("generated code should contain runAdd function")
	}

	// Check flag getters
	if !strings.Contains(code, `cmd.Flags().GetBool("force")`) {
		t.Error("generated code should contain GetBool for force flag")
	}

	if !strings.Contains(code, `cmd.Flags().GetString("output-format")`) {
		t.Error("generated code should contain GetString for output-format flag")
	}

	if !strings.Contains(code, `cmd.Flags().GetInt("count")`) {
		t.Error("generated code should contain GetInt for count flag")
	}

	if !strings.Contains(code, `cmd.Flags().GetStringSlice("tags")`) {
		t.Error("generated code should contain GetStringSlice for tags flag")
	}

	// Check args extraction
	if !strings.Contains(code, "arg0 := args[0]") {
		t.Error("generated code should extract arg0")
	}
	if !strings.Contains(code, "arg1 := args[1]") {
		t.Error("generated code should extract arg1")
	}

	// Check camelCase conversion for hyphenated flag name
	if !strings.Contains(code, "outputFormat") {
		t.Error("generated code should convert output-format to outputFormat")
	}
}

func TestGenerator_GenerateHandlers_NoFunctions(t *testing.T) {
	yamlContent := `
name: test
description: test
root:
  use: test
commands:
  add:
    use: add
    short: Add item
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	_, err = gen.GenerateHandlers("main")
	if err == nil {
		t.Error("expected error when no run_func is defined")
	}
}

func TestGenerator_GenerateHandlersToFile(t *testing.T) {
	yamlContent := `
name: test
description: test
root:
  use: test
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
	outputPath := filepath.Join(tmpDir, "handlers.go")

	err = gen.GenerateHandlersToFile("main", outputPath)
	if err != nil {
		t.Fatalf("GenerateHandlersToFile() error = %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	if !strings.Contains(string(content), "func runHello") {
		t.Error("generated file should contain runHello function")
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"name", "name"},
		{"output-format", "outputFormat"},
		{"my_flag", "myFlag"},
		{"force", "force"},
		{"some-long-flag-name", "someLongFlagName"},
		{"UPPER", "uPPER"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIterate(t *testing.T) {
	tests := []struct {
		n        int
		expected []int
	}{
		{0, []int{}},
		{1, []int{0}},
		{3, []int{0, 1, 2}},
		{5, []int{0, 1, 2, 3, 4}},
	}

	for _, tt := range tests {
		result := iterate(tt.n)
		if len(result) != len(tt.expected) {
			t.Errorf("iterate(%d) length = %d, want %d", tt.n, len(result), len(tt.expected))
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("iterate(%d)[%d] = %d, want %d", tt.n, i, v, tt.expected[i])
			}
		}
	}
}

func TestGenerator_ArgsTypes(t *testing.T) {
	tests := []struct {
		name     string
		argsYAML string
		contains string
	}{
		{
			name: "min args",
			argsYAML: `args:
      type: min
      min: 2`,
			contains: "at least 2",
		},
		{
			name: "range args",
			argsYAML: `args:
      type: range
      min: 1
      max: 3`,
			contains: "1 to 3",
		},
		{
			name: "any args",
			argsYAML: `args:
      type: any`,
			contains: "any number of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yamlContent := `
name: test
description: test
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

			code, err := gen.GenerateHandlers("main")
			if err != nil {
				t.Fatalf("GenerateHandlers() error = %v", err)
			}

			if !strings.Contains(code, tt.contains) {
				t.Errorf("generated code should contain %q for %s", tt.contains, tt.name)
			}
		})
	}
}

func TestGenerator_GenerateMain(t *testing.T) {
	yamlContent := `
name: test
description: test
root:
  use: test
commands:
  hello:
    use: hello
    short: Say hello
    run_func: runHello
  goodbye:
    use: goodbye
    short: Say goodbye
    run_func: runGoodbye
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	code, err := gen.GenerateMain("main", "commands.yaml")
	if err != nil {
		t.Fatalf("GenerateMain() error = %v", err)
	}

	// Check package declaration
	if !strings.Contains(code, "package main") {
		t.Error("generated code should contain 'package main'")
	}

	// Check imports
	if !strings.Contains(code, `"github.com/S-mishina/cobrayaml"`) {
		t.Error("generated code should import cobrayaml")
	}

	// Check config path
	if !strings.Contains(code, `NewCommandBuilder("commands.yaml")`) {
		t.Error("generated code should contain config path")
	}

	// Check function registrations
	if !strings.Contains(code, `RegisterFunction("runHello", runHello)`) {
		t.Error("generated code should register runHello")
	}
	if !strings.Contains(code, `RegisterFunction("runGoodbye", runGoodbye)`) {
		t.Error("generated code should register runGoodbye")
	}

	// Check BuildRootCommand call
	if !strings.Contains(code, "BuildRootCommand()") {
		t.Error("generated code should call BuildRootCommand")
	}

	// Check Execute call
	if !strings.Contains(code, "rootCmd.Execute()") {
		t.Error("generated code should call Execute")
	}
}

func TestGenerator_GenerateMain_WithRootRunFunc(t *testing.T) {
	yamlContent := `
name: test
description: test
root:
  use: test
  run_func: runRoot
commands:
  sub:
    use: sub
    short: Subcommand
    run_func: runSub
`
	gen, err := NewGeneratorFromString(yamlContent)
	if err != nil {
		t.Fatalf("NewGeneratorFromString() error = %v", err)
	}

	code, err := gen.GenerateMain("main", "config.yaml")
	if err != nil {
		t.Fatalf("GenerateMain() error = %v", err)
	}

	// Check both root and sub command functions are registered
	if !strings.Contains(code, `RegisterFunction("runRoot", runRoot)`) {
		t.Error("generated code should register runRoot")
	}
	if !strings.Contains(code, `RegisterFunction("runSub", runSub)`) {
		t.Error("generated code should register runSub")
	}
}

func TestGenerator_GenerateMainToFile(t *testing.T) {
	yamlContent := `
name: test
description: test
root:
  use: test
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
	outputPath := filepath.Join(tmpDir, "main.go")

	err = gen.GenerateMainToFile("main", "commands.yaml", outputPath)
	if err != nil {
		t.Fatalf("GenerateMainToFile() error = %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	if !strings.Contains(string(content), "func main()") {
		t.Error("generated file should contain main function")
	}

	if !strings.Contains(string(content), "runHello") {
		t.Error("generated file should contain runHello registration")
	}
}

func TestNewGenerator_FileNotFound(t *testing.T) {
	_, err := NewGenerator("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read config file") {
		t.Errorf("error should mention file read failure, got: %v", err)
	}
}

func TestNewGenerator_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML content
	invalidYAML := `
name: test
  invalid indentation here
    broken: yaml
`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	_, err := NewGenerator(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "failed to unmarshal YAML") {
		t.Errorf("error should mention YAML unmarshal failure, got: %v", err)
	}
}

func TestNewGeneratorFromString_InvalidYAML(t *testing.T) {
	invalidYAML := `
name: test
  invalid: indentation
    broken: yaml
`
	_, err := NewGeneratorFromString(invalidYAML)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "failed to unmarshal YAML") {
		t.Errorf("error should mention YAML unmarshal failure, got: %v", err)
	}
}

func TestGenerator_GenerateHandlersToFile_WriteError(t *testing.T) {
	yamlContent := `
name: test
description: test
root:
  use: test
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

	// Try to write to an invalid path (nonexistent directory)
	err = gen.GenerateHandlersToFile("main", "/nonexistent/path/handlers.go")
	if err == nil {
		t.Error("expected error when writing to invalid path")
	}
}

func TestGenerator_GenerateMainToFile_WriteError(t *testing.T) {
	yamlContent := `
name: test
description: test
root:
  use: test
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

	// Try to write to an invalid path (nonexistent directory)
	err = gen.GenerateMainToFile("main", "commands.yaml", "/nonexistent/path/main.go")
	if err == nil {
		t.Error("expected error when writing to invalid path")
	}
}
