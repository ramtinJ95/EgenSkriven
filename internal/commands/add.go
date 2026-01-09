package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
	"github.com/ramtinJ95/EgenSkriven/internal/config"
	"github.com/ramtinJ95/EgenSkriven/internal/output"
	"github.com/ramtinJ95/EgenSkriven/internal/resolver"
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
	DueDate     string   `json:"due_date,omitempty"`
	Parent      string   `json:"parent,omitempty"`
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
		boardRef  string
		dueDate   string
		parent    string
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
				return addBatch(app, out, stdin, file, agentName, boardRef)
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
				// Validate custom ID format
				if !isValidCustomID(customID) {
					return out.Error(ExitValidation, formatCustomIDError(customID), nil)
				}
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

			// Resolve board
			boardRecord, err := resolveBoard(app, boardRef)
			if err != nil {
				return out.Error(ExitValidation, fmt.Sprintf("invalid board: %v", err), nil)
			}

			// Get next sequence number for this board
			var seq int
			if boardRecord != nil {
				seq, err = board.GetNextSequence(app, boardRecord.Id)
				if err != nil {
					return out.Error(ExitGeneralError, fmt.Sprintf("failed to get sequence: %v", err), nil)
				}
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
			if boardRecord != nil {
				record.Set("board", boardRecord.Id)
				record.Set("seq", seq)
			}

			// Handle due date
			if dueDate != "" {
				parsedDate, err := parseDate(dueDate)
				if err != nil {
					return out.Error(ExitValidation, fmt.Sprintf("invalid due date: %v", err), nil)
				}
				record.Set("due_date", parsedDate)
			}

			// Handle parent (sub-task)
			if parent != "" {
				// Resolve parent task using full resolver (supports display IDs like TST-4)
				parentTask, err := resolver.MustResolve(app, parent)
				if err != nil {
					if ambErr, ok := err.(*resolver.AmbiguousError); ok {
						return out.AmbiguousError(parent, ambErr.Matches)
					}
					return out.Error(ExitValidation, fmt.Sprintf("invalid parent: %v", err), nil)
				}
				record.Set("parent", parentTask.Id)
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

			// Save the record using hybrid approach (API first, then fallback to direct)
			if err := saveRecordHybrid(app, record, out); err != nil {
				return out.Error(ExitGeneralError,
					fmt.Sprintf("failed to create task: %v", err), nil)
			}

			// Format display ID for output
			var displayID string
			if boardRecord != nil {
				displayID = board.FormatDisplayID(boardRecord.GetString("prefix"), seq)
			} else {
				displayID = output.ShortID(record.Id)
			}

			if out.JSON {
				out.Task(record, "Created")
			} else {
				fmt.Printf("Created: %s [%s]\n", title, displayID)
			}
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
		"Custom ID for idempotency (must be exactly 15 lowercase alphanumeric chars)")
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
	cmd.Flags().StringVarP(&boardRef, "board", "b", "",
		"Board to create task in (name or prefix)")
	cmd.Flags().StringVar(&dueDate, "due", "",
		"Due date (ISO 8601 format: YYYY-MM-DD, or relative: 'tomorrow', 'next week')")
	cmd.Flags().StringVar(&parent, "parent", "",
		"Parent task ID (creates sub-task)")

	return cmd
}

