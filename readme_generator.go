package cobrayaml

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"
)

// CommandDoc holds documentation for a single command
type CommandDoc struct {
	Name        string
	Use         string
	Short       string
	Long        string
	FullPath    string
	Aliases     []string
	Flags       []FlagConfig
	Args        *ArgsConfig
	Subcommands []CommandDoc
	Depth       int
}

// DocsConfig holds all configuration needed for documentation generation
type DocsConfig struct {
	ToolName        string
	ToolDescription string
	Version         string
	RootCommand     CommandDoc
	Commands        []CommandDoc
}

const docsTemplate = `# {{ .ToolName }}

{{ if .ToolDescription }}{{ .ToolDescription }}{{ end }}

{{ if .Version }}**Version:** {{ .Version }}{{ end }}

## Installation

` + "```" + `bash
go install github.com/your-username/{{ .ToolName }}@latest
` + "```" + `

## Usage

` + "```" + `bash
{{ .RootCommand.Use }}{{ if .Commands }} [command]{{ end }}
` + "```" + `

{{ if .RootCommand.Long }}{{ .RootCommand.Long }}{{ end }}

{{ if .RootCommand.Flags }}### Global Flags

| Flag | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
{{ range .RootCommand.Flags }}| ` + "`" + `--{{ .Name }}` + "`" + ` | {{ if .Shorthand }}` + "`" + `-{{ .Shorthand }}` + "`" + `{{ end }} | {{ .Type }} | {{ if .DefaultValue }}` + "`" + `{{ .DefaultValue }}` + "`" + `{{ end }} | {{ .Usage }}{{ if .Required }} **(required)**{{ end }} |
{{ end }}{{ end }}

## Commands

{{ range .Commands }}{{ template "command" . }}{{ end }}
`

const commandTemplate = `{{ $heading := repeat "#" (add .Depth 3) }}{{ $heading }} {{ .Name }}

{{ .Short }}

` + "```" + `bash
{{ .FullPath }}
` + "```" + `

{{ if .Long }}{{ .Long }}

{{ end }}{{ if .Aliases }}**Aliases:** {{ join .Aliases ", " }}

{{ end }}{{ if .Args }}**Arguments:** {{ argsDescription .Args }}

{{ end }}{{ if .Flags }}**Flags:**

| Flag | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
{{ range .Flags }}| ` + "`" + `--{{ .Name }}` + "`" + ` | {{ if .Shorthand }}` + "`" + `-{{ .Shorthand }}` + "`" + `{{ end }} | {{ .Type }} | {{ if .DefaultValue }}` + "`" + `{{ .DefaultValue }}` + "`" + `{{ end }} | {{ .Usage }}{{ if .Required }} **(required)**{{ end }} |
{{ end }}{{ end }}{{ if .Subcommands }}
{{ range .Subcommands }}{{ template "command" . }}{{ end }}{{ end }}`

// GenerateDocs generates README documentation from the YAML configuration
func (g *Generator) GenerateDocs() (string, error) {
	config := g.collectDocsConfig()
	return renderDocsTemplate(config)
}

// GenerateDocsToFile generates README documentation and writes to file
func (g *Generator) GenerateDocsToFile(path string) error {
	docs, err := g.GenerateDocs()
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(docs), 0644)
}

// collectDocsConfig collects all documentation configuration from the tool config
func (g *Generator) collectDocsConfig() *DocsConfig {
	config := &DocsConfig{
		ToolName:        g.config.Name,
		ToolDescription: g.config.Description,
		Version:         g.config.Version,
	}

	// Collect root command documentation
	config.RootCommand = CommandDoc{
		Name:    g.config.Root.Use,
		Use:     g.config.Root.Use,
		Short:   g.config.Root.Short,
		Long:    g.config.Root.Long,
		Flags:   filterVisibleFlags(g.config.Root.Flags),
		Args:    g.config.Root.Args,
		Aliases: g.config.Root.Aliases,
		Depth:   0,
	}

	// Collect all commands
	var commands []CommandDoc

	// Get sorted command names for consistent output
	cmdNames := make([]string, 0, len(g.config.Commands))
	for name := range g.config.Commands {
		cmdNames = append(cmdNames, name)
	}
	sort.Strings(cmdNames)

	for _, name := range cmdNames {
		cmdConfig := g.config.Commands[name]
		if !cmdConfig.Hidden {
			commands = append(commands, g.collectCommandDoc(cmdConfig, name, 0))
		}
	}

	config.Commands = commands
	return config
}

