// Package resume provides functionality for resuming AI agent sessions.
package resume

import (
	"fmt"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// Comment represents a comment record for building context
type Comment struct {
	Content    string
	AuthorType string
	AuthorId   string
	Created    time.Time
}

// BuildContextPrompt creates the full context prompt for resume.
// It includes task information and the complete conversation thread.
// The displayId should be pre-computed using the proper display ID logic
// (e.g., getTaskDisplayID from commands package).
func BuildContextPrompt(task *core.Record, displayId string, comments []Comment) string {
	var sb strings.Builder

	title := task.GetString("title")
	priority := task.GetString("priority")
	description := task.GetString("description")

	// Header
	sb.WriteString("## Task Context (from EgenSkriven)\n\n")

	// Task info
	sb.WriteString(fmt.Sprintf("**Task**: %s - %s\n", displayId, title))
	sb.WriteString("**Status**: need_input -> in_progress\n")
	sb.WriteString(fmt.Sprintf("**Priority**: %s\n", priority))

	// Include description if present (truncated if >500 chars)
	if description != "" {
		sb.WriteString("**Description**:\n")
		if len(description) > 500 {
			sb.WriteString(description[:500] + "...\n")
		} else {
			sb.WriteString(description + "\n")
		}
	}
	sb.WriteString("\n")

	// Comments thread
	sb.WriteString("## Conversation Thread\n\n")

	if len(comments) == 0 {
		sb.WriteString("_No comments yet_\n\n")
	} else {
		for _, c := range comments {
			authorLabel := formatAuthorLabel(c.AuthorType, c.AuthorId)
			timeLabel := c.Created.Format("15:04")

			sb.WriteString(fmt.Sprintf("[%s @ %s]: %s\n\n",
				authorLabel, timeLabel, c.Content))
		}
	}

	// Instructions
	sb.WriteString("## Instructions\n\n")
	sb.WriteString("Continue working on the task based on the human's response above. ")
	sb.WriteString("The conversation context should help you understand what was discussed. ")
	sb.WriteString("If you need more clarification, you can block the task again with a new question.\n")

	return sb.String()
}

// BuildMinimalPrompt creates a shorter prompt for token-constrained scenarios.
// It only includes the last 3 comments and truncates long comments at 200 chars.
// The displayId should be pre-computed using the proper display ID logic
// (e.g., getTaskDisplayID from commands package).
func BuildMinimalPrompt(task *core.Record, displayId string, comments []Comment) string {
	var sb strings.Builder

	title := task.GetString("title")

	sb.WriteString(fmt.Sprintf("Task %s: %s\n\n", displayId, title))
	sb.WriteString("Recent comments:\n")

	if len(comments) == 0 {
		sb.WriteString("_No comments yet_\n")
	} else {
		// Only include last 3 comments for minimal version
		start := 0
		if len(comments) > 3 {
			start = len(comments) - 3
		}

		for _, c := range comments[start:] {
			authorLabel := formatAuthorLabel(c.AuthorType, c.AuthorId)
			// Truncate long comments at 200 chars
			content := c.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			sb.WriteString(fmt.Sprintf("- %s: %s\n", authorLabel, content))
		}
	}

	sb.WriteString("\nContinue based on the above context.\n")

	return sb.String()
}

// formatAuthorLabel returns the author identifier for display.
// It prefers authorId if present, otherwise falls back to authorType.
func formatAuthorLabel(authorType, authorId string) string {
	if authorId != "" {
		return authorId
	}
	return authorType
}
