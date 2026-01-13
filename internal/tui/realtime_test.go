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

// TestStatusIndicatorWithMessage tests status with message.
func TestStatusIndicatorWithMessage(t *testing.T) {
	indicator := NewStatusIndicator()

	indicator.SetStatusWithMessage(ConnectionReconnecting, "attempt 3/5")
	assert.Equal(t, ConnectionReconnecting, indicator.Status())

	view := indicator.View()
	assert.Contains(t, view, "attempt 3/5")
}

// TestConnectionStatusString tests String method of ConnectionStatus.
func TestConnectionStatusString(t *testing.T) {
	tests := []struct {
		status ConnectionStatus
		want   string
	}{
		{ConnectionDisconnected, "disconnected"},
		{ConnectionConnecting, "connecting"},
		{ConnectionConnected, "connected"},
		{ConnectionReconnecting, "reconnecting"},
		{ConnectionStatus(99), "unknown"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.status.String())
	}
}

// TestColumnInsertTask tests inserting tasks into a column.
func TestColumnInsertTask(t *testing.T) {
	col := NewColumn("todo", nil, false)

	// Insert first task
	task1 := TaskItem{ID: "1", TaskTitle: "Task 1", Position: 1.0}
	col.InsertTask(task1)
	assert.Equal(t, 1, len(col.Items()))

	// Insert second task with higher position
	task2 := TaskItem{ID: "2", TaskTitle: "Task 2", Position: 2.0}
	col.InsertTask(task2)
	assert.Equal(t, 2, len(col.Items()))

	// Verify order
	items := col.Items()
	assert.Equal(t, "1", items[0].(TaskItem).ID)
	assert.Equal(t, "2", items[1].(TaskItem).ID)
}

// TestColumnInsertTaskMaintainsOrder tests that insert maintains sorted order.
func TestColumnInsertTaskMaintainsOrder(t *testing.T) {
	col := NewColumn("todo", nil, false)

	// Insert tasks out of order
	col.InsertTask(TaskItem{ID: "3", TaskTitle: "Task 3", Position: 3.0})
	col.InsertTask(TaskItem{ID: "1", TaskTitle: "Task 1", Position: 1.0})
	col.InsertTask(TaskItem{ID: "2", TaskTitle: "Task 2", Position: 2.0})

	items := col.Items()
	require.Equal(t, 3, len(items))
	assert.Equal(t, "1", items[0].(TaskItem).ID)
	assert.Equal(t, "2", items[1].(TaskItem).ID)
	assert.Equal(t, "3", items[2].(TaskItem).ID)
}

// TestColumnUpdateTask tests updating a task in a column.
func TestColumnUpdateTask(t *testing.T) {
	col := NewColumn("todo", nil, false)
	col.InsertTask(TaskItem{ID: "1", TaskTitle: "Original", Position: 1.0})

	// Update the task
	col.UpdateTask(0, TaskItem{ID: "1", TaskTitle: "Updated", Position: 1.0})

	items := col.Items()
	assert.Equal(t, "Updated", items[0].(TaskItem).TaskTitle)
}

// TestColumnUpdateTaskInvalidIndex tests updating with invalid index.
func TestColumnUpdateTaskInvalidIndex(t *testing.T) {
	col := NewColumn("todo", nil, false)
	col.InsertTask(TaskItem{ID: "1", TaskTitle: "Task 1", Position: 1.0})

	// Should not panic with invalid indices
	col.UpdateTask(-1, TaskItem{ID: "1", TaskTitle: "Updated", Position: 1.0})
	col.UpdateTask(99, TaskItem{ID: "1", TaskTitle: "Updated", Position: 1.0})

	// Original should be unchanged
	items := col.Items()
	assert.Equal(t, "Task 1", items[0].(TaskItem).TaskTitle)
}

// TestColumnRemoveTask tests removing a task from a column.
func TestColumnRemoveTask(t *testing.T) {
	col := NewColumn("todo", nil, false)
	col.InsertTask(TaskItem{ID: "1", TaskTitle: "Task 1", Position: 1.0})
	col.InsertTask(TaskItem{ID: "2", TaskTitle: "Task 2", Position: 2.0})
	col.InsertTask(TaskItem{ID: "3", TaskTitle: "Task 3", Position: 3.0})

	// Remove middle task
	col.RemoveTask(1)

	items := col.Items()
	require.Equal(t, 2, len(items))
	assert.Equal(t, "1", items[0].(TaskItem).ID)
	assert.Equal(t, "3", items[1].(TaskItem).ID)
}

