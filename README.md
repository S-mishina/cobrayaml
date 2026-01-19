# cobrayaml

YAML-based command builder for [cobra](https://github.com/spf13/cobra) CLI applications.

## Install

```bash
go get github.com/S-mishina/cobrayaml
```

## Quick Start

<!-- QUICK_START_START -->

1. Create `commands.yaml`:

```yaml
name: "my-tool"
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
```

1. Create `main.go`:

```go
package main

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
```

1. Run:

```bash
go run . --help
go run . list
go run . add myitem
go run . --version
```

<!-- QUICK_START_END -->

## YAML Reference

<!-- YAML_REFERENCE_START -->

### Flag Types

| Type          | Go Type    | Example        |
| ------------- | ---------- | -------------- |
| `string`      | `string`   | `--name foo`   |
| `bool`        | `bool`     | `--debug`      |
| `int`         | `int`      | `--count 10`   |
| `stringSlice` | `[]string` | `--tags a,b,c` |

### Args Validation

| Type    | Description             | Config                            |
| ------- | ----------------------- | --------------------------------- |
| `none`  | No arguments allowed    | `type: none`                      |
| `any`   | Any number of arguments | `type: any`                       |
| `exact` | Exact number required   | `type: exact`, `count: N`         |
| `min`   | Minimum number          | `type: min`, `min: N`             |
| `max`   | Maximum number          | `type: max`, `max: N`             |
| `range` | Range of arguments      | `type: range`, `min: N`, `max: N` |

### ToolConfig (Root)

| YAML Key      | Type                       | Description                         |
| ------------- | -------------------------- | ----------------------------------- |
| `name`        | `string`                   | Tool name                           |
| `description` | `string`                   | Tool description                    |
| `version`     | `string`                   | Tool version (shown with --version) |
| `root`        | `CommandConfig`            | Root command configuration          |
| `commands`    | `map[string]CommandConfig` | Top-level subcommands               |

### CommandConfig

| YAML Key   | Type                       | Description                                            |
| ---------- | -------------------------- | ------------------------------------------------------ |
| `use`      | `string`                   | Command name and argument pattern (e.g., `add <name>`) |
| `aliases`  | `[]string`                 | Alternative command names                              |
| `short`    | `string`                   | Brief description shown in help                        |
| `long`     | `string`                   | Detailed description                                   |
| `args`     | `ArgsConfig`               | Argument validation configuration                      |
| `run_func` | `string`                   | Name of the handler function                           |
| `flags`    | `[]FlagConfig`             | List of flag definitions                               |
| `commands` | `map[string]CommandConfig` | Nested subcommands                                     |
| `hidden`   | `bool`                     | Hide command from help output                          |

### FlagConfig

| YAML Key     | Type     | Required | Description                                   |
| ------------ | -------- | -------- | --------------------------------------------- |
| `name`       | `string` | Yes      | Flag name (e.g., `namespace` for --namespace) |
| `shorthand`  | `string` |          | Short flag (e.g., `n` for -n)                 |
| `type`       | `string` | Yes      | Flag type (string, bool, int, stringSlice)    |
| `default`    | `string` |          | Default value                                 |
| `usage`      | `string` | Yes      | Description shown in help                     |
| `required`   | `bool`   |          | Mark flag as required                         |
| `persistent` | `bool`   |          | Inherit flag to all subcommands               |
| `hidden`     | `bool`   |          | Hide flag from help output                    |

### Hidden Commands/Flags

```yaml
commands:
  internal:
    use: internal
    short: Internal command
    hidden: true

root:
  flags:
    - name: debug
      type: bool
      hidden: true
```

<!-- YAML_REFERENCE_END -->

## License

MIT
