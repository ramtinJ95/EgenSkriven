package commands

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/ramtinj/egenskriven/internal/resolver"
)

func newShowCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <task>",
		Short: "Show task details",
		Long: `Show detailed information about a task.

The task can be referenced by:
- Full ID: abc123def456
- ID prefix: abc123 (must be unique)
- Title: "fix login" (case-insensitive, must be unique)

Examples:
  egenskriven show abc123
  egenskriven show "login crash"
  egenskriven show abc123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			ref := args[0]

			// Resolve the task
			resolution, err := resolver.ResolveTask(app, ref)
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to resolve task: %v", err), nil)
			}

			if resolution.IsNotFound() {
				return out.Error(ExitNotFound, fmt.Sprintf("no task found matching: %s", ref), nil)
			}

			if resolution.IsAmbiguous() {
				return out.AmbiguousError(ref, resolution.Matches)
			}

			out.TaskDetail(resolution.Task)
			return nil
		},
	}

	return cmd
}
