package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SessionState holds state that persists during a TUI session.
// This allows filters and UI state to survive across app restarts
// within the same system session (temp directory is cleared on reboot).
type SessionState struct {
	FilterState    *FilterState `json:"filterState"`
	CurrentBoardID string       `json:"currentBoardID"`
	FocusedColumn  int          `json:"focusedColumn"`
}

// sessionFilePath returns the path to the session state file.
// Stored in temp directory so it's cleared on system reboot.
func sessionFilePath() string {
	return filepath.Join(os.TempDir(), "egenskriven-tui-session.json")
}

// SaveSession saves current session state to a file.
// Called when the TUI exits to preserve filter state.
func SaveSession(state SessionState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sessionFilePath(), data, 0644)
}

// LoadSession loads session state if available.
// Returns nil with no error if no session file exists.
func LoadSession() (*SessionState, error) {
	data, err := os.ReadFile(sessionFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No session to load
		}
		return nil, err
	}

	var state SessionState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// ClearSession removes saved session state.
// Can be called to start fresh.
func ClearSession() error {
	path := sessionFilePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // Nothing to clear
	}
	return os.Remove(path)
}

// SessionLoadedMsg is sent when session state is loaded.
type SessionLoadedMsg struct {
	Session *SessionState
}
