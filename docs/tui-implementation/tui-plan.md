# TUI Implementation Plan

## Overview

This document outlines the implementation plan for adding a Terminal User Interface (TUI) to EgenSkriven. The TUI will provide a full-featured kanban board experience in the terminal, with near feature parity to the web UI.

### Goals

- Full kanban board with 5 default columns (backlog, todo, in_progress, review, done)
- Multi-board support with board switching
- Real-time synchronization when server is running
- Feature parity with web UI for all kanban-related functionality
- Keyboard-driven interface with intuitive navigation

### Non-Goals

- Theme customization (use sensible defaults)
- Mouse support (keyboard-only initially)
- Offline-first architecture (leverage existing hybrid mode)

---

## Technology Stack

### Primary Framework: Bubble Tea

**Package:** `github.com/charmbracelet/bubbletea`

**Why Bubble Tea:**
- Most popular Go TUI framework (38k+ GitHub stars)
- Elm-style Model-View-Update architecture (functional, testable)
- Official kanban example (`kancli`) to learn from
- Excellent component library (`bubbles`)
- Beautiful styling with Lip Gloss
- Strong async support via Commands

### Supporting Libraries

| Library | Version | Purpose |
|---------|---------|---------|
| `github.com/charmbracelet/bubbletea` | latest | Core TUI framework |
| `github.com/charmbracelet/bubbles` | latest | Pre-built components (list, help, textinput, viewport) |
| `github.com/charmbracelet/lipgloss` | latest | Styling and layout |
| `github.com/charmbracelet/glamour` | latest | Markdown rendering (for task descriptions) |

---

## Architecture

### Directory Structure

```
internal/
├── tui/
│   ├── app.go              # Main application model, tea.Program entry
│   ├── board.go            # Board model (kanban view)
│   ├── column.go           # Column component wrapping bubbles/list
│   ├── task_item.go        # Task implementing list.Item interface
│   ├── task_detail.go      # Task detail view (slide-in panel)
│   ├── task_form.go        # Add/edit task form
│   ├── board_selector.go   # Board switching UI
│   ├── filter_bar.go       # Filter controls
│   ├── help.go             # Help overlay
│   ├── keys.go             # Keybinding definitions
│   ├── styles.go           # Lipgloss style definitions
│   ├── messages.go         # Custom message types
│   ├── commands.go         # Async command functions
│   └── realtime.go         # Real-time subscription handling
├── commands/
│   └── tui.go              # CLI command to launch TUI
```

### Component Hierarchy

```
App (root model)
├── BoardSelector (optional overlay)
├── Board (main kanban view)
│   ├── Header (board name, task count)
│   ├── FilterBar (active filters display)
│   ├── Columns (horizontal layout)
│   │   ├── Column[0] - Backlog
│   │   │   └── list.Model with TaskItems
│   │   ├── Column[1] - Todo
│   │   ├── Column[2] - In Progress
│   │   ├── Column[3] - Review
│   │   └── Column[4] - Done
│   └── StatusBar (keybindings hint)
├── TaskDetail (slide-in panel, optional)
├── TaskForm (modal overlay, optional)
└── Help (overlay, optional)
```

### State Management

```go
// Main application state
type App struct {
    // Core state
    pb          *pocketbase.PocketBase
    boards      []*Board           // All available boards
    currentBoard *Board            // Currently active board
    
    // View state
    view        ViewState          // board, detail, form, help, boardSelector
    columns     []Column           // 5 columns for current board
    focusedCol  int                // Currently focused column index
    
    // Overlay state
    taskDetail  *TaskDetail        // nil when hidden
    taskForm    *TaskForm          // nil when hidden
    boardSelector *BoardSelector   // nil when hidden
    help        help.Model
    
    // Filter state
    filters     []Filter
    searchQuery string
    
    // UI state
    width       int
    height      int
    ready       bool
    
    // Real-time
    subscription *Subscription     // SSE subscription when server running
}

type ViewState int

const (
    ViewBoard ViewState = iota
    ViewTaskDetail
    ViewTaskForm
    ViewHelp
    ViewBoardSelector
)
```

---

## Data Layer Integration

### Reusing Existing Packages

The TUI will leverage existing internal packages for all data operations:

```go
import (
    "github.com/your-org/egenskriven/internal/board"
    "github.com/your-org/egenskriven/internal/resolver"
    "github.com/your-org/egenskriven/internal/config"
    "github.com/your-org/egenskriven/internal/commands" // for position utilities
)
```

### Data Access Patterns

#### Loading Boards
```go
func loadBoards(app *pocketbase.PocketBase) tea.Cmd {
    return func() tea.Msg {
        records, err := board.GetAll(app)
        if err != nil {
            return errMsg{err}
        }
        return boardsLoadedMsg{boards: records}
    }
}
```

