package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// exitFunc is the function called to exit the program.
// It can be overridden in tests to prevent actual process termination.
var exitFunc = os.Exit

// Formatter handles output formatting for CLI commands.
// It supports both human-readable and JSON output modes.
type Formatter struct {
	// JSON enables JSON output mode
	JSON bool
	// Quiet suppresses non-essential output
	Quiet bool
}

// New creates a new Formatter with the given options.
func New(jsonMode, quiet bool) *Formatter {
	return &Formatter{
		JSON:  jsonMode,
		Quiet: quiet,
	}
}

// Task outputs a single task.
// Human mode: "Created task: Title [id]"
// JSON mode: Full task object
func (f *Formatter) Task(task *core.Record, action string) {
	if f.Quiet && !f.JSON {
		return
	}

	if f.JSON {
		f.writeJSON(taskToMap(task))
		return
	}

	fmt.Printf("%s task: %s [%s]\n", action, task.GetString("title"), ShortID(task.Id))
}

// Tasks outputs a list of tasks.
// Human mode: Grouped by column
// JSON mode: Array with count
func (f *Formatter) Tasks(tasks []*core.Record) {
	if f.JSON {
		f.writeJSON(map[string]any{
			"tasks": tasksToMaps(tasks),
			"count": len(tasks),
		})
		return
	}

	// Group tasks by column
	grouped := groupByColumn(tasks)
	columns := []string{"backlog", "todo", "in_progress", "need_input", "review", "done"}

	for _, col := range columns {
		colTasks := grouped[col]
		fmt.Printf("\n%s\n", strings.ToUpper(col))

		if len(colTasks) == 0 {
			fmt.Println("  (empty)")
			continue
		}

		for _, task := range colTasks {
			f.printTaskLine(task)
		}
	}
	fmt.Println()
}

// TasksWithFields outputs tasks with only specified fields (JSON only).
// If not in JSON mode, falls back to regular task output.
func (f *Formatter) TasksWithFields(tasks []*core.Record, fields []string) {
	if !f.JSON {
		f.Tasks(tasks)
		return
	}

	// Build filtered task list
	result := make([]map[string]any, len(tasks))
	for i, task := range tasks {
		fullMap := taskToMap(task)
		filtered := make(map[string]any)
		for _, field := range fields {
			field = strings.TrimSpace(field)
			if val, ok := fullMap[field]; ok {
				filtered[field] = val
			}
		}
		result[i] = filtered
	}

	f.writeJSON(map[string]any{
		"tasks": result,
		"count": len(tasks),
	})
}

// TasksWithBoards outputs a list of tasks with board-prefixed display IDs.
// Human mode: Grouped by column with display IDs
// JSON mode: Array with count and display_id field
func (f *Formatter) TasksWithBoards(tasks []*core.Record, boardsMap map[string]*core.Record) {
	if f.JSON {
		f.writeJSON(map[string]any{
			"tasks": tasksToMapsWithBoards(tasks, boardsMap),
			"count": len(tasks),
		})
		return
	}

	// Group tasks by column
	grouped := groupByColumn(tasks)
	columns := []string{"backlog", "todo", "in_progress", "need_input", "review", "done"}

	for _, col := range columns {
		colTasks := grouped[col]
		fmt.Printf("\n%s\n", strings.ToUpper(col))

		if len(colTasks) == 0 {
			fmt.Println("  (empty)")
			continue
		}

		for _, task := range colTasks {
			f.printTaskLineWithBoard(task, boardsMap)
		}
	}
	fmt.Println()
}

// TasksWithFieldsAndBoards outputs tasks with only specified fields and board info (JSON only).
// If not in JSON mode, falls back to TasksWithBoards.
func (f *Formatter) TasksWithFieldsAndBoards(tasks []*core.Record, fields []string, boardsMap map[string]*core.Record) {
	if !f.JSON {
		f.TasksWithBoards(tasks, boardsMap)
		return
	}

	// Build filtered task list
	result := make([]map[string]any, len(tasks))
	for i, task := range tasks {
		fullMap := taskToMapWithBoard(task, boardsMap)
		filtered := make(map[string]any)
		for _, field := range fields {
			field = strings.TrimSpace(field)
			if val, ok := fullMap[field]; ok {
				filtered[field] = val
			}
		}
		result[i] = filtered
	}

	f.writeJSON(map[string]any{
		"tasks": result,
		"count": len(tasks),
	})
}

