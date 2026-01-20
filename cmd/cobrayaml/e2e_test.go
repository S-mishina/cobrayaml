package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var (
	binaryPath string
)

func TestMain(m *testing.M) {
	// Build the cobrayaml binary for E2E tests
	tmpDir, err := os.MkdirTemp("", "cobrayaml-e2e-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(tmpDir)

	binaryName := "cobrayaml"
	if runtime.GOOS == "windows" {
		binaryName = "cobrayaml.exe"
	}
	binaryPath = filepath.Join(tmpDir, binaryName)

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = filepath.Join(getProjectRoot(), "cmd", "cobrayaml")
	if output, err := cmd.CombinedOutput(); err != nil {
		panic("failed to build binary: " + err.Error() + "\nOutput: " + string(output))
	}

	os.Exit(m.Run())
}

func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}

// runCobrayaml executes the cobrayaml binary with the given args and working directory
func runCobrayaml(t *testing.T, workDir string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	t.Logf(">>> Running: cobrayaml %s", strings.Join(args, " "))
	t.Logf("    Working directory: %s", workDir)

	err := cmd.Run()

	if stdout.Len() > 0 {
		t.Logf("<<< STDOUT:\n%s", stdout.String())
	}
	if stderr.Len() > 0 {
		t.Logf("<<< STDERR:\n%s", stderr.String())
	}
	if err != nil {
		t.Logf("<<< Exit error: %v", err)
	}

	return stdout.String(), stderr.String(), err
}

// ============================================================================
// init command E2E tests
// ============================================================================

func TestE2E_Init_Default(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, stderr, err := runCobrayaml(t, tmpDir, "init")
	if err != nil {
		t.Fatalf("init command failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	// Check output message
	if !strings.Contains(stdout, "Created commands.yaml") {
		t.Errorf("expected output to contain 'Created commands.yaml', got: %s", stdout)
	}

	// Check that commands.yaml was created
	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Fatal("commands.yaml was not created")
	}

	// Log the generated file content
	logFileContent(t, yamlPath)

	// Check that the file contains expected content
	content, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("failed to read commands.yaml: %v", err)
	}

	expectedContents := []string{"name: my-tool", "root:", "short:"}
	for _, expected := range expectedContents {
		if !strings.Contains(string(content), expected) {
			t.Errorf("commands.yaml should contain %q", expected)
		}
	}
}

