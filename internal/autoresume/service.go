// Package autoresume provides functionality for automatically resuming AI agent sessions.
//
// Auto-resume triggers when ALL conditions are met:
//  1. Task is in 'need_input' column
//  2. Task has a linked agent_session
//  3. Board's resume_mode is set to 'auto'
//  4. Comment contains '@agent' mention
//  5. Comment is from a human (not from agent)
package autoresume

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/resume"
)

// Service handles auto-resume logic for AI agent sessions.
type Service struct {
	app *pocketbase.PocketBase
}

// NewService creates a new auto-resume service.
func NewService(app *pocketbase.PocketBase) *Service {
	return &Service{app: app}
}

// CheckAndResume evaluates whether a comment should trigger auto-resume.
// This is the main entry point called after a comment is created.
// Returns nil if auto-resume was triggered or skipped (not an error case).
// Returns an error only if something unexpected failed.
func (s *Service) CheckAndResume(comment *core.Record) error {
	// 1. Check if comment is from human
	if comment.GetString("author_type") != "human" {
		return nil // Agent comments don't trigger
	}

	// 2. Check for @agent mention
	if !hasAgentMention(comment) {
		return nil // No @agent mention
	}

	// 3. Get the task
	taskId := comment.GetString("task")
	task, err := s.app.FindRecordById("tasks", taskId)
	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	// 4. Check task is in need_input
	if task.GetString("column") != "need_input" {
		return nil // Task not blocked
	}

	// 5. Check task has session
	sessionData := task.Get("agent_session")
	if sessionData == nil {
		return nil // No session linked
	}
	// JSON fields return types.JSONRaw even when null, check for actual value
	sessionStr := fmt.Sprintf("%s", sessionData)
	if sessionStr == "" || sessionStr == "null" || sessionStr == "{}" {
		return nil // No session linked
	}

	// 6. Check board resume mode
	boardId := task.GetString("board")
	if boardId == "" {
		return nil // No board linked
	}

	board, err := s.app.FindRecordById("boards", boardId)
	if err != nil {
		return fmt.Errorf("failed to find board: %w", err)
	}

	resumeMode := board.GetString("resume_mode")
	if resumeMode != "auto" {
		return nil // Auto-resume not enabled
	}

	// All conditions met - trigger resume
	return s.triggerResume(task, comment)
}

// triggerResume executes the resume flow.
func (s *Service) triggerResume(task *core.Record, triggerComment *core.Record) error {
	sessionMap, ok := task.Get("agent_session").(map[string]any)
	if !ok {
		return fmt.Errorf("invalid agent_session format")
	}

	tool, _ := sessionMap["tool"].(string)
	ref, _ := sessionMap["ref"].(string)
	workingDir, _ := sessionMap["working_dir"].(string)

	if tool == "" || ref == "" {
		return fmt.Errorf("invalid agent_session: missing tool or ref")
	}

	if workingDir == "" {
		workingDir = "."
	}

	// Fetch all comments for context
	comments, err := s.fetchComments(task.Id)
	if err != nil {
		return fmt.Errorf("failed to fetch comments: %w", err)
	}

	// Get display ID for context
	displayId := s.getTaskDisplayID(task)

	// Build context prompt
	prompt := resume.BuildContextPrompt(task, displayId, comments)

	// Build resume command
	resumeCmd, err := resume.BuildResumeCommand(tool, ref, workingDir, prompt)
	if err != nil {
		return fmt.Errorf("failed to build resume command: %w", err)
	}

	// Update task state
	if err := s.updateTaskForResume(task, triggerComment); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Execute resume in background
	go s.executeResume(resumeCmd, task.Id)

	return nil
}

// updateTaskForResume updates task state before resume.
func (s *Service) updateTaskForResume(task *core.Record, triggerComment *core.Record) error {
	task.Set("column", "in_progress")

	// Add history entry
	history := ensureHistorySlice(task.Get("history"))
	history = append(history, map[string]any{
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"action":       "auto_resumed",
		"actor":        "system",
		"actor_detail": "auto-resume",
		"changes": map[string]any{
			"column": map[string]any{
				"from": "need_input",
				"to":   "in_progress",
			},
		},
		"metadata": map[string]any{
			"trigger_comment": triggerComment.Id,
		},
	})
	task.Set("history", history)

	return s.app.Save(task)
}

// executeResume runs the resume command in background.
func (s *Service) executeResume(resumeCmd *resume.ResumeCommand, taskId string) {
	// Log start
	s.logAutoResume(taskId, "started", "")

	// Execute command
	if len(resumeCmd.Args) == 0 {
		s.logAutoResume(taskId, "failed", "empty command args")
		return
	}

	cmd := exec.Command(resumeCmd.Args[0], resumeCmd.Args[1:]...)
	cmd.Dir = resumeCmd.WorkingDir

	err := cmd.Run()

	if err != nil {
		s.logAutoResume(taskId, "failed", err.Error())
	} else {
		s.logAutoResume(taskId, "completed", "")
	}
}

// fetchComments gets all comments for a task.
func (s *Service) fetchComments(taskId string) ([]resume.Comment, error) {
	records, err := s.app.FindRecordsByFilter(
		"comments",
		fmt.Sprintf("task = '%s'", taskId),
		"+created",
		100,
		0,
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

// getTaskDisplayID returns the display ID (e.g., "WRK-123") for a task.
func (s *Service) getTaskDisplayID(task *core.Record) string {
	boardID := task.GetString("board")
	seq := task.GetInt("seq")

	if boardID != "" && seq > 0 {
		boardRecord, err := s.app.FindRecordById("boards", boardID)
		if err == nil {
			prefix := boardRecord.GetString("prefix")
			return fmt.Sprintf("%s-%d", prefix, seq)
		}
	}

	// Fallback to short ID
	id := task.Id
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// logAutoResume logs auto-resume events.
func (s *Service) logAutoResume(taskId, status, errorMsg string) {
	if errorMsg != "" {
		log.Printf("[auto-resume] task=%s status=%s error=%s", taskId, status, errorMsg)
	} else {
		log.Printf("[auto-resume] task=%s status=%s", taskId, status)
	}
}

// hasAgentMention checks if comment mentions @agent.
func hasAgentMention(comment *core.Record) bool {
	metadata := comment.Get("metadata")
	if metadata == nil {
		return false
	}

	// Handle both map[string]any (in-memory) and types.JSONRaw (from DB)
	var metaMap map[string]any

	switch v := metadata.(type) {
	case map[string]any:
		metaMap = v
	default:
		// Try to parse as JSON (handles types.JSONRaw and other stringifiable types)
		raw := []byte(fmt.Sprintf("%s", metadata))
		if err := json.Unmarshal(raw, &metaMap); err != nil {
			return false
		}
	}

	mentions, ok := metaMap["mentions"].([]any)
	if !ok {
		return false
	}

	for _, m := range mentions {
		if mention, ok := m.(string); ok && mention == "@agent" {
			return true
		}
	}
	return false
}

// ensureHistorySlice converts history to proper type.
func ensureHistorySlice(history any) []map[string]any {
	if history == nil {
		return []map[string]any{}
	}
	if slice, ok := history.([]any); ok {
		result := make([]map[string]any, 0, len(slice))
		for _, item := range slice {
			if m, ok := item.(map[string]any); ok {
				result = append(result, m)
			}
		}
		return result
	}
	return []map[string]any{}
}
