package commands

import (
	"fmt"
	"strings"
	"time"
)

// parseDate converts user input to ISO 8601 date string.
// Supports:
// - ISO 8601: "2025-01-15"
// - Relative: "today", "tomorrow", "next week", "next month"
// - Shorthand: "jan 15", "january 15"
func parseDate(input string) (string, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return "", fmt.Errorf("empty date input")
	}

	now := time.Now()

	// Handle relative dates
	switch input {
	case "today":
		return now.Format("2006-01-02"), nil
	case "tomorrow":
		return now.AddDate(0, 0, 1).Format("2006-01-02"), nil
	case "next week":
		return now.AddDate(0, 0, 7).Format("2006-01-02"), nil
	case "next month":
		return now.AddDate(0, 1, 0).Format("2006-01-02"), nil
	}

	// Try ISO 8601 format (YYYY-MM-DD)
	if t, err := time.Parse("2006-01-02", input); err == nil {
		return t.Format("2006-01-02"), nil
	}

	// Try common formats
	formats := []string{
		"Jan 2",           // "Jan 15"
		"January 2",       // "January 15"
		"Jan 2, 2006",     // "Jan 15, 2025"
		"January 2, 2006", // "January 15, 2025"
		"1/2/2006",        // "1/15/2025"
		"1/2",             // "1/15"
	}

	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			// For formats without year, use current year
			// If date is in past, use next year
			if t.Year() == 0 {
				t = t.AddDate(now.Year(), 0, 0)
				if t.Before(now) {
					t = t.AddDate(1, 0, 0)
				}
			}
			return t.Format("2006-01-02"), nil
		}
	}

	return "", fmt.Errorf("could not parse date: %s", input)
}