func TestE2E_Init_CustomName(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, stderr, err := runCobrayaml(t, tmpDir, "init", "custom-cli")
	if err != nil {
		t.Fatalf("init command failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	content, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("failed to read commands.yaml: %v", err)
	}

	if !strings.Contains(string(content), "name: custom-cli") {
		t.Errorf("commands.yaml should contain custom tool name, got: %s", string(content))
	}
}

func TestE2E_Init_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing commands.yaml
	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if err := os.WriteFile(yamlPath, []byte("existing content"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	_, _, err := runCobrayaml(t, tmpDir, "init")
	if err == nil {
		t.Fatal("expected error when commands.yaml already exists")
	}
}

// ============================================================================
// gen command E2E tests
// ============================================================================

func TestE2E_Gen_Basic(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple YAML config
	yamlContent := `name: test-cli
description: A test CLI
root:
  use: test-cli
  short: Test CLI application
commands:
  greet:
    use: greet [name]
    short: Greet someone
    run_func: handleGreet
    flags:
      - name: loud
        type: bool
        shorthand: l
        usage: Greet loudly
`
	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write commands.yaml: %v", err)
	}

	t.Log("--- Input YAML ---")
	t.Log(yamlContent)

	stdout, stderr, err := runCobrayaml(t, tmpDir, "gen", "commands.yaml")
	if err != nil {
		t.Fatalf("gen command failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	// Check that files were generated
	handlersPath := filepath.Join(tmpDir, "handlers.go")
	mainPath := filepath.Join(tmpDir, "main.go")

	if _, err := os.Stat(handlersPath); os.IsNotExist(err) {
		t.Fatal("handlers.go was not created")
	}
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		t.Fatal("main.go was not created")
	}

	// Log generated files
	logFileContent(t, handlersPath)
	logFileContent(t, mainPath)

	// Check handlers.go content
	handlersContent, err := os.ReadFile(handlersPath)
	if err != nil {
		t.Fatalf("failed to read handlers.go: %v", err)
	}

	expectedHandlerContents := []string{
		"package main",
		"handleGreet",
		"func",
	}
	for _, expected := range expectedHandlerContents {
		if !strings.Contains(string(handlersContent), expected) {
			t.Errorf("handlers.go should contain %q", expected)
		}
	}

	// Check main.go content
	mainContent, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}

	expectedMainContents := []string{
		"package main",
		"cobrayaml",
		"commands.yaml",
	}
	for _, expected := range expectedMainContents {
		if !strings.Contains(string(mainContent), expected) {
			t.Errorf("main.go should contain %q", expected)
		}
	}
}

func TestE2E_Gen_WithOutputPaths(t *testing.T) {
	tmpDir := t.TempDir()

	yamlContent := `name: test-cli
description: A test CLI
root:
  use: test-cli
  short: Test CLI
commands:
  hello:
    use: hello
    short: Say hello
    run_func: handleHello
`
	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write commands.yaml: %v", err)
	}

	customHandlers := filepath.Join(tmpDir, "custom_handlers.go")
	customMain := filepath.Join(tmpDir, "custom_main.go")

	stdout, stderr, err := runCobrayaml(t, tmpDir, "gen", "commands.yaml",
		"-o", customHandlers,
		"-m", customMain,
		"-p", "mypackage")
	if err != nil {
		t.Fatalf("gen command failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	if _, err := os.Stat(customHandlers); os.IsNotExist(err) {
		t.Fatal("custom_handlers.go was not created")
	}
	if _, err := os.Stat(customMain); os.IsNotExist(err) {
		t.Fatal("custom_main.go was not created")
	}

	// Check package name
	handlersContent, _ := os.ReadFile(customHandlers)
	if !strings.Contains(string(handlersContent), "package mypackage") {
		t.Error("handlers should have custom package name")
	}
}

func TestE2E_Gen_NoOverwriteWithoutForce(t *testing.T) {
	tmpDir := t.TempDir()

	yamlContent := `name: test-cli
description: A test CLI
root:
  use: test-cli
  short: Test CLI
commands:
  hello:
    use: hello
    short: Say hello
    run_func: handleHello
`
	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write commands.yaml: %v", err)
	}

	// Create existing handlers.go
	handlersPath := filepath.Join(tmpDir, "handlers.go")
	originalContent := "// original content"
	if err := os.WriteFile(handlersPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to create existing handlers.go: %v", err)
	}

	stdout, _, _ := runCobrayaml(t, tmpDir, "gen", "commands.yaml")

	// Check that warning was displayed
	if !strings.Contains(stdout, "Warning") && !strings.Contains(stdout, "already exist") {
		t.Error("expected warning about existing files")
	}

	// Check that file was NOT overwritten
	content, _ := os.ReadFile(handlersPath)
	if string(content) != originalContent {
		t.Error("handlers.go should not have been overwritten without --force")
	}
}

func TestE2E_Gen_ForceOverwrite(t *testing.T) {
	tmpDir := t.TempDir()

	yamlContent := `name: test-cli
description: A test CLI
root:
  use: test-cli
  short: Test CLI
commands:
  hello:
    use: hello
    short: Say hello
    run_func: handleHello
`
	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write commands.yaml: %v", err)
	}

	// Create existing handlers.go
	handlersPath := filepath.Join(tmpDir, "handlers.go")
	if err := os.WriteFile(handlersPath, []byte("// original content"), 0644); err != nil {
		t.Fatalf("failed to create existing handlers.go: %v", err)
	}

	_, stderr, err := runCobrayaml(t, tmpDir, "gen", "commands.yaml", "--force")
	if err != nil {
		t.Fatalf("gen --force command failed: %v\nstderr: %s", err, stderr)
	}

	// Check that file WAS overwritten
	content, _ := os.ReadFile(handlersPath)
	if !strings.Contains(string(content), "handleHello") {
		t.Error("handlers.go should have been overwritten with --force")
	}
}

