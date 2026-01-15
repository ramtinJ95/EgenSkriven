package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
	"github.com/ramtinJ95/EgenSkriven/internal/config"
	"github.com/ramtinJ95/EgenSkriven/internal/position"
)

// debugLog logs a message only when EGENSKRIVEN_DEBUG is set
func debugLog(format string, args ...any) {
	if os.Getenv("EGENSKRIVEN_DEBUG") != "" {
		log.Printf("[tui-debug] "+format, args...)
	}
}

const (
	// DefaultServerURL for API requests
	DefaultServerURL = "http://localhost:8090"
	// HealthCheckTimeout for checking if server is running
	HealthCheckTimeout = 500 * time.Millisecond
	// APIRequestTimeout for actual API operations
	APIRequestTimeout = 5 * time.Second
)

// =============================================================================
// Load Commands
// =============================================================================

// loadBoards creates a command that loads all boards from the database.
// Returns boardsLoadedMsg on success, errMsg on failure.
func loadBoards(app *pocketbase.PocketBase) tea.Cmd {
	return func() tea.Msg {
		records, err := board.GetAll(app)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to load boards: %w", err), context: "loading boards"}
		}
		return boardsLoadedMsg{boards: records}
	}
}

// loadTasks creates a command that loads all tasks for a specific board.
// Tasks are loaded and will be grouped by column in the Update handler.
func loadTasks(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return func() tea.Msg {
		// Build query for tasks in this board
		records, err := app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to load tasks: %w", err), context: "loading tasks"}
		}

		// Sort by column and position
		sort.Slice(records, func(i, j int) bool {
			colI := records[i].GetString("column")
			colJ := records[j].GetString("column")
			if colI != colJ {
				return getColumnOrder(colI) < getColumnOrder(colJ)
			}
			return records[i].GetFloat("position") < records[j].GetFloat("position")
		})

		return tasksLoadedMsg{tasks: records}
	}
}

// getColumnOrder returns the display order for a column.
// Used for sorting tasks by column.
func getColumnOrder(column string) int {
	order := map[string]int{
		"backlog":     0,
		"todo":        1,
		"in_progress": 2,
		"need_input":  3,
		"review":      4,
		"done":        5,
	}
	if o, ok := order[column]; ok {
		return o
	}
	return 99 // Unknown columns go to end
}

// loadBoardAndTasks creates a command that loads a specific board and its tasks.
// This is a convenience function for initial load.
func loadBoardAndTasks(app *pocketbase.PocketBase, boardRef string) tea.Cmd {
	return func() tea.Msg {
		// First, find the board
		var boardRecord *core.Record
		var err error

		if boardRef != "" {
			// Use board reference
			boardRecord, err = board.GetByNameOrPrefix(app, boardRef)
			if err != nil {
				return errMsg{err: fmt.Errorf("board not found: %s", boardRef), context: "finding board"}
			}
		} else {
			// Get all boards and use the first one
			boards, err := board.GetAll(app)
			if err != nil {
				return errMsg{err: fmt.Errorf("failed to load boards: %w", err), context: "loading boards"}
			}
			if len(boards) == 0 {
				return errMsg{err: fmt.Errorf("no boards found - create one with 'egenskriven board create'"), context: "loading boards"}
			}
			boardRecord = boards[0]
		}

		// Load tasks for this board
		records, err := app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardRecord.Id}),
		)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to load tasks: %w", err), context: "loading tasks"}
		}

		// Sort by position within each column
		sort.Slice(records, func(i, j int) bool {
			colI := records[i].GetString("column")
			colJ := records[j].GetString("column")
			if colI != colJ {
				return getColumnOrder(colI) < getColumnOrder(colJ)
			}
			return records[i].GetFloat("position") < records[j].GetFloat("position")
		})

		// Return combined message
		return boardAndTasksLoadedMsg{
			board: boardRecord,
			tasks: records,
		}
	}
}

// loadBoardTasks returns a command that loads tasks for a specific board.
// Returns boardTasksLoadedMsg (used during board switching).
func loadBoardTasks(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return func() tea.Msg {
		records, err := app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to load tasks: %w", err), context: "loading tasks"}
		}

		// Sort by column and position
		sort.Slice(records, func(i, j int) bool {
			colI := records[i].GetString("column")
			colJ := records[j].GetString("column")
			if colI != colJ {
				return getColumnOrder(colI) < getColumnOrder(colJ)
			}
			return records[i].GetFloat("position") < records[j].GetFloat("position")
		})

		return boardTasksLoadedMsg{tasks: records}
	}
}

