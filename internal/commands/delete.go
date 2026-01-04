package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/resolver"
)

func newDeleteCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		force bool
		stdin bool
	)

	cmd := &cobra.Command{
		Use:   "delete <task> [task...]",
		Short: "Delete tasks",
		Long: `Delete one or more tasks.

By default, asks for confirmation before deleting.
Use --force to skip confirmation (useful for scripts).
Use --stdin to read task references from stdin (one per line).

Examples:
  egenskriven delete abc123
  egenskriven delete abc123 def456 ghi789
  egenskriven delete abc123 --force
  echo -e "abc123\ndef456" | egenskriven delete --stdin --force`,
		Args: cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Collect task references
			var refs []string

			if stdin {
				// Read from stdin
				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					ref := strings.TrimSpace(scanner.Text())
					if ref != "" {
						refs = append(refs, ref)
					}
				}
				if err := scanner.Err(); err != nil {
					return out.Error(ExitGeneralError, fmt.Sprintf("failed to read stdin: %v", err), nil)
				}
			} else {
				if len(args) == 0 {
					return out.Error(ExitInvalidArguments,
						"at least one task reference is required\n\nUsage: egenskriven delete <task> [task...]\n       echo 'task-id' | egenskriven delete --stdin", nil)
				}
				refs = args
			}

			if len(refs) == 0 {
				return out.Error(ExitInvalidArguments, "no tasks to delete", nil)
			}

			// Resolve all tasks first
			type taskInfo struct {
				ref  string
				task *core.Record
			}
			var tasksToDelete []taskInfo

			for _, ref := range refs {
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
			if !force && !out.Quiet && !out.JSON {
				fmt.Printf("About to delete %d task(s):\n", len(tasksToDelete))
				for _, t := range tasksToDelete {
					fmt.Printf("  [%s] %s\n", shortID(t.task.Id), t.task.GetString("title"))
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

			out.Success(fmt.Sprintf("Deleted %d task(s)", deleted))

			return nil
		},
	}

	// Define flags
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	cmd.Flags().BoolVar(&stdin, "stdin", false, "Read task references from stdin (one per line)")

	return cmd
}
