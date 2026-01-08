package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/resolver"
)

func newBlockCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		useStdin  bool
		agentName string
	)

	cmd := &cobra.Command{
		Use:   "block <task-ref> [question]",
		Short: "Block a task and request human input",
		Long: `Move a task to the need_input column and add a comment with your question.

This is an atomic operation that ensures both the task state change and 
the comment are created together. If either fails, neither is applied.

Use this when you're blocked and need human guidance to proceed.`,
		Example: `  # Block with inline question
  egenskriven block WRK-123 "What authentication approach should I use?"
  
  # Block with question from stdin (for longer questions)
  echo "I need to decide between several options..." | egenskriven block WRK-123 --stdin
  
  # Block with JSON output
  egenskriven block abc "Should I use REST or GraphQL?" --json
  
  # Block with explicit agent name
  egenskriven block WRK-123 "Question?" --agent "opencode-build"`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			taskRef := args[0]

			// Get question from args or stdin
			var question string
			if useStdin {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return out.Error(ExitGeneralError, fmt.Sprintf("failed to read from stdin: %v", err), nil)
				}
				question = strings.TrimSpace(string(data))
			} else if len(args) > 1 {
				question = args[1]
			} else {
				return out.Error(ExitInvalidArguments, "question is required: provide as argument or use --stdin", nil)
			}

			if question == "" {
				return out.Error(ExitValidation, "question cannot be empty", nil)
			}

			// Resolve the task
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				if ambErr, ok := err.(*resolver.AmbiguousError); ok {
					return out.AmbiguousError(taskRef, ambErr.Matches)
				}
				return out.Error(ExitNotFound, err.Error(), nil)
			}

			// Validate current state - can't block if already in need_input or done
			currentColumn := task.GetString("column")
			if currentColumn == "need_input" {
				return out.Error(ExitValidation,
					fmt.Sprintf("task %s is already blocked (in need_input)", shortID(task.Id)), nil)
			}
			if currentColumn == "done" {
				return out.Error(ExitValidation, "cannot block a completed task", nil)
			}

			// Determine agent name
			if agentName == "" {
				agentName = getAgentNameFromEnv()
			}

			// Get comments collection
			commentsCollection, err := app.FindCollectionByNameOrId("comments")
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("comments collection not found: %v", err), nil)
			}

			// Execute in transaction for atomicity
			var commentId string
			err = app.RunInTransaction(func(txApp core.App) error {
				// 1. Update task column to need_input
				task.Set("column", "need_input")

				// 2. Add history entry
				addHistoryEntry(task, "blocked", agentName, map[string]any{
					"column": map[string]any{
						"from": currentColumn,
						"to":   "need_input",
					},
					"reason": question,
				})

				// Save task
				if err := txApp.Save(task); err != nil {
					return fmt.Errorf("failed to update task: %w", err)
				}

				// 3. Create comment
				comment := core.NewRecord(commentsCollection)
				comment.Set("task", task.Id)
				comment.Set("content", question)
				comment.Set("author_type", "agent")
				comment.Set("author_id", agentName)
				comment.Set("metadata", map[string]any{
					"action": "block_question",
				})

				if err := txApp.Save(comment); err != nil {
					return fmt.Errorf("failed to create comment: %w", err)
				}

				commentId = comment.Id
				return nil
			})

			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to block task: %v", err), nil)
			}

			// Get display ID for output
			displayId := getTaskDisplayID(app, task)

			// Output result
			if jsonOutput {
				result := map[string]any{
					"success":    true,
					"task_id":    task.Id,
					"display_id": displayId,
					"column":     "need_input",
					"comment_id": commentId,
					"message":    fmt.Sprintf("Task %s blocked, awaiting human input", displayId),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				encoder.Encode(result)
				return nil
			}

			out.Success(fmt.Sprintf("Task %s blocked. Awaiting human input.", displayId))
			if !quietMode {
				fmt.Printf("Question: %s\n", truncateString(question, 100))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&useStdin, "stdin", false, "Read question from stdin")
	cmd.Flags().StringVarP(&agentName, "agent", "a", "", "Agent identifier")

	return cmd
}

// getAgentNameFromEnv returns the agent name from environment variables.
// Priority: EGENSKRIVEN_AGENT env > "agent" default
func getAgentNameFromEnv() string {
	if name := os.Getenv("EGENSKRIVEN_AGENT"); name != "" {
		return name
	}
	return "agent"
}

// getTaskDisplayID returns the display ID (e.g., "WRK-123") for a task.
func getTaskDisplayID(app *pocketbase.PocketBase, task *core.Record) string {
	boardID := task.GetString("board")
	seq := task.GetInt("seq")

	if boardID != "" && seq > 0 {
		boardRecord, err := app.FindRecordById("boards", boardID)
		if err == nil {
			prefix := boardRecord.GetString("prefix")
			return fmt.Sprintf("%s-%d", prefix, seq)
		}
	}

	// Fallback to short ID
	return shortID(task.Id)
}

// truncateString truncates a string to maxLen characters, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