// loadBoardColumns returns a command that loads columns for a specific board.
func loadBoardColumns(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return func() tea.Msg {
		record, err := app.FindRecordById("boards", boardID)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to load board: %w", err), context: "loading board columns"}
		}

		b := board.RecordToBoard(record)
		return boardColumnsMsg{columns: b.Columns}
	}
}

// getTaskCountsForBoards returns task counts per board.
// This is a sync function used during board selector initialization.
func getTaskCountsForBoards(app *pocketbase.PocketBase, boardIDs []string) map[string]int {
	counts := make(map[string]int)

	for _, boardID := range boardIDs {
		records, err := app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
		if err == nil {
			counts[boardID] = len(records)
		}
	}

	return counts
}

// saveLastBoard returns a command that persists the last-used board to config.
func saveLastBoard(boardID string) tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.LoadProjectConfig()
		if err != nil {
			// If we can't load, create a new config
			cfg = config.DefaultConfig()
		}

		cfg.DefaultBoard = boardID

		if err := config.SaveConfig(".", cfg); err != nil {
			return errMsg{err: err, context: "saving last board"}
		}

		return lastBoardSavedMsg{boardID: boardID}
	}
}

// loadDefaultBoard returns a command that loads the default board from config.
func loadDefaultBoard() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.LoadProjectConfig()
		if err != nil || cfg.DefaultBoard == "" {
			// No default board configured
			return nil
		}
		return boardSwitchedMsg{boardID: cfg.DefaultBoard}
	}
}

// switchBoard returns a command sequence for switching to a new board.
// It loads the board's tasks and columns, then saves it as the last-used board.
func switchBoard(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return tea.Batch(
		loadBoardTasks(app, boardID),
		loadBoardColumns(app, boardID),
		saveLastBoard(boardID),
	)
}

// findBoardByRef finds a board by name, prefix, or ID and returns a switch message.
func findBoardByRef(app *pocketbase.PocketBase, ref string) tea.Cmd {
	return func() tea.Msg {
		record, err := board.GetByNameOrPrefix(app, ref)
		if err != nil {
			return errMsg{err: fmt.Errorf("board not found: %s", ref), context: "finding board"}
		}
		return boardSwitchedMsg{boardID: record.Id}
	}
}

// =============================================================================
// Board Data Helpers
// =============================================================================

// BoardData holds computed board information for display.
type BoardData struct {
	ID         string
	Name       string
	Prefix     string
	Color      string
	Columns    []string
	TaskCount  int
	Tasks      []*core.Record
	TasksByCol map[string][]*core.Record
}

// computeBoardData processes raw records into display-ready data.
func computeBoardData(boardRecord *core.Record, tasks []*core.Record) *BoardData {
	b := board.RecordToBoard(boardRecord)

	// Group tasks by column
	tasksByCol := make(map[string][]*core.Record)
	for _, col := range b.Columns {
		tasksByCol[col] = []*core.Record{}
	}

	for _, task := range tasks {
		col := task.GetString("column")
		if col == "" {
			col = "backlog" // Default column
		}
		tasksByCol[col] = append(tasksByCol[col], task)
	}

	return &BoardData{
		ID:         boardRecord.Id,
		Name:       b.Name,
		Prefix:     b.Prefix,
		Color:      b.Color,
		Columns:    b.Columns,
		TaskCount:  len(tasks),
		Tasks:      tasks,
		TasksByCol: tasksByCol,
	}
}

// =============================================================================
// Create Task Command
// =============================================================================

