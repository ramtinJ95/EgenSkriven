package commands

import (
	"encoding/json"
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/resolver"
)

func newUpdateCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		title           string
		description     string
		taskType        string
		priority        string
		addLabels       []string
		removeLabels    []string
		blockedBy       []string
		removeBlockedBy []string
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
  egenskriven update abc123 --blocked-by def456
  egenskriven update abc123 --remove-blocked-by def456
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
				oldLabels := task.GetStringSlice("labels")
				newLabels := updateLabels(oldLabels, addLabels, removeLabels)
				changes["labels"] = map[string]any{
					"from": oldLabels,
					"to":   newLabels,
				}
				task.Set("labels", newLabels)
			}

			// Update blocked_by
			if len(blockedBy) > 0 || len(removeBlockedBy) > 0 {
				oldBlockedBy := getTaskBlockedBy(task)

				// Resolve all blocking task references to full IDs (supports display IDs like TST-4)
				resolvedBlockedBy := make([]string, 0, len(blockedBy))
				for _, ref := range blockedBy {
					blockingTask, err := resolver.MustResolve(app, ref)
					if err != nil {
						if ambErr, ok := err.(*resolver.AmbiguousError); ok {
							return out.AmbiguousError(ref, ambErr.Matches)
						}
						return out.Error(ExitNotFound,
							fmt.Sprintf("blocking task not found: %s", ref), nil)
					}
					// Check for self-reference early and return error
					if blockingTask.Id == task.Id {
						return out.Error(ExitValidation, "task cannot block itself", nil)
					}
					resolvedBlockedBy = append(resolvedBlockedBy, blockingTask.Id)
				}

				// Resolve removeBlockedBy references too
				resolvedRemoveBlockedBy := make([]string, 0, len(removeBlockedBy))
				for _, ref := range removeBlockedBy {
					blockingTask, err := resolver.MustResolve(app, ref)
					if err != nil {
						if ambErr, ok := err.(*resolver.AmbiguousError); ok {
							return out.AmbiguousError(ref, ambErr.Matches)
						}
						return out.Error(ExitNotFound,
							fmt.Sprintf("blocking task not found: %s", ref), nil)
					}
					resolvedRemoveBlockedBy = append(resolvedRemoveBlockedBy, blockingTask.Id)
				}

				newBlockedBy := updateBlockedBy(task.Id, oldBlockedBy, resolvedBlockedBy, resolvedRemoveBlockedBy)

				// Validate that all blocking tasks exist and check for cycles
				for _, blockingID := range newBlockedBy {
					if blockingID == task.Id {
						return out.Error(ExitValidation, "task cannot block itself", nil)
					}
					blockingTask, err := app.FindRecordById("tasks", blockingID)
					if err != nil {
						return out.Error(ExitNotFound,
							fmt.Sprintf("blocking task not found: %s", blockingID), nil)
					}

					// Check for circular dependency
					if hasCircularDependency(app, task.Id, blockingTask) {
						return out.Error(ExitValidation,
							fmt.Sprintf("circular dependency detected: %s is already blocked by %s (directly or indirectly)",
								shortID(blockingID), shortID(task.Id)), nil)
					}
				}

				changes["blocked_by"] = map[string]any{
					"from": oldBlockedBy,
					"to":   newBlockedBy,
				}
				task.Set("blocked_by", newBlockedBy)
			}

			// Check if any changes were made
			if len(changes) == 0 {
				return out.Error(ExitValidation, "no changes specified", nil)
			}

			// Add to history
			addHistoryEntry(task, "updated", "", changes)

			// Save the task using hybrid approach (API first, then fallback to direct)
			if err := updateRecordHybrid(app, task, out); err != nil {
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
	cmd.Flags().StringSliceVar(&blockedBy, "blocked-by", nil, "Add blocking task ID (repeatable)")
	cmd.Flags().StringSliceVar(&removeBlockedBy, "remove-blocked-by", nil, "Remove blocking task ID (repeatable)")

	return cmd
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

// getTaskBlockedBy extracts blocked_by IDs from a task record.
func getTaskBlockedBy(task interface{ Get(string) any }) []string {
	raw := task.Get("blocked_by")
	if raw == nil {
		return []string{}
	}

	// Handle []any type (common from JSON)
	if ids, ok := raw.([]any); ok {
		result := make([]string, 0, len(ids))
		for _, id := range ids {
			if s, ok := id.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}

	// Handle []string type
	if ids, ok := raw.([]string); ok {
		return ids
	}

	// Handle types.JSONRaw (from database)
	// Try to unmarshal as JSON array of strings
	if jsonRaw, ok := raw.(interface{ String() string }); ok {
		var ids []string
		jsonStr := jsonRaw.String()
		if jsonStr == "" || jsonStr == "null" {
			return []string{}
		}
		if err := json.Unmarshal([]byte(jsonStr), &ids); err == nil {
			return ids
		}
	}

	return []string{}
}

// hasCircularDependency checks if adding a blocking relationship would create a cycle.
// It checks if targetID appears in the blocked_by chain of blockingTask.
func hasCircularDependency(app *pocketbase.PocketBase, targetID string, blockingTask interface {
	Get(string) any
	GetString(string) string
}) bool {
	// Use BFS to traverse the blocked_by chain
	visited := make(map[string]bool)
	queue := []string{}

	// Start with the blocking task's blocked_by list
	blockedBy := getTaskBlockedBy(blockingTask)
	queue = append(queue, blockedBy...)

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if visited[currentID] {
			continue
		}
		visited[currentID] = true

		// If we find the target task in the chain, there's a cycle
		if currentID == targetID {
			return true
		}

		// Get the blocked_by list of the current task
		currentTask, err := app.FindRecordById("tasks", currentID)
		if err != nil {
			continue // Task not found, skip
		}

		blockedBy := getTaskBlockedBy(currentTask)
		queue = append(queue, blockedBy...)
	}

	return false
}

// updateBlockedBy adds and removes blocking task IDs.
func updateBlockedBy(taskID string, current, add, remove []string) []string {
	// Create a set of current blocked_by
	blockedSet := make(map[string]bool)
	for _, id := range current {
		blockedSet[id] = true
	}

	// Remove IDs
	for _, id := range remove {
		delete(blockedSet, id)
	}

	// Add IDs (but not self)
	for _, id := range add {
		if id != taskID {
			blockedSet[id] = true
		}
	}

	// Convert back to slice
	result := make([]string, 0, len(blockedSet))
	for id := range blockedSet {
		result = append(result, id)
	}

	return result
}
