# Phase 2: Task Operations

**Goal**: Full CRUD operations for tasks in the TUI - create, view, edit, delete, and move tasks with proper feedback.

**Duration Estimate**: 3-4 days

**Prerequisites**: Phase 1 completed (basic board view with column navigation)

**Deliverable**: Users can create, view, edit, delete, and move tasks entirely within the TUI.

---

## Overview

Phase 2 transforms the TUI from a read-only board viewer into a fully interactive task management interface. This phase implements the complete task lifecycle:

1. **View** - Open task detail panel to see full information
2. **Create** - Add new tasks with a form interface
3. **Edit** - Modify existing tasks
4. **Delete** - Remove tasks with confirmation
5. **Move** - Move tasks between columns and reorder within columns

### Why This Matters

The TUI must provide feature parity with the CLI for task operations. Users should be able to:
- Manage tasks without leaving the terminal
- Get immediate visual feedback for all operations
- Use intuitive keyboard shortcuts for common actions

### Key Design Decisions

**Hybrid Save Pattern**: All write operations use the hybrid save pattern - try the API first (for real-time sync with web UI), fall back to direct database access if the server isn't running. This ensures the TUI works in both connected and disconnected modes.

**Position Management**: Tasks maintain their visual order through a `position` field. When moving tasks, we calculate new positions to preserve ordering without rebalancing the entire column.

**Feedback Messages**: Every operation shows a success or error message in a status bar, providing immediate feedback without modal dialogs.

---

## Tasks

### 2.1 Create Message Types [COMPLETED]

**What**: Define Bubble Tea message types for all CRUD operations.

**Why**: Bubble Tea uses messages for async operations. Each operation needs request and response message types for proper state management.

**Steps**:

1. Create the messages file:
   ```bash
   touch internal/tui/messages.go
   ```

2. Open `internal/tui/messages.go` and add the message types.

**Code**: `internal/tui/messages.go`

```go
package tui

import (
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// =============================================================================
// Task Loading Messages
// =============================================================================

// tasksLoadedMsg is sent when tasks are loaded from the database
type tasksLoadedMsg struct {
	tasks []*core.Record
}

// boardsLoadedMsg is sent when boards are loaded from the database
type boardsLoadedMsg struct {
	boards []*core.Record
}

// =============================================================================
// Task CRUD Messages
// =============================================================================

// taskCreatedMsg is sent when a task is successfully created
type taskCreatedMsg struct {
	task      *core.Record
	displayID string
}

// taskUpdatedMsg is sent when a task is successfully updated
type taskUpdatedMsg struct {
	task *core.Record
}

// taskDeletedMsg is sent when a task is successfully deleted
type taskDeletedMsg struct {
	taskID string
	title  string
}

// taskMovedMsg is sent when a task is moved between columns or reordered
type taskMovedMsg struct {
	task       *core.Record
	fromColumn string
	toColumn   string
}

// =============================================================================
// View State Messages
// =============================================================================

// openTaskDetailMsg requests opening the task detail panel
type openTaskDetailMsg struct {
	task TaskItem
}

// closeTaskDetailMsg requests closing the task detail panel
type closeTaskDetailMsg struct{}

// openTaskFormMsg requests opening the task form for add/edit
type openTaskFormMsg struct {
	mode   FormMode
	taskID string    // Empty for add mode
	task   *TaskItem // Existing task data for edit mode
}

// closeTaskFormMsg requests closing the task form
type closeTaskFormMsg struct {
	cancelled bool
}

// openConfirmDialogMsg requests opening a confirmation dialog
type openConfirmDialogMsg struct {
	title   string
	message string
	onYes   func() // Callback when user confirms
}

// closeConfirmDialogMsg requests closing the confirmation dialog
type closeConfirmDialogMsg struct {
	confirmed bool
}

// =============================================================================
// Error and Status Messages
// =============================================================================

// errMsg represents an error from an async operation
type errMsg struct {
	err     error
	context string // What operation failed
}

// statusMsg shows a temporary status message to the user
type statusMsg struct {
	text     string
	isError  bool
	duration time.Duration
}

// clearStatusMsg clears the status message
type clearStatusMsg struct{}

// =============================================================================
// Form Submission Messages
// =============================================================================

// TaskFormData contains the data from a submitted task form
type TaskFormData struct {
	Title       string
	Description string
	Type        string
	Priority    string
	Column      string
	Labels      []string
	DueDate     string
	EpicID      string
}

// submitTaskFormMsg is sent when the task form is submitted
type submitTaskFormMsg struct {
	mode   FormMode
	taskID string       // Empty for create, set for update
	data   TaskFormData
}
```

**Expected Output**: File compiles without errors.

**Common Mistakes**:
- Forgetting to export message types that need to be accessed from other files
- Not including enough context in error messages for debugging

---

### 2.2 Create Command Functions [COMPLETED]

**What**: Implement async command functions for CRUD operations using the hybrid save pattern.

**Why**: Bubble Tea commands are functions that return messages. These wrap the database operations with proper error handling and the hybrid save pattern.

**Steps**:

1. Create the commands file:
   ```bash
   touch internal/tui/commands.go
   ```

2. Open `internal/tui/commands.go` and implement the commands.

**Code**: `internal/tui/commands.go`

```go
package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
	"github.com/ramtinJ95/EgenSkriven/internal/commands"
)

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

// loadTasks loads all tasks for a specific board
func loadTasks(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return func() tea.Msg {
		var records []*core.Record
		var err error

		if boardID != "" {
			records, err = app.FindAllRecords("tasks",
				dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
			)
		} else {
			records, err = app.FindAllRecords("tasks")
		}

		if err != nil {
			return errMsg{err: err, context: "loading tasks"}
		}
		return tasksLoadedMsg{tasks: records}
	}
}

// loadBoards loads all available boards
func loadBoards(app *pocketbase.PocketBase) tea.Cmd {
	return func() tea.Msg {
		records, err := board.GetAll(app)
		if err != nil {
			return errMsg{err: err, context: "loading boards"}
		}
		return boardsLoadedMsg{boards: records}
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
		position := commands.GetNextPosition(app, data.Column)
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
			position := commands.GetNextPosition(app, data.Column)
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
		position := commands.GetNextPosition(app, targetColumn)

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
			newPosition, err = commands.GetPositionBefore(app, taskID)
		} else {
			newPosition, err = commands.GetPositionAfter(app, taskID)
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

// showStatus shows a status message for a duration
func showStatus(text string, isError bool, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		return statusMsg{
			text:     text,
			isError:  isError,
			duration: duration,
		}
	}
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
			return nil
		}
		// Fall through to direct save on API error
	}
	return app.Save(record)
}

// updateRecordHybrid attempts to update via API first, falls back to direct DB
func updateRecordHybrid(app *pocketbase.PocketBase, record *core.Record) error {
	if isServerRunning() {
		taskData := recordToAPIData(record)
		if err := updateTaskViaAPI(record.Id, taskData); err == nil {
			return nil
		}
		// Fall through to direct save on API error
	}
	return app.Save(record)
}

// deleteRecordHybrid attempts to delete via API first, falls back to direct DB
func deleteRecordHybrid(app *pocketbase.PocketBase, record *core.Record) error {
	if isServerRunning() {
		if err := deleteTaskViaAPI(record.Id); err == nil {
			return nil
		}
		// Fall through to direct delete on API error
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
```

**Expected Output**: 
```
go build ./internal/tui
# No errors
```

**Common Mistakes**:
- Not handling all error cases in the hybrid save pattern
- Forgetting to update task history when making changes
- Not calculating proper positions when moving tasks

---

### 2.3 Implement Task Detail View

**What**: Create a panel that shows full task details with markdown-rendered description.

**Why**: Users need to see complete task information without leaving the board view. The detail panel slides in from the right, showing all fields including the markdown-rendered description.

**Steps**:

1. Create the task detail file:
   ```bash
   touch internal/tui/task_detail.go
   ```

2. Open `internal/tui/task_detail.go` and implement the component.