// TaskDetail outputs detailed information about a task.
func (f *Formatter) TaskDetail(task *core.Record) {
	if f.JSON {
		f.writeJSON(taskToMap(task))
		return
	}

	fmt.Printf("\nTask: %s\n", task.Id)
	fmt.Printf("Title:       %s\n", task.GetString("title"))
	fmt.Printf("Type:        %s\n", task.GetString("type"))
	fmt.Printf("Priority:    %s\n", task.GetString("priority"))
	fmt.Printf("Column:      %s\n", task.GetString("column"))
	fmt.Printf("Position:    %.0f\n", task.GetFloat("position"))

	// Labels
	labels := getLabels(task)
	if len(labels) > 0 {
		fmt.Printf("Labels:      %s\n", strings.Join(labels, ", "))
	} else {
		fmt.Printf("Labels:      -\n")
	}

	// Blocked by
	blockedBy := getBlockedBy(task)
	if len(blockedBy) > 0 {
		fmt.Printf("Blocked by:  %s\n", strings.Join(blockedBy, ", "))
	}

	// Created by
	createdBy := task.GetString("created_by")
	if agent := task.GetString("created_by_agent"); agent != "" {
		fmt.Printf("Created by:  %s (%s)\n", createdBy, agent)
	} else {
		fmt.Printf("Created by:  %s\n", createdBy)
	}

	// Timestamps
	fmt.Printf("Created:     %s\n", formatTime(task.GetDateTime("created").Time()))
	fmt.Printf("Updated:     %s\n", formatTime(task.GetDateTime("updated").Time()))

	// Description
	if desc := task.GetString("description"); desc != "" {
		fmt.Printf("\nDescription:\n  %s\n", strings.ReplaceAll(desc, "\n", "\n  "))
	}

	fmt.Println()
}

// TaskDetailWithSubtasks outputs detailed information about a task including its sub-tasks.
func (f *Formatter) TaskDetailWithSubtasks(task *core.Record, subtasks []*core.Record) {
	if f.JSON {
		result := taskToMap(task)
		result["subtasks"] = tasksToMaps(subtasks)
		result["subtask_count"] = len(subtasks)
		// Include agent_session in JSON output
		result["agent_session"] = task.Get("agent_session")
		f.writeJSON(result)
		return
	}

	fmt.Printf("\nTask: %s\n", task.Id)
	fmt.Printf("Title:       %s\n", task.GetString("title"))
	fmt.Printf("Type:        %s\n", task.GetString("type"))
	fmt.Printf("Priority:    %s\n", task.GetString("priority"))
	fmt.Printf("Column:      %s\n", task.GetString("column"))
	fmt.Printf("Position:    %.0f\n", task.GetFloat("position"))

	// Due date
	if dueDate := task.GetDateTime("due_date"); !dueDate.IsZero() {
		fmt.Printf("Due:         %s\n", dueDate.Time().Format("2006-01-02"))
	}

	// Parent task
	if parent := task.GetString("parent"); parent != "" {
		fmt.Printf("Parent:      %s\n", ShortID(parent))
	}

	// Labels
	labels := getLabels(task)
	if len(labels) > 0 {
		fmt.Printf("Labels:      %s\n", strings.Join(labels, ", "))
	} else {
		fmt.Printf("Labels:      -\n")
	}

	// Blocked by
	blockedBy := getBlockedBy(task)
	if len(blockedBy) > 0 {
		fmt.Printf("Blocked by:  %s\n", strings.Join(blockedBy, ", "))
	}

	// Created by
	createdBy := task.GetString("created_by")
	if agent := task.GetString("created_by_agent"); agent != "" {
		fmt.Printf("Created by:  %s (%s)\n", createdBy, agent)
	} else {
		fmt.Printf("Created by:  %s\n", createdBy)
	}

	// Timestamps
	fmt.Printf("Created:     %s\n", formatTime(task.GetDateTime("created").Time()))
	fmt.Printf("Updated:     %s\n", formatTime(task.GetDateTime("updated").Time()))

	// Description
	if desc := task.GetString("description"); desc != "" {
		fmt.Printf("\nDescription:\n  %s\n", strings.ReplaceAll(desc, "\n", "\n  "))
	}

	// Agent Session
	sessionData := task.Get("agent_session")
	if session := parseAgentSession(sessionData); session != nil {
		fmt.Printf("\nAgent Session:\n")
		if tool, ok := session["tool"].(string); ok {
			fmt.Printf("  Tool:        %s\n", tool)
		}
		if ref, ok := session["ref"].(string); ok {
			fmt.Printf("  Ref:         %s\n", truncateMiddle(ref, 40))
		}
		if workingDir, ok := session["working_dir"].(string); ok {
			fmt.Printf("  Working Dir: %s\n", workingDir)
		}
		if linkedAt, ok := session["linked_at"].(string); ok {
			if t, err := time.Parse(time.RFC3339, linkedAt); err == nil {
				fmt.Printf("  Linked:      %s\n", formatTime(t))
			}
		}
		// Show resume hint if task is in need_input
		if task.GetString("column") == "need_input" {
			fmt.Printf("  (Use 'egenskriven resume <task>' to continue)\n")
		}
	}

	// Sub-tasks
	if len(subtasks) > 0 {
		fmt.Printf("\nSub-tasks (%d):\n", len(subtasks))
		for _, st := range subtasks {
			status := " "
			if st.GetString("column") == "done" {
				status = "x"
			}
			fmt.Printf("  [%s] [%s] %s (%s)\n",
				status,
				ShortID(st.Id),
				st.GetString("title"),
				st.GetString("column"),
			)
		}
	}

	fmt.Println()
}

