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
	columns := []string{"backlog", "todo", "in_progress", "review", "done"}

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
