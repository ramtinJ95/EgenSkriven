package tui

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
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
		log.Printf("realtime: failed to parse event data for %q: %v", eventType, err)
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
		// Event sent successfully
	default:
		// Channel full, event dropped. This can happen during high-frequency
		// updates. The polling fallback will eventually sync any missed changes.
		log.Printf("realtime: event channel full, dropped %s event for %s", event.Action, event.Collection)
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
