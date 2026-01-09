package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/output"
	"github.com/ramtinJ95/EgenSkriven/internal/resolver"
	"github.com/ramtinJ95/EgenSkriven/internal/resume"
)

func newResumeCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		execFlag     bool
		minimal      bool
		customPrompt string
		dryRun       bool
	)

	cmd := &cobra.Command{
		Use:   "resume <task-ref>",
		Short: "Resume work on a blocked task",
		Long: `Generate or execute the command to resume an AI agent session
for a task that is in the need_input state.

By default, this prints the resume command that you can copy and run.
Use --exec to execute the command directly.

The resume command includes context from the task and comment thread,
which is injected into the agent's session.`,
		Example: `  # Print the resume command
  egenskriven resume WRK-123
  
  # Execute the resume directly
  egenskriven resume WRK-123 --exec
  
  # Dry run - show what would be executed
  egenskriven resume WRK-123 --exec --dry-run
  
  # Use minimal prompt (fewer tokens)
  egenskriven resume WRK-123 --minimal
  
  # Output as JSON (for scripting)
  egenskriven resume WRK-123 --json
  
  # Custom prompt override
  egenskriven resume WRK-123 --prompt "Continue with the JWT implementation"`,
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

			displayId := getTaskDisplayID(app, task)

			// Validate task state
			column := task.GetString("column")
			if column != "need_input" {
				return out.Error(ExitValidation,
					fmt.Sprintf("task %s is not in need_input state (current: %s)", displayId, column), nil)
			}

			// Get session info
			sessionData := task.Get("agent_session")
			if sessionData == nil || sessionData == "" {
				return out.Error(ExitValidation,
					fmt.Sprintf("no agent session linked to task %s\n\nTo resume, first link a session:\n  egenskriven session link %s --tool <tool> --ref <session-id>",
						displayId, displayId), nil)
			}

			session, err := output.ParseAgentSession(sessionData)
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("invalid session data: %v", err), nil)
			}
			if session == nil {
				return out.Error(ExitValidation,
					fmt.Sprintf("no agent session linked to task %s\n\nTo resume, first link a session:\n  egenskriven session link %s --tool <tool> --ref <session-id>",
						displayId, displayId), nil)
			}

			tool, _ := session["tool"].(string)
			sessionRef, _ := session["ref"].(string)
			workingDir, _ := session["working_dir"].(string)

			// Validate session ref
			if err := resume.ValidateSessionRef(tool, sessionRef); err != nil {
				return out.Error(ExitValidation, fmt.Sprintf("invalid session: %v", err), nil)
			}

			// Fetch comments
			comments, err := fetchCommentsForResume(app, task.Id)
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to fetch comments: %v", err), nil)
			}

			// Build context prompt
			var prompt string
			if customPrompt != "" {
				prompt = customPrompt
			} else if minimal {
				prompt = resume.BuildMinimalPrompt(task, comments)
			} else {
				prompt = resume.BuildContextPrompt(task, comments)
			}

			// Build resume command
			resumeCmd, err := resume.BuildResumeCommand(tool, sessionRef, workingDir, prompt)
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to build resume command: %v", err), nil)
			}

			// JSON output
			if jsonOutput {
				result := map[string]any{
					"task_id":       task.Id,
					"display_id":    displayId,
					"tool":          tool,
					"session_ref":   sessionRef,
					"working_dir":   workingDir,
					"command":       resumeCmd.Command,
					"prompt":        prompt,
					"prompt_length": len(prompt),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(result)
			}

			// Execute mode
			if execFlag {
				if dryRun {
					fmt.Printf("Would execute in %s:\n\n", workingDir)
					fmt.Printf("  %s\n\n", resumeCmd.Command)
					fmt.Printf("Prompt (%d chars):\n%s\n", len(prompt), indentText(prompt, "  "))
					return nil
				}

				// Update task status BEFORE executing
				if err := updateTaskForResume(app, task); err != nil {
					return out.Error(ExitGeneralError, fmt.Sprintf("failed to update task: %v", err), nil)
				}

				// Update session status in history
				updateSessionStatusInHistory(app, task.Id, sessionRef, "active")

				fmt.Printf("Resuming session for %s...\n", displayId)
				fmt.Printf("Tool: %s\n", tool)
				fmt.Printf("Working directory: %s\n\n", workingDir)

				return executeResumeCommand(resumeCmd)
			}

			// Print mode (default)
			fmt.Printf("Resume command for %s:\n\n", displayId)
			fmt.Printf("  %s\n\n", resumeCmd.Command)
			fmt.Printf("Working directory: %s\n", workingDir)
			fmt.Printf("Prompt length: %d characters\n\n", len(prompt))
			fmt.Printf("To execute directly, run:\n")
			fmt.Printf("  egenskriven resume %s --exec\n", displayId)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&execFlag, "exec", "e", false, "Execute the resume command")
	cmd.Flags().BoolVarP(&minimal, "minimal", "m", false, "Use minimal prompt (fewer tokens)")
	cmd.Flags().StringVarP(&customPrompt, "prompt", "p", "", "Custom prompt override")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show command without executing (use with --exec)")

	return cmd
}

// fetchCommentsForResume gets comments formatted for the resume context.
func fetchCommentsForResume(app *pocketbase.PocketBase, taskId string) ([]resume.Comment, error) {
	records, err := app.FindRecordsByFilter(
		"comments",
		"task = {:taskId}",
		"+created", // Oldest first for chronological order
		100,        // Reasonable limit
		0,
		dbx.Params{"taskId": taskId},
	)
	if err != nil {
		return nil, err
	}

	comments := make([]resume.Comment, len(records))
	for i, r := range records {
		comments[i] = resume.Comment{
			Content:    r.GetString("content"),
			AuthorType: r.GetString("author_type"),
			AuthorId:   r.GetString("author_id"),
			Created:    r.GetDateTime("created").Time(),
		}
	}

	return comments, nil
}

// updateTaskForResume moves task to in_progress and adds history.
func updateTaskForResume(app *pocketbase.PocketBase, task *core.Record) error {
	task.Set("column", "in_progress")

	// Add history entry
	addHistoryEntry(task, "resumed", "", map[string]any{
		"column": map[string]any{
			"from": "need_input",
			"to":   "in_progress",
		},
	})

	return app.Save(task)
}

// executeResumeCommand runs the resume command.
func executeResumeCommand(rc *resume.ResumeCommand) error {
	// Change to working directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if rc.WorkingDir != "" {
		if err := os.Chdir(rc.WorkingDir); err != nil {
			return fmt.Errorf("failed to change to working directory %s: %w", rc.WorkingDir, err)
		}
		defer os.Chdir(originalDir)
	}

	// Execute command
	// We use the first arg as the command and rest as arguments
	if len(rc.Args) == 0 {
		return fmt.Errorf("no command arguments provided")
	}

	execCmd := exec.Command(rc.Args[0], rc.Args[1:]...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

// updateSessionStatusInHistory updates the session record status.
func updateSessionStatusInHistory(app *pocketbase.PocketBase, taskId, externalRef, status string) {
	records, err := app.FindRecordsByFilter(
		"sessions",
		"task = {:taskId} && external_ref = {:ref}",
		"-created",
		1,
		0,
		dbx.Params{"taskId": taskId, "ref": externalRef},
	)
	if err != nil || len(records) == 0 {
		return
	}

	record := records[0]
	record.Set("status", status)
	if err := app.Save(record); err != nil {
		log.Printf("warning: failed to update session status: %v", err)
	}
}

// indentText adds prefix to each line of text.
func indentText(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}
