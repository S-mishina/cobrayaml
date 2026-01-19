package cobrayaml

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
