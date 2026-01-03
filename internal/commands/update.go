package commands

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"github.com/ramtinj/egenskriven/internal/resolver"
)

func newUpdateCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		title        string
		description  string
		taskType     string
		priority     string
		addLabels    []string
		removeLabels []string
	)

	cmd := &cobra.Command{
		Use:   "update <task>",
		Short: "Update task properties",
		Long: `Update properties of an existing task.

Only specified fields will be updated. Use empty string to clear optional fields.

Examples:
  egenskriven update abc123 --title "New title"
  egenskriven update abc123 --priority urgent
  egenskriven update abc123 --add-label critical --remove-label backlog
  egenskriven update abc123 --description ""  # clears description`,
		Args: cobra.ExactArgs(1),
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

			// Track changes
			changes := make(map[string]any)

			// Update title
			if cmd.Flags().Changed("title") {
				if title == "" {
					return out.Error(ExitValidation, "title cannot be empty", nil)
				}
				changes["title"] = map[string]any{
					"from": task.GetString("title"),
					"to":   title,
				}
				task.Set("title", title)
			}

			// Update description
			if cmd.Flags().Changed("description") {
				changes["description"] = map[string]any{
					"from": task.GetString("description"),
					"to":   description,
				}
				task.Set("description", description)
			}

			// Update type
			if cmd.Flags().Changed("type") {
				if !isValidType(taskType) {
					return out.Error(ExitValidation,
						fmt.Sprintf("invalid type '%s', must be one of: %v", taskType, ValidTypes), nil)
				}
				changes["type"] = map[string]any{
					"from": task.GetString("type"),
					"to":   taskType,
				}
				task.Set("type", taskType)
			}

			// Update priority
			if cmd.Flags().Changed("priority") {
				if !isValidPriority(priority) {
					return out.Error(ExitValidation,
						fmt.Sprintf("invalid priority '%s', must be one of: %v", priority, ValidPriorities), nil)
				}
				changes["priority"] = map[string]any{
					"from": task.GetString("priority"),
					"to":   priority,
				}
				task.Set("priority", priority)
			}

			// Update labels
			if len(addLabels) > 0 || len(removeLabels) > 0 {
				oldLabels := getTaskLabels(task)
				newLabels := updateLabels(oldLabels, addLabels, removeLabels)
				changes["labels"] = map[string]any{
					"from": oldLabels,
					"to":   newLabels,
				}
				task.Set("labels", newLabels)
			}

			// Check if any changes were made
			if len(changes) == 0 {
				return out.Error(ExitValidation, "no changes specified", nil)
			}

			// Add to history
			addHistoryEntry(task, "updated", "", changes)

			// Save the task
			if err := app.Save(task); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to update task: %v", err), nil)
			}

			out.Task(task, "Updated")
			return nil
		},
	}

	// Define flags
	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&description, "description", "", "New description")
	cmd.Flags().StringVarP(&taskType, "type", "t", "", "New type (bug, feature, chore)")
	cmd.Flags().StringVarP(&priority, "priority", "p", "", "New priority (low, medium, high, urgent)")
	cmd.Flags().StringSliceVar(&addLabels, "add-label", nil, "Add label (repeatable)")
	cmd.Flags().StringSliceVar(&removeLabels, "remove-label", nil, "Remove label (repeatable)")

	return cmd
}

// getTaskLabels extracts labels from a task record.
func getTaskLabels(task *core.Record) []string {
	// Use GetStringSlice which properly handles types.JSONRaw
	return task.GetStringSlice("labels")
}

// updateLabels adds and removes labels from a list.
func updateLabels(current, add, remove []string) []string {
	// Create a set of current labels
	labelSet := make(map[string]bool)
	for _, l := range current {
		labelSet[l] = true
	}

	// Remove labels
	for _, l := range remove {
		delete(labelSet, l)
	}

	// Add labels
	for _, l := range add {
		labelSet[l] = true
	}

	// Convert back to slice
	result := make([]string, 0, len(labelSet))
	for l := range labelSet {
		result = append(result, l)
	}

	return result
}
