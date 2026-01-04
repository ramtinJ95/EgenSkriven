package output

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockExit captures the exit code instead of terminating the process.
// Used in tests to verify exit codes.
func mockExit(t *testing.T) (getCode func() int, restore func()) {
	var capturedCode int
	originalExit := exitFunc
	exitFunc = func(code int) {
		capturedCode = code
	}
	return func() int { return capturedCode }, func() { exitFunc = originalExit }
}

func TestFormatter_JSON_Mode(t *testing.T) {
	f := New(true, false)
	assert.True(t, f.JSON)
	assert.False(t, f.Quiet)
}

func TestFormatter_Quiet_Mode(t *testing.T) {
	f := New(false, true)
	assert.False(t, f.JSON)
	assert.True(t, f.Quiet)
}

func TestShortID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abc123def456", "abc123de"},
		{"short", "short"},
		{"12345678", "12345678"},
		{"123456789", "12345678"},
	}

	for _, tt := range tests {
		result := ShortID(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestFormatter_Error_Output(t *testing.T) {
	// Mock exit to prevent process termination
	getCode, restore := mockExit(t)
	defer restore()

	// Capture stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f := New(true, false)
	_ = f.Error(1, "test error", nil)

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result map[string]any
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	errorObj := result["error"].(map[string]any)
	assert.Equal(t, float64(1), errorObj["code"])
	assert.Equal(t, "test error", errorObj["message"])
	assert.Equal(t, 1, getCode(), "exit code should be 1")
}

func TestFormatter_Error_WithData(t *testing.T) {
	// Mock exit to prevent process termination
	getCode, restore := mockExit(t)
	defer restore()

	// Capture stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f := New(true, false)
	_ = f.Error(5, "validation error", map[string]any{"field": "title"})

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result map[string]any
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	errorObj := result["error"].(map[string]any)
	assert.Equal(t, float64(5), errorObj["code"])
	assert.Equal(t, "validation error", errorObj["message"])
	assert.NotNil(t, errorObj["data"])
	assert.Equal(t, 5, getCode(), "exit code should be 5")
}

func TestFormatter_Error_ExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		message  string
		wantCode int
	}{
		{"general error", 1, "general error", 1},
		{"not found", 3, "task not found", 3},
		{"ambiguous", 4, "ambiguous reference", 4},
		{"validation", 5, "invalid input", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getCode, restore := mockExit(t)
			defer restore()

			// Suppress stderr output
			old := os.Stderr
			_, w, _ := os.Pipe()
			os.Stderr = w

			f := New(false, false)
			f.Error(tt.code, tt.message, nil)

			w.Close()
			os.Stderr = old

			assert.Equal(t, tt.wantCode, getCode())
		})
	}
}
