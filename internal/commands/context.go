package commands

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

// ContextSummary holds project state summary.
type ContextSummary struct {
	Summary      Summary `json:"summary"`
	BlockedCount int     `json:"blocked_count"`
	ReadyCount   int     `json:"ready_count"`
}

// Summary holds task counts.
type Summary struct {
	Total      int            `json:"total"`
	ByColumn   map[string]int `json:"by_column"`
	ByPriority map[string]int `json:"by_priority"`
	ByType     map[string]int `json:"by_type"`
}

func newContextCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Output project state summary",
		Long: `Output a summary of the project's task state for agent context.

This provides agents with a quick overview of:
- Total task count
- Tasks by column/status
- Tasks by priority
- Number of blocked vs ready tasks

Examples:
  egenskriven context
  egenskriven context --json`,
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

			// Build summary
			summary := buildContextSummary(tasks)

			if jsonOutput {
				out.WriteJSON(summary)
			} else {
				printContextSummary(summary)
			}

			return nil
		},
	}

	return cmd
}

func buildContextSummary(tasks []*core.Record) ContextSummary {
	summary := ContextSummary{
		Summary: Summary{
			Total:      len(tasks),
			ByColumn:   make(map[string]int),
			ByPriority: make(map[string]int),
			ByType:     make(map[string]int),
		},
	}

	for _, task := range tasks {
		// Count by column
		col := task.GetString("column")
		summary.Summary.ByColumn[col]++

		// Count by priority
		priority := task.GetString("priority")
		summary.Summary.ByPriority[priority]++

		// Count by type
		taskType := task.GetString("type")
		summary.Summary.ByType[taskType]++

		// Count blocked
		blockedBy := getTaskBlockedBy(task)
		if len(blockedBy) > 0 {
			summary.BlockedCount++
		}

		// Count ready (unblocked in todo/backlog)
		if (col == "todo" || col == "backlog") && len(blockedBy) == 0 {
			summary.ReadyCount++
		}
	}

	return summary
}

func printContextSummary(s ContextSummary) {
	fmt.Printf("Project Summary\n")
	fmt.Printf("===============\n\n")

	fmt.Printf("Total tasks: %d\n", s.Summary.Total)
	fmt.Printf("Ready to work: %d\n", s.ReadyCount)
	fmt.Printf("Blocked: %d\n", s.BlockedCount)

	fmt.Printf("\nBy Column:\n")
	for _, col := range ValidColumns {
		count := s.Summary.ByColumn[col]
		if count > 0 {
			fmt.Printf("  %-12s %d\n", col+":", count)
		}
	}

	fmt.Printf("\nBy Priority:\n")
	for _, p := range ValidPriorities {
		count := s.Summary.ByPriority[p]
		if count > 0 {
			fmt.Printf("  %-12s %d\n", p+":", count)
		}
	}

	fmt.Printf("\nBy Type:\n")
	for _, t := range ValidTypes {
		count := s.Summary.ByType[t]
		if count > 0 {
			fmt.Printf("  %-12s %d\n", t+":", count)
		}
	}

	fmt.Println()
}
