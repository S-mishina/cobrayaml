package cobrayaml

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

// DocGenerator generates documentation from code definitions
type DocGenerator struct{}

// NewDocGenerator creates a new documentation generator
func NewDocGenerator() *DocGenerator {
	return &DocGenerator{}
}

// GenerateYAMLReference generates the YAML reference documentation
func (d *DocGenerator) GenerateYAMLReference() string {
	var buf bytes.Buffer

	// Flag Types (from actual constants)
	buf.WriteString("### Flag Types\n\n")
	buf.WriteString("| Type | Go Type | Example |\n")
	buf.WriteString("|------|---------|--------|\n")
	for _, ft := range SupportedFlagTypes {
		fmt.Fprintf(&buf, "| `%s` | `%s` | `%s` |\n",
			ft, flagTypeGoType(ft), flagTypeExample(ft))
	}
	buf.WriteString("\n")

	// Args Validation (from actual constants)
	buf.WriteString("### Args Validation\n\n")
	buf.WriteString("| Type | Description | Config |\n")
	buf.WriteString("|------|-------------|--------|\n")
	for _, at := range SupportedArgsTypes {
		fmt.Fprintf(&buf, "| `%s` | %s | %s |\n",
			at, argsTypeDescription(at), argsTypeConfig(at))
	}
	buf.WriteString("\n")

	// ToolConfig (from reflection)
	buf.WriteString("### ToolConfig (Root)\n\n")
	buf.WriteString("| YAML Key | Type | Description |\n")
	buf.WriteString("|----------|------|-------------|\n")
	toolFields := extractFieldDocs(reflect.TypeOf(ToolConfig{}))
	for _, f := range toolFields {
		if f.YAMLKey == "functions" {
			continue
		}
		desc := fieldDescription("ToolConfig", f.YAMLKey)
		fmt.Fprintf(&buf, "| `%s` | `%s` | %s |\n", f.YAMLKey, f.GoType, desc)
	}
	buf.WriteString("\n")

	// CommandConfig (from reflection)
	buf.WriteString("### CommandConfig\n\n")
	buf.WriteString("| YAML Key | Type | Description |\n")
	buf.WriteString("|----------|------|-------------|\n")
	cmdFields := extractFieldDocs(reflect.TypeOf(CommandConfig{}))
	for _, f := range cmdFields {
		desc := fieldDescription("CommandConfig", f.YAMLKey)
		fmt.Fprintf(&buf, "| `%s` | `%s` | %s |\n", f.YAMLKey, f.GoType, desc)
	}
	buf.WriteString("\n")

	// FlagConfig (from reflection)
	buf.WriteString("### FlagConfig\n\n")
	buf.WriteString("| YAML Key | Type | Required | Description |\n")
	buf.WriteString("|----------|------|----------|-------------|\n")
	flagFields := extractFieldDocs(reflect.TypeOf(FlagConfig{}))
	for _, f := range flagFields {
		req := ""
		if f.Required {
			req = "Yes"
		}
		desc := fieldDescription("FlagConfig", f.YAMLKey)
		fmt.Fprintf(&buf, "| `%s` | `%s` | %s | %s |\n", f.YAMLKey, f.GoType, req, desc)
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

	return buf.String()
}

// GenerateQuickStart generates the quick start documentation
func (d *DocGenerator) GenerateQuickStart() string {
	var buf bytes.Buffer
	buf.WriteString("\n\n1. Create `commands.yaml`:\n\n")
	buf.WriteString("```yaml\n")
	buf.WriteString(ExampleCommandsYAML)
	buf.WriteString("```\n\n")
	buf.WriteString("1. Create `main.go`:\n\n")
	buf.WriteString("```go\n")
	buf.WriteString(ExampleMainGo)
	buf.WriteString("```\n\n")
	buf.WriteString("1. Run:\n\n")
	buf.WriteString("```bash\n")
	buf.WriteString("go run . --help\n")
	buf.WriteString("go run . list\n")
	buf.WriteString("go run . add myitem\n")
	buf.WriteString("go run . --version\n")
	buf.WriteString("```\n\n")
	return buf.String()
}

// GenerateCodeGenSection generates the code generation documentation
func (d *DocGenerator) GenerateCodeGenSection() (string, error) {
	exampleYAML := `name: "example"
root:
  use: "example"
  short: "Example CLI tool"
commands:
  add:
    use: "add <name>"
    short: "Add an item"
    run_func: "runAdd"
    flags:
      - name: "force"
        shorthand: "f"
        type: bool
        usage: "Force the operation"
    args:
      type: exact
      count: 1`

	gen, err := NewGeneratorFromString(exampleYAML)
	if err != nil {
		return "", err
	}

	code, err := gen.GenerateHandlers("main")
	if err != nil {
		return "", err
	}

	// Extract just the function part
	lines := strings.Split(code, "\n")
	var funcLines []string
	inFunc := false
	for _, line := range lines {
		if strings.HasPrefix(line, "func ") {
			inFunc = true
		}
		if inFunc {
			funcLines = append(funcLines, line)
		}
	}
	generatedCode := strings.Join(funcLines, "\n")

	var buf bytes.Buffer
	buf.WriteString("\n\nGenerate handler function stubs from your YAML configuration:\n\n")
	buf.WriteString("```bash\n")
	buf.WriteString("# Create a new commands.yaml template\n")
	buf.WriteString("cobrayaml init my-app\n")
	buf.WriteString("\n")
	buf.WriteString("# Generate handler stubs\n")
	buf.WriteString("cobrayaml gen commands.yaml\n")
	buf.WriteString("\n")
	buf.WriteString("# Specify output file and package name\n")
	buf.WriteString("cobrayaml gen commands.yaml -o handlers.go -p main\n")
	buf.WriteString("```\n\n")
	buf.WriteString("### Generated Code Example\n\n")
	buf.WriteString("From this YAML:\n\n")
	buf.WriteString("```yaml\n")
	buf.WriteString(exampleYAML)
	buf.WriteString("\n```\n\n")
	buf.WriteString("Generates:\n\n")
	buf.WriteString("```go\n")
	buf.WriteString(generatedCode)
	buf.WriteString("```\n\n")

	return buf.String(), nil
}

// FieldDoc holds documentation extracted from struct fields
type FieldDoc struct {
	YAMLKey  string
	GoType   string
	Required bool
}

func extractFieldDocs(t reflect.Type) []FieldDoc {
	var docs []FieldDoc

	for i := range t.NumField() {
		field := t.Field(i)

		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			continue
		}

		parts := strings.Split(yamlTag, ",")
		yamlKey := parts[0]

		required := true
		for _, part := range parts[1:] {
			if part == "omitempty" {
				required = false
				break
			}
		}

		goType := formatGoType(field.Type)

		docs = append(docs, FieldDoc{
			YAMLKey:  yamlKey,
			GoType:   goType,
			Required: required,
		})
	}

	return docs
}