**Code**: `internal/tui/task_detail.go`

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// TaskDetail displays full task information in a side panel
type TaskDetail struct {
	task     TaskItem
	viewport viewport.Model
	width    int
	height   int
	ready    bool
	keys     taskDetailKeyMap
}

type taskDetailKeyMap struct {
	Close key.Binding
	Edit  key.Binding
	Up    key.Binding
	Down  key.Binding
}

func defaultTaskDetailKeyMap() taskDetailKeyMap {
	return taskDetailKeyMap{
		Close: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "close"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k/up", "scroll up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j/down", "scroll down"),
		),
	}
}

// NewTaskDetail creates a new task detail panel
func NewTaskDetail(task TaskItem, width, height int) *TaskDetail {
	vp := viewport.New(width-4, height-6) // Account for borders and header
	vp.Style = lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	td := &TaskDetail{
		task:     task,
		viewport: vp,
		width:    width,
		height:   height,
		keys:     defaultTaskDetailKeyMap(),
	}

	td.updateContent()
	td.ready = true

	return td
}

// Init initializes the task detail component
func (td *TaskDetail) Init() tea.Cmd {
	return nil
}

// Update handles messages for the task detail
func (td *TaskDetail) Update(msg tea.Msg) (*TaskDetail, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, td.keys.Close):
			return td, func() tea.Msg { return closeTaskDetailMsg{} }
		case key.Matches(msg, td.keys.Edit):
			return td, func() tea.Msg {
				return openTaskFormMsg{
					mode:   FormModeEdit,
					taskID: td.task.ID,
					task:   &td.task,
				}
			}
		}
	case tea.WindowSizeMsg:
		td.SetSize(msg.Width/2, msg.Height-4)
	}

	// Update viewport for scrolling
	td.viewport, cmd = td.viewport.Update(msg)
	return td, cmd
}

// View renders the task detail panel
func (td *TaskDetail) View() string {
	if !td.ready {
		return "Loading..."
	}

	// Header with task ID and type badge
	header := td.renderHeader()

	// Viewport with scrollable content
	content := td.viewport.View()

	// Footer with scroll position and keybindings
	footer := td.renderFooter()

	// Combine all parts
	body := lipgloss.JoinVertical(lipgloss.Left, header, content, footer)

	// Apply border style
	return taskDetailStyle.
		Width(td.width).
		Height(td.height).
		Render(body)
}

// SetSize updates the panel dimensions
func (td *TaskDetail) SetSize(width, height int) {
	td.width = width
	td.height = height
	td.viewport.Width = width - 4
	td.viewport.Height = height - 8 // Account for header and footer
	td.updateContent()
}

// updateContent rebuilds the viewport content
func (td *TaskDetail) updateContent() {
	var sections []string

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Render(td.task.Title)
	sections = append(sections, title)

	// Metadata line
	meta := td.renderMetadata()
	sections = append(sections, meta)

	// Separator
	separator := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render(strings.Repeat("-", td.width-6))
	sections = append(sections, separator)

	// Description (markdown rendered)
	desc := td.renderDescription()
	sections = append(sections, desc)

	// Additional fields
	if len(td.task.Labels) > 0 {
		sections = append(sections, "")
		labels := lipgloss.NewStyle().
			Foreground(secondaryColor).
			Render("Labels: "+strings.Join(td.task.Labels, ", "))
		sections = append(sections, labels)
	}

	if td.task.DueDate != "" {
		dueStyle := lipgloss.NewStyle()
		// Could add overdue styling here
		sections = append(sections, dueStyle.Render("Due: "+td.task.DueDate))
	}

	if td.task.EpicTitle != "" {
		sections = append(sections, "Epic: "+td.task.EpicTitle)
	}

	if len(td.task.BlockedBy) > 0 {
		blockedStyle := lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)
		sections = append(sections, blockedStyle.Render("Blocked by: "+strings.Join(td.task.BlockedBy, ", ")))
	}

	td.viewport.SetContent(strings.Join(sections, "\n"))
}

func (td *TaskDetail) renderHeader() string {
	// Display ID with type badge
	idStyle := lipgloss.NewStyle().
		Foreground(mutedColor).
		Bold(true)

	typeColors := map[string]lipgloss.Color{
		"bug":     typeBug,
		"feature": typeFeature,
		"chore":   typeChore,
	}
	typeColor := typeColors[td.task.Type]
	if typeColor == "" {
		typeColor = mutedColor
	}

	typeStyle := lipgloss.NewStyle().
		Foreground(typeColor).
		Bold(true)

	left := idStyle.Render(td.task.DisplayID) + " " + typeStyle.Render("["+td.task.Type+"]")

	// Priority indicator
	priorityColors := map[string]lipgloss.Color{
		"urgent": priorityUrgent,
		"high":   priorityHigh,
		"medium": priorityMedium,
		"low":    priorityLow,
	}
	prioColor := priorityColors[td.task.Priority]
	if prioColor == "" {
		prioColor = mutedColor
	}
	prioStyle := lipgloss.NewStyle().
		Foreground(prioColor).
		Bold(true)
	right := prioStyle.Render(td.task.Priority)

	// Join with spacing
	gap := td.width - lipgloss.Width(left) - lipgloss.Width(right) - 6
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + right
}

func (td *TaskDetail) renderMetadata() string {
	parts := []string{
		"Column: " + td.task.Column,
	}

	return lipgloss.NewStyle().
		Foreground(mutedColor).
		Render(strings.Join(parts, " | "))
}

func (td *TaskDetail) renderDescription() string {
	if td.task.Description == "" {
		return lipgloss.NewStyle().
			Italic(true).
			Foreground(mutedColor).
			Render("No description")
	}

	// Render markdown with glamour
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(td.width-8),
	)
	if err != nil {
		return td.task.Description
	}

	rendered, err := renderer.Render(td.task.Description)
	if err != nil {
		return td.task.Description
	}

	return strings.TrimSpace(rendered)
}

func (td *TaskDetail) renderFooter() string {
	// Scroll indicator
	scrollInfo := ""
	if td.viewport.TotalLineCount() > td.viewport.Height {
		percent := td.viewport.ScrollPercent() * 100
		scrollInfo = fmt.Sprintf("%.0f%%", percent)
	}

	// Keybindings hint
	hint := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render("esc:close  e:edit  j/k:scroll")

	gap := td.width - lipgloss.Width(scrollInfo) - lipgloss.Width(hint) - 6
	if gap < 1 {
		gap = 1
	}

	return hint + strings.Repeat(" ", gap) + scrollInfo
}

// Task returns the current task
func (td *TaskDetail) Task() TaskItem {
	return td.task
}

// UpdateTask updates the displayed task
func (td *TaskDetail) UpdateTask(task TaskItem) {
	td.task = task
	td.updateContent()
}
```

**Expected Output**: Component compiles and renders task details with markdown.

**Common Mistakes**:
- Not handling long descriptions that need scrolling
- Forgetting to update content when task changes
- Not accounting for border widths in size calculations

---

### 2.4 Implement Task Form

**What**: Create a form for adding and editing tasks with text inputs and select fields.

**Why**: Users need to input task data. The form provides a familiar interface with Tab navigation between fields and selection for enumerated values like type, priority, and column.

**Steps**:

1. Create the task form file:
   ```bash
   touch internal/tui/task_form.go
   ```

2. Open `internal/tui/task_form.go` and implement the form.

**Code**: `internal/tui/task_form.go`

```go
package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormMode indicates whether the form is for adding or editing
type FormMode int

const (
	FormModeAdd FormMode = iota
	FormModeEdit
)

// FormField represents which field is currently focused
type FormField int

const (
	FieldTitle FormField = iota
	FieldDescription
	FieldType
	FieldPriority
	FieldColumn
	FieldLabels
	FieldDueDate
	FieldSubmit
	FieldCancel
)

const numFields = 9 // Total number of focusable fields

// TaskForm handles task creation and editing
type TaskForm struct {
	mode   FormMode
	taskID string // Only set in edit mode

	// Text inputs
	titleInput   textinput.Model
	descInput    textarea.Model
	labelsInput  textinput.Model
	dueDateInput textinput.Model

	// Select fields (index into options)
	typeSelect     int
	prioritySelect int
	columnSelect   int

	// Options for select fields
	types      []string
	priorities []string
	columns    []string

	// Form state
	focusIndex int
	width      int
	height     int
	keys       taskFormKeyMap
}