// createTask creates a new task with the given data
func createTask(app *pocketbase.PocketBase, boardRecord *core.Record, data TaskFormData) tea.Cmd {
	return func() tea.Msg {
		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return errMsg{err: err, context: "finding tasks collection"}
		}

		record := core.NewRecord(collection)

		// Set task fields
		record.Set("title", data.Title)
		record.Set("description", data.Description)
		record.Set("type", data.Type)
		record.Set("priority", data.Priority)
		record.Set("column", data.Column)
		record.Set("labels", data.Labels)
		record.Set("blocked_by", []string{})
		record.Set("created_by", "tui")

		// Set board and sequence
		var displayID string
		if boardRecord != nil {
			seq, err := board.GetAndIncrementSequence(app, boardRecord.Id)
			if err != nil {
				return errMsg{err: err, context: "getting sequence number"}
			}
			record.Set("board", boardRecord.Id)
			record.Set("seq", seq)
			displayID = board.FormatDisplayID(boardRecord.GetString("prefix"), seq)
		}

		// Set position at end of column
		position := position.GetNext(app, data.Column)
		record.Set("position", position)

		// Handle due date
		if data.DueDate != "" {
			record.Set("due_date", data.DueDate)
		}

		// Handle epic
		if data.EpicID != "" {
			record.Set("epic", data.EpicID)
		}

		// Initialize history
		history := []map[string]any{
			{
				"timestamp":    time.Now().UTC().Format(time.RFC3339),
				"action":       "created",
				"actor":        "tui",
				"actor_detail": "",
				"changes":      nil,
			},
		}
		record.Set("history", history)

		// Save using hybrid pattern
		if err := saveRecordHybrid(app, record); err != nil {
			return errMsg{err: err, context: "creating task"}
		}

		return taskCreatedMsg{
			task:      record,
			displayID: displayID,
		}
	}
}

// =============================================================================
// Update Task Command
// =============================================================================

// updateTask updates an existing task with the given data
func updateTask(app *pocketbase.PocketBase, taskID string, data TaskFormData) tea.Cmd {
	return func() tea.Msg {
		record, err := app.FindRecordById("tasks", taskID)
		if err != nil {
			return errMsg{err: err, context: "finding task"}
		}

		// Track changes for history
		changes := make(map[string]any)

		// Update fields that changed
		if data.Title != "" && data.Title != record.GetString("title") {
			changes["title"] = map[string]string{
				"from": record.GetString("title"),
				"to":   data.Title,
			}
			record.Set("title", data.Title)
		}

		if data.Description != record.GetString("description") {
			record.Set("description", data.Description)
			// Don't track description changes in detail (too verbose)
		}

		if data.Type != "" && data.Type != record.GetString("type") {
			changes["type"] = map[string]string{
				"from": record.GetString("type"),
				"to":   data.Type,
			}
			record.Set("type", data.Type)
		}

		if data.Priority != "" && data.Priority != record.GetString("priority") {
			changes["priority"] = map[string]string{
				"from": record.GetString("priority"),
				"to":   data.Priority,
			}
			record.Set("priority", data.Priority)
		}

		if data.Column != "" && data.Column != record.GetString("column") {
			changes["column"] = map[string]string{
				"from": record.GetString("column"),
				"to":   data.Column,
			}
			// When changing column, get new position at end of target column
			position := position.GetNext(app, data.Column)
			record.Set("column", data.Column)
			record.Set("position", position)
		}

		record.Set("labels", data.Labels)

		if data.DueDate != "" {
			record.Set("due_date", data.DueDate)
		} else {
			record.Set("due_date", "")
		}

		if data.EpicID != "" {
			record.Set("epic", data.EpicID)
		} else {
			record.Set("epic", "")
		}

		// Add history entry if there were changes
		if len(changes) > 0 {
			history := getRecordHistory(record)
			history = append(history, map[string]any{
				"timestamp":    time.Now().UTC().Format(time.RFC3339),
				"action":       "updated",
				"actor":        "tui",
				"actor_detail": "",
				"changes":      changes,
			})
			record.Set("history", history)
		}

		// Save using hybrid pattern
		if err := updateRecordHybrid(app, record); err != nil {
			return errMsg{err: err, context: "updating task"}
		}

		return taskUpdatedMsg{task: record}
	}
}

// =============================================================================
// Delete Task Command
// =============================================================================

// deleteTask deletes a task by ID
func deleteTask(app *pocketbase.PocketBase, taskID string) tea.Cmd {
	return func() tea.Msg {
		record, err := app.FindRecordById("tasks", taskID)
		if err != nil {
			return errMsg{err: err, context: "finding task to delete"}
		}

		title := record.GetString("title")

		if err := deleteRecordHybrid(app, record); err != nil {
			return errMsg{err: err, context: "deleting task"}
		}

		return taskDeletedMsg{
			taskID: taskID,
			title:  title,
		}
	}
}

