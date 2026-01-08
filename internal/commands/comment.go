package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/resolver"
)

func newCommentCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		useStdin bool
		author   string
	)

	cmd := &cobra.Command{
		Use:   "comment <task-ref> [text]",
		Short: "Add a comment to a task",
		Long: `Add a comment to a task. Use this to respond to agent questions
or to add notes to any task.

If the task is in the need_input state and you include @agent in your
comment, it may trigger an auto-resume (depending on board configuration).`,
		Example: `  # Add a simple comment
  egenskriven comment WRK-123 "Use JWT with refresh tokens"
  
  # Add comment that mentions the agent (may trigger auto-resume)
  egenskriven comment WRK-123 "@agent I've decided to use OAuth2"
  
  # Add comment from stdin for longer responses
  cat response.txt | egenskriven comment WRK-123 --stdin
  
  # Specify author
  egenskriven comment WRK-123 "Approved" --author "jane.doe"`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			taskRef := args[0]

			// Get comment text from args or stdin
			var text string
			if useStdin {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return out.Error(ExitGeneralError, fmt.Sprintf("failed to read from stdin: %v", err), nil)
				}
				text = strings.TrimSpace(string(data))
			} else if len(args) > 1 {
				text = args[1]
			} else {
				return out.Error(ExitInvalidArguments, "comment text is required: provide as argument or use --stdin", nil)
			}

			if text == "" {
				return out.Error(ExitValidation, "comment text cannot be empty", nil)
			}

			// Resolve task
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				if ambErr, ok := err.(*resolver.AmbiguousError); ok {
					return out.AmbiguousError(taskRef, ambErr.Matches)
				}
				return out.Error(ExitNotFound, err.Error(), nil)
			}

			// Determine author
			authorId := resolveAuthor(author)
			authorType := "human"

			// Check if running in agent context
			if isAgentContext() {
				authorType = "agent"
				if authorId == "" {
					authorId = getAgentNameFromEnv()
				}
			}

			// Extract mentions from text
			mentions := extractMentions(text)

			// Get comments collection
			commentsCollection, err := app.FindCollectionByNameOrId("comments")
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("comments collection not found: %v", err), nil)
			}

			// Create comment
			comment := core.NewRecord(commentsCollection)
			comment.Set("task", task.Id)
			comment.Set("content", text)
			comment.Set("author_type", authorType)
			comment.Set("author_id", authorId)
			comment.Set("metadata", map[string]any{
				"mentions": mentions,
			})

			if err := app.Save(comment); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to save comment: %v", err), nil)
			}

			// Get display ID for output
			displayId := getTaskDisplayID(app, task)

			// Output result
			if jsonOutput {
				result := map[string]any{
					"success":     true,
					"comment_id":  comment.Id,
					"task_id":     task.Id,
					"display_id":  displayId,
					"author_type": authorType,
					"author_id":   authorId,
					"mentions":    mentions,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(result); err != nil {
					return out.Error(ExitGeneralError, fmt.Sprintf("failed to encode JSON output: %v", err), nil)
				}
				return nil
			}

			out.Success(fmt.Sprintf("Comment added to %s", displayId))
			if !quietMode && len(mentions) > 0 {
				fmt.Printf("Mentions: %s\n", strings.Join(mentions, ", "))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&useStdin, "stdin", false, "Read comment from stdin")
	cmd.Flags().StringVarP(&author, "author", "a", "", "Author identifier")

	return cmd
}

// resolveAuthor returns the author identifier from various sources.
// Priority: --author flag > EGENSKRIVEN_AUTHOR env > USER env > empty
func resolveAuthor(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if author := os.Getenv("EGENSKRIVEN_AUTHOR"); author != "" {
		return author
	}
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	return ""
}

// isAgentContext checks if we're running in an AI agent context.
func isAgentContext() bool {
	// Check for common agent environment indicators
	indicators := []string{
		"OPENCODE_SESSION_ID",
		"CLAUDE_SESSION_ID",
		"CODEX_THREAD_ID",
		"EGENSKRIVEN_AGENT",
	}
	for _, env := range indicators {
		if os.Getenv(env) != "" {
			return true
		}
	}
	return false
}

// extractMentions finds @mentions in text.
func extractMentions(text string) []string {
	// Match @word patterns (but not email addresses)
	re := regexp.MustCompile(`(?:^|\s)@(\w+)`)
	matches := re.FindAllStringSubmatch(text, -1)

	seen := make(map[string]bool)
	var mentions []string

	for _, match := range matches {
		mention := "@" + match[1]
		if !seen[mention] {
			seen[mention] = true
			mentions = append(mentions, mention)
		}
	}

	return mentions
}