type taskFormKeyMap struct {
	Submit key.Binding
	Cancel key.Binding
	Next   key.Binding
	Prev   key.Binding
	Left   key.Binding
	Right  key.Binding
}

func defaultTaskFormKeyMap() taskFormKeyMap {
	return taskFormKeyMap{
		Submit: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Next: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		Prev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev field"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("h/left", "prev option"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("l/right", "next option"),
		),
	}
}

// NewTaskForm creates a new task form
func NewTaskForm(mode FormMode, width, height int) *TaskForm {
	// Title input
	ti := textinput.New()
	ti.Placeholder = "Task title..."
	ti.CharLimit = 200
	ti.Width = width - 20
	ti.Focus()

	// Description textarea
	ta := textarea.New()
	ta.Placeholder = "Description (markdown supported)..."
	ta.SetWidth(width - 20)
	ta.SetHeight(5)
	ta.CharLimit = 5000

	// Labels input
	li := textinput.New()
	li.Placeholder = "Labels (comma-separated)..."
	li.CharLimit = 200
	li.Width = width - 20

	// Due date input
	di := textinput.New()
	di.Placeholder = "YYYY-MM-DD"
	di.CharLimit = 10
	di.Width = 15

	return &TaskForm{
		mode:           mode,
		titleInput:     ti,
		descInput:      ta,
		labelsInput:    li,
		dueDateInput:   di,
		types:          []string{"feature", "bug", "chore"},
		priorities:     []string{"low", "medium", "high", "urgent"},
		columns:        []string{"backlog", "todo", "in_progress", "need_input", "review", "done"},
		typeSelect:     0, // Default: feature
		prioritySelect: 1, // Default: medium
		columnSelect:   0, // Default: backlog
		focusIndex:     0,
		width:          width,
		height:         height,
		keys:           defaultTaskFormKeyMap(),
	}
}

// NewTaskFormWithData creates a form pre-filled with task data (for editing)
func NewTaskFormWithData(task *TaskItem, width, height int) *TaskForm {
	f := NewTaskForm(FormModeEdit, width, height)
	f.taskID = task.ID

	// Pre-fill fields
	f.titleInput.SetValue(task.Title)
	f.descInput.SetValue(task.Description)
	f.labelsInput.SetValue(strings.Join(task.Labels, ", "))
	f.dueDateInput.SetValue(task.DueDate)

	// Set select indices
	for i, t := range f.types {
		if t == task.Type {
			f.typeSelect = i
			break
		}
	}
	for i, p := range f.priorities {
		if p == task.Priority {
			f.prioritySelect = i
			break
		}
	}
	for i, c := range f.columns {
		if c == task.Column {
			f.columnSelect = i
			break
		}
	}

	return f
}

// Init initializes the form
func (f *TaskForm) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the form
func (f *TaskForm) Update(msg tea.Msg) (*TaskForm, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keys.Cancel):
			return f, func() tea.Msg {
				return closeTaskFormMsg{cancelled: true}
			}

		case key.Matches(msg, f.keys.Submit):
			return f, f.submit()

		case key.Matches(msg, f.keys.Next):
			f.nextField()
			return f, nil

		case key.Matches(msg, f.keys.Prev):
			f.prevField()
			return f, nil

		case key.Matches(msg, f.keys.Left):
			if f.isSelectField() {
				f.selectPrev()
				return f, nil
			}

		case key.Matches(msg, f.keys.Right):
			if f.isSelectField() {
				f.selectNext()
				return f, nil
			}

		case msg.String() == "enter":
			if FormField(f.focusIndex) == FieldSubmit {
				return f, f.submit()
			} else if FormField(f.focusIndex) == FieldCancel {
				return f, func() tea.Msg {
					return closeTaskFormMsg{cancelled: true}
				}
			}
		}

	case tea.WindowSizeMsg:
		f.SetSize(msg.Width/2, msg.Height-10)
	}

	// Update the currently focused input
	cmd := f.updateFocusedInput(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return f, tea.Batch(cmds...)
}

// View renders the form
func (f *TaskForm) View() string {
	title := "Add Task"
	if f.mode == FormModeEdit {
		title = "Edit Task"
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1).
		Render(title)

	// Build form fields
	var fields []string

	fields = append(fields, f.renderField("Title", f.titleInput.View(), FieldTitle))
	fields = append(fields, f.renderField("Description", f.descInput.View(), FieldDescription))
	fields = append(fields, f.renderSelect("Type", f.types, f.typeSelect, FieldType))
	fields = append(fields, f.renderSelect("Priority", f.priorities, f.prioritySelect, FieldPriority))
	fields = append(fields, f.renderSelect("Column", f.columns, f.columnSelect, FieldColumn))
	fields = append(fields, f.renderField("Labels", f.labelsInput.View(), FieldLabels))
	fields = append(fields, f.renderField("Due Date", f.dueDateInput.View(), FieldDueDate))

	// Buttons
	buttons := f.renderButtons()

	// Help text
	help := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render("tab:next field  ctrl+s:save  esc:cancel")

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		strings.Join(fields, "\n"),
		"",
		buttons,
		"",
		help,
	)

	return formStyle.
		Width(f.width).
		MaxHeight(f.height).
		Render(content)
}

// SetSize updates form dimensions
func (f *TaskForm) SetSize(width, height int) {
	f.width = width
	f.height = height
	f.titleInput.Width = width - 20
	f.descInput.SetWidth(width - 20)
	f.labelsInput.Width = width - 20
}

func (f *TaskForm) renderField(label, input string, field FormField) string {
	focused := FormField(f.focusIndex) == field

	labelStyle := lipgloss.NewStyle().Width(12)
	if focused {
		labelStyle = labelStyle.Foreground(primaryColor).Bold(true)
	}

	return labelStyle.Render(label+":") + " " + input
}

func (f *TaskForm) renderSelect(label string, options []string, selected int, field FormField) string {
	focused := FormField(f.focusIndex) == field

	labelStyle := lipgloss.NewStyle().Width(12)
	if focused {
		labelStyle = labelStyle.Foreground(primaryColor).Bold(true)
	}

	var rendered []string
	for i, opt := range options {
		style := lipgloss.NewStyle()
		if i == selected {
			style = style.Bold(true).Foreground(primaryColor)
			if focused {
				opt = "[" + opt + "]"
			} else {
				opt = " " + opt + " "
			}
		} else {
			style = style.Foreground(mutedColor)
			opt = " " + opt + " "
		}
		rendered = append(rendered, style.Render(opt))
	}

	return labelStyle.Render(label+":") + " " + strings.Join(rendered, "")
}

func (f *TaskForm) renderButtons() string {
	submitFocused := FormField(f.focusIndex) == FieldSubmit
	cancelFocused := FormField(f.focusIndex) == FieldCancel

	submitStyle := buttonStyle
	if submitFocused {
		submitStyle = buttonFocusedStyle
	}
	cancelStyle := buttonStyle
	if cancelFocused {
		cancelStyle = buttonFocusedStyle
	}

	submit := submitStyle.Render("[ Save ]")
	cancel := cancelStyle.Render("[ Cancel ]")

	return lipgloss.JoinHorizontal(lipgloss.Center, submit, "  ", cancel)
}

func (f *TaskForm) nextField() {
	f.blurCurrent()
	f.focusIndex = (f.focusIndex + 1) % numFields
	f.focusCurrent()
}

func (f *TaskForm) prevField() {
	f.blurCurrent()
	f.focusIndex = (f.focusIndex - 1 + numFields) % numFields
	f.focusCurrent()
}

func (f *TaskForm) blurCurrent() {
	switch FormField(f.focusIndex) {
	case FieldTitle:
		f.titleInput.Blur()
	case FieldDescription:
		f.descInput.Blur()
	case FieldLabels:
		f.labelsInput.Blur()
	case FieldDueDate:
		f.dueDateInput.Blur()
	}
}

