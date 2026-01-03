package commands

import (
	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/ramtinj/egenskriven/internal/output"
)

var (
	// Global flags
	jsonOutput bool
	quietMode  bool
	dataDir    string
)

// Register adds all CLI commands to the PocketBase app.
func Register(app *pocketbase.PocketBase) {
	// Create formatter that will be used by all commands
	// Note: The formatter is created fresh for each command execution
	// to respect the current flag values.

	// Add global flags to root command
	app.RootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false,
		"Output in JSON format")
	app.RootCmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false,
		"Suppress non-essential output")
	app.RootCmd.PersistentFlags().StringVar(&dataDir, "data", "",
		"Path to data directory (default: pb_data)")

	// Register all commands
	app.RootCmd.AddCommand(newAddCmd(app))
	app.RootCmd.AddCommand(newListCmd(app))
	app.RootCmd.AddCommand(newShowCmd(app))
	app.RootCmd.AddCommand(newMoveCmd(app))
	app.RootCmd.AddCommand(newUpdateCmd(app))
	app.RootCmd.AddCommand(newDeleteCmd(app))
}

// getFormatter creates a new output formatter with current flag values.
// This should be called at the start of each command's RunE function.
func getFormatter() *output.Formatter {
	return output.New(jsonOutput, quietMode)
}

// Exit codes for CLI commands
const (
	ExitSuccess          = 0
	ExitGeneralError     = 1
	ExitInvalidArguments = 2
	ExitNotFound         = 3
	ExitAmbiguous        = 4
	ExitValidation       = 5
)

// ValidColumns is the list of valid column values
var ValidColumns = []string{"backlog", "todo", "in_progress", "review", "done"}

// ValidTypes is the list of valid task types
var ValidTypes = []string{"bug", "feature", "chore"}

// ValidPriorities is the list of valid priority values
var ValidPriorities = []string{"low", "medium", "high", "urgent"}

// isValidColumn checks if a column name is valid.
func isValidColumn(col string) bool {
	for _, valid := range ValidColumns {
		if col == valid {
			return true
		}
	}
	return false
}

// isValidType checks if a type is valid.
func isValidType(t string) bool {
	for _, valid := range ValidTypes {
		if t == valid {
			return true
		}
	}
	return false
}

// isValidPriority checks if a priority is valid.
func isValidPriority(p string) bool {
	for _, valid := range ValidPriorities {
		if p == valid {
			return true
		}
	}
	return false
}

// Placeholder command functions - will be replaced as commands are implemented

func newMoveCmd(app *pocketbase.PocketBase) *cobra.Command {
	return &cobra.Command{
		Use:   "move <task> [column]",
		Short: "Move task to column/position",
	}
}

func newUpdateCmd(app *pocketbase.PocketBase) *cobra.Command {
	return &cobra.Command{
		Use:   "update <task>",
		Short: "Update task properties",
	}
}

func newDeleteCmd(app *pocketbase.PocketBase) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <task> [task...]",
		Short: "Delete tasks",
	}
}