func TestE2E_Gen_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if err := os.WriteFile(yamlPath, []byte("invalid: yaml: ["), 0644); err != nil {
		t.Fatalf("failed to write commands.yaml: %v", err)
	}

	_, _, err := runCobrayaml(t, tmpDir, "gen", "commands.yaml")
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestE2E_Gen_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	_, _, err := runCobrayaml(t, tmpDir, "gen", "nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// ============================================================================
// docs command E2E tests
// ============================================================================

func TestE2E_Docs_Stdout(t *testing.T) {
	tmpDir := t.TempDir()

	yamlContent := `name: test-cli
description: A test CLI for documentation
version: "1.0.0"
root:
  use: test-cli
  short: Test CLI application
  long: This is a longer description of the test CLI application
commands:
  greet:
    use: greet [name]
    short: Greet someone
    long: Greet a person by name
    run_func: handleGreet
    flags:
      - name: loud
        type: bool
        shorthand: l
        usage: Greet loudly
      - name: times
        type: int
        shorthand: t
        default: 1
        usage: Number of times to greet
`
	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write commands.yaml: %v", err)
	}

	stdout, stderr, err := runCobrayaml(t, tmpDir, "docs", "commands.yaml")
	if err != nil {
		t.Fatalf("docs command failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	// Check that documentation contains expected content
	expectedContents := []string{
		"test-cli",
		"greet",
		"--loud",
		"--times",
		"Greet someone",
	}
	for _, expected := range expectedContents {
		if !strings.Contains(stdout, expected) {
			t.Errorf("documentation should contain %q", expected)
		}
	}
}

func TestE2E_Docs_OutputFile(t *testing.T) {
	tmpDir := t.TempDir()

	yamlContent := `name: test-cli
description: A test CLI
root:
  use: test-cli
  short: Test CLI
commands:
  hello:
    use: hello
    short: Say hello
    run_func: handleHello
`
	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write commands.yaml: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "README.md")
	stdout, stderr, err := runCobrayaml(t, tmpDir, "docs", "commands.yaml", "-o", outputPath)
	if err != nil {
		t.Fatalf("docs command failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("README.md was not created")
	}

	content, _ := os.ReadFile(outputPath)
	if !strings.Contains(string(content), "test-cli") {
		t.Error("README.md should contain tool name")
	}
}

func TestE2E_Docs_NestedCommands(t *testing.T) {
	tmpDir := t.TempDir()

	yamlContent := `name: test-cli
description: A test CLI with nested commands
root:
  use: test-cli
  short: Test CLI
commands:
  db:
    use: db
    short: Database commands
    commands:
      migrate:
        use: migrate
        short: Migration commands
        commands:
          up:
            use: up
            short: Run migrations up
            run_func: handleMigrateUp
          down:
            use: down
            short: Run migrations down
            run_func: handleMigrateDown
`
	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write commands.yaml: %v", err)
	}

	stdout, stderr, err := runCobrayaml(t, tmpDir, "docs", "commands.yaml")
	if err != nil {
		t.Fatalf("docs command failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	// Check that nested commands are documented
	expectedCommands := []string{"db", "migrate", "up", "down"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(stdout, cmd) {
			t.Errorf("documentation should contain nested command %q", cmd)
		}
	}
}

// ============================================================================
// Generated code compile and execute E2E tests
// ============================================================================

func TestE2E_GeneratedCode_Compiles(t *testing.T) {
	tmpDir := t.TempDir()

	yamlContent := `name: test-cli
description: A test CLI
root:
  use: test-cli
  short: Test CLI application
commands:
  greet:
    use: greet [name]
    short: Greet someone
    run_func: handleGreet
    flags:
      - name: loud
        type: bool
        shorthand: l
        usage: Greet loudly
  config:
    use: config
    short: Configuration commands
    commands:
      show:
        use: show
        short: Show configuration
        run_func: handleConfigShow
      set:
        use: set <key> <value>
        short: Set configuration value
        run_func: handleConfigSet
        args:
          type: exact
          count: 2
`
	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write commands.yaml: %v", err)
	}

	// Generate code
	_, stderr, err := runCobrayaml(t, tmpDir, "gen", "commands.yaml")
	if err != nil {
		t.Fatalf("gen command failed: %v\nstderr: %s", err, stderr)
	}

	// Initialize go module
	cmd := exec.Command("go", "mod", "init", "test-cli")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod init failed: %v\nOutput: %s", err, string(output))
	}

	// Add required dependencies
	cmd = exec.Command("go", "get", "github.com/S-mishina/cobrayaml")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go get cobrayaml failed: %v\nOutput: %s", err, string(output))
	}

	cmd = exec.Command("go", "get", "github.com/spf13/cobra")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go get cobra failed: %v\nOutput: %s", err, string(output))
	}

	// Build the generated code
	binaryName := "test-cli"
	if runtime.GOOS == "windows" {
		binaryName = "test-cli.exe"
	}
	binaryPath := filepath.Join(tmpDir, binaryName)

	cmd = exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\nOutput: %s", err, string(output))
	}

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatal("binary was not created")
	}
}

