package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"

	"github.com/ramtinJ95/EgenSkriven/internal/output"
)

var (
	// Global flags
	jsonOutput  bool
	quietMode   bool
	directMode  bool // Skip HTTP API, use direct database access
	verboseMode bool // Show detailed output including API/direct mode
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
	app.RootCmd.PersistentFlags().BoolVar(&directMode, "direct", false,
		"Skip HTTP API, use direct database access (faster offline, no real-time updates)")
	app.RootCmd.PersistentFlags().BoolVarP(&verboseMode, "verbose", "v", false,
		"Show detailed output including connection method")

	// Register all commands
	app.RootCmd.AddCommand(newAddCmd(app))
	app.RootCmd.AddCommand(newListCmd(app))
	app.RootCmd.AddCommand(newShowCmd(app))
	app.RootCmd.AddCommand(newMoveCmd(app))
	app.RootCmd.AddCommand(newUpdateCmd(app))
	app.RootCmd.AddCommand(newDeleteCmd(app))

	// Phase 1.5 commands
	app.RootCmd.AddCommand(newInitCmd(app))
	app.RootCmd.AddCommand(newPrimeCmd(app))
	app.RootCmd.AddCommand(newContextCmd(app))
	app.RootCmd.AddCommand(newSuggestCmd(app))

	// Phase 3 commands
	app.RootCmd.AddCommand(newEpicCmd(app))
	app.RootCmd.AddCommand(newVersionCmd())

	// Phase 5 commands
	app.RootCmd.AddCommand(newBoardCmd(app))

	// Phase 8 commands
	app.RootCmd.AddCommand(newExportCmd(app))
	app.RootCmd.AddCommand(newImportCmd(app))
	app.RootCmd.AddCommand(newBackupCmd(app))

	// Phase 9 commands
	app.RootCmd.AddCommand(newCompletionCmd(app.RootCmd))
	app.RootCmd.AddCommand(newSelfUpgradeCmd())

	// AI agent integration commands
	app.RootCmd.AddCommand(newSkillCmd(app))
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
var ValidColumns = []string{"backlog", "todo", "in_progress", "need_input", "review", "done"}

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

// shortID is a convenience wrapper around output.ShortID.
func shortID(id string) string {
	return output.ShortID(id)
}

// escapeLikePattern escapes SQL LIKE special characters (% and _)
// to treat them as literal characters in search patterns.
func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

// isDirectMode returns true if direct database access should be used.
func isDirectMode() bool {
	return directMode
}

// isVerboseMode returns true if verbose output is enabled.
func isVerboseMode() bool {
	return verboseMode
}

// verboseLog prints a message if verbose mode is enabled.
func verboseLog(format string, args ...any) {
	if verboseMode && !quietMode {
		fmt.Fprintf(os.Stderr, "[verbose] "+format+"\n", args...)
	}
}

// warnLog prints a warning message to stderr.
func warnLog(format string, args ...any) {
	if !quietMode {
		fmt.Fprintf(os.Stderr, "Warning: "+format+"\n", args...)
	}
}

// Note: All command functions are now implemented in separate files
