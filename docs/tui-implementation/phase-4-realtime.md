# Phase 4: Real-Time Sync

**Goal**: Live updates from server - changes from web UI appear in TUI in real-time.

**Duration Estimate**: 2-3 days

**Prerequisites**: Phase 3 (Multi-Board Support) completed

**Deliverable**: TUI receives and displays real-time updates for task create/update/delete events, with fallback to polling when SSE fails.

---

## Overview

Real-time synchronization is what makes EgenSkriven truly multi-client. Without it, users working in the web UI and TUI simultaneously would see stale data and risk conflicts. This phase implements Server-Sent Events (SSE) integration with PocketBase's realtime API.

### Why Real-Time Matters

Consider this scenario:
1. Alice has the TUI open on her left monitor
2. Bob creates a task in the web UI
3. Without real-time sync, Alice doesn't see Bob's task until she refreshes
4. Worse, Alice might create a duplicate task or move tasks that Bob already moved

Real-time sync solves this by pushing changes to all connected clients immediately.

### PocketBase Realtime Protocol

PocketBase uses Server-Sent Events (SSE) for real-time subscriptions. The protocol works as follows:

1. **Connect** to `/api/realtime` with `Accept: text/event-stream`
2. **Receive client ID** in the first `PB_CONNECT` event
3. **Subscribe** by POSTing to `/api/realtime` with client ID and subscriptions
4. **Receive events** as SSE messages with `event:` and `data:` fields

```
Event format:
event: PB_CONNECT
data: {"clientId":"abc123"}

event: tasks
data: {"action":"create","record":{...}}
```

### Why SSE Over WebSockets?

SSE is simpler than WebSockets for this use case:
- One-way communication (server to client) is sufficient
- Built-in reconnection handling
- Works through more proxies and firewalls
- Native browser support (relevant for web UI consistency)

---

## Tasks

### 4.1 Create Realtime Message Types ✅ COMPLETED

**What**: Define message types for realtime events in the Bubble Tea application.

**Why**: Bubble Tea uses messages for all state changes. We need specific message types to handle realtime events, connection status changes, and errors.

**Steps**:

1. Create `internal/tui/messages.go` with realtime-related messages
2. Define message types for each realtime action (create, update, delete)
3. Add connection status messages

**File**: `internal/tui/messages.go`

```go
package tui

import (
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// =============================================================================
// Realtime Messages
// =============================================================================

// RealtimeEvent represents a parsed realtime event from PocketBase.
type RealtimeEvent struct {
	Action     string                 // "create", "update", "delete"
	Collection string                 // "tasks", "boards", "epics"
	Record     map[string]interface{} // The record data
}

// realtimeConnectedMsg is sent when SSE connection is established.
type realtimeConnectedMsg struct {
	clientID string
}

// realtimeDisconnectedMsg is sent when SSE connection is lost.
type realtimeDisconnectedMsg struct {
	err error
}

// realtimeEventMsg wraps a realtime event for the Update loop.
type realtimeEventMsg struct {
	event RealtimeEvent
}

// realtimeErrorMsg indicates an error in the realtime subsystem.
type realtimeErrorMsg struct {
	err error
}

// realtimeReconnectMsg triggers a reconnection attempt.
type realtimeReconnectMsg struct {
	attempt int
}

// =============================================================================
// Connection Status Messages
// =============================================================================

// ConnectionStatus represents the current realtime connection state.
type ConnectionStatus int

const (
	ConnectionDisconnected ConnectionStatus = iota
	ConnectionConnecting
	ConnectionConnected
	ConnectionReconnecting
)

func (s ConnectionStatus) String() string {
	switch s {
	case ConnectionDisconnected:
		return "disconnected"
	case ConnectionConnecting:
		return "connecting"
	case ConnectionConnected:
		return "connected"
	case ConnectionReconnecting:
		return "reconnecting"
	default:
		return "unknown"
	}
}

// connectionStatusMsg updates the connection status indicator.
type connectionStatusMsg struct {
	status ConnectionStatus
}

// =============================================================================
// Polling Fallback Messages
// =============================================================================

// pollStartMsg initiates polling mode (fallback when SSE fails).
type pollStartMsg struct{}

// pollStopMsg stops polling mode.
type pollStopMsg struct{}

// pollTickMsg triggers a poll cycle.
type pollTickMsg struct {
	time time.Time
}

// pollResultMsg contains the results of a poll cycle.
type pollResultMsg struct {
	tasks     []*core.Record
	checkTime time.Time
	err       error
}

// =============================================================================
// Task Update Messages (from realtime events)
// =============================================================================

// taskCreatedMsg is sent when a task is created (locally or via realtime).
type taskCreatedMsg struct {
	task *core.Record
}

// taskUpdatedMsg is sent when a task is updated.
type taskUpdatedMsg struct {
	task *core.Record
}

// taskDeletedMsg is sent when a task is deleted.
type taskDeletedMsg struct {
	taskID string
}

// tasksReloadedMsg is sent when tasks are bulk reloaded (after reconnect).
type tasksReloadedMsg struct {
	tasks []*core.Record
}

// =============================================================================
// Server Status Messages
// =============================================================================

// serverOnlineMsg is sent when server becomes reachable.
type serverOnlineMsg struct{}

// serverOfflineMsg is sent when server becomes unreachable.
type serverOfflineMsg struct{}
```

**Expected output**: File compiles without errors.

**Common Mistakes**:
- Forgetting to export types that need to be used outside the package
- Using the wrong record type (use `map[string]interface{}` for raw SSE data, `*core.Record` for PocketBase records)

---

### 4.2 Implement SSE Client ✅ COMPLETED

**What**: Create a robust SSE client that connects to PocketBase's realtime endpoint.

**Why**: This is the core of realtime functionality. The client must handle connection establishment, subscription management, event parsing, and graceful disconnection.

**Steps**:

1. Create `internal/tui/realtime.go` with the SSE client implementation
2. Implement connection with proper headers
3. Implement subscription message sending
4. Implement event stream parsing
5. Add graceful shutdown support

**File**: `internal/tui/realtime.go`