func TestE2E_GeneratedCode_HelpWorks(t *testing.T) {
	tmpDir := t.TempDir()

	yamlContent := `name: test-cli
description: A test CLI application
version: "1.0.0"
root:
  use: test-cli
  short: Test CLI application
  long: This is a test CLI application for E2E testing
commands:
  greet:
    use: greet [name]
    short: Greet someone
    long: Greet a person by their name
    run_func: handleGreet
    flags:
      - name: loud
        type: bool
        shorthand: l
        usage: Greet loudly
`
	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write commands.yaml: %v", err)
	}

	t.Log("--- Input YAML ---")
	t.Log(yamlContent)

	// Generate code
	_, stderr, err := runCobrayaml(t, tmpDir, "gen", "commands.yaml")
	if err != nil {
		t.Fatalf("gen command failed: %v\nstderr: %s", err, stderr)
	}

	// Log generated files
	logFileContent(t, filepath.Join(tmpDir, "handlers.go"))
	logFileContent(t, filepath.Join(tmpDir, "main.go"))

	// Initialize go module and get dependencies
	setupGoModule(t, tmpDir)

	// Build the generated code
	binaryName := "test-cli"
	if runtime.GOOS == "windows" {
		binaryName = "test-cli.exe"
	}
	genBinaryPath := buildGeneratedCode(t, tmpDir, filepath.Join(tmpDir, binaryName))

	// Test help output
	output, err := runGeneratedBinary(t, genBinaryPath, tmpDir, "--help")
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	expectedInHelp := []string{
		"test-cli",
		"E2E testing",
		"greet",
	}
	for _, expected := range expectedInHelp {
		if !strings.Contains(output, expected) {
			t.Errorf("help output should contain %q", expected)
		}
	}

	// Test subcommand help
	output, err = runGeneratedBinary(t, genBinaryPath, tmpDir, "greet", "--help")
	if err != nil {
		t.Fatalf("greet help failed: %v", err)
	}

	if !strings.Contains(output, "--loud") {
		t.Errorf("greet help should contain --loud flag")
	}
}

func TestE2E_GeneratedCode_VersionWorks(t *testing.T) {
	tmpDir := t.TempDir()

	yamlContent := `name: test-cli
description: A test CLI
version: "2.5.0"
root:
  use: test-cli
  short: Test CLI
commands:
  hello:
    use: hello
    short: Say hello
    run_func: handleHello
`
	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write commands.yaml: %v", err)
	}

	t.Log("--- Input YAML ---")
	t.Log(yamlContent)

	// Generate and build
	_, stderr, err := runCobrayaml(t, tmpDir, "gen", "commands.yaml")
	if err != nil {
		t.Fatalf("gen command failed: %v\nstderr: %s", err, stderr)
	}

	setupGoModule(t, tmpDir)

	binaryName := "test-cli"
	if runtime.GOOS == "windows" {
		binaryName = "test-cli.exe"
	}
	genBinaryPath := buildGeneratedCode(t, tmpDir, filepath.Join(tmpDir, binaryName))

	// Test version output
	output, err := runGeneratedBinary(t, genBinaryPath, tmpDir, "--version")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	if !strings.Contains(output, "2.5.0") {
		t.Errorf("version output should contain '2.5.0', got: %s", output)
	}
}

// ============================================================================
// Full workflow E2E tests
// ============================================================================

