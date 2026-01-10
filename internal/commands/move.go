package commands

import (
	"fmt"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/resolver"
)

func newMoveCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		position int
		afterID  string
		beforeID string
	)

	cmd := &cobra.Command{
		Use:   "move <task> [column]",
		Short: "Move task to column/position",
		Long: `Move a task to a different column and/or position.

If a column is specified, the task is moved to that column.
Position can be controlled with --position, --after, or --before.

Position values:
  0  = top of column
  -1 = bottom of column (default)

Examples:
  egenskriven move abc123 in_progress
  egenskriven move abc123 todo --position 0
  egenskriven move abc123 --after def456
  egenskriven move abc123 --before ghi789`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			taskRef := args[0]

			// Resolve the task
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				if ambErr, ok := err.(*resolver.AmbiguousError); ok {
					return out.AmbiguousError(taskRef, ambErr.Matches)
				}
				return out.Error(ExitNotFound, err.Error(), nil)
			}

			// Determine target column
			currentColumn := task.GetString("column")
			targetColumn := currentColumn

			if len(args) > 1 {
				targetColumn = args[1]
				if !isValidColumn(targetColumn) {
					return out.Error(ExitValidation,
						fmt.Sprintf("invalid column '%s', must be one of: %v", targetColumn, ValidColumns), nil)
				}
			}

			// Calculate new position
			var newPosition float64

			if afterID != "" {
				// Resolve the reference task (supports display IDs like TST-4)
				refTask, err := resolver.MustResolve(app, afterID)
				if err != nil {
					if ambErr, ok := err.(*resolver.AmbiguousError); ok {
						return out.AmbiguousError(afterID, ambErr.Matches)
					}
					return out.Error(ExitNotFound,
						fmt.Sprintf("task not found: %s", afterID), nil)
				}

				// Position after specific task
				pos, err := GetPositionAfter(app, refTask.Id)
				if err != nil {
					return out.Error(ExitNotFound,
						fmt.Sprintf("task not found: %s", afterID), nil)
				}
				newPosition = pos

				// Get target column from reference task
				if len(args) < 2 {
					targetColumn = refTask.GetString("column")
				}
			} else if beforeID != "" {
				// Resolve the reference task (supports display IDs like TST-4)
				refTask, err := resolver.MustResolve(app, beforeID)
				if err != nil {
					if ambErr, ok := err.(*resolver.AmbiguousError); ok {
						return out.AmbiguousError(beforeID, ambErr.Matches)
					}
					return out.Error(ExitNotFound,
						fmt.Sprintf("task not found: %s", beforeID), nil)
				}

				// Position before specific task
				pos, err := GetPositionBefore(app, refTask.Id)
				if err != nil {
					return out.Error(ExitNotFound,
						fmt.Sprintf("task not found: %s", beforeID), nil)
				}
				newPosition = pos

				// Get target column from reference task
				if len(args) < 2 {
					targetColumn = refTask.GetString("column")
				}
			} else if position >= 0 {
				// Specific position index
				newPosition = GetPositionAtIndex(app, targetColumn, position)
			} else if position == -1 {
				// Explicitly bottom of column
				newPosition = GetNextPosition(app, targetColumn)
			} else {
				// Invalid negative position
				return out.Error(ExitValidation,
					fmt.Sprintf("invalid position %d, use 0 for top or -1 for bottom", position), nil)
			}

			// Track changes for history
			oldColumn := currentColumn
			oldPosition := task.GetFloat("position")

			// Update the task
			task.Set("column", targetColumn)
			task.Set("position", newPosition)

			// Add to history
			addHistoryEntry(task, "moved", "", map[string]any{
				"column": map[string]any{
					"from": oldColumn,
					"to":   targetColumn,
				},
				"position": map[string]any{
					"from": oldPosition,
					"to":   newPosition,
				},
			})

			if err := updateRecordHybrid(app, task, out); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to move task: %v", err), nil)
			}

			if targetColumn != oldColumn {
				out.Success(fmt.Sprintf("Moved task [%s] from %s to %s", shortID(task.Id), oldColumn, targetColumn))
			} else {
				out.Success(fmt.Sprintf("Repositioned task [%s] in %s", shortID(task.Id), targetColumn))
			}

			return nil
		},
	}

	// Define flags
	cmd.Flags().IntVarP(&position, "position", "", -1,
		"Position index (0=top, -1=bottom)")
	cmd.Flags().StringVar(&afterID, "after", "",
		"Position after this task")
	cmd.Flags().StringVar(&beforeID, "before", "",
		"Position before this task")

	return cmd
}

// addHistoryEntry appends an entry to the task's history.
func addHistoryEntry(task interface {
	Get(string) any
	Set(string, any)
}, action, agent string, changes any) {
	var history []map[string]any

	// Get existing history
	raw := task.Get("history")
	if raw != nil {
		if h, ok := raw.([]any); ok {
			for _, entry := range h {
				if m, ok := entry.(map[string]any); ok {
					history = append(history, m)
				}
			}
		}
	}

	// Create new entry
	entry := map[string]any{
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"action":       action,
		"actor":        "cli",
		"actor_detail": agent,
		"changes":      changes,
	}

	history = append(history, entry)
	task.Set("history", history)
}
