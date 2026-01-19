// Package main generates YAML reference documentation from code.
//
//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// FlagTypeInfo holds documentation for a flag type
type FlagTypeInfo struct {
	Name    string
	GoType  string
	Example string
}

// ArgsTypeInfo holds documentation for an args validation type
type ArgsTypeInfo struct {
	Name        string
	Description string
	Config      string
}

// FlagFieldInfo holds documentation for a flag config field
type FlagFieldInfo struct {
	Field       string
	YAMLKey     string
	Type        string
	Required    bool
	Description string
}

// CommandFieldInfo holds documentation for a command config field
type CommandFieldInfo struct {
	Field       string
	YAMLKey     string
	Type        string
	Description string
}

var flagTypes = []FlagTypeInfo{
	{"string", "string", "--name foo"},
	{"bool", "bool", "--debug"},
	{"int", "int", "--count 10"},
	{"stringSlice", "[]string", "--tags a,b,c"},
}

var argsTypes = []ArgsTypeInfo{
	{"none", "No arguments allowed", "`type: none`"},
	{"any", "Any number of arguments", "`type: any`"},
	{"exact", "Exact number required", "`type: exact`, `count: N`"},
	{"min", "Minimum number", "`type: min`, `min: N`"},
	{"max", "Maximum number", "`type: max`, `max: N`"},
	{"range", "Range of arguments", "`type: range`, `min: N`, `max: N`"},
}

var flagFields = []FlagFieldInfo{
	{"Name", "name", "string", true, "Flag name (e.g., `namespace` for --namespace)"},
	{"Shorthand", "shorthand", "string", false, "Short flag (e.g., `n` for -n)"},
	{"Type", "type", "string", true, "Flag type (string, bool, int, stringSlice)"},
	{"DefaultValue", "default", "string", false, "Default value"},
	{"Usage", "usage", "string", true, "Description shown in help"},
	{"Required", "required", "bool", false, "Mark flag as required"},
	{"Persistent", "persistent", "bool", false, "Inherit flag to all subcommands"},
	{"Hidden", "hidden", "bool", false, "Hide flag from help output"},
}

var commandFields = []CommandFieldInfo{
	{"Use", "use", "string", "Command name and argument pattern (e.g., `add <name>`)"},
	{"Aliases", "aliases", "[]string", "Alternative command names"},
	{"Short", "short", "string", "Brief description shown in help"},
	{"Long", "long", "string", "Detailed description"},
	{"Args", "args", "ArgsConfig", "Argument validation configuration"},
	{"RunFunc", "run_func", "string", "Name of the handler function"},
	{"Flags", "flags", "[]FlagConfig", "List of flag definitions"},
	{"Commands", "commands", "map[string]CommandConfig", "Nested subcommands"},
	{"Hidden", "hidden", "bool", "Hide command from help output"},
}

var toolConfigFields = []CommandFieldInfo{
	{"Name", "name", "string", "Tool name"},
	{"Description", "description", "string", "Tool description"},
	{"Version", "version", "string", "Tool version (shown with --version)"},
	{"Root", "root", "CommandConfig", "Root command configuration"},
	{"Commands", "commands", "map[string]CommandConfig", "Top-level subcommands"},
}

// extractConstant extracts a raw string constant from examples.go
func extractConstant(content, constName string) (string, error) {
	// Match: const ConstName = `...`
	pattern := fmt.Sprintf(`const %s = `+"`"+`([^`+"`"+`]*)`+"`", constName)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) < 2 {
		return "", fmt.Errorf("constant %s not found", constName)
	}
	return matches[1], nil
}