// =============================================================================
// Move Task Commands
// =============================================================================

// moveTaskToColumn moves a task to a different column
func moveTaskToColumn(app *pocketbase.PocketBase, taskID, targetColumn string) tea.Cmd {
	return func() tea.Msg {
		record, err := app.FindRecordById("tasks", taskID)
		if err != nil {
			return errMsg{err: err, context: "finding task to move"}
		}

		fromColumn := record.GetString("column")
		if fromColumn == targetColumn {
			// Already in target column, nothing to do
			return taskMovedMsg{
				task:       record,
				fromColumn: fromColumn,
				toColumn:   targetColumn,
			}
		}

		// Get position at end of target column
		position := position.GetNext(app, targetColumn)

		// Track change in history
		history := getRecordHistory(record)
		history = append(history, map[string]any{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"action":    "moved",
			"actor":     "tui",
			"changes": map[string]any{
				"column": map[string]string{
					"from": fromColumn,
					"to":   targetColumn,
				},
			},
		})

		record.Set("column", targetColumn)
		record.Set("position", position)
		record.Set("history", history)

		if err := updateRecordHybrid(app, record); err != nil {
			return errMsg{err: err, context: "moving task"}
		}

		return taskMovedMsg{
			task:       record,
			fromColumn: fromColumn,
			toColumn:   targetColumn,
		}
	}
}

// reorderTaskInColumn moves a task up or down within its column
func reorderTaskInColumn(app *pocketbase.PocketBase, taskID string, moveUp bool) tea.Cmd {
	return func() tea.Msg {
		record, err := app.FindRecordById("tasks", taskID)
		if err != nil {
			return errMsg{err: err, context: "finding task to reorder"}
		}

		column := record.GetString("column")

		// Get new position
		var newPosition float64
		if moveUp {
			newPosition, err = position.GetBefore(app, taskID)
		} else {
			newPosition, err = position.GetAfter(app, taskID)
		}
		if err != nil {
			return errMsg{err: err, context: "calculating new position"}
		}

		record.Set("position", newPosition)

		if err := updateRecordHybrid(app, record); err != nil {
			return errMsg{err: err, context: "reordering task"}
		}

		return taskMovedMsg{
			task:       record,
			fromColumn: column,
			toColumn:   column,
		}
	}
}

// =============================================================================
// Status Message Commands
// =============================================================================

// setStatus creates a command that sends a status message.
// Used for displaying temporary feedback messages.
func setStatus(message string, isError bool) tea.Cmd {
	return func() tea.Msg {
		return statusMsg{
			message: message,
			isError: isError,
		}
	}
}

// showStatus shows a status message for a duration
func showStatus(text string, isError bool, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		return statusMsg{
			message:  text,
			isError:  isError,
			duration: duration,
		}
	}
}

// clearStatus creates a delayed command that clears the status message.
// Called after displaying a status to auto-clear it.
func clearStatus() tea.Cmd {
	return tea.Tick(
		3*time.Second,
		func(_ time.Time) tea.Msg {
			return clearStatusMsg{}
		},
	)
}

// clearStatusAfter clears the status message after a delay
func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

// =============================================================================
// Hybrid Save Pattern
// =============================================================================

// saveRecordHybrid attempts to save via API first, falls back to direct DB
func saveRecordHybrid(app *pocketbase.PocketBase, record *core.Record) error {
	if isServerRunning() {
		taskData := recordToAPIData(record)
		if err := createTaskViaAPI(taskData); err == nil {
			debugLog("created task via API")
			return nil
		} else {
			debugLog("API create failed, falling back to direct DB: %v", err)
		}
	} else {
		debugLog("server not running, using direct DB save")
	}
	return app.Save(record)
}

// updateRecordHybrid attempts to update via API first, falls back to direct DB
func updateRecordHybrid(app *pocketbase.PocketBase, record *core.Record) error {
	if isServerRunning() {
		taskData := recordToAPIData(record)
		if err := updateTaskViaAPI(record.Id, taskData); err == nil {
			debugLog("updated task via API: %s", record.Id)
			return nil
		} else {
			debugLog("API update failed, falling back to direct DB: %v", err)
		}
	} else {
		debugLog("server not running, using direct DB save")
	}
	return app.Save(record)
}

