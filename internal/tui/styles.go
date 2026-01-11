package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette - using ANSI 256 colors for terminal compatibility.
// These colors work well on both light and dark terminal backgrounds.
var (
	// Brand colors
	primaryColor   = lipgloss.Color("62")  // Blue - used for focused elements
	secondaryColor = lipgloss.Color("205") // Pink - used for accents

	// Priority colors - visual indicators for task urgency
	priorityUrgent = lipgloss.Color("196") // Red
	priorityHigh   = lipgloss.Color("208") // Orange
	priorityMedium = lipgloss.Color("226") // Yellow
	priorityLow    = lipgloss.Color("240") // Gray

	// Type colors - distinguish task types visually
	typeBug     = lipgloss.Color("196") // Red - bugs stand out
	typeFeature = lipgloss.Color("39")  // Cyan - features are positive
	typeChore   = lipgloss.Color("240") // Gray - chores are neutral

	// Column header colors - each column has a distinct color
	columnBacklog    = lipgloss.Color("240") // Gray
	columnTodo       = lipgloss.Color("39")  // Cyan
	columnInProgress = lipgloss.Color("214") // Orange
	columnNeedInput  = lipgloss.Color("196") // Red (needs attention)
	columnReview     = lipgloss.Color("205") // Pink
	columnDone       = lipgloss.Color("82")  // Green

	// UI element colors
	borderColor      = lipgloss.Color("240") // Gray border
	focusBorderColor = lipgloss.Color("62")  // Blue border for focused
	mutedColor       = lipgloss.Color("240") // Gray for secondary text
	textColor        = lipgloss.Color("252") // Light gray for text

	// Status colors
	successColor = lipgloss.Color("82")  // Green
	warningColor = lipgloss.Color("214") // Orange
	errorColor   = lipgloss.Color("196") // Red
)

// Column styles - different border colors for focused vs unfocused columns.
var (
	// focusedColumnStyle is applied to the currently selected column.
	// It has a highlighted border and slightly different background.
	focusedColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(focusBorderColor).
				Padding(0, 1)

	// blurredColumnStyle is applied to non-selected columns.
	// Uses a dimmer border color.
	blurredColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderColor).
				Padding(0, 1)
)

// Header and title styles.
var (
	// headerStyle is used for the main application header.
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	// boardTitleStyle is used for the board name in the header.
	boardTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(secondaryColor)

	// columnTitleStyle is the base style for column headers.
	// The foreground color is set dynamically based on column type.
	columnTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1)
)

// Status bar styles.
var (
	// statusBarStyle is the background style for the status bar.
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(textColor).
			Padding(0, 1)

	// statusErrorStyle is used for error messages in the status bar.
	statusErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)

	// statusSuccessStyle is used for success messages.
	statusSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("82"))
)

// Task item styles.
var (
	// selectedTaskStyle highlights the currently selected task.
	selectedTaskStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Background(lipgloss.Color("236")).
				Bold(true)

	// normalTaskStyle is used for non-selected tasks.
	normalTaskStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// blockedIndicatorStyle styles the [BLOCKED] indicator.
	blockedIndicatorStyle = lipgloss.NewStyle().
				Foreground(priorityUrgent).
				Bold(true)
)

// Panel styles - used for modal/overlay panels.
var (
	// taskDetailStyle is used for the task detail side panel.
	taskDetailStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// modalStyle is used for modal dialogs like forms and confirmations.
	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// formStyle is used for task add/edit forms.
	formStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)
)

// Button styles - used for form buttons.
var (
	// buttonStyle is the default button style.
	buttonStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(lipgloss.Color("236")).
			Padding(0, 2)

	// buttonFocusedStyle is used when a button is focused.
	buttonFocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(primaryColor).
				Bold(true).
				Padding(0, 2)

	// buttonDangerStyle is used for destructive action buttons when focused.
	buttonDangerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Background(errorColor).
				Bold(true).
				Padding(0, 2)
)

// Dialog styles - used for confirmation and other dialogs.
var (
	// confirmDialogStyle is used for confirmation dialogs.
	confirmDialogStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(warningColor).
				Padding(1, 2).
				Align(lipgloss.Center)
)

// GetColumnHeaderColor returns the appropriate color for a column header.
// Each column status has a distinct color for quick visual identification.
func GetColumnHeaderColor(status string) lipgloss.Color {
	switch status {
	case "backlog":
		return columnBacklog
	case "todo":
		return columnTodo
	case "in_progress":
		return columnInProgress
	case "need_input":
		return columnNeedInput
	case "review":
		return columnReview
	case "done":
		return columnDone
	default:
		return mutedColor
	}
}

// GetPriorityIndicator returns a styled priority indicator string.
// Higher priority = more prominent indicator.
func GetPriorityIndicator(priority string) string {
	var color lipgloss.Color
	var indicator string

	switch priority {
	case "urgent":
		color = priorityUrgent
		indicator = "!!!"
	case "high":
		color = priorityHigh
		indicator = "!!"
	case "medium":
		color = priorityMedium
		indicator = "!"
	default: // low
		color = priorityLow
		indicator = ""
	}

	if indicator == "" {
		return ""
	}

	return lipgloss.NewStyle().Foreground(color).Render(indicator)
}

// GetTypeIndicator returns a styled type badge.
func GetTypeIndicator(taskType string) string {
	var color lipgloss.Color

	switch taskType {
	case "bug":
		color = typeBug
	case "feature":
		color = typeFeature
	case "chore":
		color = typeChore
	default:
		color = mutedColor
	}

	return lipgloss.NewStyle().Foreground(color).Render("[" + taskType + "]")
}