#### Loading Tasks for Board
```go
func loadTasks(app *pocketbase.PocketBase, boardID string) tea.Cmd {
    return func() tea.Msg {
        records, err := app.FindAllRecords("tasks",
            dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
        )
        if err != nil {
            return errMsg{err}
        }
        return tasksLoadedMsg{tasks: records}
    }
}
```

#### Task Operations (Hybrid Mode)

Reuse the existing hybrid save pattern from `internal/commands/`:

```go
// Create task - uses existing saveRecordHybrid pattern
func createTask(app *pocketbase.PocketBase, input TaskInput) tea.Cmd {
    return func() tea.Msg {
        collection, _ := app.FindCollectionByNameOrId("tasks")
        record := core.NewRecord(collection)
        
        // Set fields
        record.Set("title", input.Title)
        record.Set("description", input.Description)
        record.Set("type", input.Type)
        record.Set("priority", input.Priority)
        record.Set("column", input.Column)
        record.Set("board", input.BoardID)
        record.Set("created_by", "cli") // or "tui"
        
        // Get position and sequence
        position := commands.GetNextPosition(app, input.Column)
        record.Set("position", position)
        
        seq, _ := board.GetAndIncrementSequence(app, input.BoardID)
        record.Set("seq", seq)
        
        // Save (hybrid mode)
        if err := saveRecordHybrid(app, record); err != nil {
            return errMsg{err}
        }
        
        return taskCreatedMsg{task: record}
    }
}

// Move task between columns
func moveTask(app *pocketbase.PocketBase, taskID, targetColumn string, position float64) tea.Cmd {
    return func() tea.Msg {
        record, err := app.FindRecordById("tasks", taskID)
        if err != nil {
            return errMsg{err}
        }
        
        record.Set("column", targetColumn)
        record.Set("position", position)
        
        if err := updateRecordHybrid(app, record); err != nil {
            return errMsg{err}
        }
        
        return taskMovedMsg{task: record}
    }
}
```

---

## Real-Time Synchronization

### Architecture

When the server is running, the TUI should receive real-time updates for:
- Tasks created/updated/deleted (from web UI or other CLI instances)
- Board changes
- Epic changes

### Implementation Strategy

#### Option 1: SSE Client (Recommended)

PocketBase uses Server-Sent Events for real-time. Implement a Go SSE client:

```go
// internal/tui/realtime.go

type Subscription struct {
    cancel context.CancelFunc
    events chan RealtimeEvent
}

type RealtimeEvent struct {
    Action     string // "create", "update", "delete"
    Collection string // "tasks", "boards", "epics"
    Record     map[string]interface{}
}

func subscribeToRealtime(serverURL string) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithCancel(context.Background())
        events := make(chan RealtimeEvent, 100)
        
        go func() {
            // Connect to SSE endpoint
            url := serverURL + "/api/realtime"
            req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
            req.Header.Set("Accept", "text/event-stream")
            
            resp, err := http.DefaultClient.Do(req)
            if err != nil {
                return
            }
            defer resp.Body.Close()
            
            // Subscribe to collections
            // Send subscription message...
            
            // Read events
            scanner := bufio.NewScanner(resp.Body)
            for scanner.Scan() {
                line := scanner.Text()
                if strings.HasPrefix(line, "data:") {
                    // Parse and send event
                    var event RealtimeEvent
                    json.Unmarshal([]byte(line[5:]), &event)
                    events <- event
                }
            }
        }()
        
        return subscriptionStartedMsg{
            subscription: &Subscription{cancel: cancel, events: events},
        }
    }
}

// Command to listen for events
func listenForRealtimeEvents(sub *Subscription) tea.Cmd {
    return func() tea.Msg {
        event := <-sub.events
        return realtimeEventMsg{event: event}
    }
}
```

#### Option 2: Polling Fallback

If SSE is complex, implement polling as fallback:

```go
func pollForChanges(app *pocketbase.PocketBase, lastCheck time.Time) tea.Cmd {
    return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
        // Query for records updated since lastCheck
        records, _ := app.FindAllRecords("tasks",
            dbx.NewExp("updated > {:time}", dbx.Params{"time": lastCheck}),
        )
        return pollResultMsg{tasks: records, checkTime: t}
    })
}
```

### Handling Real-Time Updates

```go
func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case realtimeEventMsg:
        switch msg.event.Collection {
        case "tasks":
            switch msg.event.Action {
            case "create":
                // Add task to appropriate column
                return m, m.addTaskToColumn(msg.event.Record)
            case "update":
                // Update task in place
                return m, m.updateTaskInColumn(msg.event.Record)
            case "delete":
                // Remove task from column
                return m, m.removeTaskFromColumn(msg.event.Record["id"].(string))
            }
        case "boards":
            // Refresh board list
            return m, loadBoards(m.pb)
        }
        // Continue listening
        return m, listenForRealtimeEvents(m.subscription)
    }
    return m, nil
}
```

