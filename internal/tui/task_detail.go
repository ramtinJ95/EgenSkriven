package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// TaskDetail displays full task information in a side panel
type TaskDetail struct {
	task     TaskItem
	viewport viewport.Model
	width    int
	height   int
	ready    bool
	keys     taskDetailKeyMap
}

type taskDetailKeyMap struct {
	Close key.Binding
	Edit  key.Binding
	Up    key.Binding
	Down  key.Binding
}

func defaultTaskDetailKeyMap() taskDetailKeyMap {
	return taskDetailKeyMap{
		Close: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "close"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k/up", "scroll up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j/down", "scroll down"),
		),
	}
}

// NewTaskDetail creates a new task detail panel
func NewTaskDetail(task TaskItem, width, height int) *TaskDetail {
	vp := viewport.New(width-4, height-6) // Account for borders and header
	vp.Style = lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	td := &TaskDetail{
		task:     task,
		viewport: vp,
		width:    width,
		height:   height,
		keys:     defaultTaskDetailKeyMap(),
	}

	td.updateContent()
	td.ready = true

	return td
}

// Init initializes the task detail component
func (td *TaskDetail) Init() tea.Cmd {
	return nil
}

// Update handles messages for the task detail
func (td *TaskDetail) Update(msg tea.Msg) (*TaskDetail, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, td.keys.Close):
			return td, func() tea.Msg { return closeTaskDetailMsg{} }
		case key.Matches(msg, td.keys.Edit):
			return td, func() tea.Msg {
				return openTaskFormMsg{
					mode:   FormModeEdit,
					taskID: td.task.ID,
					task:   &td.task,
				}
			}
		}
	case tea.WindowSizeMsg:
		td.SetSize(msg.Width/2, msg.Height-4)
	}

	// Update viewport for scrolling
	td.viewport, cmd = td.viewport.Update(msg)
	return td, cmd
}

// View renders the task detail panel
func (td *TaskDetail) View() string {
	if !td.ready {
		return "Loading..."
	}

	// Header with task ID and type badge
	header := td.renderHeader()

	// Viewport with scrollable content
	content := td.viewport.View()

	// Footer with scroll position and keybindings
	footer := td.renderFooter()

	// Combine all parts
	body := lipgloss.JoinVertical(lipgloss.Left, header, content, footer)

	// Apply border style
	return taskDetailStyle.
		Width(td.width).
		Height(td.height).
		Render(body)
}

// SetSize updates the panel dimensions
func (td *TaskDetail) SetSize(width, height int) {
	td.width = width
	td.height = height
	td.viewport.Width = width - 4
	td.viewport.Height = height - 8 // Account for header and footer
	td.updateContent()
}

// updateContent rebuilds the viewport content
func (td *TaskDetail) updateContent() {
	var sections []string

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Render(td.task.TaskTitle)
	sections = append(sections, title)

	// Metadata line
	meta := td.renderMetadata()
	sections = append(sections, meta)

	// Separator
	separator := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render(strings.Repeat("-", td.width-6))
	sections = append(sections, separator)

	// Description (markdown rendered)
	desc := td.renderDescription()
	sections = append(sections, desc)

	// Additional fields
	if len(td.task.Labels) > 0 {
		sections = append(sections, "")
		labels := lipgloss.NewStyle().
			Foreground(secondaryColor).
			Render("Labels: " + strings.Join(td.task.Labels, ", "))
		sections = append(sections, labels)
	}

	if td.task.DueDate != "" {
		dueStyle := lipgloss.NewStyle()
		// Could add overdue styling here
		sections = append(sections, dueStyle.Render("Due: "+td.task.DueDate))
	}

	if td.task.EpicTitle != "" {
		sections = append(sections, "Epic: "+td.task.EpicTitle)
	}

	if len(td.task.BlockedBy) > 0 {
		blockedStyle := lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)
		sections = append(sections, blockedStyle.Render("Blocked by: "+strings.Join(td.task.BlockedBy, ", ")))
	}

	td.viewport.SetContent(strings.Join(sections, "\n"))
}

func (td *TaskDetail) renderHeader() string {
	// Display ID with type badge
	idStyle := lipgloss.NewStyle().
		Foreground(mutedColor).
		Bold(true)

	typeColors := map[string]lipgloss.Color{
		"bug":     typeBug,
		"feature": typeFeature,
		"chore":   typeChore,
	}
	typeColor := typeColors[td.task.Type]
	if typeColor == "" {
		typeColor = mutedColor
	}

	typeStyle := lipgloss.NewStyle().
		Foreground(typeColor).
		Bold(true)

	left := idStyle.Render(td.task.DisplayID) + " " + typeStyle.Render("["+td.task.Type+"]")

	// Priority indicator
	priorityColors := map[string]lipgloss.Color{
		"urgent": priorityUrgent,
		"high":   priorityHigh,
		"medium": priorityMedium,
		"low":    priorityLow,
	}
	prioColor := priorityColors[td.task.Priority]
	if prioColor == "" {
		prioColor = mutedColor
	}
	prioStyle := lipgloss.NewStyle().
		Foreground(prioColor).
		Bold(true)
	right := prioStyle.Render(td.task.Priority)

	// Join with spacing
	gap := td.width - lipgloss.Width(left) - lipgloss.Width(right) - 6
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + right
}

func (td *TaskDetail) renderMetadata() string {
	parts := []string{
		"Column: " + td.task.Column,
	}

	return lipgloss.NewStyle().
		Foreground(mutedColor).
		Render(strings.Join(parts, " | "))
}

func (td *TaskDetail) renderDescription() string {
	if td.task.TaskDescription == "" {
		return lipgloss.NewStyle().
			Italic(true).
			Foreground(mutedColor).
			Render("No description")
	}

	// Render markdown with glamour
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(td.width-8),
	)
	if err != nil {
		return td.task.TaskDescription
	}

	rendered, err := renderer.Render(td.task.TaskDescription)
	if err != nil {
		return td.task.TaskDescription
	}

	return strings.TrimSpace(rendered)
}

func (td *TaskDetail) renderFooter() string {
	// Scroll indicator
	scrollInfo := ""
	if td.viewport.TotalLineCount() > td.viewport.Height {
		percent := td.viewport.ScrollPercent() * 100
		scrollInfo = fmt.Sprintf("%.0f%%", percent)
	}

	// Keybindings hint
	hint := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render("esc:close  e:edit  j/k:scroll")

	gap := td.width - lipgloss.Width(scrollInfo) - lipgloss.Width(hint) - 6
	if gap < 1 {
		gap = 1
	}

	return hint + strings.Repeat(" ", gap) + scrollInfo
}

// Task returns the current task
func (td *TaskDetail) Task() TaskItem {
	return td.task
}

// UpdateTask updates the displayed task
func (td *TaskDetail) UpdateTask(task TaskItem) {
	td.task = task
	td.updateContent()
}