// deleteRecordHybrid attempts to delete via API first, falls back to direct DB
func deleteRecordHybrid(app *pocketbase.PocketBase, record *core.Record) error {
	if isServerRunning() {
		if err := deleteTaskViaAPI(record.Id); err == nil {
			debugLog("deleted task via API: %s", record.Id)
			return nil
		} else {
			debugLog("API delete failed, falling back to direct DB: %v", err)
		}
	} else {
		debugLog("server not running, using direct DB delete")
	}
	return app.Delete(record)
}

// isServerRunning checks if the PocketBase server is accessible
func isServerRunning() bool {
	client := &http.Client{Timeout: HealthCheckTimeout}
	resp, err := client.Get(DefaultServerURL + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// =============================================================================
// API Helper Functions
// =============================================================================

// APITaskData represents task data for API requests
type APITaskData struct {
	ID          string   `json:"id,omitempty"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	Column      string   `json:"column,omitempty"`
	Position    float64  `json:"position,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	BlockedBy   []string `json:"blocked_by,omitempty"`
	CreatedBy   string   `json:"created_by,omitempty"`
	Epic        string   `json:"epic,omitempty"`
	Board       string   `json:"board,omitempty"`
	Seq         int      `json:"seq,omitempty"`
	DueDate     string   `json:"due_date,omitempty"`
	History     []any    `json:"history,omitempty"`
}

func recordToAPIData(record *core.Record) APITaskData {
	var labels []string
	if rawLabels := record.Get("labels"); rawLabels != nil {
		if l, ok := rawLabels.([]any); ok {
			for _, item := range l {
				if s, ok := item.(string); ok {
					labels = append(labels, s)
				}
			}
		} else if l, ok := rawLabels.([]string); ok {
			labels = l
		}
	}

	var history []any
	if h := record.Get("history"); h != nil {
		if hSlice, ok := h.([]map[string]any); ok {
			for _, entry := range hSlice {
				history = append(history, entry)
			}
		} else if hAny, ok := h.([]any); ok {
			// Handle []any type (consistent with getRecordHistory)
			history = hAny
		}
	}

	return APITaskData{
		ID:          record.Id,
		Title:       record.GetString("title"),
		Description: record.GetString("description"),
		Type:        record.GetString("type"),
		Priority:    record.GetString("priority"),
		Column:      record.GetString("column"),
		Position:    record.GetFloat("position"),
		Labels:      labels,
		CreatedBy:   record.GetString("created_by"),
		Epic:        record.GetString("epic"),
		Board:       record.GetString("board"),
		Seq:         record.GetInt("seq"),
		DueDate:     record.GetString("due_date"),
		History:     history,
	}
}

func createTaskViaAPI(data APITaskData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: APIRequestTimeout}
	resp, err := client.Post(
		DefaultServerURL+"/api/collections/tasks/records",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func updateTaskViaAPI(id string, data APITaskData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"PATCH",
		DefaultServerURL+"/api/collections/tasks/records/"+id,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: APIRequestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func deleteTaskViaAPI(id string) error {
	req, err := http.NewRequest(
		"DELETE",
		DefaultServerURL+"/api/collections/tasks/records/"+id,
		nil,
	)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: APIRequestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// getRecordHistory extracts history from a record as a slice of maps
func getRecordHistory(record *core.Record) []map[string]any {
	var history []map[string]any
	if h := record.Get("history"); h != nil {
		if hSlice, ok := h.([]map[string]any); ok {
			history = hSlice
		} else if hAny, ok := h.([]any); ok {
			for _, item := range hAny {
				if m, ok := item.(map[string]any); ok {
					history = append(history, m)
				}
			}
		}
	}
	return history
}

// =============================================================================
// Polling Fallback Commands
// =============================================================================

// PollConfig holds configuration for the polling fallback.
type PollConfig struct {
	Interval  time.Duration
	BoardID   string
	LastCheck time.Time
}

// StartPolling initiates the polling fallback mechanism.
func StartPolling(config PollConfig) tea.Cmd {
	return tea.Tick(config.Interval, func(t time.Time) tea.Msg {
		return pollTickMsg{time: t}
	})
}

// PollForChanges checks for records updated since the last check.
func PollForChanges(app *pocketbase.PocketBase, boardID string, lastCheck time.Time) tea.Cmd {
	return func() tea.Msg {
		// Format timestamp for PocketBase query
		// PocketBase uses ISO 8601 format: 2006-01-02 15:04:05.000Z
		timestamp := lastCheck.UTC().Format("2006-01-02 15:04:05.000Z")

		// Query for tasks updated since lastCheck
		records, err := app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board} AND updated > {:time}",
				dbx.Params{
					"board": boardID,
					"time":  timestamp,
				}),
		)
		if err != nil {
			return pollResultMsg{
				tasks:     nil,
				checkTime: time.Now(),
				err:       err,
			}
		}

		return pollResultMsg{
			tasks:     records,
			checkTime: time.Now(),
			err:       nil,
		}
	}
}

// ContinuePolling schedules the next poll cycle.
func ContinuePolling(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return pollTickMsg{time: t}
	})
}