---

## UI Components

### 1. Board View (Main Kanban)

```go
// internal/tui/board.go

type Board struct {
    name    string
    prefix  string
    columns []Column
    focused int
    width   int
    height  int
}

func (b *Board) View() string {
    // Calculate column width
    colWidth := (b.width - 4) / len(b.columns) // 4 for borders
    
    // Render each column
    var cols []string
    for i, col := range b.columns {
        style := b.columnStyle(i == b.focused)
        cols = append(cols, style.Width(colWidth).Render(col.View()))
    }
    
    // Join horizontally
    return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

func (b *Board) columnStyle(focused bool) lipgloss.Style {
    if focused {
        return lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color("62"))
    }
    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("240"))
}
```

### 2. Column Component

```go
// internal/tui/column.go

type Column struct {
    status  string        // "backlog", "todo", etc.
    title   string        // Display title
    list    list.Model    // bubbles/list
    focused bool
}

func NewColumn(status, title string, items []list.Item, focused bool) Column {
    delegate := list.NewDefaultDelegate()
    delegate.ShowDescription = true
    delegate.SetHeight(3) // Compact cards
    
    l := list.New(items, delegate, 0, 0)
    l.Title = title
    l.SetShowStatusBar(true)
    l.SetFilteringEnabled(true)
    l.SetShowHelp(false) // We have global help
    
    return Column{
        status:  status,
        title:   title,
        list:    l,
        focused: focused,
    }
}

func (c *Column) SetSize(width, height int) {
    c.list.SetSize(width, height-2) // Account for title
}

func (c *Column) View() string {
    titleStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("205"))
    
    count := len(c.list.Items())
    header := titleStyle.Render(fmt.Sprintf("%s (%d)", c.title, count))
    
    return lipgloss.JoinVertical(lipgloss.Left, header, c.list.View())
}
```

### 3. Task Item (list.Item implementation)

```go
// internal/tui/task_item.go

type TaskItem struct {
    ID          string
    DisplayID   string // "WRK-123"
    Title       string
    Description string
    Type        string // bug, feature, chore
    Priority    string // low, medium, high, urgent
    Column      string
    Labels      []string
    DueDate     string
    Epic        string
    EpicTitle   string
    BlockedBy   []string
    IsBlocked   bool
    Position    float64
}

// Implement list.Item interface
func (t TaskItem) FilterValue() string {
    return t.Title + " " + t.Description
}

func (t TaskItem) Title() string {
    return t.renderTitle()
}

func (t TaskItem) Description() string {
    return t.renderDescription()
}

func (t TaskItem) renderTitle() string {
    var parts []string
    
    // Priority indicator
    priorityColors := map[string]string{
        "urgent": "196", // red
        "high":   "208", // orange
        "medium": "226", // yellow
        "low":    "240", // gray
    }
    dot := lipgloss.NewStyle().
        Foreground(lipgloss.Color(priorityColors[t.Priority])).
        Render("●")
    
    // Display ID
    id := lipgloss.NewStyle().
        Foreground(lipgloss.Color("240")).
        Render(t.DisplayID)
    
    // Title (truncated)
    title := truncate(t.Title, 40)
    
    parts = append(parts, dot, id, title)
    
    // Type badge
    typeColors := map[string]string{
        "bug":     "196",
        "feature": "39",
        "chore":   "240",
    }
    if t.Type != "" {
        badge := lipgloss.NewStyle().
            Foreground(lipgloss.Color(typeColors[t.Type])).
            Render("[" + t.Type + "]")
        parts = append(parts, badge)
    }
    
    // Blocked indicator
    if t.IsBlocked {
        blocked := lipgloss.NewStyle().
            Foreground(lipgloss.Color("196")).
            Render("[BLOCKED]")
        parts = append(parts, blocked)
    }
    
    return strings.Join(parts, " ")
}

func (t TaskItem) renderDescription() string {
    var parts []string
    
    // Labels
    if len(t.Labels) > 0 {
        for _, label := range t.Labels[:min(3, len(t.Labels))] {
            parts = append(parts, "#"+label)
        }
    }
    
    // Due date
    if t.DueDate != "" {
        parts = append(parts, "Due: "+t.DueDate)
    }
    
    // Epic
    if t.EpicTitle != "" {
        parts = append(parts, "Epic: "+t.EpicTitle)
    }
    
    return strings.Join(parts, " | ")
}
```

### 4. Task Detail Panel