func formatGoType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Slice:
		return "[]" + formatGoType(t.Elem())
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", formatGoType(t.Key()), formatGoType(t.Elem()))
	case reflect.Ptr:
		return "*" + formatGoType(t.Elem())
	default:
		name := t.Name()
		if name == "" {
			return t.String()
		}
		return name
	}
}

func flagTypeGoType(flagType string) string {
	switch flagType {
	case FlagTypeString:
		return "string"
	case FlagTypeBool:
		return "bool"
	case FlagTypeInt:
		return "int"
	case FlagTypeStringSlice:
		return "[]string"
	default:
		return "any"
	}
}

func flagTypeExample(flagType string) string {
	switch flagType {
	case FlagTypeString:
		return "--name foo"
	case FlagTypeBool:
		return "--debug"
	case FlagTypeInt:
		return "--count 10"
	case FlagTypeStringSlice:
		return "--tags a,b,c"
	default:
		return ""
	}
}

func argsTypeDescription(argsType string) string {
	switch argsType {
	case ArgsTypeNone:
		return "No arguments allowed"
	case ArgsTypeAny:
		return "Any number of arguments"
	case ArgsTypeExact:
		return "Exact number required"
	case ArgsTypeMin:
		return "Minimum number"
	case ArgsTypeMax:
		return "Maximum number"
	case ArgsTypeRange:
		return "Range of arguments"
	default:
		return ""
	}
}

func argsTypeConfig(argsType string) string {
	switch argsType {
	case ArgsTypeNone:
		return "`type: none`"
	case ArgsTypeAny:
		return "`type: any`"
	case ArgsTypeExact:
		return "`type: exact`, `count: N`"
	case ArgsTypeMin:
		return "`type: min`, `min: N`"
	case ArgsTypeMax:
		return "`type: max`, `max: N`"
	case ArgsTypeRange:
		return "`type: range`, `min: N`, `max: N`"
	default:
		return ""
	}
}

func fieldDescription(structName, yamlKey string) string {
	descriptions := map[string]map[string]string{
		"ToolConfig": {
			"name":        "Tool name",
			"description": "Tool description",
			"version":     "Tool version (shown with --version)",
			"root":        "Root command configuration",
			"commands":    "Top-level subcommands",
		},
		"CommandConfig": {
			"use":      "Command name and argument pattern (e.g., `add <name>`)",
			"aliases":  "Alternative command names",
			"short":    "Brief description shown in help",
			"long":     "Detailed description",
			"args":     "Argument validation configuration",
			"run_func": "Name of the handler function",
			"flags":    "List of flag definitions",
			"commands": "Nested subcommands",
			"hidden":   "Hide command from help output",
		},
		"FlagConfig": {
			"name":       "Flag name (e.g., `namespace` for --namespace)",
			"shorthand":  "Short flag (e.g., `n` for -n)",
			"type":       "Flag type (string, bool, int, stringSlice)",
			"default":    "Default value",
			"usage":      "Description shown in help",
			"required":   "Mark flag as required",
			"persistent": "Inherit flag to all subcommands",
			"hidden":     "Hide flag from help output",
		},
	}

	if structDescs, ok := descriptions[structName]; ok {
		if desc, ok := structDescs[yamlKey]; ok {
			return desc
		}
	}
	return ""
}