func (f *TaskForm) focusCurrent() {
	switch FormField(f.focusIndex) {
	case FieldTitle:
		f.titleInput.Focus()
	case FieldDescription:
		f.descInput.Focus()
	case FieldLabels:
		f.labelsInput.Focus()
	case FieldDueDate:
		f.dueDateInput.Focus()
	}
}

func (f *TaskForm) isSelectField() bool {
	field := FormField(f.focusIndex)
	return field == FieldType || field == FieldPriority || field == FieldColumn
}

func (f *TaskForm) selectNext() {
	switch FormField(f.focusIndex) {
	case FieldType:
		f.typeSelect = (f.typeSelect + 1) % len(f.types)
	case FieldPriority:
		f.prioritySelect = (f.prioritySelect + 1) % len(f.priorities)
	case FieldColumn:
		f.columnSelect = (f.columnSelect + 1) % len(f.columns)
	}
}

func (f *TaskForm) selectPrev() {
	switch FormField(f.focusIndex) {
	case FieldType:
		f.typeSelect = (f.typeSelect - 1 + len(f.types)) % len(f.types)
	case FieldPriority:
		f.prioritySelect = (f.prioritySelect - 1 + len(f.priorities)) % len(f.priorities)
	case FieldColumn:
		f.columnSelect = (f.columnSelect - 1 + len(f.columns)) % len(f.columns)
	}
}

func (f *TaskForm) updateFocusedInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch FormField(f.focusIndex) {
	case FieldTitle:
		f.titleInput, cmd = f.titleInput.Update(msg)
	case FieldDescription:
		f.descInput, cmd = f.descInput.Update(msg)
	case FieldLabels:
		f.labelsInput, cmd = f.labelsInput.Update(msg)
	case FieldDueDate:
		f.dueDateInput, cmd = f.dueDateInput.Update(msg)
	}

	return cmd
}

func (f *TaskForm) submit() tea.Cmd {
	// Validate
	title := strings.TrimSpace(f.titleInput.Value())
	if title == "" {
		return showStatus("Title is required", true, 3*secondDuration)
	}

	// Parse labels
	var labels []string
	labelsStr := strings.TrimSpace(f.labelsInput.Value())
	if labelsStr != "" {
		for _, l := range strings.Split(labelsStr, ",") {
			l = strings.TrimSpace(l)
			if l != "" {
				labels = append(labels, l)
			}
		}
	}

	data := TaskFormData{
		Title:       title,
		Description: f.descInput.Value(),
		Type:        f.types[f.typeSelect],
		Priority:    f.priorities[f.prioritySelect],
		Column:      f.columns[f.columnSelect],
		Labels:      labels,
		DueDate:     strings.TrimSpace(f.dueDateInput.Value()),
	}

	return func() tea.Msg {
		return submitTaskFormMsg{
			mode:   f.mode,
			taskID: f.taskID,
			data:   data,
		}
	}
}

const secondDuration = 1_000_000_000 // 1 second in nanoseconds

// Mode returns the form mode
func (f *TaskForm) Mode() FormMode {
	return f.mode
}

// TaskID returns the task ID (for edit mode)
func (f *TaskForm) TaskID() string {
	return f.taskID
}
```

**Expected Output**: Form renders with all fields and Tab navigation works.

**Common Mistakes**:
- Not blurring the previous field when focusing a new one
- Forgetting to handle both keyboard and Tab navigation for select fields
- Not validating required fields before submission

---

### 2.5 Implement Confirmation Dialog

**What**: Create a reusable confirmation dialog for destructive actions like delete.

**Why**: Users should confirm before deleting tasks to prevent accidental data loss. The dialog provides a clear yes/no choice.

**Steps**:

1. Create the confirm dialog file:
   ```bash
   touch internal/tui/confirm.go
   ```

2. Open `internal/tui/confirm.go` and implement the dialog.

**Code**: `internal/tui/confirm.go`

```go
package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmDialog presents a yes/no confirmation to the user
type ConfirmDialog struct {
	title    string
	message  string
	yesLabel string
	noLabel  string
	focused  bool // true = Yes focused, false = No focused
	width    int
	keys     confirmKeyMap
}

type confirmKeyMap struct {
	Yes   key.Binding
	No    key.Binding
	Left  key.Binding
	Right key.Binding
	Tab   key.Binding
}

func defaultConfirmKeyMap() confirmKeyMap {
	return confirmKeyMap{
		Yes: key.NewBinding(
			key.WithKeys("y", "enter"),
			key.WithHelp("y/enter", "confirm"),
		),
		No: key.NewBinding(
			key.WithKeys("n", "esc"),
			key.WithHelp("n/esc", "cancel"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("h/left", "switch"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("l/right", "switch"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch"),
		),
	}
}

// NewConfirmDialog creates a new confirmation dialog
func NewConfirmDialog(title, message string) *ConfirmDialog {
	return &ConfirmDialog{
		title:    title,
		message:  message,
		yesLabel: "Yes",
		noLabel:  "No",
		focused:  false, // Default to No for safety
		width:    50,
		keys:     defaultConfirmKeyMap(),
	}
}

// NewDeleteConfirmDialog creates a confirmation dialog for deleting a task
func NewDeleteConfirmDialog(taskTitle string) *ConfirmDialog {
	return &ConfirmDialog{
		title:    "Delete Task?",
		message:  "Delete \"" + truncateString(taskTitle, 30) + "\"?\nThis action cannot be undone.",
		yesLabel: "Delete",
		noLabel:  "Cancel",
		focused:  false, // Default to Cancel for safety
		width:    50,
		keys:     defaultConfirmKeyMap(),
	}
}

// Init initializes the dialog
func (d *ConfirmDialog) Init() tea.Cmd {
	return nil
}

// Update handles messages for the dialog
func (d *ConfirmDialog) Update(msg tea.Msg) (*ConfirmDialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keys.Yes):
			if d.focused {
				return d, func() tea.Msg {
					return closeConfirmDialogMsg{confirmed: true}
				}
			}
			// If on No button and pressed enter, treat as No
			if msg.String() == "enter" && !d.focused {
				return d, func() tea.Msg {
					return closeConfirmDialogMsg{confirmed: false}
				}
			}
			// Just 'y' always confirms
			if msg.String() == "y" {
				return d, func() tea.Msg {
					return closeConfirmDialogMsg{confirmed: true}
				}
			}

		case key.Matches(msg, d.keys.No):
			return d, func() tea.Msg {
				return closeConfirmDialogMsg{confirmed: false}
			}

		case key.Matches(msg, d.keys.Left), key.Matches(msg, d.keys.Right), key.Matches(msg, d.keys.Tab):
			d.focused = !d.focused
			return d, nil
		}
	}

	return d, nil
}

// View renders the dialog
func (d *ConfirmDialog) View() string {
	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(warningColor).
		MarginBottom(1).
		Render(d.title)

	// Message
	message := lipgloss.NewStyle().
		Width(d.width - 4).
		Render(d.message)

	// Buttons
	yesStyle := buttonStyle
	noStyle := buttonStyle

	if d.focused {
		yesStyle = buttonDangerStyle
	} else {
		noStyle = buttonFocusedStyle
	}

	yesBtn := yesStyle.Render("[ " + d.yesLabel + " ]")
	noBtn := noStyle.Render("[ " + d.noLabel + " ]")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, yesBtn, "  ", noBtn)

	// Combine all parts
	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		message,
		"",
		buttons,
	)

	return confirmDialogStyle.
		Width(d.width).
		Render(content)
}

// SetWidth sets the dialog width
func (d *ConfirmDialog) SetWidth(width int) {
	d.width = width
}

// truncateString truncates a string to a maximum length, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}
```

**Expected Output**: Dialog renders with two buttons and keyboard navigation.

**Common Mistakes**:
- Defaulting to "Yes" which could lead to accidental deletions
- Not handling both Enter and Y/N keys appropriately
- Forgetting to truncate long task titles in the message

---

### 2.6 Add Styles for New Components

**What**: Add Lipgloss styles for the task detail, form, and confirmation dialog.

**Why**: Consistent styling across components creates a polished user experience. These styles should match the existing board styling.

**Steps**:

1. If `internal/tui/styles.go` doesn't exist, create it. Otherwise, add the new styles.

**Code**: Add to `internal/tui/styles.go`

```go
package tui