```go
// internal/tui/task_detail.go

type TaskDetail struct {
    task     TaskItem
    viewport viewport.Model
    width    int
    height   int
}

func (d *TaskDetail) View() string {
    style := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("62")).
        Padding(1, 2)
    
    // Header
    header := lipgloss.NewStyle().Bold(true).Render(d.task.DisplayID + " " + d.task.Title)
    
    // Metadata
    meta := fmt.Sprintf(
        "Type: %s | Priority: %s | Column: %s",
        d.task.Type, d.task.Priority, d.task.Column,
    )
    
    // Description (rendered markdown)
    desc := d.renderDescription()
    
    // Labels
    labels := "Labels: " + strings.Join(d.task.Labels, ", ")
    
    // Blocked by
    var blockedBy string
    if len(d.task.BlockedBy) > 0 {
        blockedBy = "Blocked by: " + strings.Join(d.task.BlockedBy, ", ")
    }
    
    content := lipgloss.JoinVertical(lipgloss.Left,
        header, "", meta, "", desc, "", labels, blockedBy,
    )
    
    return style.Width(d.width).Height(d.height).Render(content)
}

func (d *TaskDetail) renderDescription() string {
    if d.task.Description == "" {
        return lipgloss.NewStyle().Faint(true).Render("No description")
    }
    
    // Use glamour for markdown rendering
    rendered, err := glamour.Render(d.task.Description, "dark")
    if err != nil {
        return d.task.Description
    }
    return rendered
}
```

### 5. Task Form (Add/Edit)

```go
// internal/tui/task_form.go

type TaskForm struct {
    mode        FormMode // Add or Edit
    taskID      string   // For edit mode
    
    titleInput  textinput.Model
    descInput   textarea.Model
    typeSelect  int      // Index into types slice
    prioritySelect int   // Index into priorities slice
    columnSelect int     // Index into columns slice
    labelsInput textinput.Model
    dueDateInput textinput.Model
    epicSelect   int     // Index into epics slice
    
    focusIndex  int      // Which field is focused
    width       int
    height      int
    
    // Options
    types      []string
    priorities []string
    columns    []string
    epics      []EpicOption
}

type FormMode int

const (
    FormModeAdd FormMode = iota
    FormModeEdit
)

type EpicOption struct {
    ID    string
    Title string
}

func NewTaskForm(mode FormMode) *TaskForm {
    ti := textinput.New()
    ti.Placeholder = "Task title..."
    ti.Focus()
    
    ta := textarea.New()
    ta.Placeholder = "Description (markdown supported)..."
    ta.SetHeight(5)
    
    li := textinput.New()
    li.Placeholder = "Labels (comma-separated)..."
    
    di := textinput.New()
    di.Placeholder = "YYYY-MM-DD"
    
    return &TaskForm{
        mode:         mode,
        titleInput:   ti,
        descInput:    ta,
        labelsInput:  li,
        dueDateInput: di,
        types:        []string{"feature", "bug", "chore"},
        priorities:   []string{"low", "medium", "high", "urgent"},
        columns:      []string{"backlog", "todo", "in_progress", "review", "done"},
        focusIndex:   0,
    }
}

func (f *TaskForm) View() string {
    style := lipgloss.NewStyle().
        Border(lipgloss.DoubleBorder()).
        BorderForeground(lipgloss.Color("62")).
        Padding(1, 2)
    
    title := "Add Task"
    if f.mode == FormModeEdit {
        title = "Edit Task"
    }
    header := lipgloss.NewStyle().Bold(true).Render(title)
    
    // Form fields
    fields := []string{
        f.renderField("Title", f.titleInput.View(), 0),
        f.renderField("Description", f.descInput.View(), 1),
        f.renderSelect("Type", f.types, f.typeSelect, 2),
        f.renderSelect("Priority", f.priorities, f.prioritySelect, 3),
        f.renderSelect("Column", f.columns, f.columnSelect, 4),
        f.renderField("Labels", f.labelsInput.View(), 5),
        f.renderField("Due Date", f.dueDateInput.View(), 6),
        f.renderEpicSelect(7),
    }
    
    // Action buttons
    buttons := f.renderButtons()
    
    content := lipgloss.JoinVertical(lipgloss.Left,
        append([]string{header, ""}, append(fields, "", buttons)...)...,
    )
    
    return style.Render(content)
}

func (f *TaskForm) renderSelect(label string, options []string, selected, index int) string {
    focused := f.focusIndex == index
    
    var rendered []string
    for i, opt := range options {
        style := lipgloss.NewStyle()
        if i == selected {
            style = style.Bold(true).Foreground(lipgloss.Color("62"))
            opt = "[" + opt + "]"
        }
        rendered = append(rendered, style.Render(opt))
    }
    
    labelStyle := lipgloss.NewStyle()
    if focused {
        labelStyle = labelStyle.Foreground(lipgloss.Color("62"))
    }
    
    return labelStyle.Render(label+": ") + strings.Join(rendered, " ")
}
```

