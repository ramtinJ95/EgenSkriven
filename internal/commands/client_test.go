package commands

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAPIClient_DefaultURL(t *testing.T) {
	client := NewAPIClient()
	assert.NotNil(t, client)
	// Default URL should be localhost:8090
	assert.Equal(t, DefaultServerURL, client.baseURL)
}

func TestNewAPIClientWithURL(t *testing.T) {
	customURL := "http://custom:9090"
	client := NewAPIClientWithURL(customURL)
	assert.NotNil(t, client)
	assert.Equal(t, customURL, client.baseURL)
}

func TestAPIClient_IsServerRunning_WhenRunning(t *testing.T) {
	// Create a test server that responds to /api/health
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":200,"message":"API is healthy."}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewAPIClientWithURL(server.URL)
	assert.True(t, client.IsServerRunning())
}

func TestAPIClient_IsServerRunning_WhenNotRunning(t *testing.T) {
	// Use an invalid URL that won't connect
	client := NewAPIClientWithURL("http://localhost:59999")
	assert.False(t, client.IsServerRunning())
}

func TestAPIError_IsValidationError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   bool
	}{
		{"400 Bad Request", 400, true},
		{"401 Unauthorized", 401, true},
		{"403 Forbidden", 403, true},
		{"404 Not Found", 404, true},
		{"422 Unprocessable Entity", 422, true},
		{"500 Internal Server Error", 500, false},
		{"502 Bad Gateway", 502, false},
		{"503 Service Unavailable", 503, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &APIError{StatusCode: tt.statusCode}
			assert.Equal(t, tt.expected, err.IsValidationError())
		})
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		StatusCode: 400,
		Message:    "Invalid data",
	}
	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "Invalid data")
}

func TestIsAPIError(t *testing.T) {
	apiErr := &APIError{StatusCode: 400, Message: "test"}

	// Test with APIError
	result, ok := IsAPIError(apiErr)
	assert.True(t, ok)
	assert.Equal(t, apiErr, result)

	// Test with non-APIError
	result, ok = IsAPIError(assert.AnError)
	assert.False(t, ok)
	assert.Nil(t, result)
}

func TestParseAPIError(t *testing.T) {
	// Test with valid JSON error response
	body := []byte(`{"status":400,"message":"Validation failed","data":{"title":"required"}}`)
	err := parseAPIError(400, body)

	assert.Equal(t, 400, err.StatusCode)
	assert.Equal(t, "Validation failed", err.Message)
	assert.NotNil(t, err.Data)

	// Test with invalid JSON
	body = []byte(`invalid json`)
	err = parseAPIError(500, body)

	assert.Equal(t, 500, err.StatusCode)
	assert.Equal(t, "invalid json", err.Message)

	// Test with empty body
	body = []byte(``)
	err = parseAPIError(500, body)

	assert.Equal(t, 500, err.StatusCode)
	assert.Equal(t, "HTTP 500", err.Message)
}

func TestAPIClient_CreateTask_Success(t *testing.T) {
	// Create a test server that handles task creation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/api/collections/tasks/records" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "test123",
				"title": "Test Task",
				"type": "feature",
				"priority": "medium",
				"column": "backlog",
				"position": 1000
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewAPIClientWithURL(server.URL)

	task, err := client.CreateTask(TaskData{
		Title:    "Test Task",
		Type:     "feature",
		Priority: "medium",
		Column:   "backlog",
	})

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "test123", task.ID)
	assert.Equal(t, "Test Task", task.Title)
}

func TestAPIClient_CreateTask_ValidationError(t *testing.T) {
	// Create a test server that returns a validation error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":400,"message":"Validation failed","data":{"title":"required"}}`))
	}))
	defer server.Close()

	client := NewAPIClientWithURL(server.URL)

	task, err := client.CreateTask(TaskData{})

	assert.Error(t, err)
	assert.Nil(t, task)

	apiErr, ok := IsAPIError(err)
	assert.True(t, ok)
	assert.True(t, apiErr.IsValidationError())
}

func TestAPIClient_UpdateTask_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PATCH" && r.URL.Path == "/api/collections/tasks/records/test123" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "test123",
				"title": "Updated Task",
				"priority": "urgent"
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewAPIClientWithURL(server.URL)

	task, err := client.UpdateTask("test123", TaskData{
		Title:    "Updated Task",
		Priority: "urgent",
	})

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "Updated Task", task.Title)
}

func TestAPIClient_DeleteTask_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/api/collections/tasks/records/test123" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewAPIClientWithURL(server.URL)

	err := client.DeleteTask("test123")
	assert.NoError(t, err)
}

func TestAPIClient_DeleteTask_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"status":404,"message":"Record not found"}`))
	}))
	defer server.Close()

	client := NewAPIClientWithURL(server.URL)

	err := client.DeleteTask("nonexistent")
	assert.Error(t, err)

	apiErr, ok := IsAPIError(err)
	assert.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestAPIClient_GetTask_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/api/collections/tasks/records/test123" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "test123",
				"title": "Test Task",
				"type": "feature",
				"priority": "medium",
				"column": "backlog"
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewAPIClientWithURL(server.URL)

	task, err := client.GetTask("test123")

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "test123", task.ID)
	assert.Equal(t, "Test Task", task.Title)
}
