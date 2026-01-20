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
	rootCmd.AddCommand(docsCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func genCmd() *cobra.Command {
	var (
		packageName    string
		outputPath     string
		mainOutputPath string
		force          bool
	)

	cmd := &cobra.Command{
		Use:   "gen <commands.yaml>",
		Short: "Generate handler function stubs and main.go from YAML",
		Long: `Generate Go handler function stubs and main.go based on the run_func definitions in your YAML file.

Example:
  cobrayaml gen commands.yaml
  cobrayaml gen commands.yaml -p mypackage -o handlers.go -m main.go
  cobrayaml gen commands.yaml --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			yamlPath := args[0]

			gen, err := cobrayaml.NewGenerator(yamlPath)
			if err != nil {
				return fmt.Errorf("failed to load YAML: %w", err)
			}

			dir := filepath.Dir(yamlPath)
			if outputPath == "" {
				outputPath = filepath.Join(dir, "handlers.go")
			}
			if mainOutputPath == "" {
				mainOutputPath = filepath.Join(dir, "main.go")
			}

			// Check if files already exist
			handlersExist := false
			mainExist := false
			if _, err := os.Stat(outputPath); err == nil {
				handlersExist = true
			}
			if _, err := os.Stat(mainOutputPath); err == nil {
				mainExist = true
			}

			if (handlersExist || mainExist) && !force {
				var existingFiles []string
				if handlersExist {
					existingFiles = append(existingFiles, outputPath)
				}
				if mainExist {
					existingFiles = append(existingFiles, mainOutputPath)
				}
				fmt.Printf("Warning: %v already exist(s). Use --force to overwrite.\n", existingFiles)
				fmt.Println("Generated code preview:")
				fmt.Println("------------------------")
				fmt.Println("// handlers.go")
				code, err := gen.GenerateHandlers(packageName)
				if err != nil {
					return err
				}
				fmt.Println(code)
				fmt.Println("// main.go")
				mainCode, err := gen.GenerateMain(packageName, filepath.Base(yamlPath))
				if err != nil {
					return err
				}
				fmt.Println(mainCode)
				return nil
			}

			// Generate handlers.go
			if !handlersExist || force {
				if err := gen.GenerateHandlersToFile(packageName, outputPath); err != nil {
					return fmt.Errorf("failed to generate handlers: %w", err)
				}
				fmt.Printf("Generated handlers at: %s\n", outputPath)
			}

			// Generate main.go
			if !mainExist || force {
				if err := gen.GenerateMainToFile(packageName, filepath.Base(yamlPath), mainOutputPath); err != nil {
					return fmt.Errorf("failed to generate main: %w", err)
				}
				fmt.Printf("Generated main at: %s\n", mainOutputPath)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&packageName, "package", "p", "main", "Package name for generated code")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path for handlers (default: handlers.go)")
	cmd.Flags().StringVarP(&mainOutputPath, "main", "m", "", "Output file path for main.go (default: main.go)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files")

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
			fmt.Println("  4. Run: go run . [command]")
			return nil
		},
	}

	return cmd
}

func docsCmd() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "docs <commands.yaml>",
		Short: "Generate README documentation from YAML",
		Long: `Generate comprehensive README documentation based on your YAML configuration.

Example:
  cobrayaml docs commands.yaml
  cobrayaml docs commands.yaml -o README.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			yamlPath := args[0]

			gen, err := cobrayaml.NewGenerator(yamlPath)
			if err != nil {
				return fmt.Errorf("failed to load YAML: %w", err)
			}

			if outputPath == "" {
				// Output to stdout
				docs, err := gen.GenerateDocs()
				if err != nil {
					return fmt.Errorf("failed to generate docs: %w", err)
				}
				fmt.Print(docs)
				return nil
			}

			// Output to file
			if err := gen.GenerateDocsToFile(outputPath); err != nil {
				return fmt.Errorf("failed to generate docs: %w", err)
			}

			fmt.Printf("Generated documentation at: %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: stdout)")

	return cmd
}
