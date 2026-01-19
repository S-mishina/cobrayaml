package cobrayaml

import (
	"reflect"
	"strings"
	"testing"
)

func TestDocGenerator_GenerateYAMLReference(t *testing.T) {
	gen := NewDocGenerator()
	result := gen.GenerateYAMLReference()

	// Verify it contains sections derived from actual constants
	for _, flagType := range SupportedFlagTypes {
		if !strings.Contains(result, flagType) {
			t.Errorf("YAML reference should contain flag type %q", flagType)
		}
	}

	for _, argsType := range SupportedArgsTypes {
		if !strings.Contains(result, argsType) {
			t.Errorf("YAML reference should contain args type %q", argsType)
		}
	}

	// Verify it contains struct field documentation
	expectedFields := []string{"name", "use", "short", "flags", "commands"}
	for _, field := range expectedFields {
		if !strings.Contains(result, field) {
			t.Errorf("YAML reference should contain field %q", field)
		}
	}
}

func TestDocGenerator_GenerateQuickStart(t *testing.T) {
	gen := NewDocGenerator()
	result := gen.GenerateQuickStart()

	// Should contain the example YAML
	if !strings.Contains(result, "my-tool") {
		t.Error("Quick start should contain example tool name")
	}

	// Should contain the example Go code
	if !strings.Contains(result, "NewCommandBuilderFromString") {
		t.Error("Quick start should contain example Go code")
	}

	// Should contain run instructions
	if !strings.Contains(result, "go run") {
		t.Error("Quick start should contain run instructions")
	}
}

func TestDocGenerator_GenerateCodeGenSection(t *testing.T) {
	gen := NewDocGenerator()
	result, err := gen.GenerateCodeGenSection()
	if err != nil {
		t.Fatalf("GenerateCodeGenSection() error = %v", err)
	}

	// Should contain CLI usage examples
	if !strings.Contains(result, "cobrayaml init") {
		t.Error("Code gen section should contain init command")
	}
	if !strings.Contains(result, "cobrayaml gen") {
		t.Error("Code gen section should contain gen command")
	}

	// Should contain actual generated code (from running the generator)
	if !strings.Contains(result, "func runAdd") {
		t.Error("Code gen section should contain generated function")
	}
	if !strings.Contains(result, "GetBool") {
		t.Error("Code gen section should contain flag getter")
	}
}

func TestDocGenerator_AllSectionsCanBeGenerated(t *testing.T) {
	gen := NewDocGenerator()

	// This test verifies that all documentation sections can be generated
	// without errors - used by CI to ensure README generation will succeed

	yamlRef := gen.GenerateYAMLReference()
	if yamlRef == "" {
		t.Error("YAML reference generation returned empty string")
	}

	quickStart := gen.GenerateQuickStart()
	if quickStart == "" {
		t.Error("Quick start generation returned empty string")
	}

	codeGen, err := gen.GenerateCodeGenSection()
	if err != nil {
		t.Errorf("Code gen section generation failed: %v", err)
	}
	if codeGen == "" {
		t.Error("Code gen section generation returned empty string")
	}

	// Verify GenerateInitTemplate works
	initTemplate := GenerateInitTemplate("test-app")
	if initTemplate == "" {
		t.Error("Init template generation returned empty string")
	}
	if !strings.Contains(initTemplate, "test-app") {
		t.Error("Init template should contain the app name")
	}
}

func TestExtractFieldDocs(t *testing.T) {
	// Test that we can extract field docs from our structs
	toolFields := extractFieldDocs(reflect.TypeOf(ToolConfig{}))
	if len(toolFields) == 0 {
		t.Error("Should extract fields from ToolConfig")
	}

	cmdFields := extractFieldDocs(reflect.TypeOf(CommandConfig{}))
	if len(cmdFields) == 0 {
		t.Error("Should extract fields from CommandConfig")
	}

	flagFields := extractFieldDocs(reflect.TypeOf(FlagConfig{}))
	if len(flagFields) == 0 {
		t.Error("Should extract fields from FlagConfig")
	}
}

func TestFieldDocumentation_MatchesStructFields(t *testing.T) {
	// Verify that all struct fields have documentation
	toolFields := extractFieldDocs(reflect.TypeOf(ToolConfig{}))
	for _, f := range toolFields {
		if f.YAMLKey == "functions" {
			continue // internal field
		}
		desc := fieldDescription("ToolConfig", f.YAMLKey)
		if desc == "" {
			t.Errorf("ToolConfig field %q has no description", f.YAMLKey)
		}
	}

	cmdFields := extractFieldDocs(reflect.TypeOf(CommandConfig{}))
	for _, f := range cmdFields {
		desc := fieldDescription("CommandConfig", f.YAMLKey)
		if desc == "" {
			t.Errorf("CommandConfig field %q has no description", f.YAMLKey)
		}
	}

	flagFields := extractFieldDocs(reflect.TypeOf(FlagConfig{}))
	for _, f := range flagFields {
		desc := fieldDescription("FlagConfig", f.YAMLKey)
		if desc == "" {
			t.Errorf("FlagConfig field %q has no description", f.YAMLKey)
		}
	}
}
