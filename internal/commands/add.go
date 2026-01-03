package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

func newAddCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		taskType  string
		priority  string
		column    string
		labels    []string
		customID  string
		createdBy string
		agentName string
	)

	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a new task",
		Long: `Add a new task to the kanban board.

The task will be created with the specified properties and added to the 
end of the target column.

Examples:
  egenskriven add "Implement dark mode"
  egenskriven add "Fix login crash" --type bug --priority urgent
  egenskriven add "Setup CI" --id ci-setup-001
  egenskriven add "Refactor auth" --agent claude`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app to ensure database is ready
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			title := args[0]

			// Validate inputs
			if !isValidType(taskType) {
				return out.Error(ExitValidation,
					fmt.Sprintf("invalid type '%s', must be one of: %v", taskType, ValidTypes), nil)
			}
			if !isValidPriority(priority) {
				return out.Error(ExitValidation,
					fmt.Sprintf("invalid priority '%s', must be one of: %v", priority, ValidPriorities), nil)
			}
			if !isValidColumn(column) {
				return out.Error(ExitValidation,
					fmt.Sprintf("invalid column '%s', must be one of: %v", column, ValidColumns), nil)
			}

			// Determine creator
			if createdBy == "" {
				if agentName != "" {
					createdBy = "agent"
				} else {
					// Detect if running in a TTY
					if fileInfo, _ := os.Stdin.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
						createdBy = "user"
					} else {
						createdBy = "cli"
					}
				}
			}

			// Find the tasks collection
			collection, err := app.FindCollectionByNameOrId("tasks")
			if err != nil {
				return out.Error(ExitGeneralError,
					"tasks collection not found - run migrations first", nil)
			}

			// Create the record
			record := core.NewRecord(collection)

			// Set custom ID if provided
			if customID != "" {
				// Check if task with this ID already exists (idempotency)
				existing, err := app.FindRecordById("tasks", customID)
				if err == nil {
					// Task exists, return it (idempotent behavior)
					out.Task(existing, "Existing")
					return nil
				}
				record.Id = customID
			}

			// Set task fields
			record.Set("title", title)
			record.Set("type", taskType)
			record.Set("priority", priority)
			record.Set("column", column)
			record.Set("position", GetNextPosition(app, column))
			record.Set("labels", labels)
			record.Set("blocked_by", []string{})
			record.Set("created_by", createdBy)
			if agentName != "" {
				record.Set("created_by_agent", agentName)
			}

			// Initialize history
			history := []map[string]any{
				{
					"timestamp":    time.Now().UTC().Format(time.RFC3339),
					"action":       "created",
					"actor":        createdBy,
					"actor_detail": agentName,
					"changes":      nil,
				},
			}
			record.Set("history", history)

			// Save the record
			if err := app.Save(record); err != nil {
				return out.Error(ExitGeneralError,
					fmt.Sprintf("failed to create task: %v", err), nil)
			}

			out.Task(record, "Created")
			return nil
		},
	}

	// Define flags
	cmd.Flags().StringVarP(&taskType, "type", "t", "feature",
		"Task type (bug, feature, chore)")
	cmd.Flags().StringVarP(&priority, "priority", "p", "medium",
		"Priority (low, medium, high, urgent)")
	cmd.Flags().StringVarP(&column, "column", "c", "backlog",
		"Initial column")
	cmd.Flags().StringSliceVarP(&labels, "label", "l", nil,
		"Labels (repeatable)")
	cmd.Flags().StringVar(&customID, "id", "",
		"Custom ID for idempotency")
	cmd.Flags().StringVar(&createdBy, "created-by", "",
		"Creator type (user, agent, cli)")
	cmd.Flags().StringVar(&agentName, "agent", "",
		"Agent identifier (implies --created-by agent)")

	return cmd
}
