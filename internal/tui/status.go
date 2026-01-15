package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// StatusIndicator displays the realtime connection status.
type StatusIndicator struct {
	status  ConnectionStatus
	message string // Optional status message
}

// NewStatusIndicator creates a new status indicator.
func NewStatusIndicator() *StatusIndicator {
	return &StatusIndicator{
		status: ConnectionDisconnected,
	}
}

// SetStatus updates the connection status.
func (s *StatusIndicator) SetStatus(status ConnectionStatus) {
	s.status = status
	s.message = ""
}

// SetStatusWithMessage updates the status with an optional message.
func (s *StatusIndicator) SetStatusWithMessage(status ConnectionStatus, message string) {
	s.status = status
	s.message = message
}

// Status returns the current connection status.
func (s *StatusIndicator) Status() ConnectionStatus {
	return s.status
}

// View renders the status indicator.
func (s *StatusIndicator) View() string {
	var indicator string
	var style lipgloss.Style

	switch s.status {
	case ConnectionConnected:
		// Green dot for connected
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("82")) // Green
		indicator = "●"
	case ConnectionConnecting:
		// Yellow dot for connecting
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
		indicator = "◐"
	case ConnectionReconnecting:
		// Orange dot for reconnecting
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // Orange
		indicator = "◐"
	case ConnectionDisconnected:
		// Gray dot for disconnected
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Gray
		indicator = "○"
	default:
		// Unknown status - fallback to gray question mark
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		indicator = "?"
	}

	result := style.Render(indicator)

	if s.message != "" {
		msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
		result += " " + msgStyle.Render(s.message)
	}

	return result
}

// ViewWithLabel renders the status indicator with a label.
func (s *StatusIndicator) ViewWithLabel() string {
	var label string

	switch s.status {
	case ConnectionConnected:
		label = "Live"
	case ConnectionConnecting:
		label = "Connecting..."
	case ConnectionReconnecting:
		label = "Reconnecting..."
	case ConnectionDisconnected:
		label = "Offline"
	default:
		label = "Unknown"
	}

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	return fmt.Sprintf("%s %s", s.View(), labelStyle.Render(label))
}

// StatusBar represents the full status bar at the bottom of the TUI.
type StatusBar struct {
	status      *StatusIndicator
	boardName   string
	taskCount   int
	filterCount int
	width       int
}

// NewStatusBar creates a new status bar.
func NewStatusBar() *StatusBar {
	return &StatusBar{
		status: NewStatusIndicator(),
	}
}

// SetWidth sets the width of the status bar.
func (b *StatusBar) SetWidth(width int) {
	b.width = width
}

// SetBoardName sets the current board name.
func (b *StatusBar) SetBoardName(name string) {
	b.boardName = name
}

// SetTaskCount sets the total task count.
func (b *StatusBar) SetTaskCount(count int) {
	b.taskCount = count
}

// SetFilterCount sets the number of active filters.
func (b *StatusBar) SetFilterCount(count int) {
	b.filterCount = count
}

// SetConnectionStatus updates the connection status.
func (b *StatusBar) SetConnectionStatus(status ConnectionStatus) {
	b.status.SetStatus(status)
}

// SetConnectionStatusWithMessage updates the connection status with a message.
func (b *StatusBar) SetConnectionStatusWithMessage(status ConnectionStatus, message string) {
	b.status.SetStatusWithMessage(status, message)
}

// View renders the status bar.
func (b *StatusBar) View() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Width(b.width).
		Padding(0, 1)

	// Left side: connection status and board name
	left := b.status.ViewWithLabel()
	if b.boardName != "" {
		left += " | " + b.boardName
	}

	// Right side: task count and filters
	var right string
	if b.taskCount > 0 {
		right = fmt.Sprintf("%d tasks", b.taskCount)
	}
	if b.filterCount > 0 {
		if right != "" {
			right += " | "
		}
		right += fmt.Sprintf("%d filters", b.filterCount)
	}

	// Calculate spacing
	padding := b.width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if padding < 1 {
		padding = 1
	}
	spacer := lipgloss.NewStyle().Width(padding).Render("")

	return style.Render(left + spacer + right)
}