import "github.com/charmbracelet/lipgloss"

// =============================================================================
// Color Palette
// =============================================================================

var (
	// Brand colors
	primaryColor   = lipgloss.Color("62")  // Blue
	secondaryColor = lipgloss.Color("205") // Pink

	// Status colors
	successColor = lipgloss.Color("82")  // Green
	warningColor = lipgloss.Color("214") // Orange
	errorColor   = lipgloss.Color("196") // Red

	// Priority colors
	priorityUrgent = lipgloss.Color("196") // Red
	priorityHigh   = lipgloss.Color("208") // Orange
	priorityMedium = lipgloss.Color("226") // Yellow
	priorityLow    = lipgloss.Color("240") // Gray

	// Type colors
	typeBug     = lipgloss.Color("196") // Red
	typeFeature = lipgloss.Color("39")  // Cyan
	typeChore   = lipgloss.Color("240") // Gray

	// Column colors
	columnBacklog    = lipgloss.Color("240")
	columnTodo       = lipgloss.Color("39")
	columnInProgress = lipgloss.Color("214")
	columnNeedInput  = lipgloss.Color("205")
	columnReview     = lipgloss.Color("205")
	columnDone       = lipgloss.Color("82")

	// UI colors
	borderColor      = lipgloss.Color("240")
	focusBorderColor = lipgloss.Color("62")
	mutedColor       = lipgloss.Color("240")
)

// =============================================================================
// Component Styles
// =============================================================================

// Task Detail Panel
var taskDetailStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(focusBorderColor).
	Padding(1, 2)

// Task Form
var formStyle = lipgloss.NewStyle().
	Border(lipgloss.DoubleBorder()).
	BorderForeground(primaryColor).
	Padding(1, 2)

// Confirmation Dialog
var confirmDialogStyle = lipgloss.NewStyle().
	Border(lipgloss.DoubleBorder()).
	BorderForeground(warningColor).
	Padding(1, 2).
	Align(lipgloss.Center)

// Buttons
var buttonStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("252")).
	Background(lipgloss.Color("238")).
	Padding(0, 2)

var buttonFocusedStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("0")).
	Background(primaryColor).
	Padding(0, 2).
	Bold(true)

var buttonDangerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("0")).
	Background(errorColor).
	Padding(0, 2).
	Bold(true)

// Status Bar
var statusBarStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("236")).
	Foreground(lipgloss.Color("252")).
	Padding(0, 1)

var statusBarErrorStyle = lipgloss.NewStyle().
	Background(errorColor).
	Foreground(lipgloss.Color("255")).
	Padding(0, 1).
	Bold(true)

var statusBarSuccessStyle = lipgloss.NewStyle().
	Background(successColor).
	Foreground(lipgloss.Color("0")).
	Padding(0, 1).
	Bold(true)

// =============================================================================
// Column Styles
// =============================================================================

var focusedColumnStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(focusBorderColor).
	Padding(0, 1)

var blurredColumnStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(borderColor).
	Padding(0, 1)

// =============================================================================
// Helper Functions
// =============================================================================

// ColumnColor returns the color for a given column status
func ColumnColor(status string) lipgloss.Color {
	colors := map[string]lipgloss.Color{
		"backlog":     columnBacklog,
		"todo":        columnTodo,
		"in_progress": columnInProgress,
		"need_input":  columnNeedInput,
		"review":      columnReview,
		"done":        columnDone,
	}
	if c, ok := colors[status]; ok {
		return c
	}
	return borderColor
}

// PriorityColor returns the color for a given priority
func PriorityColor(priority string) lipgloss.Color {
	colors := map[string]lipgloss.Color{
		"urgent": priorityUrgent,
		"high":   priorityHigh,
		"medium": priorityMedium,
		"low":    priorityLow,
	}
	if c, ok := colors[priority]; ok {
		return c
	}
	return mutedColor
}

// TypeColor returns the color for a given task type
func TypeColor(taskType string) lipgloss.Color {
	colors := map[string]lipgloss.Color{
		"bug":     typeBug,
		"feature": typeFeature,
		"chore":   typeChore,
	}
	if c, ok := colors[taskType]; ok {
		return c
	}
	return mutedColor
}
```

**Expected Output**: Styles compile and produce visually distinct components.

**Common Mistakes**:
- Using colors that don't have enough contrast
- Not accounting for terminal color limitations
- Inconsistent padding/margin across related components

---

### 2.7 Update App Model for CRUD Operations

**What**: Integrate all new components into the main App model and handle CRUD messages.

**Why**: The App model is the central hub that coordinates all components. It needs to handle the new message types and manage the overlay state for forms and dialogs.

**Steps**:

1. Update the existing App model (typically in `internal/tui/app.go`) to handle new messages.

**Code**: Key additions to `internal/tui/app.go`

```go
// Add these fields to the App struct
type App struct {
	// ... existing fields ...

	// Overlays
	taskDetail    *TaskDetail
	taskForm      *TaskForm
	confirmDialog *ConfirmDialog

	// Status
	statusText  string
	statusError bool

	// Current board record (needed for creating tasks)
	currentBoardRecord *core.Record
}

// Add to the Update function's switch statement
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// =================================================================
	// Task CRUD Messages
	// =================================================================

	case taskCreatedMsg:
		// Reload tasks and show success message
		cmds = append(cmds,
			loadTasks(a.pb, a.currentBoardRecord.Id),
			showStatus("Created: "+msg.task.GetString("title")+" ["+msg.displayID+"]", false, 3*time.Second),
		)
		// Close form
		a.taskForm = nil
		a.view = ViewBoard

	case taskUpdatedMsg:
		// Reload tasks and show success message
		cmds = append(cmds,
			loadTasks(a.pb, a.currentBoardRecord.Id),
			showStatus("Updated: "+msg.task.GetString("title"), false, 3*time.Second),
		)
		// Close form and detail if open
		a.taskForm = nil
		if a.taskDetail != nil {
			// Update the detail view with new data
			a.taskDetail.UpdateTask(recordToTaskItem(msg.task, a.currentBoardRecord))
		}
		a.view = ViewBoard

	case taskDeletedMsg:
		// Reload tasks and show success message
		cmds = append(cmds,
			loadTasks(a.pb, a.currentBoardRecord.Id),
			showStatus("Deleted: "+msg.title, false, 3*time.Second),
		)
		// Close any overlays
		a.taskDetail = nil
		a.confirmDialog = nil
		a.view = ViewBoard

	case taskMovedMsg:
		// Reload tasks and show status
		cmds = append(cmds,
			loadTasks(a.pb, a.currentBoardRecord.Id),
		)
		if msg.fromColumn != msg.toColumn {
			cmds = append(cmds,
				showStatus("Moved to "+msg.toColumn, false, 2*time.Second),
			)
		}

	case tasksLoadedMsg:
		// Distribute tasks to columns
		a.distributeTasks(msg.tasks)

	// =================================================================
	// View State Messages
	// =================================================================

	case openTaskDetailMsg:
		a.taskDetail = NewTaskDetail(msg.task, a.width/2, a.height-4)
		a.view = ViewTaskDetail

	case closeTaskDetailMsg:
		a.taskDetail = nil
		a.view = ViewBoard

	case openTaskFormMsg:
		if msg.mode == FormModeEdit && msg.task != nil {
			a.taskForm = NewTaskFormWithData(msg.task, a.width/2, a.height-10)
		} else {
			a.taskForm = NewTaskForm(FormModeAdd, a.width/2, a.height-10)
		}
		a.view = ViewTaskForm

	case closeTaskFormMsg:
		a.taskForm = nil
		a.view = ViewBoard

	case openConfirmDialogMsg:
		a.confirmDialog = NewConfirmDialog(msg.title, msg.message)
		a.view = ViewConfirm

	case closeConfirmDialogMsg:
		if msg.confirmed {
			// Execute the pending action (delete)
			if a.pendingDeleteTaskID != "" {
				cmds = append(cmds, deleteTask(a.pb, a.pendingDeleteTaskID))
				a.pendingDeleteTaskID = ""
			}
		}
		a.confirmDialog = nil
		a.view = ViewBoard

	case submitTaskFormMsg:
		if msg.mode == FormModeAdd {
			cmds = append(cmds, createTask(a.pb, a.currentBoardRecord, msg.data))
		} else {
			cmds = append(cmds, updateTask(a.pb, msg.taskID, msg.data))
		}

	// =================================================================
	// Status Messages
	// =================================================================

	case statusMsg:
		a.statusText = msg.text
		a.statusError = msg.isError
		cmds = append(cmds, clearStatusAfter(msg.duration))

	case clearStatusMsg:
		a.statusText = ""
		a.statusError = false

	// =================================================================
	// Error Messages
	// =================================================================

	case errMsg:
		a.statusText = msg.context + ": " + msg.err.Error()
		a.statusError = true
		cmds = append(cmds, clearStatusAfter(5*time.Second))

	// =================================================================
	// Keyboard Input
	// =================================================================

	case tea.KeyMsg:
		// Handle based on current view
		switch a.view {
		case ViewBoard:
			return a.handleBoardKeys(msg)
		case ViewTaskDetail:
			return a.handleDetailKeys(msg)
		case ViewTaskForm:
			return a.handleFormKeys(msg)
		case ViewConfirm:
			return a.handleConfirmKeys(msg)
		}
	}

	return a, tea.Batch(cmds...)
}

