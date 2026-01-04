package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/output"
)

// TaskInput represents a task for batch creation
type TaskInput struct {
	ID          string   `json:"id,omitempty"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	Column      string   `json:"column,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Epic        string   `json:"epic,omitempty"`
}

func newAddCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		taskType  string
		priority  string
		column    string
		labels    []string
		customID  string
		createdBy string
		agentName string
		epic      string
		stdin     bool
		file      string
	)

	cmd := &cobra.Command{
		Use:   "add [title]",
		Short: "Add a new task",
		Long: `Add a new task to the kanban board.

Supports batch creation via --stdin or --file for agent workflows.
Batch input accepts JSON lines (one JSON object per line) or a JSON array.

Examples:
  egenskriven add "Implement dark mode"
  egenskriven add "Fix login crash" --type bug --priority urgent
  egenskriven add "Setup CI" --id ci-setup-001
  egenskriven add "Refactor auth" --agent claude
  egenskriven add "Add login" --epic "Auth Refactor"
  
  # Batch from stdin (JSON lines)
  echo '{"title":"Task 1"}
{"title":"Task 2","priority":"high"}' | egenskriven add --stdin
  
  # Batch from file
  egenskriven add --file tasks.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app to ensure database is ready
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Handle batch input
			if stdin || file != "" {
				return addBatch(app, out, stdin, file, agentName)
			}

			// Single task creation requires title argument
			if len(args) == 0 {
				return out.Error(ExitInvalidArguments,
					"title is required\n\nUsage: egenskriven add <title>\n       egenskriven add --stdin < tasks.json", nil)
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

			// Determine and validate creator
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
			} else if createdBy != "user" && createdBy != "agent" && createdBy != "cli" {
				return out.Error(ExitValidation,
					fmt.Sprintf("invalid created-by '%s', must be one of: user, agent, cli", createdBy), nil)
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

			// Resolve epic if provided
			var epicID string
			if epic != "" {
				epicRecord, err := resolveEpic(app, epic)
				if err != nil {
					return out.Error(ExitValidation, fmt.Sprintf("invalid epic: %v", err), nil)
				}
				epicID = epicRecord.Id
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
			if epicID != "" {
				record.Set("epic", epicID)
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
	cmd.Flags().StringVarP(&epic, "epic", "e", "",
		"Link task to epic (ID or title)")
	cmd.Flags().BoolVar(&stdin, "stdin", false,
		"Read tasks from stdin (JSON lines or array)")
	cmd.Flags().StringVarP(&file, "file", "f", "",
		"Read tasks from JSON file")

	return cmd
}

// addBatch handles batch task creation from stdin or file
func addBatch(app *pocketbase.PocketBase, out *output.Formatter, useStdin bool, filePath string, agent string) error {
	var reader io.Reader

	if useStdin {
		reader = os.Stdin
	} else {
		f, err := os.Open(filePath)
		if err != nil {
			return out.Error(ExitGeneralError, fmt.Sprintf("failed to open file: %v", err), nil)
		}
		defer f.Close()
		reader = f
	}

	// Read all content to detect format
	content, err := io.ReadAll(reader)
	if err != nil {
		return out.Error(ExitGeneralError, fmt.Sprintf("failed to read input: %v", err), nil)
	}

	trimmed := strings.TrimSpace(string(content))
	if trimmed == "" {
		return out.Error(ExitInvalidArguments, "empty input", nil)
	}

	var inputs []TaskInput

	// Detect format: JSON array or JSON lines
	if strings.HasPrefix(trimmed, "[") {
		// JSON array format
		if err := json.Unmarshal([]byte(trimmed), &inputs); err != nil {
			return out.Error(ExitInvalidArguments, fmt.Sprintf("invalid JSON array: %v", err), nil)
		}
	} else {
		// JSON lines format
		scanner := bufio.NewScanner(strings.NewReader(trimmed))
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var input TaskInput
			if err := json.Unmarshal([]byte(line), &input); err != nil {
				return out.Error(ExitInvalidArguments,
					fmt.Sprintf("line %d: invalid JSON: %v", lineNum, err), nil)
			}
			inputs = append(inputs, input)
		}
		if err := scanner.Err(); err != nil {
			return out.Error(ExitGeneralError, fmt.Sprintf("failed to read input: %v", err), nil)
		}
	}

	if len(inputs) == 0 {
		return out.Error(ExitInvalidArguments, "no tasks found in input", nil)
	}

	// Find the tasks collection
	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		return out.Error(ExitGeneralError, "tasks collection not found - run migrations first", nil)
	}

	// Create all tasks
	var created []*core.Record
	var errors []string

	for i, input := range inputs {
		if input.Title == "" {
			errors = append(errors, fmt.Sprintf("task %d: title is required", i+1))
			continue
		}

		// Validate type, priority, column with defaults
		taskType := defaultString(input.Type, "feature")
		if !isValidType(taskType) {
			errors = append(errors, fmt.Sprintf("task %d (%s): invalid type '%s', must be one of: %v",
				i+1, input.Title, taskType, ValidTypes))
			continue
		}

		priority := defaultString(input.Priority, "medium")
		if !isValidPriority(priority) {
			errors = append(errors, fmt.Sprintf("task %d (%s): invalid priority '%s', must be one of: %v",
				i+1, input.Title, priority, ValidPriorities))
			continue
		}

		column := defaultString(input.Column, "backlog")
		if !isValidColumn(column) {
			errors = append(errors, fmt.Sprintf("task %d (%s): invalid column '%s', must be one of: %v",
				i+1, input.Title, column, ValidColumns))
			continue
		}

		record := core.NewRecord(collection)

		if input.ID != "" {
			// Check idempotency
			existing, err := app.FindRecordById("tasks", input.ID)
			if err == nil {
				created = append(created, existing)
				continue
			}
			record.Id = input.ID
		}

		// Set fields with validated values
		record.Set("title", input.Title)
		record.Set("type", taskType)
		record.Set("priority", priority)
		record.Set("column", column)
		record.Set("position", GetNextPosition(app, column))
		record.Set("labels", input.Labels)
		record.Set("blocked_by", []string{})

		if input.Description != "" {
			record.Set("description", input.Description)
		}

		// Handle epic
		if input.Epic != "" {
			epicRecord, err := resolveEpic(app, input.Epic)
			if err != nil {
				errors = append(errors, fmt.Sprintf("task %d (%s): invalid epic '%s': %v",
					i+1, input.Title, input.Epic, err))
				continue
			}
			record.Set("epic", epicRecord.Id)
		}

		// Set creator info
		createdBy := "cli"
		if agent != "" {
			createdBy = "agent"
			record.Set("created_by_agent", agent)
		}
		record.Set("created_by", createdBy)

		// Initialize history
		history := []map[string]any{
			{
				"timestamp":    time.Now().UTC().Format(time.RFC3339),
				"action":       "created",
				"actor":        createdBy,
				"actor_detail": agent,
				"changes":      nil,
			},
		}
		record.Set("history", history)

		if err := app.Save(record); err != nil {
			errors = append(errors, fmt.Sprintf("task %d (%s): failed to save: %v", i+1, input.Title, err))
			continue
		}
		created = append(created, record)
	}

	// Output results
	if out.JSON {
		tasks := make([]map[string]any, 0, len(created))
		for _, record := range created {
			tasks = append(tasks, map[string]any{
				"id":       record.Id,
				"title":    record.GetString("title"),
				"type":     record.GetString("type"),
				"priority": record.GetString("priority"),
				"column":   record.GetString("column"),
			})
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"created": len(created),
			"failed":  len(errors),
			"tasks":   tasks,
			"errors":  errors,
		})
	}

	// Human output
	for _, record := range created {
		fmt.Printf("Created: %s [%s]\n", record.GetString("title"), shortID(record.Id))
	}

	if len(errors) > 0 {
		fmt.Println("\nErrors:")
		for _, e := range errors {
			fmt.Printf("  %s\n", e)
		}
	}

	fmt.Printf("\nCreated %d tasks", len(created))
	if len(errors) > 0 {
		fmt.Printf(", %d failed", len(errors))
	}
	fmt.Println()

	return nil
}

// defaultString returns the value if non-empty, otherwise the default
func defaultString(value, defaultVal string) string {
	if value == "" {
		return defaultVal
	}
	return value
}