func TestE2E_FullWorkflow_InitGenBuild(t *testing.T) {
	tmpDir := t.TempDir()

	// Step 1: Initialize
	stdout, stderr, err := runCobrayaml(t, tmpDir, "init", "my-awesome-cli")
	if err != nil {
		t.Fatalf("init command failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	// Verify commands.yaml was created
	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Fatal("commands.yaml was not created by init")
	}

	// Step 2: Generate code
	stdout, stderr, err = runCobrayaml(t, tmpDir, "gen", "commands.yaml")
	if err != nil {
		t.Fatalf("gen command failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	// Verify handlers.go and main.go were created
	handlersPath := filepath.Join(tmpDir, "handlers.go")
	mainPath := filepath.Join(tmpDir, "main.go")
	if _, err := os.Stat(handlersPath); os.IsNotExist(err) {
		t.Fatal("handlers.go was not created by gen")
	}
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		t.Fatal("main.go was not created by gen")
	}

	// Step 3: Initialize go module and build
	setupGoModule(t, tmpDir)

	binaryName := "my-awesome-cli"
	if runtime.GOOS == "windows" {
		binaryName = "my-awesome-cli.exe"
	}
	binaryPath := filepath.Join(tmpDir, binaryName)

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\nOutput: %s", err, string(output))
	}

	// Step 4: Run the built binary
	cmd = exec.Command(binaryPath, "--help")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("binary execution failed: %v\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), "my-awesome-cli") {
		t.Errorf("help output should contain tool name, got: %s", string(output))
	}
}

func TestE2E_FullWorkflow_ComplexCLI(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a complex YAML config
	yamlContent := `name: complex-cli
description: A complex CLI for E2E testing
version: "1.2.3"
root:
  use: complex-cli
  short: Complex CLI application
  long: This is a complex CLI application with multiple subcommands and features
  flags:
    - name: verbose
      type: bool
      shorthand: v
      usage: Enable verbose output
      persistent: true
    - name: config
      type: string
      shorthand: c
      usage: Config file path
      persistent: true
commands:
  db:
    use: db
    short: Database operations
    commands:
      migrate:
        use: migrate
        short: Run database migrations
        run_func: handleMigrate
        flags:
          - name: dry-run
            type: bool
            usage: Show what would be migrated
      seed:
        use: seed
        short: Seed the database
        run_func: handleSeed
  user:
    use: user
    short: User management
    commands:
      create:
        use: create <username>
        short: Create a new user
        run_func: handleUserCreate
        args:
          type: exact
          count: 1
        flags:
          - name: email
            type: string
            shorthand: e
            usage: User email
          - name: roles
            type: stringSlice
            shorthand: r
            usage: User roles
      list:
        use: list
        short: List all users
        run_func: handleUserList
        flags:
          - name: limit
            type: int
            shorthand: l
            default: 10
            usage: Maximum number of users to list
`
	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write commands.yaml: %v", err)
	}

	// Generate code
	_, stderr, err := runCobrayaml(t, tmpDir, "gen", "commands.yaml")
	if err != nil {
		t.Fatalf("gen command failed: %v\nstderr: %s", err, stderr)
	}

	// Setup and build
	setupGoModule(t, tmpDir)

	binaryName := "complex-cli"
	if runtime.GOOS == "windows" {
		binaryName = "complex-cli.exe"
	}
	binaryPath := filepath.Join(tmpDir, binaryName)

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\nOutput: %s", err, string(output))
	}

	// Test various command helps
	testCases := []struct {
		args     []string
		expected []string
	}{
		{
			args:     []string{"--help"},
			expected: []string{"complex-cli", "db", "user", "--verbose", "--config"},
		},
		{
			args:     []string{"db", "--help"},
			expected: []string{"migrate", "seed"},
		},
		{
			args:     []string{"db", "migrate", "--help"},
			expected: []string{"--dry-run"},
		},
		{
			args:     []string{"user", "--help"},
			expected: []string{"create", "list"},
		},
		{
			args:     []string{"user", "create", "--help"},
			expected: []string{"--email", "--roles"},
		},
		{
			args:     []string{"user", "list", "--help"},
			expected: []string{"--limit"},
		},
		{
			args:     []string{"--version"},
			expected: []string{"1.2.3"},
		},
	}

	for _, tc := range testCases {
		t.Run(strings.Join(tc.args, "_"), func(t *testing.T) {
			cmd := exec.Command(binaryPath, tc.args...)
			cmd.Dir = tmpDir
			output, err := cmd.CombinedOutput()
			if err != nil {
				// --version returns exit code 0 but some commands may fail
				// Only fail if we expected certain content and don't have it
			}

			for _, expected := range tc.expected {
				if !strings.Contains(string(output), expected) {
					t.Errorf("output for %v should contain %q, got: %s", tc.args, expected, string(output))
				}
			}
		})
	}

	// Generate and verify documentation
	stdout, stderr, err := runCobrayaml(t, tmpDir, "docs", "commands.yaml")
	if err != nil {
		t.Fatalf("docs command failed: %v\nstderr: %s", err, stderr)
	}

	docsExpected := []string{
		"complex-cli",
		"db",
		"user",
		"migrate",
		"seed",
		"create",
		"list",
		"--verbose",
		"--config",
		"--dry-run",
		"--email",
		"--roles",
		"--limit",
	}
	for _, expected := range docsExpected {
		if !strings.Contains(stdout, expected) {
			t.Errorf("documentation should contain %q", expected)
		}
	}
}