// handleBoardKeys handles keyboard input when in board view
func (a *App) handleBoardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "n":
		// New task
		return a, func() tea.Msg {
			return openTaskFormMsg{mode: FormModeAdd}
		}

	case "enter":
		// Open task detail
		task := a.getSelectedTask()
		if task != nil {
			return a, func() tea.Msg {
				return openTaskDetailMsg{task: *task}
			}
		}

	case "e":
		// Edit task
		task := a.getSelectedTask()
		if task != nil {
			return a, func() tea.Msg {
				return openTaskFormMsg{
					mode:   FormModeEdit,
					taskID: task.ID,
					task:   task,
				}
			}
		}

	case "d":
		// Delete task (with confirmation)
		task := a.getSelectedTask()
		if task != nil {
			a.pendingDeleteTaskID = task.ID
			a.confirmDialog = NewDeleteConfirmDialog(task.Title)
			a.view = ViewConfirm
		}

	case "H":
		// Move task to previous column
		return a.moveTaskLeft()

	case "L":
		// Move task to next column
		return a.moveTaskRight()

	case "K":
		// Move task up in column
		return a.reorderTaskUp()

	case "J":
		// Move task down in column
		return a.reorderTaskDown()

	case "1", "2", "3", "4", "5", "6":
		// Move task to column by number
		columnIndex := int(msg.String()[0] - '1')
		return a.moveTaskToColumnIndex(columnIndex)
	}

	// ... handle navigation keys (h, j, k, l) ...

	return a, nil
}

// Helper methods for task operations
func (a *App) moveTaskLeft() (tea.Model, tea.Cmd) {
	task := a.getSelectedTask()
	if task == nil {
		return a, nil
	}

	currentIndex := a.getColumnIndex(task.Column)
	if currentIndex <= 0 {
		return a, showStatus("Already in first column", true, 2*time.Second)
	}

	targetColumn := a.getColumnByIndex(currentIndex - 1)
	return a, moveTaskToColumn(a.pb, task.ID, targetColumn)
}

func (a *App) moveTaskRight() (tea.Model, tea.Cmd) {
	task := a.getSelectedTask()
	if task == nil {
		return a, nil
	}

	currentIndex := a.getColumnIndex(task.Column)
	if currentIndex >= len(a.columns)-1 {
		return a, showStatus("Already in last column", true, 2*time.Second)
	}

	targetColumn := a.getColumnByIndex(currentIndex + 1)
	return a, moveTaskToColumn(a.pb, task.ID, targetColumn)
}

func (a *App) reorderTaskUp() (tea.Model, tea.Cmd) {
	task := a.getSelectedTask()
	if task == nil {
		return a, nil
	}
	return a, reorderTaskInColumn(a.pb, task.ID, true)
}

func (a *App) reorderTaskDown() (tea.Model, tea.Cmd) {
	task := a.getSelectedTask()
	if task == nil {
		return a, nil
	}
	return a, reorderTaskInColumn(a.pb, task.ID, false)
}

func (a *App) moveTaskToColumnIndex(index int) (tea.Model, tea.Cmd) {
	if index < 0 || index >= len(a.columns) {
		return a, nil
	}

	task := a.getSelectedTask()
	if task == nil {
		return a, nil
	}

	targetColumn := a.getColumnByIndex(index)
	if targetColumn == task.Column {
		return a, nil // Already in this column
	}

	return a, moveTaskToColumn(a.pb, task.ID, targetColumn)
}

// recordToTaskItem converts a database record to a TaskItem
func recordToTaskItem(record *core.Record, boardRecord *core.Record) TaskItem {
	var labels []string
	if rawLabels := record.Get("labels"); rawLabels != nil {
		if l, ok := rawLabels.([]any); ok {
			for _, item := range l {
				if s, ok := item.(string); ok {
					labels = append(labels, s)
				}
			}
		}
	}

	var blockedBy []string
	if rawBlocked := record.Get("blocked_by"); rawBlocked != nil {
		if b, ok := rawBlocked.([]any); ok {
			for _, item := range b {
				if s, ok := item.(string); ok {
					blockedBy = append(blockedBy, s)
				}
			}
		}
	}

	displayID := record.Id[:7] // Short ID fallback
	if boardRecord != nil {
		displayID = board.FormatDisplayID(
			boardRecord.GetString("prefix"),
			record.GetInt("seq"),
		)
	}

	return TaskItem{
		ID:          record.Id,
		DisplayID:   displayID,
		Title:       record.GetString("title"),
		Description: record.GetString("description"),
		Type:        record.GetString("type"),
		Priority:    record.GetString("priority"),
		Column:      record.GetString("column"),
		Labels:      labels,
		DueDate:     record.GetString("due_date"),
		Epic:        record.GetString("epic"),
		BlockedBy:   blockedBy,
		IsBlocked:   len(blockedBy) > 0,
		Position:    record.GetFloat("position"),
	}
}
```

**Expected Output**: App correctly routes messages and manages view state.

**Common Mistakes**:
- Forgetting to reset overlay state when switching views
- Not reloading tasks after CRUD operations
- Leaving pending operations (like delete confirmations) in state

---

### 2.8 Write Tests for CRUD Operations

**What**: Create tests to verify the CRUD operations work correctly.

**Why**: Tests ensure the hybrid save pattern works, position calculations are correct, and message handling is proper.

**Steps**:

1. Create the test file:
   ```bash
   touch internal/tui/commands_test.go
   ```

2. Implement the tests.

**Code**: `internal/tui/commands_test.go`

```go
package tui

