package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"github.com/ramtinj/egenskriven/internal/resolver"
)

func newDeleteCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		force bool
	)

	cmd := &cobra.Command{
		Use:   "delete <task> [task...]",
		Short: "Delete tasks",
		Long: `Delete one or more tasks.

By default, asks for confirmation before deleting.
Use --force to skip confirmation (useful for scripts).

Examples:
  egenskriven delete abc123
  egenskriven delete abc123 def456 ghi789
  egenskriven delete abc123 --force`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Resolve all tasks first
			type taskInfo struct {
				ref  string
				task *core.Record
			}
			var tasksToDelete []taskInfo

			for _, ref := range args {
				task, err := resolver.MustResolve(app, ref)
				if err != nil {
					if ambErr, ok := err.(*resolver.AmbiguousError); ok {
						return out.AmbiguousError(ref, ambErr.Matches)
					}
					return out.Error(ExitNotFound, err.Error(), nil)
				}
				tasksToDelete = append(tasksToDelete, taskInfo{ref, task})
			}

			// Confirm deletion unless --force
			if !force && !jsonOutput {
				fmt.Printf("About to delete %d task(s):\n", len(tasksToDelete))
				for _, t := range tasksToDelete {
					shortId := t.task.Id
					if len(shortId) > 8 {
						shortId = shortId[:8]
					}
					fmt.Printf("  [%s] %s\n", shortId, t.task.GetString("title"))
				}
				fmt.Print("\nConfirm deletion? [y/N]: ")

				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))

				if response != "y" && response != "yes" {
					fmt.Println("Deletion cancelled.")
					return nil
				}
			}

			// Delete tasks
			var deleted int
			for _, t := range tasksToDelete {
				// Need to get the actual record for deletion
				record, err := app.FindRecordById("tasks", t.task.Id)
				if err != nil {
					continue
				}

				if err := app.Delete(record); err != nil {
					return out.Error(ExitGeneralError,
						fmt.Sprintf("failed to delete task %s: %v", t.ref, err), nil)
				}
				deleted++
			}

			if jsonOutput {
				out.Success(fmt.Sprintf("Deleted %d task(s)", deleted))
			} else {
				fmt.Printf("Deleted %d task(s)\n", deleted)
			}

			return nil
		},
	}

	// Define flags
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}
