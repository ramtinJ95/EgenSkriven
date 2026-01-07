package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ramtinJ95/EgenSkriven/internal/config"
)

const (
	// DefaultServerURL is the default PocketBase server URL
	DefaultServerURL = "http://localhost:8090"

	// HealthCheckTimeout is the timeout for health check requests (quick fail for offline detection)
	HealthCheckTimeout = 500 * time.Millisecond

	// APIRequestTimeout is the timeout for actual API operations
	APIRequestTimeout = 5 * time.Second
)

// APIClient handles HTTP requests to the PocketBase server.
// It provides methods for CRUD operations on tasks that trigger real-time events.
type APIClient struct {
	baseURL       string
	healthClient  *http.Client // Quick timeout for health checks
	requestClient *http.Client // Longer timeout for actual requests
}

// NewAPIClient creates a new API client with the configured or default server URL.
func NewAPIClient() *APIClient {
	baseURL := DefaultServerURL

	// Try to load URL from config
	if cfg, err := config.LoadProjectConfig(); err == nil && cfg.Server.URL != "" {
		baseURL = cfg.Server.URL
	}

	return &APIClient{
		baseURL: baseURL,
		healthClient: &http.Client{
			Timeout: HealthCheckTimeout,
		},
		requestClient: &http.Client{
			Timeout: APIRequestTimeout,
		},
	}
}

// NewAPIClientWithURL creates a new API client with a specific server URL.
func NewAPIClientWithURL(url string) *APIClient {
	return &APIClient{
		baseURL: url,
		healthClient: &http.Client{
			Timeout: HealthCheckTimeout,
		},
		requestClient: &http.Client{
			Timeout: APIRequestTimeout,
		},
	}
}

// IsServerRunning checks if the PocketBase server is accessible.
// Uses a quick timeout to avoid blocking when server is offline.
func (c *APIClient) IsServerRunning() bool {
	resp, err := c.healthClient.Get(c.baseURL + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// APIError represents an error response from the PocketBase API.
type APIError struct {
	StatusCode int
	Status     int            `json:"status"`
	Message    string         `json:"message"`
	Data       map[string]any `json:"data"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// IsValidationError returns true if this is a client-side validation error (4xx).
// Validation errors should NOT trigger fallback to direct DB access.
func (e *APIError) IsValidationError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// TaskData represents the data for creating or updating a task via API.
type TaskData struct {
	ID             string   `json:"id,omitempty"`
	Title          string   `json:"title,omitempty"`
	Description    string   `json:"description,omitempty"`
	Type           string   `json:"type,omitempty"`
	Priority       string   `json:"priority,omitempty"`
	Column         string   `json:"column,omitempty"`
	Position       float64  `json:"position,omitempty"`
	Labels         []string `json:"labels,omitempty"`
	BlockedBy      []string `json:"blocked_by,omitempty"`
	CreatedBy      string   `json:"created_by,omitempty"`
	CreatedByAgent string   `json:"created_by_agent,omitempty"`
	Epic           string   `json:"epic,omitempty"`
	Board          string   `json:"board,omitempty"`
	Seq            int      `json:"seq,omitempty"`
	Parent         string   `json:"parent,omitempty"`
	History        []any    `json:"history,omitempty"`
}

// TaskResponse represents a task returned from the API.
type TaskResponse struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Type           string   `json:"type"`
	Priority       string   `json:"priority"`
	Column         string   `json:"column"`
	Position       float64  `json:"position"`
	Labels         []string `json:"labels"`
	BlockedBy      []string `json:"blocked_by"`
	CreatedBy      string   `json:"created_by"`
	CreatedByAgent string   `json:"created_by_agent"`
	Epic           string   `json:"epic"`
	Board          string   `json:"board"`
	Seq            int      `json:"seq"`
	Created        string   `json:"created"`
	Updated        string   `json:"updated"`
}

// CreateTask creates a task via the HTTP API.
// Returns the created task or an error.
func (c *APIClient) CreateTask(data TaskData) (*TaskResponse, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task data: %w", err)
	}

	resp, err := c.requestClient.Post(
		c.baseURL+"/api/collections/tasks/records",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, body)
	}

	var task TaskResponse
	if err := json.Unmarshal(body, &task); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &task, nil
}

// UpdateTask updates a task via the HTTP API.
// Only non-zero fields in data will be updated.
func (c *APIClient) UpdateTask(id string, data TaskData) (*TaskResponse, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task data: %w", err)
	}

	req, err := http.NewRequest(
		"PATCH",
		c.baseURL+"/api/collections/tasks/records/"+id,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.requestClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, body)
	}

	var task TaskResponse
	if err := json.Unmarshal(body, &task); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &task, nil
}

// DeleteTask deletes a task via the HTTP API.
func (c *APIClient) DeleteTask(id string) error {
	req, err := http.NewRequest(
		"DELETE",
		c.baseURL+"/api/collections/tasks/records/"+id,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.requestClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return parseAPIError(resp.StatusCode, body)
	}

	return nil
}

// GetTask fetches a single task by ID via the HTTP API.
func (c *APIClient) GetTask(id string) (*TaskResponse, error) {
	resp, err := c.requestClient.Get(c.baseURL + "/api/collections/tasks/records/" + id)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, body)
	}

	var task TaskResponse
	if err := json.Unmarshal(body, &task); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &task, nil
}

// parseAPIError parses a PocketBase error response.
func parseAPIError(statusCode int, body []byte) *APIError {
	var apiErr APIError
	apiErr.StatusCode = statusCode

	if err := json.Unmarshal(body, &apiErr); err != nil {
		// If we can't parse the error, create a generic one
		apiErr.Message = string(body)
		if apiErr.Message == "" {
			apiErr.Message = fmt.Sprintf("HTTP %d", statusCode)
		}
	}

	return &apiErr
}

// IsAPIError checks if an error is an APIError and returns it.
func IsAPIError(err error) (*APIError, bool) {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr, true
	}
	return nil, false
}