import (
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// TestCreateTaskCommand verifies task creation works correctly
func TestCreateTaskCommand(t *testing.T) {
	app := testutil.NewTestApp(t)

	// Create tasks collection
	testutil.CreateTestCollection(t, app, "tasks",
		&core.TextField{Name: "title", Required: true},
		&core.TextField{Name: "description"},
		&core.TextField{Name: "type"},
		&core.TextField{Name: "priority"},
		&core.TextField{Name: "column"},
		&core.NumberField{Name: "position"},
		&core.JSONField{Name: "labels"},
		&core.JSONField{Name: "blocked_by"},
		&core.TextField{Name: "created_by"},
		&core.TextField{Name: "board"},
		&core.NumberField{Name: "seq"},
		&core.TextField{Name: "due_date"},
		&core.TextField{Name: "epic"},
		&core.JSONField{Name: "history"},
	)

	// Create boards collection for sequence tracking
	testutil.CreateTestCollection(t, app, "boards",
		&core.TextField{Name: "name", Required: true},
		&core.TextField{Name: "prefix", Required: true},
		&core.NumberField{Name: "next_seq"},
		&core.JSONField{Name: "columns"},
	)

	// Create a test board
	boardRecord, err := board.Create(app, board.CreateInput{
		Name:   "Test",
		Prefix: "TST",
	})
	require.NoError(t, err)

	// Get the board record
	boardRec, err := app.FindRecordById("boards", boardRecord.ID)
	require.NoError(t, err)

	// Execute the create command
	data := TaskFormData{
		Title:       "Test Task",
		Description: "This is a test",
		Type:        "feature",
		Priority:    "medium",
		Column:      "backlog",
		Labels:      []string{"test", "unit"},
	}

	cmd := createTask(app, boardRec, data)
	msg := cmd()

	// Verify result
	created, ok := msg.(taskCreatedMsg)
	require.True(t, ok, "Expected taskCreatedMsg, got %T", msg)
	assert.NotEmpty(t, created.task.Id)
	assert.Equal(t, "Test Task", created.task.GetString("title"))
	assert.Equal(t, "This is a test", created.task.GetString("description"))
	assert.Equal(t, "feature", created.task.GetString("type"))
	assert.Equal(t, "medium", created.task.GetString("priority"))
	assert.Equal(t, "backlog", created.task.GetString("column"))
	assert.Equal(t, "TST-1", created.displayID)
	assert.Equal(t, "tui", created.task.GetString("created_by"))
}

// TestUpdateTaskCommand verifies task updates work correctly
func TestUpdateTaskCommand(t *testing.T) {
	app := testutil.NewTestApp(t)

	// Create collection and board (same as above)
	testutil.CreateTestCollection(t, app, "tasks",
		&core.TextField{Name: "title", Required: true},
		&core.TextField{Name: "description"},
		&core.TextField{Name: "type"},
		&core.TextField{Name: "priority"},
		&core.TextField{Name: "column"},
		&core.NumberField{Name: "position"},
		&core.JSONField{Name: "labels"},
		&core.JSONField{Name: "blocked_by"},
		&core.TextField{Name: "created_by"},
		&core.TextField{Name: "board"},
		&core.NumberField{Name: "seq"},
		&core.TextField{Name: "due_date"},
		&core.TextField{Name: "epic"},
		&core.JSONField{Name: "history"},
	)

	testutil.CreateTestCollection(t, app, "boards",
		&core.TextField{Name: "name", Required: true},
		&core.TextField{Name: "prefix", Required: true},
		&core.NumberField{Name: "next_seq"},
		&core.JSONField{Name: "columns"},
	)

	boardRecord, _ := board.Create(app, board.CreateInput{Name: "Test", Prefix: "TST"})
	boardRec, _ := app.FindRecordById("boards", boardRecord.ID)

	// Create a task first
	data := TaskFormData{
		Title:    "Original Title",
		Type:     "feature",
		Priority: "low",
		Column:   "backlog",
	}
	cmd := createTask(app, boardRec, data)
	created := cmd().(taskCreatedMsg)

	// Update the task
	updateData := TaskFormData{
		Title:       "Updated Title",
		Description: "Added description",
		Type:        "bug",
		Priority:    "high",
		Column:      "backlog",
	}

	updateCmd := updateTask(app, created.task.Id, updateData)
	updateMsg := updateCmd()

	// Verify result
	updated, ok := updateMsg.(taskUpdatedMsg)
	require.True(t, ok, "Expected taskUpdatedMsg, got %T", updateMsg)
	assert.Equal(t, "Updated Title", updated.task.GetString("title"))
	assert.Equal(t, "Added description", updated.task.GetString("description"))
	assert.Equal(t, "bug", updated.task.GetString("type"))
	assert.Equal(t, "high", updated.task.GetString("priority"))
}

// TestDeleteTaskCommand verifies task deletion works correctly
func TestDeleteTaskCommand(t *testing.T) {
	app := testutil.NewTestApp(t)

	// Create collection
	collection := testutil.CreateTestCollection(t, app, "tasks",
		&core.TextField{Name: "title", Required: true},
		&core.TextField{Name: "column"},
		&core.NumberField{Name: "position"},
	)

	// Create a task directly
	record := core.NewRecord(collection)
	record.Set("title", "Task to Delete")
	record.Set("column", "backlog")
	record.Set("position", 1000.0)
	require.NoError(t, app.Save(record))

	// Delete the task
	cmd := deleteTask(app, record.Id)
	msg := cmd()

	// Verify result
	deleted, ok := msg.(taskDeletedMsg)
	require.True(t, ok, "Expected taskDeletedMsg, got %T", msg)
	assert.Equal(t, record.Id, deleted.taskID)
	assert.Equal(t, "Task to Delete", deleted.title)

	// Verify task no longer exists
	_, err := app.FindRecordById("tasks", record.Id)
	assert.Error(t, err)
}

// TestMoveTaskToColumnCommand verifies task column movement
func TestMoveTaskToColumnCommand(t *testing.T) {
	app := testutil.NewTestApp(t)

	// Create collection
	collection := testutil.CreateTestCollection(t, app, "tasks",
		&core.TextField{Name: "title", Required: true},
		&core.TextField{Name: "column"},
		&core.NumberField{Name: "position"},
		&core.JSONField{Name: "history"},
	)

	// Create a task in backlog
	record := core.NewRecord(collection)
	record.Set("title", "Task to Move")
	record.Set("column", "backlog")
	record.Set("position", 1000.0)
	record.Set("history", []map[string]any{})
	require.NoError(t, app.Save(record))

	// Move to todo
	cmd := moveTaskToColumn(app, record.Id, "todo")
	msg := cmd()

	// Verify result
	moved, ok := msg.(taskMovedMsg)
	require.True(t, ok, "Expected taskMovedMsg, got %T", msg)
	assert.Equal(t, "backlog", moved.fromColumn)
	assert.Equal(t, "todo", moved.toColumn)
	assert.Equal(t, "todo", moved.task.GetString("column"))

	// Verify in database
	updated, err := app.FindRecordById("tasks", record.Id)
	require.NoError(t, err)
	assert.Equal(t, "todo", updated.GetString("column"))
}

// TestReorderTaskInColumnCommand verifies task reordering
func TestReorderTaskInColumnCommand(t *testing.T) {
	app := testutil.NewTestApp(t)

	// Create collection
	collection := testutil.CreateTestCollection(t, app, "tasks",
		&core.TextField{Name: "title", Required: true},
		&core.TextField{Name: "column"},
		&core.NumberField{Name: "position"},
	)

	// Create three tasks with known positions
	positions := []float64{1000, 2000, 3000}
	var tasks []*core.Record
	for i, pos := range positions {
		record := core.NewRecord(collection)
		record.Set("title", "Task "+string(rune('A'+i)))
		record.Set("column", "backlog")
		record.Set("position", pos)
		require.NoError(t, app.Save(record))
		tasks = append(tasks, record)
	}

	// Move middle task up (should go between first task and top)
	cmd := reorderTaskInColumn(app, tasks[1].Id, true)
	msg := cmd()

	moved, ok := msg.(taskMovedMsg)
	require.True(t, ok, "Expected taskMovedMsg, got %T", msg)

	// New position should be less than original
	assert.Less(t, moved.task.GetFloat("position"), 2000.0)
}

// TestStatusMessages verifies status message handling
func TestStatusMessages(t *testing.T) {
	// Test success status
	cmd := showStatus("Task created", false, time.Second)
	msg := cmd()

	status, ok := msg.(statusMsg)
	require.True(t, ok)
	assert.Equal(t, "Task created", status.text)
	assert.False(t, status.isError)

	// Test error status
	cmd = showStatus("Failed to save", true, time.Second)
	msg = cmd()

	status, ok = msg.(statusMsg)
	require.True(t, ok)
	assert.Equal(t, "Failed to save", status.text)
	assert.True(t, status.isError)
}

// TestTaskFormValidation verifies form validation
func TestTaskFormValidation(t *testing.T) {
	form := NewTaskForm(FormModeAdd, 80, 40)

	// Try to submit without title
	cmd := form.submit()
	msg := cmd()

	// Should return status error about missing title
	status, ok := msg.(statusMsg)
	require.True(t, ok)
	assert.True(t, status.isError)
	assert.Contains(t, status.text, "Title")
}

// TestTaskFormLabels verifies label parsing
func TestTaskFormLabels(t *testing.T) {
	form := NewTaskForm(FormModeAdd, 80, 40)
	form.titleInput.SetValue("Test Task")
	form.labelsInput.SetValue("bug, frontend, urgent")

	cmd := form.submit()
	msg := cmd()

	submit, ok := msg.(submitTaskFormMsg)
	require.True(t, ok)
	assert.Equal(t, []string{"bug", "frontend", "urgent"}, submit.data.Labels)
}

// TestConfirmDialogDefault verifies confirm dialog defaults to No
func TestConfirmDialogDefault(t *testing.T) {
	dialog := NewDeleteConfirmDialog("Test Task")

	// Default should be focused on No/Cancel
	assert.False(t, dialog.focused)
}
```

**Expected Output**:
```bash
go test ./internal/tui -v -run "TestCreate|TestUpdate|TestDelete|TestMove|TestReorder|TestStatus|TestTaskForm|TestConfirm"
# All tests should pass
```

**Common Mistakes**:
- Not setting up all required collection fields in tests
- Forgetting to check both success and error cases
- Not testing edge cases like empty columns or single-task columns

---

## Verification Checklist

Complete each item to verify Phase 2 is working correctly.

### Component Verification

- [ ] **Task Detail renders correctly**
  ```
  Launch TUI, select a task, press Enter
  Should see task detail panel with all fields
  Description should be markdown-rendered
  ```

- [ ] **Task Form works for adding**
  ```
  Press 'n' to open add form
  Fill in fields, Tab between them
  Press Ctrl+S to save
  New task should appear in column
  ```

- [ ] **Task Form works for editing**
  ```
  Select a task, press 'e' to edit
  Form should be pre-filled with task data
  Change fields and save
  Task should update
  ```

- [ ] **Delete confirmation works**
  ```
  Select a task, press 'd'
  Confirmation dialog should appear
  Press 'n' or Esc to cancel
  Press 'y' to confirm deletion
  ```

### Movement Verification

- [ ] **H/L moves task between columns**
  ```
  Select a task in "backlog"
  Press 'L' to move right
  Task should move to "todo"
  Press 'H' to move left
  Task should return to "backlog"
  ```

- [ ] **1-5 keys move to specific column**
  ```
  Select a task
  Press '3' to move to in_progress (3rd column)
  Task should appear in in_progress
  ```

- [ ] **Shift+J/K reorders within column**
  ```
  Have multiple tasks in a column
  Select middle task
  Press Shift+K to move up
  Press Shift+J to move down
  Order should persist after reload
  ```

### Feedback Verification

- [ ] **Success messages show**
  ```
  Create a task
  Green status bar should show "Created: [title]"
  Message should disappear after 3 seconds
  ```

- [ ] **Error messages show**
  ```
  Try to submit form with empty title
  Red status bar should show error
  ```

### Hybrid Save Verification

- [ ] **Works with server running**
  ```
  Start server: make dev
  Create/edit/delete tasks in TUI
  Check web UI - changes should appear immediately
  ```

- [ ] **Works without server**
  ```
  Stop the server
  Create/edit/delete tasks in TUI
  Restart TUI - changes should persist
  Start server - tasks should be visible in web UI
  ```

### Tests Pass

- [ ] **All tests pass**
  ```bash
  go test ./internal/tui -v
  ```

---

## File Summary

| File | Lines (approx) | Purpose |
|------|----------------|---------|
| `internal/tui/messages.go` | ~80 | Message types for CRUD operations |
| `internal/tui/commands.go` | ~400 | Async command functions with hybrid save |
| `internal/tui/task_detail.go` | ~200 | Task detail panel with markdown |
| `internal/tui/task_form.go` | ~350 | Add/edit task form |
| `internal/tui/confirm.go` | ~130 | Confirmation dialog |
| `internal/tui/styles.go` | ~120 | Lipgloss styles |
| `internal/tui/commands_test.go` | ~250 | Tests for CRUD operations |

**Total new code**: ~1,530 lines

---

## What You Should Have Now

After completing Phase 2, your TUI should:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ Board: Work (WRK)                                                      [?]  │
├─────────────┬─────────────┬─────────────┬─────────────┬─────────────────────┤
│ Backlog (2) │ Todo (1)    │ In Progress │ Review (1)  │ Done (3)            │
├─────────────┼─────────────┼─────────────┼─────────────┼─────────────────────┤
│ ● WRK-1     │ ● WRK-4     │             │ ● WRK-6     │ ● WRK-2             │
│   Add dark  │   Fix bug   │             │   Code rev  │   Initial setup     │
│   mode      │   [bug]     │             │             │                     │
│             │             │             │             │ ● WRK-3             │
│ ● WRK-5     │             │             │             │   Add tests         │
│   Refactor  │             │             │             │                     │
│             │             │             │             │ ● WRK-7             │
│             │             │             │             │   Documentation     │
└─────────────┴─────────────┴─────────────┴─────────────┴─────────────────────┘
│ Created: Add dark mode [WRK-8]                                              │
└─────────────────────────────────────────────────────────────────────────────┘
│ n:new  e:edit  d:delete  H/L:move  J/K:reorder  enter:detail  ?:help  q:quit│
└─────────────────────────────────────────────────────────────────────────────┘
```

**Capabilities**:
- Press `Enter` to view task details with markdown description
- Press `n` to create a new task
- Press `e` to edit selected task
- Press `d` to delete with confirmation
- Press `H`/`L` to move task between columns
- Press `Shift+J`/`Shift+K` to reorder within column
- Press `1-5` to move to specific column
- Status bar shows success/error feedback
- All operations use hybrid save (API first, DB fallback)

---

## Next Phase

**Phase 3: Multi-Board Support** will add:
- Board selector component (b key to open)
- Load and switch between multiple boards
- Display current board name in header
- Persist last-used board in config
- Support board-specific columns

---

## Troubleshooting

### "Task not appearing after create"

**Problem**: Task was created but doesn't show in the column.

**Solution**: 
1. Check if tasks are being reloaded after create: `loadTasks` should be called
2. Verify the task's column matches a visible column
3. Check console for errors during save

### "Form fields not accepting input"

**Problem**: Typing in form fields doesn't work.

**Solution**:
1. Ensure the correct field is focused (`Focus()` called)
2. Check that key messages are being routed to the form
3. Verify the view state is `ViewTaskForm`

### "Confirmation dialog accepts wrong keys"

**Problem**: Pressing any key confirms or cancels unexpectedly.

**Solution**:
1. Check key bindings match expected keys
2. Ensure Enter only confirms when Yes is focused
3. Verify default focus is on "No/Cancel"

### "Task movement not persisting"

**Problem**: Task moves visually but returns to original position after reload.

**Solution**:
1. Check that `updateRecordHybrid` is being called
2. Verify position is being calculated and set
3. Check for save errors in status bar

### "Markdown not rendering in detail view"

**Problem**: Description shows raw markdown instead of rendered.

**Solution**:
1. Ensure glamour is imported and working
2. Check viewport width is being set correctly
3. Try simpler markdown to test rendering

### "API errors when server is running"

**Problem**: Hybrid save fails even with server running.

**Solution**:
1. Verify server URL is correct (`http://localhost:8090`)
2. Check server is actually running (`/api/health` returns 200)
3. Look for CORS or authentication issues in server logs

---

## Glossary

| Term | Definition |
|------|------------|
| **CRUD** | Create, Read, Update, Delete - the four basic data operations |
| **Hybrid Save** | Pattern that tries API first, falls back to direct database |
| **Position** | Float64 value determining task order within a column |
| **Display ID** | Human-readable task ID like "WRK-123" |
| **Overlay** | UI component that appears above the main board (detail, form, dialog) |
| **Focus Index** | Which form field is currently active for input |
| **Select Field** | Form field with predefined options (type, priority, column) |
| **Glamour** | Go library for rendering markdown in terminal |
| **Viewport** | Scrollable content area within a component |
| **tea.Cmd** | Bubble Tea command - a function that returns a message |
