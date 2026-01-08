package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/resolver"
)

func newCommentsCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		since string
		limit int
	)

	cmd := &cobra.Command{
		Use:   "comments <task-ref>",
		Short: "List comments for a task",
		Long: `Display all comments for a task, sorted by creation time.

This is useful for checking if there are new responses to your questions,
or for reviewing the discussion history on a task.`,
		Example: `  # List all comments
  egenskriven comments WRK-123
  
  # List as JSON
  egenskriven comments WRK-123 --json
  
  # Only show comments from a specific time
  egenskriven comments WRK-123 --since "2026-01-07T10:00:00Z"
  
  # Show only the last 5 comments
  egenskriven comments WRK-123 --limit 5`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			taskRef := args[0]

			// Resolve task
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				if ambErr, ok := err.(*resolver.AmbiguousError); ok {
					return out.AmbiguousError(taskRef, ambErr.Matches)
				}
				return out.Error(ExitNotFound, err.Error(), nil)
			}

			// Build filter
			filter := fmt.Sprintf("task = '%s'", task.Id)

			if since != "" {
				// Parse and validate timestamp
				t, err := time.Parse(time.RFC3339, since)
				if err != nil {
					return out.Error(ExitValidation,
						fmt.Sprintf("invalid --since timestamp (use RFC3339 format like 2026-01-07T10:00:00Z): %v", err), nil)
				}
				filter += fmt.Sprintf(" && created > '%s'", t.Format(time.RFC3339))
			}

			// Query comments
			records, err := app.FindRecordsByFilter(
				"comments",
				filter,
				"+created", // Sort ascending (oldest first)
				limit,      // 0 means no limit
				0,          // No offset
			)
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to fetch comments: %v", err), nil)
			}

			// Get display ID for output
			displayId := getTaskDisplayID(app, task)

			// JSON output
			if jsonOutput {
				comments := make([]map[string]any, len(records))
				for i, r := range records {
					comments[i] = map[string]any{
						"id":          r.Id,
						"content":     r.GetString("content"),
						"author_type": r.GetString("author_type"),
						"author_id":   r.GetString("author_id"),
						"metadata":    r.Get("metadata"),
						"created":     r.GetDateTime("created").Time().Format(time.RFC3339),
					}
				}
				result := map[string]any{
					"task_id":    task.Id,
					"display_id": displayId,
					"count":      len(comments),
					"comments":   comments,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				encoder.Encode(result)
				return nil
			}

			// Human-readable output
			if len(records) == 0 {
				fmt.Printf("No comments on %s\n", displayId)
				return nil
			}

			fmt.Printf("Comments on %s (%d):\n\n", displayId, len(records))

			for _, r := range records {
				authorType := r.GetString("author_type")
				authorId := r.GetString("author_id")
				content := r.GetString("content")
				created := r.GetDateTime("created").Time()

				// Format author display
				var authorDisplay string
				if authorId != "" {
					authorDisplay = fmt.Sprintf("%s (%s)", authorId, authorType)
				} else {
					authorDisplay = authorType
				}

				// Format timestamp
				timeDisplay := formatRelativeTime(created)

				// Print comment
				fmt.Printf("[%s] %s:\n", timeDisplay, authorDisplay)

				// Indent content
				lines := strings.Split(content, "\n")
				for _, line := range lines {
					fmt.Printf("  %s\n", line)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&since, "since", "", "Only show comments after timestamp (RFC3339)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 0, "Limit number of comments (0 = all)")

	return cmd
}

// formatRelativeTime formats a time as relative (e.g., "2h ago") or absolute if > 24h
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		mins := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	} else {
		return t.Format("Jan 2, 15:04")
	}
}