func TestE2E_AllFlagTypes(t *testing.T) {
	tmpDir := t.TempDir()

	yamlContent := `name: flag-test
description: Test all flag types
root:
  use: flag-test
  short: Flag test CLI
commands:
  test:
    use: test
    short: Test command with all flag types
    run_func: handleTest
    flags:
      - name: str-flag
        type: string
        shorthand: s
        default: "default"
        usage: A string flag
      - name: bool-flag
        type: bool
        shorthand: b
        usage: A boolean flag
      - name: int-flag
        type: int
        shorthand: i
        default: 42
        usage: An integer flag
      - name: slice-flag
        type: stringSlice
        shorthand: l
        usage: A string slice flag
`
	yamlPath := filepath.Join(tmpDir, "commands.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write commands.yaml: %v", err)
	}

	// Generate and build
	_, stderr, err := runCobrayaml(t, tmpDir, "gen", "commands.yaml")
	if err != nil {
		t.Fatalf("gen command failed: %v\nstderr: %s", err, stderr)
	}

	setupGoModule(t, tmpDir)

	binaryName := "flag-test"
	if runtime.GOOS == "windows" {
		binaryName = "flag-test.exe"
	}
	binaryPath := filepath.Join(tmpDir, binaryName)

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\nOutput: %s", err, string(output))
	}

	// Check help contains all flags
	cmd = exec.Command(binaryPath, "test", "--help")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("test help failed: %v\nOutput: %s", err, string(output))
	}

	expectedFlags := []string{
		"--str-flag", "-s",
		"--bool-flag", "-b",
		"--int-flag", "-i",
		"--slice-flag", "-l",
	}
	for _, expected := range expectedFlags {
		if !strings.Contains(string(output), expected) {
			t.Errorf("help should contain %q, got: %s", expected, string(output))
		}
	}
}

// ============================================================================
// Helper functions
// ============================================================================

func setupGoModule(t *testing.T, dir string) {
	t.Helper()

	t.Log(">>> Setting up Go module...")

	cmd := exec.Command("go", "mod", "init", "test-cli")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod init failed: %v\nOutput: %s", err, string(output))
	}
	t.Log("    go mod init: OK")

	cmd = exec.Command("go", "get", "github.com/S-mishina/cobrayaml")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go get cobrayaml failed: %v\nOutput: %s", err, string(output))
	}
	t.Log("    go get cobrayaml: OK")

	cmd = exec.Command("go", "get", "github.com/spf13/cobra")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go get cobra failed: %v\nOutput: %s", err, string(output))
	}
	t.Log("    go get cobra: OK")
}

// runGeneratedBinary executes a generated binary with logging
func runGeneratedBinary(t *testing.T, binaryPath string, workDir string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = workDir

	t.Logf(">>> Running generated binary: %s %s", filepath.Base(binaryPath), strings.Join(args, " "))

	output, err := cmd.CombinedOutput()

	if len(output) > 0 {
		t.Logf("<<< OUTPUT:\n%s", string(output))
	}
	if err != nil {
		t.Logf("<<< Exit error: %v", err)
	}

	return string(output), err
}

// buildGeneratedCode builds the generated Go code with logging
func buildGeneratedCode(t *testing.T, dir string, outputBinary string) string {
	t.Helper()

	t.Logf(">>> Building generated code in %s", dir)

	cmd := exec.Command("go", "build", "-o", outputBinary, ".")
	cmd.Dir = dir

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\nOutput: %s", err, string(output))
	}

	t.Logf("<<< Build successful: %s", outputBinary)
	return outputBinary
}

// logFileContent logs the content of a file for debugging
func logFileContent(t *testing.T, filePath string) {
	t.Helper()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Logf("    Could not read %s: %v", filePath, err)
		return
	}

	t.Logf("--- Content of %s ---\n%s\n--- End of %s ---", filepath.Base(filePath), string(content), filepath.Base(filePath))
}
