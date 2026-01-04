package commands

import (
	"bufio"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBatchInput_JSONLines(t *testing.T) {
	input := `{"title":"Task 1","type":"bug"}
{"title":"Task 2","priority":"high"}
{"title":"Task 3","column":"todo"}`

	var inputs []TaskInput

	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var ti TaskInput
		err := json.Unmarshal([]byte(line), &ti)
		require.NoError(t, err)
		inputs = append(inputs, ti)
	}
	require.NoError(t, scanner.Err())

	assert.Len(t, inputs, 3)
	assert.Equal(t, "Task 1", inputs[0].Title)
	assert.Equal(t, "bug", inputs[0].Type)
	assert.Equal(t, "Task 2", inputs[1].Title)
	assert.Equal(t, "high", inputs[1].Priority)
	assert.Equal(t, "Task 3", inputs[2].Title)
	assert.Equal(t, "todo", inputs[2].Column)
}

func TestParseBatchInput_JSONLines_WithEmptyLines(t *testing.T) {
	input := `{"title":"Task 1"}

{"title":"Task 2"}

`

	var inputs []TaskInput

	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var ti TaskInput
		err := json.Unmarshal([]byte(line), &ti)
		require.NoError(t, err)
		inputs = append(inputs, ti)
	}
	require.NoError(t, scanner.Err())

	assert.Len(t, inputs, 2)
	assert.Equal(t, "Task 1", inputs[0].Title)
	assert.Equal(t, "Task 2", inputs[1].Title)
}

func TestParseBatchInput_JSONArray(t *testing.T) {
	input := `[
		{"title":"Task 1"},
		{"title":"Task 2","labels":["frontend","ui"]}
	]`

	var inputs []TaskInput
	err := json.Unmarshal([]byte(input), &inputs)
	require.NoError(t, err)

	assert.Len(t, inputs, 2)
	assert.Equal(t, "Task 1", inputs[0].Title)
	assert.Equal(t, "Task 2", inputs[1].Title)
	assert.Len(t, inputs[1].Labels, 2)
	assert.Contains(t, inputs[1].Labels, "frontend")
	assert.Contains(t, inputs[1].Labels, "ui")
}

func TestParseBatchInput_JSONArray_AllFields(t *testing.T) {
	input := `[
		{
			"id": "custom-id-001",
			"title": "Full Task",
			"description": "A complete task",
			"type": "bug",
			"priority": "urgent",
			"column": "todo",
			"labels": ["critical", "backend"],
			"epic": "Q1 Launch"
		}
	]`

	var inputs []TaskInput
	err := json.Unmarshal([]byte(input), &inputs)
	require.NoError(t, err)

	assert.Len(t, inputs, 1)
	task := inputs[0]
	assert.Equal(t, "custom-id-001", task.ID)
	assert.Equal(t, "Full Task", task.Title)
	assert.Equal(t, "A complete task", task.Description)
	assert.Equal(t, "bug", task.Type)
	assert.Equal(t, "urgent", task.Priority)
	assert.Equal(t, "todo", task.Column)
	assert.Equal(t, []string{"critical", "backend"}, task.Labels)
	assert.Equal(t, "Q1 Launch", task.Epic)
}

func TestParseBatchInput_DetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		isArray  bool
		expected int
	}{
		{
			name:     "JSON array",
			input:    `[{"title":"Task 1"},{"title":"Task 2"}]`,
			isArray:  true,
			expected: 2,
		},
		{
			name:     "JSON lines",
			input:    `{"title":"Task 1"}` + "\n" + `{"title":"Task 2"}`,
			isArray:  false,
			expected: 2,
		},
		{
			name:     "JSON array with whitespace prefix",
			input:    `  [{"title":"Task 1"}]`,
			isArray:  true,
			expected: 1,
		},
		{
			name:     "Single JSON object (lines format)",
			input:    `{"title":"Task 1"}`,
			isArray:  false,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trimmed := strings.TrimSpace(tt.input)
			var inputs []TaskInput

			if strings.HasPrefix(trimmed, "[") {
				// JSON array format
				assert.True(t, tt.isArray, "expected JSON array format")
				err := json.Unmarshal([]byte(trimmed), &inputs)
				require.NoError(t, err)
			} else {
				// JSON lines format
				assert.False(t, tt.isArray, "expected JSON lines format")
				scanner := bufio.NewScanner(strings.NewReader(trimmed))
				for scanner.Scan() {
					line := strings.TrimSpace(scanner.Text())
					if line == "" {
						continue
					}
					var ti TaskInput
					err := json.Unmarshal([]byte(line), &ti)
					require.NoError(t, err)
					inputs = append(inputs, ti)
				}
				require.NoError(t, scanner.Err())
			}

			assert.Len(t, inputs, tt.expected)
		})
	}
}

func TestDefaultString(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		defaultVal string
		expected   string
	}{
		{
			name:       "empty value returns default",
			value:      "",
			defaultVal: "default",
			expected:   "default",
		},
		{
			name:       "non-empty value returns value",
			value:      "value",
			defaultVal: "default",
			expected:   "value",
		},
		{
			name:       "whitespace value returns whitespace",
			value:      "  ",
			defaultVal: "default",
			expected:   "  ",
		},
		{
			name:       "empty default with empty value",
			value:      "",
			defaultVal: "",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultString(tt.value, tt.defaultVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTaskInput_JSONUnmarshal_OptionalFields(t *testing.T) {
	// Test that optional fields can be omitted
	input := `{"title":"Minimal Task"}`

	var ti TaskInput
	err := json.Unmarshal([]byte(input), &ti)
	require.NoError(t, err)

	assert.Equal(t, "Minimal Task", ti.Title)
	assert.Empty(t, ti.ID)
	assert.Empty(t, ti.Description)
	assert.Empty(t, ti.Type)
	assert.Empty(t, ti.Priority)
	assert.Empty(t, ti.Column)
	assert.Nil(t, ti.Labels)
	assert.Empty(t, ti.Epic)
}

func TestTaskInput_JSONUnmarshal_EmptyLabels(t *testing.T) {
	input := `{"title":"Task","labels":[]}`

	var ti TaskInput
	err := json.Unmarshal([]byte(input), &ti)
	require.NoError(t, err)

	assert.Equal(t, "Task", ti.Title)
	assert.NotNil(t, ti.Labels)
	assert.Empty(t, ti.Labels)
}