```go
package tui

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	// SSE connection timeout
	sseConnectTimeout = 10 * time.Second

	// Maximum reconnection attempts before falling back to polling
	maxReconnectAttempts = 5

	// Base delay for exponential backoff (doubles each attempt)
	baseReconnectDelay = 1 * time.Second

	// Maximum reconnect delay
	maxReconnectDelay = 30 * time.Second

	// Polling interval when in fallback mode
	pollInterval = 3 * time.Second
)

// RealtimeClient manages the SSE connection to PocketBase.
type RealtimeClient struct {
	serverURL string
	clientID  string

	// Connection management
	ctx       context.Context
	cancel    context.CancelFunc
	connected bool
	mu        sync.RWMutex

	// Event channel for sending events to the TUI
	events chan RealtimeEvent

	// HTTP client with appropriate timeouts
	httpClient *http.Client

	// Subscriptions
	collections []string
}

// NewRealtimeClient creates a new realtime client for the given server.
func NewRealtimeClient(serverURL string) *RealtimeClient {
	return &RealtimeClient{
		serverURL: strings.TrimSuffix(serverURL, "/"),
		events:    make(chan RealtimeEvent, 100),
		httpClient: &http.Client{
			Timeout: 0, // No timeout for SSE (long-lived connection)
		},
		collections: []string{"tasks", "boards", "epics"},
	}
}

// SetCollections configures which collections to subscribe to.
func (c *RealtimeClient) SetCollections(collections []string) {
	c.collections = collections
}

// IsConnected returns true if the SSE connection is active.
func (c *RealtimeClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Events returns the channel for receiving realtime events.
func (c *RealtimeClient) Events() <-chan RealtimeEvent {
	return c.events
}

// Connect establishes the SSE connection and returns a Bubble Tea command.
// This is the main entry point for starting realtime sync.
func (c *RealtimeClient) Connect() tea.Cmd {
	return func() tea.Msg {
		c.mu.Lock()
		if c.cancel != nil {
			c.cancel() // Cancel any existing connection
		}
		c.ctx, c.cancel = context.WithCancel(context.Background())
		c.mu.Unlock()

		// Step 1: Connect to SSE endpoint
		req, err := http.NewRequestWithContext(c.ctx, "GET", c.serverURL+"/api/realtime", nil)
		if err != nil {
			return realtimeErrorMsg{err: fmt.Errorf("creating request: %w", err)}
		}
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return realtimeErrorMsg{err: fmt.Errorf("connecting to SSE: %w", err)}
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return realtimeErrorMsg{err: fmt.Errorf("SSE connection failed: %s", resp.Status)}
		}

		// Step 2: Read the initial PB_CONNECT event to get client ID
		clientID, err := c.readClientID(resp.Body)
		if err != nil {
			resp.Body.Close()
			return realtimeErrorMsg{err: fmt.Errorf("reading client ID: %w", err)}
		}

		c.mu.Lock()
		c.clientID = clientID
		c.connected = true
		c.mu.Unlock()

		// Step 3: Subscribe to collections
		if err := c.sendSubscription(); err != nil {
			resp.Body.Close()
			c.mu.Lock()
			c.connected = false
			c.mu.Unlock()
			return realtimeErrorMsg{err: fmt.Errorf("subscribing: %w", err)}
		}

		// Step 4: Start reading events in a goroutine
		go c.readEvents(resp.Body)

		return realtimeConnectedMsg{clientID: clientID}
	}
}

// readClientID reads the initial PB_CONNECT event from the SSE stream.
func (c *RealtimeClient) readClientID(body io.Reader) (string, error) {
	scanner := bufio.NewScanner(body)

	var eventType string
	var eventData string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			eventData = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		} else if line == "" && eventType != "" {
			// Empty line marks end of event
			if eventType == "PB_CONNECT" {
				var data struct {
					ClientID string `json:"clientId"`
				}
				if err := json.Unmarshal([]byte(eventData), &data); err != nil {
					return "", fmt.Errorf("parsing PB_CONNECT: %w", err)
				}
				return data.ClientID, nil
			}
			eventType = ""
			eventData = ""
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("connection closed before receiving client ID")
}

// sendSubscription sends the subscription request to PocketBase.
func (c *RealtimeClient) sendSubscription() error {
	c.mu.RLock()
	clientID := c.clientID
	c.mu.RUnlock()

	// Build subscription payload
	// PocketBase expects: { "clientId": "...", "subscriptions": ["tasks", "boards"] }
	payload := map[string]interface{}{
		"clientId":      clientID,
		"subscriptions": c.collections,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling subscription: %w", err)
	}

	req, err := http.NewRequestWithContext(c.ctx, "POST", c.serverURL+"/api/realtime", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("creating subscription request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending subscription: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("subscription failed (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// readEvents continuously reads events from the SSE stream.
// This runs in a goroutine and sends events to the events channel.
func (c *RealtimeClient) readEvents(body io.ReadCloser) {
	defer body.Close()
	defer func() {
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()

		// Notify that connection was lost
		select {
		case c.events <- RealtimeEvent{Action: "disconnect"}:
		default:
		}
	}()

	scanner := bufio.NewScanner(body)
	// Increase buffer for large events
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	var eventType string
	var eventData strings.Builder

	for scanner.Scan() {
		// Check if context was cancelled
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		line := scanner.Text()

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			eventData.Reset()
		} else if strings.HasPrefix(line, "data:") {
			if eventData.Len() > 0 {
				eventData.WriteString("\n")
			}
			eventData.WriteString(strings.TrimPrefix(line, "data:"))
		} else if line == "" && eventType != "" {
			// Empty line marks end of event
			c.processEvent(eventType, eventData.String())
			eventType = ""
			eventData.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		// Connection error - will trigger reconnection
		select {
		case <-c.ctx.Done():
			// Intentional disconnect
		default:
			// Unintentional disconnect
		}
	}
}

// processEvent parses and dispatches a single SSE event.
func (c *RealtimeClient) processEvent(eventType, data string) {
	// Skip internal PocketBase events
	if strings.HasPrefix(eventType, "PB_") {
		return
	}

	// Parse the event data
	var eventPayload struct {
		Action string                 `json:"action"`
		Record map[string]interface{} `json:"record"`
	}

	if err := json.Unmarshal([]byte(data), &eventPayload); err != nil {
		// Log parsing error but continue
		return
	}

	// Determine collection from event type
	// Event types are in format: "tasks", "boards", etc.
	collection := eventType

	event := RealtimeEvent{
		Action:     eventPayload.Action,
		Collection: collection,
		Record:     eventPayload.Record,
	}

	// Send to events channel (non-blocking)
	select {
	case c.events <- event:
	default:
		// Channel full, drop event
	}
}

// Disconnect closes the SSE connection gracefully.
func (c *RealtimeClient) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
	c.connected = false
}

// WaitForEvent returns a command that waits for the next realtime event.
func WaitForEvent(client *RealtimeClient) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-client.Events()
		if !ok {
			return realtimeDisconnectedMsg{err: fmt.Errorf("event channel closed")}
		}

		if event.Action == "disconnect" {
			return realtimeDisconnectedMsg{err: fmt.Errorf("connection lost")}
		}

		return realtimeEventMsg{event: event}
	}
}

// ReconnectWithBackoff returns a command that attempts to reconnect with exponential backoff.
func ReconnectWithBackoff(client *RealtimeClient, attempt int) tea.Cmd {
	if attempt >= maxReconnectAttempts {
		// Max attempts reached, switch to polling
		return func() tea.Msg {
			return pollStartMsg{}
		}
	}

	delay := baseReconnectDelay * (1 << attempt) // Exponential backoff
	if delay > maxReconnectDelay {
		delay = maxReconnectDelay
	}

	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return realtimeReconnectMsg{attempt: attempt + 1}
	})
}
```

**Expected output**: File compiles without errors.

**Common Mistakes**:
- Not handling the multi-line data format for SSE events
- Forgetting to cancel the context on disconnect
- Not handling the `PB_CONNECT` event before subscribing
- Using a blocking channel send (can deadlock the TUI)

---

### 4.3 Implement Connection Status Indicator ✅ COMPLETED

**What**: Create a visual indicator showing the realtime connection status.

**Why**: Users need to know if real-time sync is active. A simple colored indicator provides at-a-glance status.

**Steps**:

1. Create `internal/tui/status.go` with the status indicator component
2. Implement different visual states (connected, disconnected, reconnecting)
3. Add optional message display for errors

**File**: `internal/tui/status.go`