// resolveBoard determines which board to use for task creation
// Priority: 1) explicit boardRef, 2) config default, 3) create/use first board
func resolveBoard(app *pocketbase.PocketBase, boardRef string) (*core.Record, error) {
	// If explicit board reference provided, use it
	if boardRef != "" {
		return board.GetByNameOrPrefix(app, boardRef)
	}

	// Check config for default board
	cfg, _ := config.LoadProjectConfig()
	if cfg != nil && cfg.DefaultBoard != "" {
		record, err := board.GetByNameOrPrefix(app, cfg.DefaultBoard)
		if err == nil {
			return record, nil
		}
		// Default board not found, fall through
	}

	// Get existing boards
	boards, err := board.GetAll(app)
	if err != nil || len(boards) == 0 {
		// No boards exist - create a default board
		b, err := board.Create(app, board.CreateInput{
			Name:   "Default",
			Prefix: "DEF",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create default board: %w", err)
		}
		return app.FindRecordById("boards", b.ID)
	}

	// Use first board
	return boards[0], nil
}

// resolveTaskByID finds a task by ID or ID prefix
func resolveTaskByID(app *pocketbase.PocketBase, ref string) (*core.Record, error) {
	// Try exact ID match
	record, err := app.FindRecordById("tasks", ref)
	if err == nil {
		return record, nil
	}

	// Try ID prefix match (for short IDs like "abc1234")
	records, err := app.FindAllRecords("tasks",
		dbx.NewExp("id LIKE {:prefix}", dbx.Params{"prefix": ref + "%"}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}

	switch len(records) {
	case 0:
		return nil, fmt.Errorf("no task found with ID: %s", ref)
	case 1:
		return records[0], nil
	default:
		var matches []string
		for _, r := range records {
			matches = append(matches, fmt.Sprintf("[%s] %s", shortID(r.Id), r.GetString("title")))
		}
		return nil, fmt.Errorf("ambiguous task ID prefix '%s' matches multiple tasks:\n  %s",
			ref, strings.Join(matches, "\n  "))
	}
}

// addBatch handles batch task creation from stdin or file
func addBatch(app *pocketbase.PocketBase, out *output.Formatter, useStdin bool, filePath string, agent string, boardRef string) error {
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

	// Resolve board for batch
	boardRecord, err := resolveBoard(app, boardRef)
	if err != nil {
		return out.Error(ExitValidation, fmt.Sprintf("invalid board: %v", err), nil)
	}

	// Create all tasks
	var created []*core.Record
	var createdDisplayIDs []string
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
			// Validate custom ID format
			if !isValidCustomID(input.ID) {
				errors = append(errors, fmt.Sprintf("task %d (%s): %s",
					i+1, input.Title, formatCustomIDError(input.ID)))
				continue
			}
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

		// Set board and sequence
		var displayID string
		if boardRecord != nil {
			seq, err := board.GetNextSequence(app, boardRecord.Id)
			if err != nil {
				errors = append(errors, fmt.Sprintf("task %d (%s): failed to get sequence: %v", i+1, input.Title, err))
				continue
			}
			record.Set("board", boardRecord.Id)
			record.Set("seq", seq)
			displayID = board.FormatDisplayID(boardRecord.GetString("prefix"), seq)
		} else {
			displayID = output.ShortID(record.Id)
		}

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

		if err := saveRecordHybrid(app, record, out); err != nil {
			errors = append(errors, fmt.Sprintf("task %d (%s): failed to save: %v", i+1, input.Title, err))
			continue
		}
		created = append(created, record)
		createdDisplayIDs = append(createdDisplayIDs, displayID)
	}

	// Output results
	if out.JSON {
		tasks := make([]map[string]any, 0, len(created))
		for i, record := range created {
			taskData := map[string]any{
				"id":       record.Id,
				"title":    record.GetString("title"),
				"type":     record.GetString("type"),
				"priority": record.GetString("priority"),
				"column":   record.GetString("column"),
			}
			if i < len(createdDisplayIDs) {
				taskData["display_id"] = createdDisplayIDs[i]
			}
			if record.GetString("board") != "" {
				taskData["board"] = record.GetString("board")
				taskData["seq"] = record.GetInt("seq")
			}
			tasks = append(tasks, taskData)
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"created": len(created),
			"failed":  len(errors),
			"tasks":   tasks,
			"errors":  errors,
		})
	}

	// Human output
	for i, record := range created {
		displayID := shortID(record.Id)
		if i < len(createdDisplayIDs) {
			displayID = createdDisplayIDs[i]
		}
		fmt.Printf("Created: %s [%s]\n", record.GetString("title"), displayID)
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

// isValidCustomID validates a custom task ID.
// PocketBase requires IDs to be exactly 15 alphanumeric characters.
func isValidCustomID(id string) bool {
	if len(id) != 15 {
		return false
	}
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

// formatCustomIDError returns a helpful error message for invalid custom IDs
func formatCustomIDError(id string) string {
	if len(id) != 15 {
		return fmt.Sprintf("invalid id '%s': must be exactly 15 characters (got %d)", id, len(id))
	}
	return fmt.Sprintf("invalid id '%s': must contain only lowercase letters (a-z) and digits (0-9)", id)
}

// saveRecordHybrid attempts to save a record via HTTP API first (for real-time updates),
// falling back to direct database access if the server is not running.
func saveRecordHybrid(app *pocketbase.PocketBase, record *core.Record, out *output.Formatter) error {
	// If direct mode is enabled, skip API attempt
	if isDirectMode() {
		verboseLog("Using direct database access (--direct flag)")
		return app.Save(record)
	}

	// Try HTTP API first
	client := NewAPIClient()
	if client.IsServerRunning() {
		verboseLog("Server is running, using HTTP API for real-time updates")

		// Convert record to TaskData for API
		taskData := recordToTaskData(record)

		// Create via API
		_, err := client.CreateTask(taskData)
		if err != nil {
			// Check if it's a validation error (don't fall back)
			if apiErr, ok := IsAPIError(err); ok && apiErr.IsValidationError() {
				return fmt.Errorf("validation error: %s", apiErr.Message)
			}

			// Network or other error - fall back with warning
			warnLog("API request failed, falling back to direct database: %v", err)
			verboseLog("Falling back to direct database access")
			return app.Save(record)
		}

		verboseLog("Task created via API (real-time updates enabled)")
		return nil
	}

	// Server not running - fall back to direct
	verboseLog("Server not running, using direct database access")
	return app.Save(record)
}

// updateRecordHybrid attempts to update a record via HTTP API first (for real-time updates),
// falling back to direct database access if the server is not running.
func updateRecordHybrid(app *pocketbase.PocketBase, record *core.Record, out *output.Formatter) error {
	// If direct mode is enabled, skip API attempt
	if isDirectMode() {
		verboseLog("Using direct database access (--direct flag)")
		return app.Save(record)
	}

	// Try HTTP API first
	client := NewAPIClient()
	if client.IsServerRunning() {
		verboseLog("Server is running, using HTTP API for real-time updates")

		// Convert record to TaskData for API
		taskData := recordToTaskData(record)

		// Update via API
		_, err := client.UpdateTask(record.Id, taskData)
		if err != nil {
			// Check if it's a validation error (don't fall back)
			if apiErr, ok := IsAPIError(err); ok && apiErr.IsValidationError() {
				return fmt.Errorf("validation error: %s", apiErr.Message)
			}

			// Network or other error - fall back with warning
			warnLog("API request failed, falling back to direct database: %v", err)
			verboseLog("Falling back to direct database access")
			return app.Save(record)
		}

		verboseLog("Task updated via API (real-time updates enabled)")
		return nil
	}

	// Server not running - fall back to direct
	verboseLog("Server not running, using direct database access")
	return app.Save(record)
}

// deleteRecordHybrid attempts to delete a record via HTTP API first (for real-time updates),
// falling back to direct database access if the server is not running.
func deleteRecordHybrid(app *pocketbase.PocketBase, record *core.Record, out *output.Formatter) error {
	// If direct mode is enabled, skip API attempt
	if isDirectMode() {
		verboseLog("Using direct database access (--direct flag)")
		return app.Delete(record)
	}

	// Try HTTP API first
	client := NewAPIClient()
	if client.IsServerRunning() {
		verboseLog("Server is running, using HTTP API for real-time updates")

		// Delete via API
		err := client.DeleteTask(record.Id)
		if err != nil {
			// Check if it's a validation error (don't fall back)
			if apiErr, ok := IsAPIError(err); ok && apiErr.IsValidationError() {
				return fmt.Errorf("error: %s", apiErr.Message)
			}

			// Network or other error - fall back with warning
			warnLog("API request failed, falling back to direct database: %v", err)
			verboseLog("Falling back to direct database access")
			return app.Delete(record)
		}

		verboseLog("Task deleted via API (real-time updates enabled)")
		return nil
	}

	// Server not running - fall back to direct
	verboseLog("Server not running, using direct database access")
	return app.Delete(record)
}

// recordToTaskData converts a core.Record to TaskData for API calls.
func recordToTaskData(record *core.Record) TaskData {
	// Get labels as string slice (handle multiple types like getTaskBlockedBy)
	var labels []string
	if rawLabels := record.Get("labels"); rawLabels != nil {
		// Handle []any type (common from JSON)
		if l, ok := rawLabels.([]any); ok {
			labels = make([]string, 0, len(l))
			for _, item := range l {
				if s, ok := item.(string); ok {
					labels = append(labels, s)
				}
			}
		} else if l, ok := rawLabels.([]string); ok {
			// Handle []string type
			labels = l
		} else if jsonRaw, ok := rawLabels.(interface{ String() string }); ok {
			// Handle types.JSONRaw (from database)
			jsonStr := jsonRaw.String()
			if jsonStr != "" && jsonStr != "null" {
				_ = json.Unmarshal([]byte(jsonStr), &labels)
			}
		}
	}

	// Get blocked_by as string slice (use getTaskBlockedBy for proper type handling)
	blockedBy := getTaskBlockedBy(record)

	// Get history
	var history []any
	if rawHistory := record.Get("history"); rawHistory != nil {
		if h, ok := rawHistory.([]map[string]any); ok {
			for _, entry := range h {
				history = append(history, entry)
			}
		}
	}

	return TaskData{
		ID:             record.Id,
		Title:          record.GetString("title"),
		Description:    record.GetString("description"),
		Type:           record.GetString("type"),
		Priority:       record.GetString("priority"),
		Column:         record.GetString("column"),
		Position:       record.GetFloat("position"),
		Labels:         labels,
		BlockedBy:      blockedBy,
		CreatedBy:      record.GetString("created_by"),
		CreatedByAgent: record.GetString("created_by_agent"),
		Epic:           record.GetString("epic"),
		Board:          record.GetString("board"),
		Seq:            record.GetInt("seq"),
		Parent:         record.GetString("parent"),
		History:        history,
	}
}
