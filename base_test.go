package cobrayaml

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewBaseCommand(t *testing.T) {
	tests := []struct {
		name      string
		use       string
		short     string
		long      string
		wantUse   string
		wantShort string
		wantLong  string
	}{
		{
			name:      "basic command",
			use:       "test",
			short:     "Short description",
			long:      "Long description for test command",
			wantUse:   "test",
			wantShort: "Short description",
			wantLong:  "Long description for test command",
		},
		{
			name:      "empty descriptions",
			use:       "empty",
			short:     "",
			long:      "",
			wantUse:   "empty",
			wantShort: "",
			wantLong:  "",
		},
		{
			name:      "complex command",
			use:       "complex [flags]",
			short:     "A complex command",
			long:      "This is a more complex command with multiple features",
			wantUse:   "complex [flags]",
			wantShort: "A complex command",
			wantLong:  "This is a more complex command with multiple features",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := NewBaseCommand(tt.use, tt.short, tt.long)

			if base == nil {
				t.Fatal("NewBaseCommand returned nil")
			}

			if base.RootCmd == nil {
				t.Fatal("RootCmd is nil")
			}

			if base.RootCmd.Use != tt.wantUse {
				t.Errorf("Use = %q, want %q", base.RootCmd.Use, tt.wantUse)
			}

			if base.RootCmd.Short != tt.wantShort {
				t.Errorf("Short = %q, want %q", base.RootCmd.Short, tt.wantShort)
			}

			if base.RootCmd.Long != tt.wantLong {
				t.Errorf("Long = %q, want %q", base.RootCmd.Long, tt.wantLong)
			}
		})
	}
}

func TestBaseCommand_CommonFlags(t *testing.T) {
	base := NewBaseCommand("test", "short", "long")

	// Check that common flags are set up
	configFlag := base.RootCmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Error("config flag not found")
	} else {
		if configFlag.Usage == "" {
			t.Error("config flag has no usage description")
		}
	}

	configPathFlag := base.RootCmd.PersistentFlags().Lookup("config-path")
	if configPathFlag == nil {
		t.Error("config-path flag not found")
	} else {
		if configPathFlag.Usage == "" {
			t.Error("config-path flag has no usage description")
		}
	}
}

func TestBaseCommand_AddCommand(t *testing.T) {
	base := NewBaseCommand("parent", "parent command", "parent long")

	// Create a subcommand
	subCmd := NewBaseCommand("child", "child command", "child long")

	// Add the subcommand
	base.AddCommand(subCmd.RootCmd)

	// Verify the subcommand was added
	commands := base.RootCmd.Commands()
	if len(commands) != 1 {
		t.Errorf("expected 1 subcommand, got %d", len(commands))
	}

	if commands[0].Use != "child" {
		t.Errorf("subcommand Use = %q, want %q", commands[0].Use, "child")
	}
}

func TestBaseCommand_Execute(t *testing.T) {
	base := NewBaseCommand("test", "short", "long")

	// Set a simple Run function
	executed := false
	base.RootCmd.Run = func(cmd *cobra.Command, args []string) {
		executed = true
	}

	// Execute should not return an error
	base.RootCmd.SetArgs([]string{})
	err := base.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !executed {
		t.Error("command was not executed")
	}
}
