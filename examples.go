package cobrayaml

import "gopkg.in/yaml.v2"

// ExampleCommandsYAML is the example YAML configuration used in documentation.
// This is also used in tests to ensure the example stays valid.
const ExampleCommandsYAML = `name: "my-tool"
version: "1.0.0"
root:
  use: "my-tool"
  short: "My CLI tool"
  flags:
    - name: "config"
      shorthand: "c"
      type: "string"
      usage: "Config file path"
      persistent: true

commands:
  list:
    use: "list"
    short: "List items"
    args:
      type: none
    run_func: "runList"

  add:
    use: "add <name>"
    short: "Add an item"
    args:
      type: exact
      count: 1
    run_func: "runAdd"

  delete:
    use: "delete <name>"
    short: "Delete an item"
    aliases:
      - rm
    args:
      type: exact
      count: 1
    run_func: "runDelete"
`

// ExampleMainGo is the example main.go code used in documentation.
const ExampleMainGo = `package main

import (
    _ "embed"
    "fmt"

    "github.com/S-mishina/cobrayaml"
    "github.com/spf13/cobra"
)

//go:embed commands.yaml
var commandsYAML string

func main() {
    builder, _ := cobrayaml.NewCommandBuilderFromString(commandsYAML)

    builder.RegisterFunction("runList", runList)
    builder.RegisterFunction("runAdd", runAdd)
    builder.RegisterFunction("runDelete", runDelete)

    rootCmd, _ := builder.BuildRootCommand()
    rootCmd.Execute()
}

func runList(cmd *cobra.Command, args []string) error {
    fmt.Println("Listing items...")
    return nil
}

func runAdd(cmd *cobra.Command, args []string) error {
    fmt.Printf("Adding: %s\n", args[0])
    return nil
}

func runDelete(cmd *cobra.Command, args []string) error {
    fmt.Printf("Deleting: %s\n", args[0])
    return nil
}
`

// GenerateInitTemplate generates a commands.yaml template for the given tool name.
// This ensures the template always matches the current YAML schema.
func GenerateInitTemplate(name string) string {
	config := ToolConfig{
		Name:    name,
		Version: "0.1.0",
		Root: CommandConfig{
			Use:   name,
			Short: name + " CLI",
			Flags: []FlagConfig{
				{
					Name:       "config",
					Shorthand:  "c",
					Type:       FlagTypeString,
					Usage:      "Config file path",
					Persistent: true,
				},
			},
		},
		Commands: map[string]CommandConfig{
			"hello": {
				Use:     "hello <name>",
				Short:   "Say hello",
				RunFunc: "runHello",
				Args: &ArgsConfig{
					Type:  ArgsTypeExact,
					Count: 1,
				},
				Flags: []FlagConfig{
					{
						Name:      "loud",
						Shorthand: "l",
						Type:      FlagTypeBool,
						Usage:     "Say it loudly",
					},
				},
			},
		},
	}

	return config.ToYAML()
}

// ToYAML converts ToolConfig to YAML string
func (c *ToolConfig) ToYAML() string {
	data, err := yaml.Marshal(c)
	if err != nil {
		return ""
	}
	return string(data)
}