### 6. Board Selector

```go
// internal/tui/board_selector.go

type BoardSelector struct {
    boards  []BoardOption
    list    list.Model
    width   int
    height  int
}

type BoardOption struct {
    ID     string
    Name   string
    Prefix string
    Color  string
}

func (b BoardOption) FilterValue() string { return b.Name }
func (b BoardOption) Title() string       { return b.Prefix + " - " + b.Name }
func (b BoardOption) Description() string { return "" }

func NewBoardSelector(boards []BoardOption) *BoardSelector {
    items := make([]list.Item, len(boards))
    for i, b := range boards {
        items[i] = b
    }
    
    l := list.New(items, list.NewDefaultDelegate(), 0, 0)
    l.Title = "Select Board"
    l.SetShowStatusBar(false)
    l.SetFilteringEnabled(true)
    
    return &BoardSelector{
        boards: boards,
        list:   l,
    }
}

func (s *BoardSelector) View() string {
    style := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("205")).
        Padding(1, 2)
    
    return style.Width(40).Render(s.list.View())
}
```

### 7. Filter Bar

```go
// internal/tui/filter_bar.go

type FilterBar struct {
    filters     []Filter
    searchInput textinput.Model
    isSearching bool
}

type Filter struct {
    Field    string // column, priority, type, label, epic
    Operator string // is, is_not, includes
    Value    string
}

func (f *FilterBar) View() string {
    if len(f.filters) == 0 && !f.isSearching {
        return ""
    }
    
    style := lipgloss.NewStyle().
        Background(lipgloss.Color("236")).
        Padding(0, 1)
    
    var parts []string
    
    if f.isSearching {
        parts = append(parts, "Search: "+f.searchInput.View())
    }
    
    for _, filter := range f.filters {
        chip := lipgloss.NewStyle().
            Background(lipgloss.Color("62")).
            Foreground(lipgloss.Color("0")).
            Padding(0, 1).
            Render(filter.Field + ":" + filter.Value)
        parts = append(parts, chip)
    }
    
    return style.Render(strings.Join(parts, " "))
}
```

---

## Keybindings

### Global Keys

| Key | Action |
|-----|--------|
| `q`, `Ctrl+C` | Quit application |
| `?` | Toggle help overlay |
| `b` | Open board selector |
| `/` | Focus search/filter |
| `Esc` | Close overlay / cancel |
| `Ctrl+R` | Force refresh |

### Board Navigation

| Key | Action |
|-----|--------|
| `h`, `Left` | Focus previous column |
| `l`, `Right` | Focus next column |
| `j`, `Down` | Select next task in column |
| `k`, `Up` | Select previous task in column |
| `g` | Go to first task in column |
| `G` | Go to last task in column |
| `Tab` | Cycle through columns |

### Task Actions

| Key | Action |
|-----|--------|
| `Enter` | Open task detail |
| `n` | New task (in current column) |
| `e` | Edit selected task |
| `d` | Delete selected task (with confirm) |
| `Space` | Toggle task selection (multi-select) |

### Task Movement

| Key | Action |
|-----|--------|
| `m` + `h/l` | Move task to adjacent column |
| `Shift+H` | Move task to previous column |
| `Shift+L` | Move task to next column |
| `Shift+K` | Move task up in column |
| `Shift+J` | Move task down in column |
| `1-5` | Move task to column by number |

### Quick Filters

| Key | Action |
|-----|--------|
| `fp` | Filter by priority |
| `ft` | Filter by type |
| `fl` | Filter by label |
| `fe` | Filter by epic |
| `fc` | Clear all filters |

### Implementation