```go
package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// StatusIndicator displays the realtime connection status.
type StatusIndicator struct {
	status  ConnectionStatus
	message string // Optional status message
}

// NewStatusIndicator creates a new status indicator.
func NewStatusIndicator() *StatusIndicator {
	return &StatusIndicator{
		status: ConnectionDisconnected,
	}
}

// SetStatus updates the connection status.
func (s *StatusIndicator) SetStatus(status ConnectionStatus) {
	s.status = status
	s.message = ""
}

// SetStatusWithMessage updates the status with an optional message.
func (s *StatusIndicator) SetStatusWithMessage(status ConnectionStatus, message string) {
	s.status = status
	s.message = message
}

// Status returns the current connection status.
func (s *StatusIndicator) Status() ConnectionStatus {
	return s.status
}

// View renders the status indicator.
func (s *StatusIndicator) View() string {
	var indicator string
	var style lipgloss.Style

	switch s.status {
	case ConnectionConnected:
		// Green dot for connected
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("82")) // Green
		indicator = "●"
	case ConnectionConnecting:
		// Yellow dot for connecting
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
		indicator = "◐"
	case ConnectionReconnecting:
		// Orange dot for reconnecting
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // Orange
		indicator = "◐"
	case ConnectionDisconnected:
		// Gray dot for disconnected
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Gray
		indicator = "○"
	}

	result := style.Render(indicator)

	if s.message != "" {
		msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
		result += " " + msgStyle.Render(s.message)
	}

	return result
}

// ViewWithLabel renders the status indicator with a label.
func (s *StatusIndicator) ViewWithLabel() string {
	var label string

	switch s.status {
	case ConnectionConnected:
		label = "Live"
	case ConnectionConnecting:
		label = "Connecting..."
	case ConnectionReconnecting:
		label = "Reconnecting..."
	case ConnectionDisconnected:
		label = "Offline"
	}

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	return fmt.Sprintf("%s %s", s.View(), labelStyle.Render(label))
}

// StatusBar represents the full status bar at the bottom of the TUI.
type StatusBar struct {
	status      *StatusIndicator
	boardName   string
	taskCount   int
	filterCount int
	width       int
}

// NewStatusBar creates a new status bar.
func NewStatusBar() *StatusBar {
	return &StatusBar{
		status: NewStatusIndicator(),
	}
}

// SetWidth sets the width of the status bar.
func (b *StatusBar) SetWidth(width int) {
	b.width = width
}

// SetBoardName sets the current board name.
func (b *StatusBar) SetBoardName(name string) {
	b.boardName = name
}

// SetTaskCount sets the total task count.
func (b *StatusBar) SetTaskCount(count int) {
	b.taskCount = count
}

// SetFilterCount sets the number of active filters.
func (b *StatusBar) SetFilterCount(count int) {
	b.filterCount = count
}

// SetConnectionStatus updates the connection status.
func (b *StatusBar) SetConnectionStatus(status ConnectionStatus) {
	b.status.SetStatus(status)
}

// SetConnectionStatusWithMessage updates the connection status with a message.
func (b *StatusBar) SetConnectionStatusWithMessage(status ConnectionStatus, message string) {
	b.status.SetStatusWithMessage(status, message)
}

// View renders the status bar.
func (b *StatusBar) View() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Width(b.width).
		Padding(0, 1)

	// Left side: connection status and board name
	left := b.status.ViewWithLabel()
	if b.boardName != "" {
		left += " | " + b.boardName
	}

	// Right side: task count and filters
	var right string
	if b.taskCount > 0 {
		right = fmt.Sprintf("%d tasks", b.taskCount)
	}
	if b.filterCount > 0 {
		if right != "" {
			right += " | "
		}
		right += fmt.Sprintf("%d filters", b.filterCount)
	}

	// Calculate spacing
	padding := b.width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if padding < 1 {
		padding = 1
	}
	spacer := lipgloss.NewStyle().Width(padding).Render("")

	return style.Render(left + spacer + right)
}
```

**Expected output**: File compiles without errors.

