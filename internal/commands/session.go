package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/output"
	"github.com/ramtinJ95/EgenSkriven/internal/resolver"
)

// validSessionTools are the supported AI agent tools
var validSessionTools = []string{"opencode", "claude-code", "codex"}

func newSessionCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage agent sessions linked to tasks",
		Long: `Commands for linking and viewing AI agent sessions on tasks.

Sessions allow tracking which AI tool (OpenCode, Claude Code, Codex) is 
working on a task, enabling context-preserving resume when a task is blocked.`,
	}

	// Add subcommands
	cmd.AddCommand(newSessionLinkCmd(app))
	cmd.AddCommand(newSessionShowCmd(app))
	cmd.AddCommand(newSessionHistoryCmd(app))
	cmd.AddCommand(newSessionUnlinkCmd(app))

	return cmd
}

// ========== Session Link ==========

func newSessionLinkCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		tool       string
		ref        string
		workingDir string
	)

	cmd := &cobra.Command{
		Use:   "link <task-ref>",
		Short: "Link an agent session to a task",
		Long: `Link an AI agent session to a task for tracking and resume support.

If a session is already linked to the task, it will be marked as "abandoned"
and moved to history before linking the new session.

The session reference is typically a UUID (for OpenCode/Claude Code/Codex) 
or a file path (for some tools).`,
		Example: `  # Link an OpenCode session
  egenskriven session link WRK-123 --tool opencode --ref abc-123-def
  
  # Link a Claude Code session with explicit working directory
  egenskriven session link WRK-123 --tool claude-code --ref 550e8400-... --working-dir /home/user/project
  
  # Link and output JSON
  egenskriven session link WRK-123 -t codex -r xyz-789 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			taskRef := args[0]

			// Validate tool
			if !containsString(validSessionTools, tool) {
				return out.Error(ExitValidation,
					fmt.Sprintf("invalid tool %q: must be one of %v", tool, validSessionTools), nil)
			}

			// Validate ref
			if ref == "" {
				return out.Error(ExitInvalidArguments, "--ref is required", nil)
			}

			// Resolve working directory
			if workingDir == "" {
				var err error
				workingDir, err = os.Getwd()
				if err != nil {
					return out.Error(ExitGeneralError, fmt.Sprintf("failed to get working directory: %v", err), nil)
				}
			}
			// Make absolute
			workingDir, _ = filepath.Abs(workingDir)

			// Determine ref type
			refType := determineRefType(ref)

			// Resolve task
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				if ambErr, ok := err.(*resolver.AmbiguousError); ok {
					return out.AmbiguousError(taskRef, ambErr.Matches)
				}
				return out.Error(ExitNotFound, err.Error(), nil)
			}

			displayId := getTaskDisplayID(app, task)
			now := time.Now()

			// Get sessions collection
			sessionsCollection, err := app.FindCollectionByNameOrId("sessions")
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("sessions collection not found: %v", err), nil)
			}

			// Execute in transaction
			var sessionId string
			err = app.RunInTransaction(func(txApp core.App) error {
				// Check for existing session
				existingSessionData := task.Get("agent_session")
				if existingSessionData != nil && existingSessionData != "" {
					oldSession, _ := output.ParseAgentSession(existingSessionData)
					if oldSession != nil {
						// Mark old session as abandoned in history
						if oldRef, ok := oldSession["ref"].(string); ok {
							if err := markSessionStatus(txApp, task.Id, oldRef, "abandoned", now); err != nil {
								// Log but don't fail - old session status is non-critical
								warnLog("could not mark old session as abandoned: %v", err)
							}
						}
					}
				}

				// Create new agent_session on task
				newSession := map[string]any{
					"tool":        tool,
					"ref":         ref,
					"ref_type":    refType,
					"working_dir": workingDir,
					"linked_at":   now.Format(time.RFC3339),
				}
				task.Set("agent_session", newSession)

				// Add to history
				addHistoryEntry(task, "session_linked", "agent", map[string]any{
					"tool":        tool,
					"session_ref": ref,
				})

				if err := txApp.Save(task); err != nil {
					return fmt.Errorf("failed to update task: %w", err)
				}

				// Create session record in sessions table
				sessionRecord := core.NewRecord(sessionsCollection)
				sessionRecord.Set("task", task.Id)
				sessionRecord.Set("tool", tool)
				sessionRecord.Set("external_ref", ref)
				sessionRecord.Set("ref_type", refType)
				sessionRecord.Set("working_dir", workingDir)
				sessionRecord.Set("status", "active")

				if err := txApp.Save(sessionRecord); err != nil {
					return fmt.Errorf("failed to create session record: %w", err)
				}

				sessionId = sessionRecord.Id
				return nil
			})

			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to link session: %v", err), nil)
			}

			// Output
			if jsonOutput {
				result := map[string]any{
					"success":     true,
					"task_id":     task.Id,
					"display_id":  displayId,
					"session_id":  sessionId,
					"tool":        tool,
					"ref":         ref,
					"ref_type":    refType,
					"working_dir": workingDir,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(result)
			}

			fmt.Printf("Session linked to %s\n", displayId)
			fmt.Printf("  Tool: %s\n", tool)
			fmt.Printf("  Ref:  %s\n", output.TruncateMiddle(ref, 40))
			return nil
		},
	}

	cmd.Flags().StringVarP(&tool, "tool", "t", "", "Tool name (opencode, claude-code, codex)")
	cmd.Flags().StringVarP(&ref, "ref", "r", "", "Session/thread reference")
	cmd.Flags().StringVarP(&workingDir, "working-dir", "w", "", "Working directory (defaults to current)")

	cmd.MarkFlagRequired("tool")
	cmd.MarkFlagRequired("ref")

	return cmd
}

// ========== Session Show ==========

func newSessionShowCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <task-ref>",
		Short: "Show the current session linked to a task",
		Long: `Display information about the AI agent session currently linked to a task.

This shows which tool is working on the task, the session reference,
and when it was linked.`,
		Example: `  egenskriven session show WRK-123
  egenskriven session show abc --json`,
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
			sessionData := task.Get("agent_session")

			// Handle nil or empty session
			if sessionData == nil || sessionData == "" {
				if jsonOutput {
					result := map[string]any{
						"task_id":     task.Id,
						"display_id":  displayId,
						"has_session": false,
						"session":     nil,
					}
					encoder := json.NewEncoder(os.Stdout)
					encoder.SetIndent("", "  ")
					return encoder.Encode(result)
				}
				fmt.Printf("No session linked to %s\n", displayId)
				return nil
			}

			// Parse session data - handle both direct map and JSON string
			session, err := output.ParseAgentSession(sessionData)
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("invalid session data: %v", err), nil)
			}
			if session == nil {
				if jsonOutput {
					result := map[string]any{
						"task_id":     task.Id,
						"display_id":  displayId,
						"has_session": false,
						"session":     nil,
					}
					encoder := json.NewEncoder(os.Stdout)
					encoder.SetIndent("", "  ")
					return encoder.Encode(result)
				}
				fmt.Printf("No session linked to %s\n", displayId)
				return nil
			}

			if jsonOutput {
				result := map[string]any{
					"task_id":     task.Id,
					"display_id":  displayId,
					"has_session": true,
					"session":     session,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(result)
			}

			// Human-readable output
			fmt.Printf("Session for %s:\n\n", displayId)
			fmt.Printf("  Tool:        %s\n", session["tool"])
			fmt.Printf("  Reference:   %s\n", session["ref"])
			fmt.Printf("  Type:        %s\n", session["ref_type"])
			fmt.Printf("  Working Dir: %s\n", session["working_dir"])

			if linkedAt, ok := session["linked_at"].(string); ok {
				t, parseErr := time.Parse(time.RFC3339, linkedAt)
				if parseErr == nil {
					fmt.Printf("  Linked:      %s\n", formatRelativeTime(t))
				}
			}

			// Show resume hint if task is in need_input
			if task.GetString("column") == "need_input" {
				fmt.Printf("\n  Tip: Run 'egenskriven resume %s' to resume this session\n", displayId)
			}

			return nil
		},
	}

	return cmd
}

// ========== Session History ==========

func newSessionHistoryCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history <task-ref>",
		Short: "Show session history for a task",
		Long: `Display all AI agent sessions that have been linked to a task,
including their status (active, abandoned, completed).

This helps track which tools have worked on a task over time.`,
		Example: `  egenskriven session history WRK-123
  egenskriven session history abc --json`,
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

			// Fetch all sessions for this task
			records, err := app.FindAllRecords("sessions",
				dbx.NewExp("task = {:taskId}", dbx.Params{"taskId": task.Id}),
			)
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to fetch sessions: %v", err), nil)
			}

			// Sort by created descending (most recent first)
			// PocketBase returns in creation order, so reverse
			for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
				records[i], records[j] = records[j], records[i]
			}

			if jsonOutput {
				sessions := make([]map[string]any, len(records))
				for i, r := range records {
					sessions[i] = map[string]any{
						"id":           r.Id,
						"tool":         r.GetString("tool"),
						"external_ref": r.GetString("external_ref"),
						"ref_type":     r.GetString("ref_type"),
						"working_dir":  r.GetString("working_dir"),
						"status":       r.GetString("status"),
						"created":      r.GetDateTime("created").Time().Format(time.RFC3339),
						"ended_at":     formatSessionEndedAt(r),
					}
				}
				result := map[string]any{
					"task_id":    task.Id,
					"display_id": displayId,
					"count":      len(sessions),
					"sessions":   sessions,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(result)
			}

			// Human-readable output
			if len(records) == 0 {
				fmt.Printf("No session history for %s\n", displayId)
				return nil
			}

			fmt.Printf("Session history for %s (%d sessions):\n\n", displayId, len(records))

			for i, r := range records {
				status := r.GetString("status")
				tool := r.GetString("tool")
				ref := r.GetString("external_ref")
				created := r.GetDateTime("created").Time()

				// Status indicator
				statusIcon := getSessionStatusIcon(status)

				fmt.Printf("%d. %s %s (%s)\n", i+1, statusIcon, tool, status)
				fmt.Printf("   Ref: %s\n", output.TruncateMiddle(ref, 50))
				fmt.Printf("   Started: %s\n", formatRelativeTime(created))

				if status != "active" {
					endedAt := r.GetDateTime("ended_at")
					if !endedAt.IsZero() {
						fmt.Printf("   Ended: %s\n", formatRelativeTime(endedAt.Time()))
					}
				}
				fmt.Println()
			}

			return nil
		},
	}

	return cmd
}

// ========== Session Unlink ==========

func newSessionUnlinkCmd(app *pocketbase.PocketBase) *cobra.Command {
	var status string

	cmd := &cobra.Command{
		Use:   "unlink <task-ref>",
		Short: "Unlink the current session from a task",
		Long: `Remove the current session link from a task, optionally marking it
with a final status (abandoned or completed).

Use this when you want to disconnect a session without linking a new one.`,
		Example: `  # Mark session as abandoned (default)
  egenskriven session unlink WRK-123
  
  # Mark session as completed
  egenskriven session unlink WRK-123 --status completed`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			taskRef := args[0]

			// Validate status
			validStatuses := []string{"abandoned", "completed"}
			if !containsString(validStatuses, status) {
				return out.Error(ExitValidation,
					fmt.Sprintf("invalid status %q: must be 'abandoned' or 'completed'", status), nil)
			}

			// Resolve task
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				if ambErr, ok := err.(*resolver.AmbiguousError); ok {
					return out.AmbiguousError(taskRef, ambErr.Matches)
				}
				return out.Error(ExitNotFound, err.Error(), nil)
			}

			displayId := getTaskDisplayID(app, task)
			sessionData := task.Get("agent_session")

			if sessionData == nil || sessionData == "" {
				return out.Error(ExitValidation, fmt.Sprintf("no session linked to %s", displayId), nil)
			}

			session, err := output.ParseAgentSession(sessionData)
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("invalid session data: %v", err), nil)
			}
			if session == nil {
				return out.Error(ExitValidation, fmt.Sprintf("no session linked to %s", displayId), nil)
			}

			now := time.Now()

			err = app.RunInTransaction(func(txApp core.App) error {
				// Update session in history table
				if ref, ok := session["ref"].(string); ok {
					if err := markSessionStatus(txApp, task.Id, ref, status, now); err != nil {
						// Log but don't fail - session status update is non-critical
						warnLog("could not update session status: %v", err)
					}
				}

				// Clear agent_session on task
				task.Set("agent_session", nil)

				// Add history entry
				addHistoryEntry(task, "session_unlinked", "user", map[string]any{
					"tool":         session["tool"],
					"session_ref":  session["ref"],
					"final_status": status,
				})

				return txApp.Save(task)
			})

			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to unlink session: %v", err), nil)
			}

			if jsonOutput {
				result := map[string]any{
					"success":      true,
					"task_id":      task.Id,
					"display_id":   displayId,
					"final_status": status,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(result)
			}

			fmt.Printf("Session unlinked from %s (marked as %s)\n", displayId, status)
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "abandoned", "Final status (abandoned, completed)")

	return cmd
}

// ========== Helper Functions ==========

// determineRefType guesses if ref is a UUID or path
func determineRefType(ref string) string {
	// If it starts with / or . or contains path separators, treat as path
	if strings.HasPrefix(ref, "/") || strings.HasPrefix(ref, ".") ||
		strings.Contains(ref, "/") || strings.Contains(ref, "\\") {
		return "path"
	}
	return "uuid"
}

// markSessionStatus updates a session's status in the history table.
// Returns an error if the save fails, allowing callers to decide how to handle it.
func markSessionStatus(app core.App, taskId, externalRef, status string, endTime time.Time) error {
	// Find the session record
	records, err := app.FindRecordsByFilter(
		"sessions",
		"task = {:taskId} && external_ref = {:ref} && status = 'active'",
		"-created",
		1,
		0,
		dbx.Params{"taskId": taskId, "ref": externalRef},
	)
	if err != nil || len(records) == 0 {
		// Session not found in history, that's okay
		return nil
	}

	record := records[0]
	record.Set("status", status)
	record.Set("ended_at", endTime)

	if err := app.Save(record); err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}
	return nil
}

// containsString checks if slice contains string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getSessionStatusIcon returns a status indicator for display
func getSessionStatusIcon(status string) string {
	switch status {
	case "active":
		return "[ACTIVE]"
	case "paused":
		return "[PAUSED]"
	case "completed":
		return "[DONE]  "
	case "abandoned":
		return "[OLD]   "
	default:
		return "[?]     "
	}
}

// formatSessionEndedAt returns the ended_at time as RFC3339 string or nil
func formatSessionEndedAt(r *core.Record) *string {
	endedAt := r.GetDateTime("ended_at")
	if endedAt.IsZero() {
		return nil
	}
	s := endedAt.Time().Format(time.RFC3339)
	return &s
}