// truncateMiddle truncates a string in the middle if too long
func truncateMiddle(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	half := (maxLen - 3) / 2
	return s[:half] + "..." + s[len(s)-half:]
}

// parseAgentSession converts various session data formats to map[string]any
// It handles nil, map[string]any, map[string]interface{}, JSON string, []byte, and fmt.Stringer types
func parseAgentSession(data any) map[string]any {
	if data == nil {
		return nil
	}

	// Handle map[string]any directly
	if m, ok := data.(map[string]any); ok {
		return m
	}

	// Handle map[string]interface{} (common from PocketBase)
	if m, ok := data.(map[string]interface{}); ok {
		result := make(map[string]any, len(m))
		for k, v := range m {
			result[k] = v
		}
		return result
	}

	// Handle JSON string
	if s, ok := data.(string); ok {
		if s == "" || s == "null" {
			return nil
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(s), &result); err != nil {
			return nil
		}
		return result
	}

	// Handle types.JSONRaw and []byte (PocketBase JSON field types)
	if b, ok := data.([]byte); ok {
		if len(b) == 0 || string(b) == "null" {
			return nil
		}
		var result map[string]any
		if err := json.Unmarshal(b, &result); err != nil {
			return nil
		}
		return result
	}

	// Try to handle any type that implements fmt.Stringer
	if stringer, ok := data.(fmt.Stringer); ok {
		s := stringer.String()
		if s == "" || s == "null" {
			return nil
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(s), &result); err != nil {
			return nil
		}
		return result
	}

	return nil
}

// Success outputs a success message.
func (f *Formatter) Success(message string) {
	if f.Quiet {
		return
	}
	if f.JSON {
		f.writeJSON(map[string]any{
			"success": true,
			"message": message,
		})
		return
	}
	fmt.Println(message)
}

// Error outputs an error with optional data and exits with the given code.
// This function does not return in normal operation - it calls os.Exit().
func (f *Formatter) Error(code int, message string, data any) error {
	if f.JSON {
		errObj := map[string]any{
			"error": map[string]any{
				"code":    code,
				"message": message,
			},
		}
		if data != nil {
			errObj["error"].(map[string]any)["data"] = data
		}
		json.NewEncoder(os.Stderr).Encode(errObj)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", message)
	}
	exitFunc(code)
	return nil // unreachable in production, but satisfies return type for tests
}

// ErrorWithSuggestion outputs an error with a helpful suggestion and exits.
// This function does not return in normal operation - it calls os.Exit().
func (f *Formatter) ErrorWithSuggestion(code int, message, suggestion string, data any) error {
	if f.JSON {
		errObj := map[string]any{
			"error": map[string]any{
				"code":    code,
				"message": message,
			},
		}
		if suggestion != "" {
			errObj["error"].(map[string]any)["suggestion"] = suggestion
		}
		if data != nil {
			errObj["error"].(map[string]any)["data"] = data
		}
		json.NewEncoder(os.Stderr).Encode(errObj)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", message)
		if suggestion != "" {
			fmt.Fprintf(os.Stderr, "\nSuggestion: %s\n", suggestion)
		}
	}
	exitFunc(code)
	return nil // unreachable in production, but satisfies return type for tests
}

// AmbiguousError outputs an error for ambiguous task references and exits.
// This function does not return in normal operation - it calls os.Exit(4).
func (f *Formatter) AmbiguousError(ref string, matches []*core.Record) error {
	data := map[string]any{
		"reference": ref,
		"matches":   tasksToMaps(matches),
	}

	if f.JSON {
		return f.Error(4, fmt.Sprintf("Ambiguous task reference: '%s' matches multiple tasks", ref), data)
	}

	fmt.Fprintf(os.Stderr, "Error: Ambiguous task reference: '%s' matches multiple tasks:\n", ref)
	for _, task := range matches {
		fmt.Fprintf(os.Stderr, "  [%s] %s\n", ShortID(task.Id), task.GetString("title"))
	}
	exitFunc(4)
	return nil // unreachable in production, but satisfies return type for tests
}