```go
// internal/tui/keys.go

type keyMap struct {
    // Navigation
    Up        key.Binding
    Down      key.Binding
    Left      key.Binding
    Right     key.Binding
    FirstItem key.Binding
    LastItem  key.Binding
    
    // Actions
    Enter     key.Binding
    New       key.Binding
    Edit      key.Binding
    Delete    key.Binding
    Select    key.Binding
    
    // Movement
    MoveLeft  key.Binding
    MoveRight key.Binding
    MoveUp    key.Binding
    MoveDown  key.Binding
    MoveTo1   key.Binding
    MoveTo2   key.Binding
    MoveTo3   key.Binding
    MoveTo4   key.Binding
    MoveTo5   key.Binding
    
    // Global
    Quit      key.Binding
    Help      key.Binding
    Board     key.Binding
    Search    key.Binding
    Refresh   key.Binding
    Escape    key.Binding
}

func defaultKeyMap() keyMap {
    return keyMap{
        Up: key.NewBinding(
            key.WithKeys("up", "k"),
            key.WithHelp("k/up", "up"),
        ),
        Down: key.NewBinding(
            key.WithKeys("down", "j"),
            key.WithHelp("j/down", "down"),
        ),
        Left: key.NewBinding(
            key.WithKeys("left", "h"),
            key.WithHelp("h/left", "prev column"),
        ),
        Right: key.NewBinding(
            key.WithKeys("right", "l"),
            key.WithHelp("l/right", "next column"),
        ),
        MoveLeft: key.NewBinding(
            key.WithKeys("H"),
            key.WithHelp("H", "move task left"),
        ),
        MoveRight: key.NewBinding(
            key.WithKeys("L"),
            key.WithHelp("L", "move task right"),
        ),
        New: key.NewBinding(
            key.WithKeys("n"),
            key.WithHelp("n", "new task"),
        ),
        Edit: key.NewBinding(
            key.WithKeys("e"),
            key.WithHelp("e", "edit task"),
        ),
        Delete: key.NewBinding(
            key.WithKeys("d"),
            key.WithHelp("d", "delete task"),
        ),
        Enter: key.NewBinding(
            key.WithKeys("enter"),
            key.WithHelp("enter", "view details"),
        ),
        Quit: key.NewBinding(
            key.WithKeys("q", "ctrl+c"),
            key.WithHelp("q", "quit"),
        ),
        Help: key.NewBinding(
            key.WithKeys("?"),
            key.WithHelp("?", "help"),
        ),
        Board: key.NewBinding(
            key.WithKeys("b"),
            key.WithHelp("b", "switch board"),
        ),
        Search: key.NewBinding(
            key.WithKeys("/"),
            key.WithHelp("/", "search"),
        ),
        Escape: key.NewBinding(
            key.WithKeys("esc"),
            key.WithHelp("esc", "cancel/close"),
        ),
    }
}
```

---

## Styling

### Color Palette

```go
// internal/tui/styles.go

var (
    // Brand colors
    primaryColor   = lipgloss.Color("62")   // Blue
    secondaryColor = lipgloss.Color("205")  // Pink
    
    // Status colors
    successColor = lipgloss.Color("82")   // Green
    warningColor = lipgloss.Color("214")  // Orange
    errorColor   = lipgloss.Color("196")  // Red
    
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
    columnReview     = lipgloss.Color("205")
    columnDone       = lipgloss.Color("82")
    
    // UI colors
    borderColor       = lipgloss.Color("240")
    focusBorderColor  = lipgloss.Color("62")
    mutedColor        = lipgloss.Color("240")
    backgroundColor   = lipgloss.Color("0")
)

// Styles
var (
    focusedColumnStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(focusBorderColor).
        Padding(0, 1)
    
    blurredColumnStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(borderColor).
        Padding(0, 1)
    
    headerStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(primaryColor)
    
    taskCardStyle = lipgloss.NewStyle().
        Border(lipgloss.NormalBorder()).
        BorderForeground(borderColor).
        Padding(0, 1)
    
    selectedTaskStyle = lipgloss.NewStyle().
        Border(lipgloss.NormalBorder()).
        BorderForeground(primaryColor).
        Background(lipgloss.Color("236")).
        Padding(0, 1)
    
    statusBarStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("236")).
        Foreground(lipgloss.Color("252")).
        Padding(0, 1)
    
    modalStyle = lipgloss.NewStyle().
        Border(lipgloss.DoubleBorder()).
        BorderForeground(primaryColor).
        Padding(1, 2)
)
```

### Column Header Styling

```go
func columnHeaderStyle(status string, focused bool) lipgloss.Style {
    colors := map[string]lipgloss.Color{
        "backlog":     columnBacklog,
        "todo":        columnTodo,
        "in_progress": columnInProgress,
        "review":      columnReview,
        "done":        columnDone,
    }
    
    style := lipgloss.NewStyle().
        Bold(true).
        Foreground(colors[status])
    
    if focused {
        style = style.Underline(true)
    }
    
    return style
}
```

---

## Implementation Phases

### Phase 1: Foundation (3-4 days)

**Goal:** Basic kanban board with navigation

**Tasks:**
1. Create `internal/tui/` directory structure
2. Add Bubble Tea dependencies to `go.mod`
3. Implement `tui.go` CLI command
4. Create basic `App` model with Init/Update/View
5. Implement `Column` component with `bubbles/list`
6. Implement `TaskItem` (list.Item interface)
7. Load tasks from PocketBase and display in columns
8. Implement column navigation (h/l, left/right)
9. Implement task navigation within column (j/k, up/down)
10. Add basic styling with Lipgloss

