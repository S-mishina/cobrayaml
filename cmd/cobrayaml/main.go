package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/S-mishina/cobrayaml"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "cobrayaml",
		Short:   "YAML-based command builder for cobra CLI applications",
		Version: version,
	}

	rootCmd.AddCommand(genCmd())
	rootCmd.AddCommand(initCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func genCmd() *cobra.Command {
	var (
		packageName string
		outputPath  string
	)

	cmd := &cobra.Command{
		Use:   "gen <commands.yaml>",
		Short: "Generate handler function stubs from YAML",
		Long: `Generate Go handler function stubs based on the run_func definitions in your YAML file.

Example:
  cobrayaml gen commands.yaml
  cobrayaml gen commands.yaml -p mypackage -o handlers.go`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			yamlPath := args[0]

			gen, err := cobrayaml.NewGenerator(yamlPath)
			if err != nil {
				return fmt.Errorf("failed to load YAML: %w", err)
			}

			if outputPath == "" {
				// Default output path: same directory as YAML, named handlers.go
				dir := filepath.Dir(yamlPath)
				outputPath = filepath.Join(dir, "handlers.go")
			}

			// Check if file already exists
			if _, err := os.Stat(outputPath); err == nil {
				fmt.Printf("Warning: %s already exists. Use --force to overwrite.\n", outputPath)
				fmt.Println("Generated code preview:")
				fmt.Println("------------------------")
				code, err := gen.GenerateHandlers(packageName)
				if err != nil {
					return err
				}
				fmt.Println(code)
				return nil
			}

			if err := gen.GenerateHandlersToFile(packageName, outputPath); err != nil {
				return fmt.Errorf("failed to generate handlers: %w", err)
			}

			fmt.Printf("Generated handlers at: %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&packageName, "package", "p", "main", "Package name for generated code")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: handlers.go)")

	return cmd
}

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Create a new commands.yaml template",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "my-tool"
			if len(args) > 0 {
				name = args[0]
			}

			// Generate template from actual types
			template := cobrayaml.GenerateInitTemplate(name)

			outputPath := "commands.yaml"
			if _, err := os.Stat(outputPath); err == nil {
				return fmt.Errorf("%s already exists", outputPath)
			}

			if err := os.WriteFile(outputPath, []byte(template), 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			fmt.Printf("Created %s\n", outputPath)
			fmt.Println("\nNext steps:")
			fmt.Println("  1. Edit commands.yaml to define your CLI structure")
			fmt.Println("  2. Run: cobrayaml gen commands.yaml")
			fmt.Println("  3. Implement your handler functions in handlers.go")
			return nil
		},
	}

	return cmd
}