// collectCommandDoc recursively collects documentation for a command and its subcommands
func (g *Generator) collectCommandDoc(cmd CommandConfig, name string, depth int) CommandDoc {
	// Extract the command name from Use field (first word)
	cmdName := name
	if fields := strings.Fields(cmd.Use); len(fields) > 0 {
		cmdName = fields[0]
	}

	doc := CommandDoc{
		Name:     cmdName,
		Use:      cmd.Use,
		Short:    cmd.Short,
		Long:     cmd.Long,
		FullPath: g.config.Root.Use + " " + cmd.Use,
		Flags:    filterVisibleFlags(cmd.Flags),
		Args:     cmd.Args,
		Aliases:  cmd.Aliases,
		Depth:    depth,
	}

	// Collect subcommands
	if len(cmd.Commands) > 0 {
		// Get sorted subcommand names for consistent output
		subNames := make([]string, 0, len(cmd.Commands))
		for subName := range cmd.Commands {
			subNames = append(subNames, subName)
		}
		sort.Strings(subNames)

		for _, subName := range subNames {
			subCmd := cmd.Commands[subName]
			if !subCmd.Hidden {
				subDoc := g.collectCommandDoc(subCmd, subName, depth+1)
				// Update full path for nested commands
				subCmdName := subName
				if fields := strings.Fields(subCmd.Use); len(fields) > 0 {
					subCmdName = fields[0]
				}
				subDoc.FullPath = doc.FullPath + " " + subCmdName
				doc.Subcommands = append(doc.Subcommands, subDoc)
			}
		}
	}

	return doc
}

// filterVisibleFlags returns only non-hidden flags
func filterVisibleFlags(flags []FlagConfig) []FlagConfig {
	var visible []FlagConfig
	for _, f := range flags {
		if !f.Hidden {
			visible = append(visible, f)
		}
	}
	return visible
}

// renderDocsTemplate renders the documentation template with the given config
func renderDocsTemplate(config *DocsConfig) (string, error) {
	funcMap := template.FuncMap{
		"join": strings.Join,
		"add": func(a, b int) int {
			return a + b
		},
		"repeat": func(s string, n int) string {
			return strings.Repeat(s, n)
		},
		"argsDescription": func(args *ArgsConfig) string {
			if args == nil {
				return ""
			}
			switch args.Type {
			case ArgsTypeNone:
				return "No arguments allowed"
			case ArgsTypeAny:
				return "Any number of arguments"
			case ArgsTypeExact:
				return fmt.Sprintf("Exactly %d argument(s) required", args.Count)
			case ArgsTypeMin:
				return fmt.Sprintf("At least %d argument(s) required", args.Min)
			case ArgsTypeMax:
				return fmt.Sprintf("At most %d argument(s) allowed", args.Max)
			case ArgsTypeRange:
				return fmt.Sprintf("%d to %d argument(s)", args.Min, args.Max)
			default:
				return ""
			}
		},
	}

	// Parse the command template first
	tmpl, err := template.New("docs").Funcs(funcMap).Parse(docsTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse docs template: %w", err)
	}

	// Parse the command template as a nested template
	tmpl, err = tmpl.New("command").Parse(commandTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse command template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "docs", config); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Clean up extra blank lines
	result := buf.String()
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}

	return result, nil
}
