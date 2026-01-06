package commands

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDate_ISO8601(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2025-01-15", "2025-01-15"},
		{"2025-12-31", "2025-12-31"},
		{"2024-02-29", "2024-02-29"}, // Leap year
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseDate(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseDate_Relative(t *testing.T) {
	now := time.Now()

	tests := []struct {
		input    string
		expected string
	}{
		{"today", now.Format("2006-01-02")},
		{"tomorrow", now.AddDate(0, 0, 1).Format("2006-01-02")},
		{"next week", now.AddDate(0, 0, 7).Format("2006-01-02")},
		{"next month", now.AddDate(0, 1, 0).Format("2006-01-02")},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseDate(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseDate_CaseInsensitive(t *testing.T) {
	tests := []string{"TODAY", "Today", "toDay", "TOMORROW", "Tomorrow"}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseDate(input)
			require.NoError(t, err, "should parse %q", input)
		})
	}
}

func TestParseDate_CommonFormats(t *testing.T) {
	tests := []struct {
		input       string
		shouldParse bool
	}{
		{"Jan 15", true},
		{"January 15", true},
		{"Jan 15, 2025", true},
		{"1/15/2025", true},
		{"1/15", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseDate(tc.input)
			if tc.shouldParse {
				require.NoError(t, err)
				assert.NotEmpty(t, result)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestParseDate_InvalidInput(t *testing.T) {
	tests := []string{
		"",
		"invalid",
		"yesterday", // Not supported
		"someday",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseDate(input)
			assert.Error(t, err, "should fail to parse %q", input)
		})
	}
}

func TestParseDate_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"  today  "},
		{"\ttomorrow\t"},
		{" 2025-01-15 "},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			_, err := parseDate(tc.input)
			require.NoError(t, err, "should handle whitespace in %q", tc.input)
		})
	}
}
