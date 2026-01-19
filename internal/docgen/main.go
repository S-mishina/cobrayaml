//go:build ignore

// Package main generates documentation from actual code definitions.
// Run with: go generate ./...
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/S-mishina/cobrayaml"
)

func main() {
	gen := cobrayaml.NewDocGenerator()

	// Read README
	readme, err := os.ReadFile("README.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading README.md: %v\n", err)
		os.Exit(1)
	}

	content := string(readme)

	// Update YAML Reference section
	yamlRef := gen.GenerateYAMLReference()
	content = replaceSection(content,
		"<!-- YAML_REFERENCE_START -->",
		"<!-- YAML_REFERENCE_END -->",
		"\n\n"+yamlRef+"\n")

	// Update Quick Start section
	quickStart := gen.GenerateQuickStart()
	content = replaceSection(content,
		"<!-- QUICK_START_START -->",
		"<!-- QUICK_START_END -->",
		quickStart)

	// Update Code Generation section
	codeGen, err := gen.GenerateCodeGenSection()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code gen section: %v\n", err)
		os.Exit(1)
	}
	content = replaceSection(content,
		"<!-- CODE_GEN_START -->",
		"<!-- CODE_GEN_END -->",
		codeGen)

	// Write updated README
	if err := os.WriteFile("README.md", []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing README.md: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Documentation generated successfully from code.")
}

// replaceSection replaces content between start and end markers
func replaceSection(content, startMarker, endMarker, newContent string) string {
	startIdx := strings.Index(content, startMarker)
	endIdx := strings.Index(content, endMarker)

	if startIdx == -1 || endIdx == -1 {
		fmt.Fprintf(os.Stderr, "Warning: markers %s / %s not found\n", startMarker, endMarker)
		return content
	}

	return content[:startIdx+len(startMarker)] + newContent + content[endIdx:]
}