**Deliverable:** Can launch TUI, see tasks in columns, navigate with keyboard

### Phase 2: Task Operations (3-4 days)

**Goal:** Full CRUD for tasks

**Tasks:**
1. Implement `TaskDetail` view (Enter to open)
2. Implement `TaskForm` for adding tasks (n key)
3. Implement task editing (e key)
4. Implement task deletion with confirmation (d key)
5. Implement task movement between columns (H/L, 1-5)
6. Implement task reordering within column (Shift+J/K)
7. Use hybrid save pattern for all operations
8. Add success/error feedback messages

**Deliverable:** Can create, view, edit, delete, and move tasks

### Phase 3: Multi-Board Support (2-3 days)

**Goal:** Board switching and management

**Tasks:**
1. Implement `BoardSelector` component
2. Load all boards on startup
3. Implement board switching (b key)
4. Persist last-used board in config
5. Display board name/prefix in header
6. Support board-specific columns (if configured)
7. Handle default board from `.egenskriven/config.json`

**Deliverable:** Can switch between boards, remembers last board

### Phase 4: Real-Time Sync (2-3 days)

**Goal:** Live updates from server

**Tasks:**
1. Implement SSE client for PocketBase realtime
2. Subscribe to tasks collection on startup
3. Handle create/update/delete events
4. Update UI without full refresh
5. Implement fallback polling if SSE fails
6. Show connection status indicator
7. Handle reconnection on disconnect

**Deliverable:** Changes from web UI appear in TUI in real-time

### Phase 5: Filtering & Search (2-3 days)

**Goal:** Filter and search tasks

**Tasks:**
1. Implement `FilterBar` component
2. Add search input (/ key)
3. Filter by column (already implicit)
4. Filter by priority (fp)
5. Filter by type (ft)
6. Filter by label (fl)
7. Filter by epic (fe)
8. Clear filters (fc)
9. Show active filters in UI
10. Persist filter state during session

**Deliverable:** Can filter/search tasks like web UI

### Phase 6: Advanced Features (3-4 days)

**Goal:** Feature parity with web UI

**Tasks:**
1. Implement epic display and filtering
2. Show blocked tasks indicator
3. Show due date with overdue highlighting
4. Implement subtask display (expandable)
5. Multi-select tasks (Space)
6. Bulk operations (move, delete)
7. Add help overlay (?)
8. Keyboard shortcut cheat sheet
9. Command palette (Ctrl+K) - optional

**Deliverable:** Full feature parity with web UI

### Phase 7: Polish & Testing (2-3 days)

**Goal:** Production-ready TUI

**Tasks:**
1. Handle terminal resize gracefully
2. Add loading indicators
3. Error handling and display
4. Edge cases (empty boards, no tasks)
5. Performance optimization for large boards
6. Write unit tests for components
7. Integration tests with PocketBase
8. Documentation and usage guide
9. Add to main CLI help

**Deliverable:** Polished, tested TUI ready for use

---

## Testing Strategy

### Unit Tests

```go
// internal/tui/column_test.go

func TestColumnNavigation(t *testing.T) {
    items := []list.Item{
        TaskItem{ID: "1", Title: "Task 1"},
        TaskItem{ID: "2", Title: "Task 2"},
        TaskItem{ID: "3", Title: "Task 3"},
    }
    
    col := NewColumn("todo", "Todo", items, true)
    
    // Test down navigation
    col, _ = col.Update(tea.KeyMsg{Type: tea.KeyDown})
    assert.Equal(t, 1, col.list.Index())
    
    // Test up navigation
    col, _ = col.Update(tea.KeyMsg{Type: tea.KeyUp})
    assert.Equal(t, 0, col.list.Index())
}

func TestTaskItemRendering(t *testing.T) {
    task := TaskItem{
        DisplayID: "WRK-123",
        Title:     "Test task",
        Priority:  "high",
        Type:      "bug",
        IsBlocked: true,
    }
    
    title := task.Title()
    assert.Contains(t, title, "WRK-123")
    assert.Contains(t, title, "Test task")
    assert.Contains(t, title, "[bug]")
    assert.Contains(t, title, "[BLOCKED]")
}
```

### Integration Tests

```go
// internal/tui/integration_test.go

func TestBoardWithRealData(t *testing.T) {
    // Setup test PocketBase instance
    app := setupTestApp(t)
    
    // Create test board and tasks
    board := createTestBoard(t, app, "Test", "TST")
    createTestTask(t, app, board.Id, "Task 1", "todo")
    createTestTask(t, app, board.Id, "Task 2", "in_progress")
    
    // Create TUI model
    model := NewApp(app)
    model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
    
    // Verify tasks loaded
    assert.Equal(t, 1, len(model.columns[1].list.Items())) // todo
    assert.Equal(t, 1, len(model.columns[2].list.Items())) // in_progress
}
```