// WriteJSON outputs any value as formatted JSON.
// This is exported for use by commands that need custom JSON output.
func (f *Formatter) WriteJSON(v any) {
	f.writeJSON(v)
}

// --- Helper functions ---

func (f *Formatter) writeJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func (f *Formatter) printTaskLine(task *core.Record) {
	priority := task.GetString("priority")
	priorityIndicator := ""
	switch priority {
	case "urgent":
		priorityIndicator = "!!!"
	case "high":
		priorityIndicator = "!!"
	case "medium":
		priorityIndicator = "!"
	}

	fmt.Printf("  [%s] %s (%s%s)\n",
		ShortID(task.Id),
		task.GetString("title"),
		task.GetString("type"),
		func() string {
			if priorityIndicator != "" {
				return ", " + priorityIndicator + priority
			}
			return ""
		}(),
	)
}

func (f *Formatter) printTaskLineWithBoard(task *core.Record, boardsMap map[string]*core.Record) {
	priority := task.GetString("priority")
	priorityIndicator := ""
	switch priority {
	case "urgent":
		priorityIndicator = "!!!"
	case "high":
		priorityIndicator = "!!"
	case "medium":
		priorityIndicator = "!"
	}

	displayID := getDisplayID(task, boardsMap)

	fmt.Printf("  [%s] %s (%s%s)\n",
		displayID,
		task.GetString("title"),
		task.GetString("type"),
		func() string {
			if priorityIndicator != "" {
				return ", " + priorityIndicator + priority
			}
			return ""
		}(),
	)
}

func taskToMap(task *core.Record) map[string]any {
	return map[string]any{
		"id":               task.Id,
		"title":            task.GetString("title"),
		"description":      task.GetString("description"),
		"type":             task.GetString("type"),
		"priority":         task.GetString("priority"),
		"column":           task.GetString("column"),
		"position":         task.GetFloat("position"),
		"labels":           getLabels(task),
		"blocked_by":       getBlockedBy(task),
		"created_by":       task.GetString("created_by"),
		"created_by_agent": task.GetString("created_by_agent"),
		"created":          task.GetDateTime("created").Time().Format(time.RFC3339),
		"updated":          task.GetDateTime("updated").Time().Format(time.RFC3339),
	}
}

func tasksToMaps(tasks []*core.Record) []map[string]any {
	result := make([]map[string]any, len(tasks))
	for i, task := range tasks {
		result[i] = taskToMap(task)
	}
	return result
}

func tasksToMapsWithBoards(tasks []*core.Record, boardsMap map[string]*core.Record) []map[string]any {
	result := make([]map[string]any, len(tasks))
	for i, task := range tasks {
		result[i] = taskToMapWithBoard(task, boardsMap)
	}
	return result
}

func taskToMapWithBoard(task *core.Record, boardsMap map[string]*core.Record) map[string]any {
	result := taskToMap(task)
	result["display_id"] = getDisplayID(task, boardsMap)

	// Add board and seq if present
	if boardID := task.GetString("board"); boardID != "" {
		result["board"] = boardID
		result["seq"] = task.GetInt("seq")
	}

	return result
}

// getDisplayID returns the board-prefixed display ID (e.g., "WRK-123") or short ID as fallback
func getDisplayID(task *core.Record, boardsMap map[string]*core.Record) string {
	boardID := task.GetString("board")
	seq := task.GetInt("seq")

	if boardID != "" && seq > 0 {
		if boardRecord, ok := boardsMap[boardID]; ok {
			prefix := boardRecord.GetString("prefix")
			return fmt.Sprintf("%s-%d", prefix, seq)
		}
	}

	// Fallback to short ID
	return ShortID(task.Id)
}

func groupByColumn(tasks []*core.Record) map[string][]*core.Record {
	grouped := make(map[string][]*core.Record)
	for _, task := range tasks {
		col := task.GetString("column")
		grouped[col] = append(grouped[col], task)
	}
	return grouped
}

func getLabels(task *core.Record) []string {
	// Use GetStringSlice which properly handles types.JSONRaw
	return task.GetStringSlice("labels")
}

func getBlockedBy(task *core.Record) []string {
	// Use GetStringSlice which properly handles types.JSONRaw
	return task.GetStringSlice("blocked_by")
}

// ShortID returns the first 8 characters of an ID for display.
// Safe to call with IDs shorter than 8 characters.
func ShortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
