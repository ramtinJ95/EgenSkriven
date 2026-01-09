// Package resume provides functionality for resuming AI agent sessions.
package resume

import (
	"fmt"
	"strings"
)

// Tool constants for supported AI coding tools
const (
	ToolOpenCode   = "opencode"
	ToolClaudeCode = "claude-code"
	ToolCodex      = "codex"
)

// ValidTools is a list of all supported tool names
var ValidTools = []string{ToolOpenCode, ToolClaudeCode, ToolCodex}

// ResumeCommand holds the details needed to resume a session
type ResumeCommand struct {
	Tool       string   // The AI tool name (opencode, claude-code, codex)
	SessionRef string   // The session/thread ID
	WorkingDir string   // Directory where the command should be executed
	Prompt     string   // The context prompt to inject
	Command    string   // The full shell command to execute
	Args       []string // Parsed arguments for exec.Command
}

// BuildResumeCommand generates the resume command for a specific tool.
// Returns an error if the tool is not supported.
func BuildResumeCommand(tool, sessionRef, workingDir, prompt string) (*ResumeCommand, error) {
	rc := &ResumeCommand{
		Tool:       tool,
		SessionRef: sessionRef,
		WorkingDir: workingDir,
		Prompt:     prompt,
	}

	switch tool {
	case ToolOpenCode:
		rc.Command = buildOpenCodeCommand(sessionRef, prompt)
		rc.Args = []string{"opencode", "run", prompt, "--session", sessionRef}

	case ToolClaudeCode:
		rc.Command = buildClaudeCodeCommand(sessionRef, prompt)
		rc.Args = []string{"claude", "--resume", sessionRef, prompt}

	case ToolCodex:
		rc.Command = buildCodexCommand(sessionRef, prompt)
		rc.Args = []string{"codex", "exec", "resume", sessionRef, prompt}

	default:
		return nil, fmt.Errorf("unsupported tool: %s (supported: %v)", tool, ValidTools)
	}

	return rc, nil
}

// buildOpenCodeCommand generates the shell command for OpenCode resume
// Format: opencode run "<prompt>" --session <session-id>
func buildOpenCodeCommand(sessionRef, prompt string) string {
	escapedPrompt := ShellQuote(prompt)
	return fmt.Sprintf("opencode run %s --session %s", escapedPrompt, sessionRef)
}

// buildClaudeCodeCommand generates the shell command for Claude Code resume
// Format: claude --resume <session-id> "<prompt>"
func buildClaudeCodeCommand(sessionRef, prompt string) string {
	escapedPrompt := ShellQuote(prompt)
	return fmt.Sprintf("claude --resume %s %s", sessionRef, escapedPrompt)
}

// buildCodexCommand generates the shell command for Codex resume
// Format: codex exec resume <session-id> "<prompt>"
func buildCodexCommand(sessionRef, prompt string) string {
	escapedPrompt := ShellQuote(prompt)
	return fmt.Sprintf("codex exec resume %s %s", sessionRef, escapedPrompt)
}

// ValidateSessionRef checks if the session ref format is valid for the tool.
// Returns an error if the reference is invalid.
func ValidateSessionRef(tool, ref string) error {
	if ref == "" {
		return fmt.Errorf("session reference is empty")
	}

	// Basic validation - refs should be reasonably long
	switch tool {
	case ToolOpenCode, ToolClaudeCode, ToolCodex:
		// These typically use UUIDs or similar identifiers
		if len(ref) < 8 {
			return fmt.Errorf("session reference seems too short: %q (minimum 8 characters)", ref)
		}
	}

	return nil
}

// IsValidTool checks if the given tool name is supported
func IsValidTool(tool string) bool {
	for _, t := range ValidTools {
		if t == tool {
			return true
		}
	}
	return false
}

// ShellQuote wraps a string in single quotes for shell safety.
// This escapes any single quotes within the string using the '\” technique.
func ShellQuote(s string) string {
	// If the string is empty, return empty quoted string
	if s == "" {
		return "''"
	}

	// Replace single quotes with '\'' (end quote, escaped quote, start quote)
	escaped := strings.ReplaceAll(s, "'", "'\\''")
	return "'" + escaped + "'"
}