---

## Risk Mitigation

### Risk 1: SSE Complexity

**Risk:** PocketBase SSE protocol may be complex to implement in Go

**Mitigation:**
- Start with polling fallback
- Research existing Go SSE clients
- Consider using PocketBase Go SDK if it supports subscriptions
- Time-box SSE implementation, fall back to polling if needed

### Risk 2: Terminal Compatibility

**Risk:** Different terminals may render differently

**Mitigation:**
- Test on common terminals (iTerm2, Terminal.app, Windows Terminal, Alacritty)
- Use Bubble Tea's built-in terminal detection
- Provide fallback styles for limited terminals
- Document supported terminals

### Risk 3: Large Board Performance

**Risk:** Boards with many tasks may be slow

**Mitigation:**
- `bubbles/list` has built-in pagination
- Implement virtualization for very large columns
- Limit initial load to active columns
- Lazy load done/backlog columns

### Risk 4: Conflict with Web UI

**Risk:** Concurrent edits may cause conflicts

**Mitigation:**
- Real-time sync prevents most conflicts
- Optimistic updates with rollback
- Show "modified by another user" warnings
- Last-write-wins for simple conflicts

---

## Dependencies to Add

```go
// go.mod additions

require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/bubbles v0.18.0
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/charmbracelet/glamour v0.6.0
)
```

---

## CLI Integration

### Command Registration

```go
// internal/commands/tui.go

func newTuiCmd(app *pocketbase.PocketBase) *cobra.Command {
    var boardRef string
    
    cmd := &cobra.Command{
        Use:     "tui",
        Aliases: []string{"ui", "board"},
        Short:   "Open interactive kanban board",
        Long:    "Launch the terminal user interface for managing tasks in a kanban board view.",
        RunE: func(cmd *cobra.Command, args []string) error {
            if err := app.Bootstrap(); err != nil {
                return err
            }
            
            // Load config for default board
            cfg, _ := config.LoadProjectConfig()
            if boardRef == "" && cfg != nil {
                boardRef = cfg.DefaultBoard
            }
            
            // Create and run TUI
            tuiApp := tui.NewApp(app, boardRef)
            p := tea.NewProgram(tuiApp, tea.WithAltScreen())
            
            _, err := p.Run()
            return err
        },
    }
    
    cmd.Flags().StringVarP(&boardRef, "board", "b", "", "Board to open (name or prefix)")
    
    return cmd
}
```

### Registration in root.go

```go
// internal/commands/root.go

func Register(app *pocketbase.PocketBase) {
    // ... existing commands ...
    app.RootCmd.AddCommand(newTuiCmd(app))
}
```

---

## Success Criteria

### MVP (Phases 1-4)
- [ ] Can launch TUI with `egenskriven tui`
- [ ] See all 5 columns with tasks
- [ ] Navigate between columns and tasks
- [ ] Create new tasks
- [ ] Edit existing tasks
- [ ] Delete tasks
- [ ] Move tasks between columns
- [ ] Switch between boards
- [ ] Real-time updates from web UI

### Full Release (All Phases)
- [ ] All MVP features
- [ ] Search and filter tasks
- [ ] Filter by priority, type, label, epic
- [ ] See blocked tasks indicator
- [ ] See due dates with highlighting
- [ ] Multi-select and bulk operations
- [ ] Help overlay with all keybindings
- [ ] Smooth resize handling
- [ ] No crashes or data loss
- [ ] Tests passing

---

## Timeline Summary

| Phase | Duration | Cumulative |
|-------|----------|------------|
| Phase 1: Foundation | 3-4 days | 3-4 days |
| Phase 2: Task Operations | 3-4 days | 6-8 days |
| Phase 3: Multi-Board | 2-3 days | 8-11 days |
| Phase 4: Real-Time | 2-3 days | 10-14 days |
| Phase 5: Filtering | 2-3 days | 12-17 days |
| Phase 6: Advanced | 3-4 days | 15-21 days |
| Phase 7: Polish | 2-3 days | 17-24 days |

**Total Estimated Time: 3-4 weeks**

---

## References

- [Bubble Tea GitHub](https://github.com/charmbracelet/bubbletea)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Lip Gloss Styling](https://github.com/charmbracelet/lipgloss)
- [kancli Official Example](https://github.com/charm-and-friends/kancli)
- [gh-dash Architecture](https://github.com/dlvhdr/gh-dash)
- [PocketBase Realtime](https://pocketbase.io/docs/api-realtime/)
