package output

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		result := shortID(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestFormatter_Error_Output(t *testing.T) {
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
}

func TestFormatter_Error_WithData(t *testing.T) {
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
}