// =============================================================================
// Server Status Commands
// =============================================================================

// CheckServerStatus checks if the PocketBase server is reachable.
func CheckServerStatus(serverURL string) tea.Cmd {
	return func() tea.Msg {
		if isServerReachable(serverURL) {
			return serverOnlineMsg{}
		}
		return serverOfflineMsg{}
	}
}

// ScheduleServerCheck schedules a server health check after a delay.
func ScheduleServerCheck(serverURL string, delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		if isServerReachable(serverURL) {
			return serverOnlineMsg{}
		}
		return serverOfflineMsg{}
	})
}

// isServerReachable checks if the server is reachable at the given URL.
func isServerReachable(serverURL string) bool {
	client := &http.Client{Timeout: HealthCheckTimeout}
	resp, err := client.Get(serverURL + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// =============================================================================
// Filter Commands
// =============================================================================

// CmdApplyFilter adds a filter and refreshes the view
func CmdApplyFilter(filterState *FilterState, filter Filter) tea.Cmd {
	return func() tea.Msg {
		filterState.AddFilter(filter)
		return FilterChangedMsg{FilterState: filterState}
	}
}

// CmdRemoveFilter removes a filter and refreshes the view
func CmdRemoveFilter(filterState *FilterState, filter Filter) tea.Cmd {
	return func() tea.Msg {
		filterState.RemoveFilter(filter)
		return FilterChangedMsg{FilterState: filterState}
	}
}

// CmdClearFilters clears all filters
func CmdClearFilters(filterState *FilterState) tea.Cmd {
	return func() tea.Msg {
		filterState.Clear()
		return FilterChangedMsg{FilterState: filterState}
	}
}

// CmdSetSearchQuery sets the search query
func CmdSetSearchQuery(filterState *FilterState, query string) tea.Cmd {
	return func() tea.Msg {
		filterState.SetSearchQuery(query)
		return FilterChangedMsg{FilterState: filterState}
	}
}

// CmdLoadLabels loads all unique labels from tasks
func CmdLoadLabels(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return func() tea.Msg {
		labels := make(map[string]bool)

		records, err := app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
		if err != nil {
			return LabelsLoadedMsg{Labels: []string{}}
		}

		for _, record := range records {
			// Extract labels (stored as JSON array)
			taskLabels := record.Get("labels")
			if labelSlice, ok := taskLabels.([]interface{}); ok {
				for _, l := range labelSlice {
					if label, ok := l.(string); ok && label != "" {
						labels[label] = true
					}
				}
			} else if labelSlice, ok := taskLabels.([]string); ok {
				for _, label := range labelSlice {
					if label != "" {
						labels[label] = true
					}
				}
			}
		}

		// Convert to sorted slice
		result := make([]string, 0, len(labels))
		for label := range labels {
			result = append(result, label)
		}
		sort.Strings(result)

		return LabelsLoadedMsg{Labels: result}
	}
}

// CmdLoadEpics loads all epics for a board
func CmdLoadEpics(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return func() tea.Msg {
		records, err := app.FindAllRecords("epics",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
		if err != nil {
			return EpicsLoadedMsg{Epics: []EpicOption{}}
		}

		epics := make([]EpicOption, 0, len(records))
		for _, record := range records {
			epics = append(epics, EpicOption{
				ID:    record.Id,
				Title: record.GetString("title"),
				Color: record.GetString("color"),
			})
		}

		// Sort by title
		sort.Slice(epics, func(i, j int) bool {
			return epics[i].Title < epics[j].Title
		})

		return EpicsLoadedMsg{Epics: epics}
	}
}