func main() {
	// Read examples.go to extract example code
	examplesContent, err := os.ReadFile("examples.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading examples.go: %v\n", err)
		os.Exit(1)
	}

	exampleYAML, err := extractConstant(string(examplesContent), "ExampleCommandsYAML")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting ExampleCommandsYAML: %v\n", err)
		os.Exit(1)
	}

	exampleMainGo, err := extractConstant(string(examplesContent), "ExampleMainGo")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting ExampleMainGo: %v\n", err)
		os.Exit(1)
	}

	var buf bytes.Buffer

	// Flag Types
	buf.WriteString("### Flag Types\n\n")
	buf.WriteString("| Type | Go Type | Example |\n")
	buf.WriteString("|------|---------|--------|\n")
	for _, ft := range flagTypes {
		fmt.Fprintf(&buf, "| `%s` | `%s` | `%s` |\n", ft.Name, ft.GoType, ft.Example)
	}
	buf.WriteString("\n")

	// Args Validation
	buf.WriteString("### Args Validation\n\n")
	buf.WriteString("| Type | Description | Config |\n")
	buf.WriteString("|------|-------------|--------|\n")
	for _, at := range argsTypes {
		fmt.Fprintf(&buf, "| `%s` | %s | %s |\n", at.Name, at.Description, at.Config)
	}
	buf.WriteString("\n")

	// ToolConfig
	buf.WriteString("### ToolConfig (Root)\n\n")
	buf.WriteString("| YAML Key | Type | Description |\n")
	buf.WriteString("|----------|------|-------------|\n")
	for _, f := range toolConfigFields {
		fmt.Fprintf(&buf, "| `%s` | `%s` | %s |\n", f.YAMLKey, f.Type, f.Description)
	}
	buf.WriteString("\n")

	// CommandConfig
	buf.WriteString("### CommandConfig\n\n")
	buf.WriteString("| YAML Key | Type | Description |\n")
	buf.WriteString("|----------|------|-------------|\n")
	for _, f := range commandFields {
		fmt.Fprintf(&buf, "| `%s` | `%s` | %s |\n", f.YAMLKey, f.Type, f.Description)
	}
	buf.WriteString("\n")

	// FlagConfig
	buf.WriteString("### FlagConfig\n\n")
	buf.WriteString("| YAML Key | Type | Required | Description |\n")
	buf.WriteString("|----------|------|----------|-------------|\n")
	for _, f := range flagFields {
		req := ""
		if f.Required {
			req = "Yes"
		}
		fmt.Fprintf(&buf, "| `%s` | `%s` | %s | %s |\n", f.YAMLKey, f.Type, req, f.Description)
	}
	buf.WriteString("\n")

	// Hidden Commands/Flags Example
	buf.WriteString("### Hidden Commands/Flags\n\n")
	buf.WriteString("```yaml\n")
	buf.WriteString("commands:\n")
	buf.WriteString("  internal:\n")
	buf.WriteString("    use: internal\n")
	buf.WriteString("    short: Internal command\n")
	buf.WriteString("    hidden: true\n")
	buf.WriteString("\n")
	buf.WriteString("root:\n")
	buf.WriteString("  flags:\n")
	buf.WriteString("    - name: debug\n")
	buf.WriteString("      type: bool\n")
	buf.WriteString("      hidden: true\n")
	buf.WriteString("```\n")

	// Read README
	readme, err := os.ReadFile("README.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading README.md: %v\n", err)
		os.Exit(1)
	}

	// Find and replace the YAML Reference section
	content := string(readme)
	startMarker := "<!-- YAML_REFERENCE_START -->"
	endMarker := "<!-- YAML_REFERENCE_END -->"

	startIdx := strings.Index(content, startMarker)
	endIdx := strings.Index(content, endMarker)

	if startIdx == -1 || endIdx == -1 {
		fmt.Fprintf(os.Stderr, "Markers not found in README.md\n")
		fmt.Fprintf(os.Stderr, "Please add %s and %s markers\n", startMarker, endMarker)
		os.Exit(1)
	}

	newContent := content[:startIdx+len(startMarker)] + "\n\n" + buf.String() + "\n" + content[endIdx:]

	// Update Quick Start section
	quickStartMarker := "<!-- QUICK_START_START -->"
	quickStartEndMarker := "<!-- QUICK_START_END -->"

	qsStartIdx := strings.Index(newContent, quickStartMarker)
	qsEndIdx := strings.Index(newContent, quickStartEndMarker)

	if qsStartIdx != -1 && qsEndIdx != -1 {
		var qsBuf bytes.Buffer
		qsBuf.WriteString("\n\n1. Create `commands.yaml`:\n\n")
		qsBuf.WriteString("```yaml\n")
		qsBuf.WriteString(exampleYAML)
		qsBuf.WriteString("```\n\n")
		qsBuf.WriteString("1. Create `main.go`:\n\n")
		qsBuf.WriteString("```go\n")
		qsBuf.WriteString(exampleMainGo)
		qsBuf.WriteString("```\n\n")
		qsBuf.WriteString("1. Run:\n\n")
		qsBuf.WriteString("```bash\n")
		qsBuf.WriteString("go run . --help\n")
		qsBuf.WriteString("go run . list\n")
		qsBuf.WriteString("go run . add myitem\n")
		qsBuf.WriteString("go run . --version\n")
		qsBuf.WriteString("```\n\n")

		newContent = newContent[:qsStartIdx+len(quickStartMarker)] + qsBuf.String() + newContent[qsEndIdx:]
	}

	if err := os.WriteFile("README.md", []byte(newContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing README.md: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Documentation generated successfully.")
}
