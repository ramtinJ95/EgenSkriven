package commands

import (
	"fmt"
	"sort"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

// Suggestion represents a task suggestion with reasoning.
type Suggestion struct {
	Task   map[string]any `json:"task"`
	Reason string         `json:"reason"`
}

// SuggestResponse holds the list of suggestions.
type SuggestResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
}

func newSuggestCmd(app *pocketbase.PocketBase) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "suggest",
		Short: "Suggest tasks to work on",
		Long: `Suggest tasks to work on next based on priority and dependencies.

Suggestion priority:
1. In-progress tasks (continue current work)
2. Urgent unblocked tasks
3. High priority unblocked tasks
4. Tasks that unblock the most other tasks

Examples:
  egenskriven suggest
  egenskriven suggest --json
  egenskriven suggest --json --limit 3`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Get all tasks
			tasks, err := app.FindAllRecords("tasks")
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to list tasks: %v", err), nil)
			}

			// Build suggestions
			suggestions := buildSuggestions(tasks, limit)

			if jsonOutput {
				out.WriteJSON(SuggestResponse{Suggestions: suggestions})
			} else {
				printSuggestions(suggestions)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 5, "Maximum number of suggestions")

	return cmd
}

func buildSuggestions(tasks []*core.Record, limit int) []Suggestion {
	var suggestions []Suggestion

	// Calculate how many tasks each task unblocks
	unblocksCount := make(map[string]int)
	for _, t := range tasks {
		blockedBy := getTaskBlockedBy(t)
		for _, blockingID := range blockedBy {
			unblocksCount[blockingID]++
		}
	}

	// Track which task IDs are already added
	addedIDs := make(map[string]bool)

	// Helper to add suggestion if not already added
	addSuggestion := func(task *core.Record, reason string) {
		if !addedIDs[task.Id] {
			suggestions = append(suggestions, Suggestion{
				Task:   taskToSuggestionMap(task),
				Reason: reason,
			})
			addedIDs[task.Id] = true
		}
	}

	// 1. In-progress tasks (continue current work)
	for _, t := range tasks {
		if t.GetString("column") == "in_progress" {
			addSuggestion(t, "Continue current work")
		}
	}

	// 2. Urgent unblocked tasks
	for _, t := range tasks {
		col := t.GetString("column")
		if col != "in_progress" && col != "done" && col != "review" &&
			t.GetString("priority") == "urgent" &&
			len(getTaskBlockedBy(t)) == 0 {
			addSuggestion(t, "Urgent priority, unblocked")
		}
	}

	// 3. High priority unblocked tasks
	for _, t := range tasks {
		col := t.GetString("column")
		if col != "in_progress" && col != "done" && col != "review" &&
			t.GetString("priority") == "high" &&
			len(getTaskBlockedBy(t)) == 0 {
			addSuggestion(t, "High priority, unblocked")
		}
	}

	// 4. Tasks that unblock others
	type unblockingTask struct {
		task  *core.Record
		count int
	}
	var unblocking []unblockingTask
	for _, t := range tasks {
		col := t.GetString("column")
		if col != "done" && col != "review" {
			count := unblocksCount[t.Id]
			if count > 0 && len(getTaskBlockedBy(t)) == 0 {
				unblocking = append(unblocking, unblockingTask{t, count})
			}
		}
	}

	// Sort by count descending
	sort.Slice(unblocking, func(i, j int) bool {
		return unblocking[i].count > unblocking[j].count
	})

	for _, ut := range unblocking {
		addSuggestion(ut.task, fmt.Sprintf("Unblocks %d other task(s)", ut.count))
	}

	// Limit results
	if limit > 0 && len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return suggestions
}

func taskToSuggestionMap(task *core.Record) map[string]any {
	return map[string]any{
		"id":       task.Id,
		"title":    task.GetString("title"),
		"type":     task.GetString("type"),
		"priority": task.GetString("priority"),
		"column":   task.GetString("column"),
	}
}

func printSuggestions(suggestions []Suggestion) {
	if len(suggestions) == 0 {
		fmt.Println("No suggestions - all caught up!")
		return
	}

	fmt.Println("Suggested tasks to work on:")
	fmt.Println()

	for i, s := range suggestions {
		id := s.Task["id"].(string)
		fmt.Printf("%d. [%s] %s\n", i+1, shortID(id), s.Task["title"])
		fmt.Printf("   Type: %s, Priority: %s, Column: %s\n",
			s.Task["type"], s.Task["priority"], s.Task["column"])
		fmt.Printf("   Reason: %s\n", s.Reason)
		fmt.Println()
	}
}