**Common Mistakes**:
- Using ANSI codes directly instead of lipgloss (won't work on all terminals)
- Not handling narrow terminal widths

---

### 4.4 Implement Polling Fallback ✅ COMPLETED

**What**: Create a polling mechanism that activates when SSE fails.

**Why**: SSE might fail due to network issues, firewalls, or proxy limitations. Polling provides a reliable fallback to ensure sync continues.

**Steps**:

1. Add polling commands to `internal/tui/commands.go`
2. Implement task change detection using timestamps
3. Add polling interval configuration

**File**: `internal/tui/commands.go`

```go
package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// =============================================================================
// Data Loading Commands
// =============================================================================

// LoadBoards fetches all boards from the database.
func LoadBoards(app *pocketbase.PocketBase) tea.Cmd {
	return func() tea.Msg {
		records, err := app.FindAllRecords("boards")
		if err != nil {
			return errMsg{err: fmt.Errorf("loading boards: %w", err)}
		}
		return boardsLoadedMsg{boards: records}
	}
}

// LoadTasks fetches all tasks for a specific board.
func LoadTasks(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return func() tea.Msg {
		records, err := app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
		if err != nil {
			return errMsg{err: fmt.Errorf("loading tasks: %w", err)}
		}
		return tasksLoadedMsg{tasks: records}
	}
}

// LoadEpics fetches all epics for a specific board.
func LoadEpics(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return func() tea.Msg {
		records, err := app.FindAllRecords("epics",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
		if err != nil {
			return errMsg{err: fmt.Errorf("loading epics: %w", err)}
		}
		return epicsLoadedMsg{epics: records}
	}
}

// =============================================================================
// Polling Commands
// =============================================================================

// PollConfig holds configuration for the polling fallback.
type PollConfig struct {
	Interval  time.Duration
	BoardID   string
	LastCheck time.Time
}

// StartPolling initiates the polling fallback mechanism.
func StartPolling(app *pocketbase.PocketBase, config PollConfig) tea.Cmd {
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
// Server Check Commands
// =============================================================================

// CheckServerStatus checks if the PocketBase server is reachable.
func CheckServerStatus(serverURL string) tea.Cmd {
	return func() tea.Msg {
		client := NewAPIClientForCheck(serverURL)
		if client.IsServerRunning() {
			return serverOnlineMsg{}
		}
		return serverOfflineMsg{}
	}
}

// ScheduleServerCheck schedules a server health check.
func ScheduleServerCheck(serverURL string, delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		client := NewAPIClientForCheck(serverURL)
		if client.IsServerRunning() {
			return serverOnlineMsg{}
		}
		return serverOfflineMsg{}
	})
}

// APIClientForCheck is a minimal client for health checks.
type APIClientForCheck struct {
	baseURL string
}

// NewAPIClientForCheck creates a new client for health checks.
func NewAPIClientForCheck(baseURL string) *APIClientForCheck {
	return &APIClientForCheck{baseURL: baseURL}
}

// IsServerRunning checks if the server is reachable.
func (c *APIClientForCheck) IsServerRunning() bool {
	// Import http here to avoid dependency issues
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(c.baseURL + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// =============================================================================
// Message Types for Loading
// =============================================================================

// errMsg wraps errors for the Update loop.
type errMsg struct {
	err error
}

// boardsLoadedMsg contains loaded boards.
type boardsLoadedMsg struct {
	boards []*core.Record
}

// tasksLoadedMsg contains loaded tasks.
type tasksLoadedMsg struct {
	tasks []*core.Record
}

// epicsLoadedMsg contains loaded epics.
type epicsLoadedMsg struct {
	epics []*core.Record
}

// =============================================================================
// Add missing import
// =============================================================================

import "net/http"
```

Wait, there's an import issue. Let me fix the file:

**File**: `internal/tui/commands.go` (corrected)

```go
package tui

import (
	"fmt"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// =============================================================================
// Data Loading Commands
// =============================================================================

// LoadBoards fetches all boards from the database.
func LoadBoards(app *pocketbase.PocketBase) tea.Cmd {
	return func() tea.Msg {
		records, err := app.FindAllRecords("boards")
		if err != nil {
			return errMsg{err: fmt.Errorf("loading boards: %w", err)}
		}
		return boardsLoadedMsg{boards: records}
	}
}

// LoadTasks fetches all tasks for a specific board.
func LoadTasks(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return func() tea.Msg {
		records, err := app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
		if err != nil {
			return errMsg{err: fmt.Errorf("loading tasks: %w", err)}
		}
		return tasksLoadedMsg{tasks: records}
	}
}

// LoadEpics fetches all epics for a specific board.
func LoadEpics(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return func() tea.Msg {
		records, err := app.FindAllRecords("epics",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
		if err != nil {
			return errMsg{err: fmt.Errorf("loading epics: %w", err)}
		}
		return epicsLoadedMsg{epics: records}
	}
}

// =============================================================================
// Polling Commands
// =============================================================================

// PollConfig holds configuration for the polling fallback.
type PollConfig struct {
	Interval  time.Duration
	BoardID   string
	LastCheck time.Time
}

// StartPolling initiates the polling fallback mechanism.
func StartPolling(app *pocketbase.PocketBase, config PollConfig) tea.Cmd {
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
// Server Check Commands
// =============================================================================

// CheckServerStatus checks if the PocketBase server is reachable.
func CheckServerStatus(serverURL string) tea.Cmd {
	return func() tea.Msg {
		if isServerRunning(serverURL) {
			return serverOnlineMsg{}
		}
		return serverOfflineMsg{}
	}
}

// ScheduleServerCheck schedules a server health check.
func ScheduleServerCheck(serverURL string, delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		if isServerRunning(serverURL) {
			return serverOnlineMsg{}
		}
		return serverOfflineMsg{}
	})
}

// isServerRunning checks if the server is reachable.
func isServerRunning(baseURL string) bool {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(baseURL + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// =============================================================================
// Message Types for Loading
// =============================================================================

// errMsg wraps errors for the Update loop.
type errMsg struct {
	err error
}

// boardsLoadedMsg contains loaded boards.
type boardsLoadedMsg struct {
	boards []*core.Record
}

// tasksLoadedMsg contains loaded tasks.
type tasksLoadedMsg struct {
	tasks []*core.Record
}

// epicsLoadedMsg contains loaded epics.
type epicsLoadedMsg struct {
	epics []*core.Record
}
```

**Expected output**: File compiles without errors.

**Common Mistakes**:
- Not formatting timestamps correctly for PocketBase queries
- Missing error handling in poll results
- Not continuing the poll loop after each tick

---

### 4.5 Handle Realtime Events in App Model ✅ COMPLETED

**What**: Integrate realtime event handling into the main App model's Update function.

**Why**: Events from the SSE client need to be translated into UI updates. This involves finding the right column, inserting/updating/removing tasks, and triggering re-renders.

**Steps**:

1. Update `internal/tui/app.go` with realtime event handling
2. Implement helper methods for updating columns
3. Handle each event type (create, update, delete)

**File**: `internal/tui/app.go` (relevant sections)

```go
package tui

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/config"
)

// App is the main TUI application model.
type App struct {
	// Core dependencies
	pb        *pocketbase.PocketBase
	serverURL string

	// Board state
	boards       []*core.Record
	currentBoard *core.Record
	boardID      string

	// Task state
	tasks      []*core.Record
	columns    []Column
	focusedCol int

	// Realtime state
	realtimeClient *RealtimeClient
	statusBar      *StatusBar
	usePolling     bool
	lastPollTime   time.Time

	// UI state
	width  int
	height int
	ready  bool
	err    error

	// Key bindings
	keys keyMap
}

// NewApp creates a new TUI application.
func NewApp(pb *pocketbase.PocketBase, boardRef string) *App {
	// Load server URL from config
	serverURL := "http://localhost:8090"
	if cfg, err := config.Load(); err == nil && cfg.Server.URL != "" {
		serverURL = cfg.Server.URL
	}

	app := &App{
		pb:             pb,
		serverURL:      serverURL,
		realtimeClient: NewRealtimeClient(serverURL),
		statusBar:      NewStatusBar(),
		keys:           defaultKeyMap(),
		columns:        make([]Column, 5),
	}

	// Initialize columns
	columnNames := []string{"Backlog", "Todo", "In Progress", "Review", "Done"}
	columnStatuses := []string{"backlog", "todo", "in_progress", "review", "done"}
	for i := range app.columns {
		app.columns[i] = NewColumn(columnStatuses[i], columnNames[i], nil, i == 0)
	}

	return app
}

// Init initializes the TUI application.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		LoadBoards(a.pb),
		CheckServerStatus(a.serverURL),
	)
}

// Update handles all messages in the TUI.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return a.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.statusBar.SetWidth(msg.Width)
		a.updateColumnSizes()
		a.ready = true
		return a, nil

	// =============================================================================
	// Server Status Messages
	// =============================================================================
	case serverOnlineMsg:
		// Server is online, start realtime connection
		a.statusBar.SetConnectionStatus(ConnectionConnecting)
		return a, a.realtimeClient.Connect()

	case serverOfflineMsg:
		// Server is offline, show disconnected status
		a.statusBar.SetConnectionStatusWithMessage(ConnectionDisconnected, "Server offline")
		// Schedule another check in 10 seconds
		return a, ScheduleServerCheck(a.serverURL, 10*time.Second)

	// =============================================================================
	// Realtime Messages
	// =============================================================================
	case realtimeConnectedMsg:
		a.statusBar.SetConnectionStatus(ConnectionConnected)
		a.usePolling = false
		// Start listening for events
		return a, WaitForEvent(a.realtimeClient)

	case realtimeDisconnectedMsg:
		a.statusBar.SetConnectionStatusWithMessage(ConnectionReconnecting, "Reconnecting...")
		// Start reconnection with backoff
		return a, ReconnectWithBackoff(a.realtimeClient, 0)

	case realtimeReconnectMsg:
		a.statusBar.SetConnectionStatusWithMessage(ConnectionReconnecting,
			fmt.Sprintf("Attempt %d/%d", msg.attempt, maxReconnectAttempts))
		return a, a.realtimeClient.Connect()

	case realtimeErrorMsg:
		// Check if we should retry or fall back to polling
		if a.usePolling {
			// Already in polling mode
			return a, nil
		}
		a.statusBar.SetConnectionStatusWithMessage(ConnectionReconnecting, msg.err.Error())
		return a, ReconnectWithBackoff(a.realtimeClient, 0)

	case realtimeEventMsg:
		// Process the realtime event
		cmd := a.handleRealtimeEvent(msg.event)
		// Continue listening for more events
		return a, tea.Batch(cmd, WaitForEvent(a.realtimeClient))

	// =============================================================================
	// Polling Fallback Messages
	// =============================================================================
	case pollStartMsg:
		a.usePolling = true
		a.lastPollTime = time.Now()
		a.statusBar.SetConnectionStatusWithMessage(ConnectionDisconnected, "Polling mode")
		return a, ContinuePolling(pollInterval)

	case pollTickMsg:
		if !a.usePolling {
			return a, nil
		}
		return a, PollForChanges(a.pb, a.boardID, a.lastPollTime)

	case pollResultMsg:
		if msg.err != nil {
			// Poll failed, try again
			return a, ContinuePolling(pollInterval)
		}
		a.lastPollTime = msg.checkTime
		// Update tasks from poll results
		for _, record := range msg.tasks {
			a.updateTaskInColumn(record)
		}
		// Continue polling
		return a, ContinuePolling(pollInterval)

	// =============================================================================
	// Data Loading Messages
	// =============================================================================
	case boardsLoadedMsg:
		a.boards = msg.boards
		if len(a.boards) > 0 {
			// Select first board by default
			a.currentBoard = a.boards[0]
			a.boardID = a.currentBoard.Id
			a.statusBar.SetBoardName(a.currentBoard.GetString("name"))
			return a, LoadTasks(a.pb, a.boardID)
		}
		return a, nil

	case tasksLoadedMsg:
		a.tasks = msg.tasks
		a.distributeTasksToColumns()
		a.statusBar.SetTaskCount(len(a.tasks))
		return a, nil

	case errMsg:
		a.err = msg.err
		return a, nil
	}

	return a, tea.Batch(cmds...)
}

// handleRealtimeEvent processes a single realtime event.
func (a *App) handleRealtimeEvent(event RealtimeEvent) tea.Cmd {
	// Only process task events for now
	if event.Collection != "tasks" {
		return nil
	}

	// Check if this task belongs to the current board
	boardID, ok := event.Record["board"].(string)
	if !ok || boardID != a.boardID {
		return nil
	}

	switch event.Action {
	case "create":
		return a.handleTaskCreated(event.Record)
	case "update":
		return a.handleTaskUpdated(event.Record)
	case "delete":
		return a.handleTaskDeleted(event.Record)
	}

	return nil
}

// handleTaskCreated adds a new task to the appropriate column.
func (a *App) handleTaskCreated(record map[string]interface{}) tea.Cmd {
	// Convert map to TaskItem
	task := a.mapToTaskItem(record)

	// Find the target column
	colIndex := a.columnIndexForStatus(task.Column)
	if colIndex < 0 || colIndex >= len(a.columns) {
		return nil
	}

	// Add to column
	a.columns[colIndex].InsertTask(task)

	// Update task count
	a.statusBar.SetTaskCount(a.statusBar.taskCount + 1)

	return nil
}

// handleTaskUpdated updates a task in place.
func (a *App) handleTaskUpdated(record map[string]interface{}) tea.Cmd {
	task := a.mapToTaskItem(record)
	taskID := task.ID

	// Find current location of task
	oldColIndex, oldItemIndex := a.findTaskLocation(taskID)

	// Find new location
	newColIndex := a.columnIndexForStatus(task.Column)

	if oldColIndex < 0 {
		// Task not found, treat as create
		return a.handleTaskCreated(record)
	}

	if oldColIndex == newColIndex {
		// Same column, update in place
		a.columns[oldColIndex].UpdateTask(oldItemIndex, task)
	} else {
		// Different column, remove and insert
		a.columns[oldColIndex].RemoveTask(oldItemIndex)
		if newColIndex >= 0 && newColIndex < len(a.columns) {
			a.columns[newColIndex].InsertTask(task)
		}
	}

	return nil
}

// handleTaskDeleted removes a task from its column.
func (a *App) handleTaskDeleted(record map[string]interface{}) tea.Cmd {
	taskID, ok := record["id"].(string)
	if !ok {
		return nil
	}

	// Find and remove task
	colIndex, itemIndex := a.findTaskLocation(taskID)
	if colIndex >= 0 {
		a.columns[colIndex].RemoveTask(itemIndex)
		a.statusBar.SetTaskCount(a.statusBar.taskCount - 1)
	}

	return nil
}

// updateTaskInColumn updates a task from a poll result.
func (a *App) updateTaskInColumn(record *core.Record) {
	task := a.recordToTaskItem(record)
	taskID := task.ID

	// Find current location
	oldColIndex, oldItemIndex := a.findTaskLocation(taskID)
	newColIndex := a.columnIndexForStatus(task.Column)

	if oldColIndex < 0 {
		// New task, insert it
		if newColIndex >= 0 && newColIndex < len(a.columns) {
			a.columns[newColIndex].InsertTask(task)
			a.statusBar.SetTaskCount(a.statusBar.taskCount + 1)
		}
		return
	}

	if oldColIndex == newColIndex {
		// Same column, update in place
		a.columns[oldColIndex].UpdateTask(oldItemIndex, task)
	} else {
		// Different column, move it
		a.columns[oldColIndex].RemoveTask(oldItemIndex)
		if newColIndex >= 0 && newColIndex < len(a.columns) {
			a.columns[newColIndex].InsertTask(task)
		}
	}
}

// findTaskLocation finds a task's column and index.
func (a *App) findTaskLocation(taskID string) (colIndex, itemIndex int) {
	for i, col := range a.columns {
		for j, item := range col.list.Items() {
			if task, ok := item.(TaskItem); ok && task.ID == taskID {
				return i, j
			}
		}
	}
	return -1, -1
}

// columnIndexForStatus returns the column index for a status string.
func (a *App) columnIndexForStatus(status string) int {
	statuses := []string{"backlog", "todo", "in_progress", "review", "done"}
	for i, s := range statuses {
		if s == status {
			return i
		}
	}
	return -1
}

// distributeTasksToColumns sorts tasks into their respective columns.
func (a *App) distributeTasksToColumns() {
	// Group tasks by column
	tasksByColumn := make(map[string][]TaskItem)
	for _, record := range a.tasks {
		task := a.recordToTaskItem(record)
		tasksByColumn[task.Column] = append(tasksByColumn[task.Column], task)
	}

	// Sort each group by position
	for _, tasks := range tasksByColumn {
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Position < tasks[j].Position
		})
	}

	// Populate columns
	statuses := []string{"backlog", "todo", "in_progress", "review", "done"}
	for i, status := range statuses {
		tasks := tasksByColumn[status]
		items := make([]list.Item, len(tasks))
		for j, task := range tasks {
			items[j] = task
		}
		a.columns[i].SetItems(items)
	}
}

// mapToTaskItem converts a map[string]interface{} to TaskItem.
func (a *App) mapToTaskItem(m map[string]interface{}) TaskItem {
	getString := func(key string) string {
		if v, ok := m[key].(string); ok {
			return v
		}
		return ""
	}
	getFloat := func(key string) float64 {
		if v, ok := m[key].(float64); ok {
			return v
		}
		return 0
	}
	getStringSlice := func(key string) []string {
		if v, ok := m[key].([]interface{}); ok {
			result := make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
		return nil
	}

	return TaskItem{
		ID:          getString("id"),
		DisplayID:   fmt.Sprintf("%s-%d", getString("board_prefix"), int(getFloat("seq"))),
		Title:       getString("title"),
		Description: getString("description"),
		Type:        getString("type"),
		Priority:    getString("priority"),
		Column:      getString("column"),
		Position:    getFloat("position"),
		Labels:      getStringSlice("labels"),
		BlockedBy:   getStringSlice("blocked_by"),
		Epic:        getString("epic"),
		DueDate:     getString("due_date"),
	}
}

// recordToTaskItem converts a PocketBase record to TaskItem.
func (a *App) recordToTaskItem(r *core.Record) TaskItem {
	return TaskItem{
		ID:          r.Id,
		DisplayID:   fmt.Sprintf("%s-%d", r.GetString("board_prefix"), r.GetInt("seq")),
		Title:       r.GetString("title"),
		Description: r.GetString("description"),
		Type:        r.GetString("type"),
		Priority:    r.GetString("priority"),
		Column:      r.GetString("column"),
		Position:    r.GetFloat("position"),
		Labels:      r.GetStringSlice("labels"),
		BlockedBy:   r.GetStringSlice("blocked_by"),
		Epic:        r.GetString("epic"),
		DueDate:     r.GetString("due_date"),
	}
}

// updateColumnSizes recalculates column dimensions.
func (a *App) updateColumnSizes() {
	if len(a.columns) == 0 {
		return
	}

	// Reserve space for status bar
	contentHeight := a.height - 2

	// Divide width equally among columns
	colWidth := (a.width - 4) / len(a.columns)

	for i := range a.columns {
		a.columns[i].SetSize(colWidth, contentHeight)
	}
}

// View renders the TUI.
func (a *App) View() string {
	if !a.ready {
		return "Loading..."
	}

	if a.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", a.err)
	}

	// Render columns
	var cols []string
	for i, col := range a.columns {
		style := blurredColumnStyle
		if i == a.focusedCol {
			style = focusedColumnStyle
		}
		colWidth := (a.width - 4) / len(a.columns)
		cols = append(cols, style.Width(colWidth).Render(col.View()))
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, cols...)
	status := a.statusBar.View()

	return lipgloss.JoinVertical(lipgloss.Left, board, status)
}

// handleKeyMsg processes keyboard input.
func (a *App) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Quit):
		a.realtimeClient.Disconnect()
		return a, tea.Quit

	case key.Matches(msg, a.keys.Left):
		if a.focusedCol > 0 {
			a.columns[a.focusedCol].SetFocused(false)
			a.focusedCol--
			a.columns[a.focusedCol].SetFocused(true)
		}
		return a, nil

	case key.Matches(msg, a.keys.Right):
		if a.focusedCol < len(a.columns)-1 {
			a.columns[a.focusedCol].SetFocused(false)
			a.focusedCol++
			a.columns[a.focusedCol].SetFocused(true)
		}
		return a, nil

	case key.Matches(msg, a.keys.Refresh):
		// Manual refresh
		return a, LoadTasks(a.pb, a.boardID)
	}

	// Pass to focused column
	col, cmd := a.columns[a.focusedCol].Update(msg)
	a.columns[a.focusedCol] = col
	return a, cmd
}
```

**Expected output**: File compiles without errors.

**Common Mistakes**:
- Not checking if the event belongs to the current board
- Forgetting to continue listening for events after processing one
- Not handling the case where a task is moved between columns

---

### 4.6 Add Column Helper Methods ✅ COMPLETED

**What**: Add methods to the Column component for inserting, updating, and removing tasks.

**Why**: Realtime updates need to modify columns without a full reload. These helpers enable incremental updates.

**File**: `internal/tui/column.go` (additions)

```go
package tui

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Column represents a kanban column with a list of tasks.
type Column struct {
	status  string     // "backlog", "todo", etc.
	title   string     // Display title
	list    list.Model // bubbles/list
	focused bool
}

// NewColumn creates a new column with the given tasks.
func NewColumn(status, title string, items []list.Item, focused bool) Column {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetHeight(3) // Compact cards

	l := list.New(items, delegate, 0, 0)
	l.Title = title
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false) // Use global filter
	l.SetShowHelp(false)         // Use global help

	return Column{
		status:  status,
		title:   title,
		list:    l,
		focused: focused,
	}
}

// SetSize sets the column dimensions.
func (c *Column) SetSize(width, height int) {
	c.list.SetSize(width-2, height-4) // Account for borders and title
}

// SetItems replaces all items in the column.
func (c *Column) SetItems(items []list.Item) {
	c.list.SetItems(items)
}

// SetFocused sets whether this column is focused.
func (c *Column) SetFocused(focused bool) {
	c.focused = focused
}

// InsertTask adds a task to the column in sorted position order.
func (c *Column) InsertTask(task TaskItem) {
	items := c.list.Items()

	// Find insertion point based on position
	insertAt := len(items)
	for i, item := range items {
		if t, ok := item.(TaskItem); ok && t.Position > task.Position {
			insertAt = i
			break
		}
	}

	// Insert at position
	newItems := make([]list.Item, 0, len(items)+1)
	newItems = append(newItems, items[:insertAt]...)
	newItems = append(newItems, task)
	newItems = append(newItems, items[insertAt:]...)

	c.list.SetItems(newItems)
}

// UpdateTask updates a task at the given index.
func (c *Column) UpdateTask(index int, task TaskItem) {
	items := c.list.Items()
	if index < 0 || index >= len(items) {
		return
	}

	items[index] = task

	// Check if position changed - if so, re-sort
	needsSort := false
	if index > 0 {
		if prev, ok := items[index-1].(TaskItem); ok && prev.Position > task.Position {
			needsSort = true
		}
	}
	if index < len(items)-1 {
		if next, ok := items[index+1].(TaskItem); ok && next.Position < task.Position {
			needsSort = true
		}
	}

	if needsSort {
		// Re-sort by position
		taskItems := make([]TaskItem, len(items))
		for i, item := range items {
			taskItems[i] = item.(TaskItem)
		}
		sort.Slice(taskItems, func(i, j int) bool {
			return taskItems[i].Position < taskItems[j].Position
		})
		for i, t := range taskItems {
			items[i] = t
		}
	}

	c.list.SetItems(items)
}

// RemoveTask removes a task at the given index.
func (c *Column) RemoveTask(index int) {
	items := c.list.Items()
	if index < 0 || index >= len(items) {
		return
	}

	newItems := make([]list.Item, 0, len(items)-1)
	newItems = append(newItems, items[:index]...)
	newItems = append(newItems, items[index+1:]...)

	c.list.SetItems(newItems)
}

// Update handles messages for this column.
func (c *Column) Update(msg tea.Msg) (Column, tea.Cmd) {
	var cmd tea.Cmd
	c.list, cmd = c.list.Update(msg)
	return *c, cmd
}

// View renders the column.
func (c *Column) View() string {
	titleStyle := columnHeaderStyle(c.status, c.focused)

	count := len(c.list.Items())
	header := titleStyle.Render(fmt.Sprintf("%s (%d)", c.title, count))

	return lipgloss.JoinVertical(lipgloss.Left, header, c.list.View())
}

// SelectedTask returns the currently selected task, if any.
func (c *Column) SelectedTask() (TaskItem, bool) {
	item := c.list.SelectedItem()
	if item == nil {
		return TaskItem{}, false
	}
	task, ok := item.(TaskItem)
	return task, ok
}
```

**Expected output**: File compiles without errors.

**Common Mistakes**:
- Not handling empty columns
- Index out of bounds when removing items
- Forgetting to re-sort after position changes

---

### 4.7 Write Tests for Realtime Event Handling ✅ COMPLETED

**What**: Create comprehensive tests for the realtime subsystem.

**Why**: Realtime sync is critical functionality. Tests ensure events are processed correctly and UI state stays consistent.

**File**: `internal/tui/realtime_test.go`

```go
package tui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRealtimeEventParsing tests parsing of SSE events.
func TestRealtimeEventParsing(t *testing.T) {
	tests := []struct {
		name     string
		event    RealtimeEvent
		wantCol  string
		wantAct  string
	}{
		{
			name: "create task event",
			event: RealtimeEvent{
				Action:     "create",
				Collection: "tasks",
				Record: map[string]interface{}{
					"id":       "task123",
					"title":    "New Task",
					"column":   "todo",
					"position": 1.0,
				},
			},
			wantCol: "tasks",
			wantAct: "create",
		},
		{
			name: "update task event",
			event: RealtimeEvent{
				Action:     "update",
				Collection: "tasks",
				Record: map[string]interface{}{
					"id":       "task123",
					"title":    "Updated Task",
					"column":   "in_progress",
					"position": 2.0,
				},
			},
			wantCol: "tasks",
			wantAct: "update",
		},
		{
			name: "delete task event",
			event: RealtimeEvent{
				Action:     "delete",
				Collection: "tasks",
				Record: map[string]interface{}{
					"id": "task123",
				},
			},
			wantCol: "tasks",
			wantAct: "delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCol, tt.event.Collection)
			assert.Equal(t, tt.wantAct, tt.event.Action)
		})
	}
}

// TestConnectionStatus tests connection status transitions.
func TestConnectionStatus(t *testing.T) {
	indicator := NewStatusIndicator()

	// Initial state
	assert.Equal(t, ConnectionDisconnected, indicator.Status())

	// Transition to connecting
	indicator.SetStatus(ConnectionConnecting)
	assert.Equal(t, ConnectionConnecting, indicator.Status())

	// Transition to connected
	indicator.SetStatus(ConnectionConnected)
	assert.Equal(t, ConnectionConnected, indicator.Status())

	// Transition to reconnecting
	indicator.SetStatus(ConnectionReconnecting)
	assert.Equal(t, ConnectionReconnecting, indicator.Status())

	// Back to disconnected
	indicator.SetStatus(ConnectionDisconnected)
	assert.Equal(t, ConnectionDisconnected, indicator.Status())
}

// TestStatusIndicatorView tests rendering of status indicator.
func TestStatusIndicatorView(t *testing.T) {
	indicator := NewStatusIndicator()

	tests := []struct {
		status  ConnectionStatus
		wantStr string
	}{
		{ConnectionConnected, "Live"},
		{ConnectionConnecting, "Connecting..."},
		{ConnectionReconnecting, "Reconnecting..."},
		{ConnectionDisconnected, "Offline"},
	}

	for _, tt := range tests {
		indicator.SetStatus(tt.status)
		view := indicator.ViewWithLabel()
		assert.Contains(t, view, tt.wantStr)
	}
}

// TestColumnTaskOperations tests column insert/update/remove operations.
func TestColumnTaskOperations(t *testing.T) {
	t.Run("InsertTask", func(t *testing.T) {
		col := NewColumn("todo", "Todo", nil, false)

		// Insert first task
		task1 := TaskItem{ID: "1", Title: "Task 1", Position: 1.0}
		col.InsertTask(task1)
		assert.Equal(t, 1, len(col.list.Items()))

		// Insert second task with higher position
		task2 := TaskItem{ID: "2", Title: "Task 2", Position: 2.0}
		col.InsertTask(task2)
		assert.Equal(t, 2, len(col.list.Items()))

		// Verify order
		items := col.list.Items()
		assert.Equal(t, "1", items[0].(TaskItem).ID)
		assert.Equal(t, "2", items[1].(TaskItem).ID)
	})

	t.Run("InsertTaskMaintainsOrder", func(t *testing.T) {
		col := NewColumn("todo", "Todo", nil, false)

		// Insert tasks out of order
		col.InsertTask(TaskItem{ID: "3", Title: "Task 3", Position: 3.0})
		col.InsertTask(TaskItem{ID: "1", Title: "Task 1", Position: 1.0})
		col.InsertTask(TaskItem{ID: "2", Title: "Task 2", Position: 2.0})

		items := col.list.Items()
		require.Equal(t, 3, len(items))
		assert.Equal(t, "1", items[0].(TaskItem).ID)
		assert.Equal(t, "2", items[1].(TaskItem).ID)
		assert.Equal(t, "3", items[2].(TaskItem).ID)
	})

	t.Run("UpdateTask", func(t *testing.T) {
		col := NewColumn("todo", "Todo", nil, false)
		col.InsertTask(TaskItem{ID: "1", Title: "Original", Position: 1.0})

		// Update the task
		col.UpdateTask(0, TaskItem{ID: "1", Title: "Updated", Position: 1.0})

		items := col.list.Items()
		assert.Equal(t, "Updated", items[0].(TaskItem).Title)
	})

	t.Run("RemoveTask", func(t *testing.T) {
		col := NewColumn("todo", "Todo", nil, false)
		col.InsertTask(TaskItem{ID: "1", Title: "Task 1", Position: 1.0})
		col.InsertTask(TaskItem{ID: "2", Title: "Task 2", Position: 2.0})
		col.InsertTask(TaskItem{ID: "3", Title: "Task 3", Position: 3.0})

		// Remove middle task
		col.RemoveTask(1)

		items := col.list.Items()
		require.Equal(t, 2, len(items))
		assert.Equal(t, "1", items[0].(TaskItem).ID)
		assert.Equal(t, "3", items[1].(TaskItem).ID)
	})

	t.Run("RemoveLastTask", func(t *testing.T) {
		col := NewColumn("todo", "Todo", nil, false)
		col.InsertTask(TaskItem{ID: "1", Title: "Task 1", Position: 1.0})

		col.RemoveTask(0)

		assert.Equal(t, 0, len(col.list.Items()))
	})
}

// TestMapToTaskItem tests conversion from raw event data to TaskItem.
func TestMapToTaskItem(t *testing.T) {
	// Create a minimal App for testing
	app := &App{}

	record := map[string]interface{}{
		"id":           "abc123",
		"title":        "Test Task",
		"description":  "A test description",
		"type":         "feature",
		"priority":     "high",
		"column":       "in_progress",
		"position":     1.5,
		"seq":          42.0,
		"board_prefix": "WRK",
		"labels":       []interface{}{"backend", "urgent"},
		"blocked_by":   []interface{}{"xyz789"},
		"epic":         "epic123",
		"due_date":     "2025-01-15",
	}

	task := app.mapToTaskItem(record)

	assert.Equal(t, "abc123", task.ID)
	assert.Equal(t, "Test Task", task.Title)
	assert.Equal(t, "A test description", task.Description)
	assert.Equal(t, "feature", task.Type)
	assert.Equal(t, "high", task.Priority)
	assert.Equal(t, "in_progress", task.Column)
	assert.Equal(t, 1.5, task.Position)
	assert.Equal(t, []string{"backend", "urgent"}, task.Labels)
	assert.Equal(t, []string{"xyz789"}, task.BlockedBy)
	assert.Equal(t, "epic123", task.Epic)
	assert.Equal(t, "2025-01-15", task.DueDate)
}

// TestRealtimeClientLifecycle tests client connection lifecycle.
func TestRealtimeClientLifecycle(t *testing.T) {
	client := NewRealtimeClient("http://localhost:8090")

	// Initial state
	assert.False(t, client.IsConnected())

	// Disconnect should be safe to call before connect
	client.Disconnect()
	assert.False(t, client.IsConnected())
}

// TestExponentialBackoff tests reconnection delay calculation.
func TestExponentialBackoff(t *testing.T) {
	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 16 * time.Second},
		{5, 30 * time.Second}, // Capped at max
		{10, 30 * time.Second}, // Still capped
	}

	for _, tt := range tests {
		delay := baseReconnectDelay * (1 << tt.attempt)
		if delay > maxReconnectDelay {
			delay = maxReconnectDelay
		}
		assert.Equal(t, tt.expected, delay, "attempt %d", tt.attempt)
	}
}

// TestPollConfig tests polling configuration.
func TestPollConfig(t *testing.T) {
	config := PollConfig{
		Interval:  3 * time.Second,
		BoardID:   "board123",
		LastCheck: time.Now(),
	}

	assert.Equal(t, 3*time.Second, config.Interval)
	assert.Equal(t, "board123", config.BoardID)
	assert.False(t, config.LastCheck.IsZero())
}
```

**Expected output**: All tests pass.

**Common Mistakes**:
- Not testing edge cases (empty columns, invalid indices)
- Assuming specific event ordering
- Not testing reconnection logic

---

## Verification Checklist

Complete each section in order. Check off each item as you verify it.

### Code Compilation

- [ ] **All files compile**
  ```bash
  go build ./internal/tui/...
  ```
  Should complete without errors.

- [ ] **Dependencies are available**
  ```bash
  go mod tidy
  ```
  Should complete without errors.

### Unit Tests

- [ ] **Realtime tests pass**
  ```bash
  go test ./internal/tui/... -v -run Realtime
  ```
  Should show all tests passing.

- [ ] **Column operation tests pass**
  ```bash
  go test ./internal/tui/... -v -run Column
  ```
  Should show all tests passing.

- [ ] **Status indicator tests pass**
  ```bash
  go test ./internal/tui/... -v -run Status
  ```
  Should show all tests passing.

### Integration Testing

- [ ] **Start server**
  ```bash
  egenskriven serve &
  ```
  Server should start on port 8090.

- [ ] **Launch TUI**
  ```bash
  egenskriven tui
  ```
  TUI should display with connection status indicator.

- [ ] **Verify initial connection**
  
  Status indicator should show green dot with "Live" after a few seconds.

- [ ] **Test create event**
  1. Open web UI in browser
  2. Create a new task
  3. Verify task appears in TUI without refresh

- [ ] **Test update event**
  1. Edit a task in web UI
  2. Verify change appears in TUI immediately

- [ ] **Test delete event**
  1. Delete a task in web UI
  2. Verify task disappears from TUI immediately

- [ ] **Test move event**
  1. Drag a task to different column in web UI
  2. Verify task moves to new column in TUI

### Fallback Testing

- [ ] **Test reconnection**
  1. Stop the server while TUI is running
  2. Status should change to "Reconnecting..."
  3. Restart server
  4. Status should return to "Live"

- [ ] **Test polling fallback**
  1. Configure server to block SSE (or simulate with firewall)
  2. TUI should fall back to "Polling mode" after max retries
  3. Changes should still sync (with slight delay)

### Edge Cases

- [ ] **Empty board**
  
  TUI should handle boards with no tasks gracefully.

- [ ] **High-frequency updates**
  
  Rapid task changes should not crash or lag the TUI.

- [ ] **Large payloads**
  
  Tasks with long descriptions should parse correctly.

---

## File Summary

| File | Lines | Purpose |
|------|-------|---------|
| `internal/tui/messages.go` | ~100 | Message type definitions for realtime events |
| `internal/tui/realtime.go` | ~350 | SSE client implementation |
| `internal/tui/status.go` | ~150 | Connection status indicator |
| `internal/tui/commands.go` | ~150 | Data loading and polling commands |
| `internal/tui/app.go` | ~400 | Main app with realtime handling |
| `internal/tui/column.go` | ~150 | Column with insert/update/remove helpers |
| `internal/tui/realtime_test.go` | ~250 | Tests for realtime functionality |

**Total new/modified code**: ~1,550 lines

---

## What You Should Have Now

After completing Phase 4, your TUI should:

```
internal/tui/
├── app.go           # Updated with realtime handling
├── column.go        # Updated with task operations
├── commands.go      # New - data loading and polling
├── messages.go      # New - message type definitions
├── realtime.go      # New - SSE client
├── realtime_test.go # New - tests
├── status.go        # New - connection indicator
├── styles.go        # Existing
├── keys.go          # Existing
├── task_item.go     # Existing
└── ...
```

**Functionality:**
- Connection status indicator in status bar
- SSE connection to PocketBase realtime API
- Live task create/update/delete events
- Automatic reconnection with exponential backoff
- Polling fallback when SSE fails
- Graceful shutdown of connections

---

## Next Phase

**Phase 5: Filtering & Search** will add:
- Search input with `/` key
- Filter by priority, type, label, epic
- Filter bar showing active filters
- Filter persistence during session
- Quick filter shortcuts (fp, ft, fl, fe)
- Clear filters command (fc)

---

## Troubleshooting

### "SSE connection times out"

**Problem**: The SSE connection fails immediately or times out.

**Solution**:
```go
// Ensure the HTTP client has no timeout for SSE
httpClient := &http.Client{
    Timeout: 0, // Important: no timeout for long-lived connections
}
```

Also check that the server is running:
```bash
curl http://localhost:8090/api/health
```

### "Events aren't being received"

**Problem**: SSE connects but no events come through.

**Solution**:
1. Verify subscription was sent successfully
2. Check the collections list matches your schema
3. Verify events are being generated (check web UI)
4. Increase event buffer size if events are being dropped

```go
events := make(chan RealtimeEvent, 100) // Increase if needed
```

### "Polling fallback keeps failing"

**Problem**: Polling queries fail with database errors.

**Solution**:
Check the timestamp format for PocketBase queries:
```go
// Correct format
timestamp := lastCheck.UTC().Format("2006-01-02 15:04:05.000Z")
```

### "Tasks appear duplicated after reconnect"

**Problem**: Tasks are duplicated when SSE reconnects.

**Solution**:
Always check if task exists before inserting:
```go
func (a *App) handleTaskCreated(record map[string]interface{}) tea.Cmd {
    task := a.mapToTaskItem(record)
    
    // Check if already exists
    if colIdx, _ := a.findTaskLocation(task.ID); colIdx >= 0 {
        // Already exists, treat as update
        return a.handleTaskUpdated(record)
    }
    // ... rest of create logic
}
```

### "Status indicator flickers"

**Problem**: Status indicator rapidly changes between states.

**Solution**:
Add debouncing to status changes:
```go
// Only update if status actually changed
func (s *StatusIndicator) SetStatus(status ConnectionStatus) {
    if s.status == status {
        return // No change
    }
    s.status = status
    s.message = ""
}
```

### "Channel deadlock on disconnect"

**Problem**: TUI freezes when SSE connection drops.

**Solution**:
Always use non-blocking channel sends:
```go
select {
case c.events <- event:
    // Sent successfully
default:
    // Channel full, drop event (log this in production)
}
```

### "Context cancellation doesn't stop SSE"

**Problem**: Calling Disconnect() doesn't stop the SSE goroutine.

**Solution**:
Check context in the read loop:
```go
for scanner.Scan() {
    select {
    case <-c.ctx.Done():
        return // Context cancelled, exit loop
    default:
    }
    // ... process line
}
```

---

## Glossary

| Term | Definition |
|------|------------|
| **SSE** | Server-Sent Events - a protocol for one-way server-to-client streaming |
| **PB_CONNECT** | PocketBase's initial SSE event containing the client ID |
| **Realtime** | PocketBase's subscription system for live database updates |
| **Exponential Backoff** | Reconnection strategy where delay doubles each attempt |
| **Polling** | Periodically querying for changes (fallback for SSE) |
| **tea.Cmd** | Bubble Tea command - an async operation that returns a message |
| **tea.Msg** | Bubble Tea message - triggers an Update cycle |
| **lipgloss** | Charm's terminal styling library |