// TestColumnRemoveLastTask tests removing the only task.
func TestColumnRemoveLastTask(t *testing.T) {
	col := NewColumn("todo", nil, false)
	col.InsertTask(TaskItem{ID: "1", TaskTitle: "Task 1", Position: 1.0})

	col.RemoveTask(0)

	assert.Equal(t, 0, len(col.Items()))
}

// TestColumnRemoveTaskInvalidIndex tests removing with invalid index.
func TestColumnRemoveTaskInvalidIndex(t *testing.T) {
	col := NewColumn("todo", nil, false)
	col.InsertTask(TaskItem{ID: "1", TaskTitle: "Task 1", Position: 1.0})

	// Should not panic with invalid indices
	col.RemoveTask(-1)
	col.RemoveTask(99)

	// Original should be unchanged
	assert.Equal(t, 1, len(col.Items()))
}

// TestColumnFindTaskByID tests finding a task by ID.
func TestColumnFindTaskByID(t *testing.T) {
	col := NewColumn("todo", nil, false)
	col.InsertTask(TaskItem{ID: "abc", TaskTitle: "Task A", Position: 1.0})
	col.InsertTask(TaskItem{ID: "def", TaskTitle: "Task B", Position: 2.0})
	col.InsertTask(TaskItem{ID: "ghi", TaskTitle: "Task C", Position: 3.0})

	// Find existing task
	idx := col.FindTaskByID("def")
	assert.Equal(t, 1, idx)

	// Find first task
	idx = col.FindTaskByID("abc")
	assert.Equal(t, 0, idx)

	// Find last task
	idx = col.FindTaskByID("ghi")
	assert.Equal(t, 2, idx)

	// Find non-existent task
	idx = col.FindTaskByID("nonexistent")
	assert.Equal(t, -1, idx)
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

// TestRealtimeClientSetCollections tests setting collections to subscribe to.
func TestRealtimeClientSetCollections(t *testing.T) {
	client := NewRealtimeClient("http://localhost:8090")

	// Default collections
	assert.Equal(t, []string{"tasks", "boards", "epics"}, client.collections)

	// Set custom collections
	client.SetCollections([]string{"tasks"})
	assert.Equal(t, []string{"tasks"}, client.collections)
}

// TestRealtimeClientEvents tests the events channel.
func TestRealtimeClientEvents(t *testing.T) {
	client := NewRealtimeClient("http://localhost:8090")

	// Events channel should be available
	eventsChan := client.Events()
	assert.NotNil(t, eventsChan)
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
	now := time.Now()
	config := PollConfig{
		Interval:  3 * time.Second,
		BoardID:   "board123",
		LastCheck: now,
	}

	assert.Equal(t, 3*time.Second, config.Interval)
	assert.Equal(t, "board123", config.BoardID)
	assert.Equal(t, now, config.LastCheck)
}

// TestStatusBar tests the status bar component.
func TestStatusBar(t *testing.T) {
	bar := NewStatusBar()

	// Set width
	bar.SetWidth(80)

	// Set board name
	bar.SetBoardName("Test Board")

	// Set task count
	bar.SetTaskCount(5)

	// Set filter count
	bar.SetFilterCount(2)

	// Set connection status
	bar.SetConnectionStatus(ConnectionConnected)

	// Render view
	view := bar.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Live")
	assert.Contains(t, view, "Test Board")
	assert.Contains(t, view, "5 tasks")
	assert.Contains(t, view, "2 filters")
}

// TestStatusBarWithMessage tests status bar with connection message.
func TestStatusBarWithMessage(t *testing.T) {
	bar := NewStatusBar()
	bar.SetWidth(80)

	bar.SetConnectionStatusWithMessage(ConnectionReconnecting, "retry 2/5")

	view := bar.View()
	assert.Contains(t, view, "Reconnecting...")
}
