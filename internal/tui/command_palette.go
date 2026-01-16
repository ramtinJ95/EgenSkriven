package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Command represents an action in the command palette.
type Command struct {
	ID          string
	Name        string            // Display name
	Description string            // Longer description
	Shortcut    string            // Keyboard shortcut hint
	Category    string            // Grouping category
	Action      func() tea.Cmd   // Action to execute
}

// CommandPalette is the quick command launcher.
type CommandPalette struct {
	visible    bool
	input      textinput.Model
	commands   []Command
	filtered   []Command
	selected   int
	width      int
	height     int
	maxResults int
}

// NewCommandPalette creates a new command palette.
func NewCommandPalette(commands []Command) *CommandPalette {
	ti := textinput.New()
	ti.Placeholder = "Type a command..."
	ti.Focus()

	return &CommandPalette{
		visible:    false,
		input:      ti,
		commands:   commands,
		filtered:   commands,
		selected:   0,
		maxResults: 10,
	}
}

// CommandActions holds action functions for commands.
// This is populated by the main App model.
type CommandActions struct {
	NewTask          func() tea.Cmd
	EditTask         func() tea.Cmd
	DeleteTask       func() tea.Cmd
	ViewTask         func() tea.Cmd
	MoveTaskLeft     func() tea.Cmd
	MoveTaskRight    func() tea.Cmd
	MoveToColumn     func(column string) func() tea.Cmd
	Search           func() tea.Cmd
	FilterByPriority func() tea.Cmd
	FilterByType     func() tea.Cmd
	FilterByEpic     func() tea.Cmd
	FilterByLabel    func() tea.Cmd
	ClearFilters     func() tea.Cmd
	SwitchBoard      func() tea.Cmd
	Refresh          func() tea.Cmd
	ToggleHelp       func() tea.Cmd
	SelectAll        func() tea.Cmd
	BulkMove         func() tea.Cmd
	BulkDelete       func() tea.Cmd
}

// DefaultCommands returns the standard command list.
func DefaultCommands(actions *CommandActions) []Command {
	return []Command{
		// Task Commands
		{ID: "new-task", Name: "New Task", Description: "Create a new task", Shortcut: "n", Category: "Tasks", Action: actions.NewTask},
		{ID: "edit-task", Name: "Edit Task", Description: "Edit selected task", Shortcut: "e", Category: "Tasks", Action: actions.EditTask},
		{ID: "delete-task", Name: "Delete Task", Description: "Delete selected task", Shortcut: "d", Category: "Tasks", Action: actions.DeleteTask},
		{ID: "view-task", Name: "View Task Details", Description: "Open task detail view", Shortcut: "Enter", Category: "Tasks", Action: actions.ViewTask},

		// Movement Commands
		{ID: "move-left", Name: "Move Task Left", Description: "Move task to previous column", Shortcut: "H", Category: "Movement", Action: actions.MoveTaskLeft},
		{ID: "move-right", Name: "Move Task Right", Description: "Move task to next column", Shortcut: "L", Category: "Movement", Action: actions.MoveTaskRight},
		{ID: "move-backlog", Name: "Move to Backlog", Description: "Move task to backlog", Shortcut: "1", Category: "Movement", Action: actions.MoveToColumn("backlog")},
		{ID: "move-todo", Name: "Move to Todo", Description: "Move task to todo", Shortcut: "2", Category: "Movement", Action: actions.MoveToColumn("todo")},
		{ID: "move-progress", Name: "Move to In Progress", Description: "Move task to in_progress", Shortcut: "3", Category: "Movement", Action: actions.MoveToColumn("in_progress")},
		{ID: "move-need-input", Name: "Move to Need Input", Description: "Move task to need_input", Shortcut: "4", Category: "Movement", Action: actions.MoveToColumn("need_input")},
		{ID: "move-review", Name: "Move to Review", Description: "Move task to review", Shortcut: "5", Category: "Movement", Action: actions.MoveToColumn("review")},
		{ID: "move-done", Name: "Move to Done", Description: "Mark task as done", Shortcut: "6", Category: "Movement", Action: actions.MoveToColumn("done")},

		// Filter Commands
		{ID: "search", Name: "Search", Description: "Search tasks", Shortcut: "/", Category: "Filter", Action: actions.Search},
		{ID: "filter-priority", Name: "Filter by Priority", Description: "Filter tasks by priority", Shortcut: "fp", Category: "Filter", Action: actions.FilterByPriority},
		{ID: "filter-type", Name: "Filter by Type", Description: "Filter tasks by type", Shortcut: "ft", Category: "Filter", Action: actions.FilterByType},
		{ID: "filter-epic", Name: "Filter by Epic", Description: "Filter tasks by epic", Shortcut: "fe", Category: "Filter", Action: actions.FilterByEpic},
		{ID: "filter-label", Name: "Filter by Label", Description: "Filter tasks by label", Shortcut: "fl", Category: "Filter", Action: actions.FilterByLabel},
		{ID: "clear-filters", Name: "Clear Filters", Description: "Remove all active filters", Shortcut: "fc", Category: "Filter", Action: actions.ClearFilters},

		// View Commands
		{ID: "switch-board", Name: "Switch Board", Description: "Change to different board", Shortcut: "b", Category: "View", Action: actions.SwitchBoard},
		{ID: "refresh", Name: "Refresh", Description: "Reload all data", Shortcut: "Ctrl+R", Category: "View", Action: actions.Refresh},
		{ID: "toggle-help", Name: "Toggle Help", Description: "Show/hide keyboard shortcuts", Shortcut: "?", Category: "View", Action: actions.ToggleHelp},

		// Bulk Commands
		{ID: "select-all", Name: "Select All in Column", Description: "Select all tasks in current column", Shortcut: "Ctrl+A", Category: "Selection", Action: actions.SelectAll},
		{ID: "bulk-move", Name: "Bulk Move", Description: "Move selected tasks", Shortcut: "m", Category: "Selection", Action: actions.BulkMove},
		{ID: "bulk-delete", Name: "Bulk Delete", Description: "Delete selected tasks", Shortcut: "d", Category: "Selection", Action: actions.BulkDelete},
	}
}

// Show opens the command palette.
func (p *CommandPalette) Show() {
	p.visible = true
	p.input.SetValue("")
	p.filtered = p.commands
	p.selected = 0
	p.input.Focus()
}

// Hide closes the command palette.
func (p *CommandPalette) Hide() {
	p.visible = false
	p.input.Blur()
}

// IsVisible returns true if palette is showing.
func (p *CommandPalette) IsVisible() bool {
	return p.visible
}

// SetSize updates the palette dimensions.
func (p *CommandPalette) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.input.Width = width - 6
}

// SetCommands updates the available commands.
func (p *CommandPalette) SetCommands(commands []Command) {
	p.commands = commands
	p.filtered = commands
}

// Update handles input for the command palette.
func (p *CommandPalette) Update(msg tea.Msg) (*CommandPalette, tea.Cmd) {
	if !p.visible {
		return p, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			p.Hide()
			return p, nil

		case "enter":
			if len(p.filtered) > 0 && p.selected < len(p.filtered) {
				cmd := p.filtered[p.selected]
				p.Hide()
				if cmd.Action != nil {
					return p, cmd.Action()
				}
			}
			return p, nil

		case "up", "ctrl+p":
			if p.selected > 0 {
				p.selected--
			}
			return p, nil

		case "down", "ctrl+n":
			if p.selected < len(p.filtered)-1 {
				p.selected++
			}
			return p, nil
		}
	}

	// Handle text input
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)

	// Filter commands based on input
	p.filterCommands(p.input.Value())

	return p, cmd
}

// filterCommands filters commands based on search query.
func (p *CommandPalette) filterCommands(query string) {
	if query == "" {
		p.filtered = p.commands
		p.selected = 0
		return
	}

	query = strings.ToLower(query)
	var matched []struct {
		cmd   Command
		score int
	}

	for _, cmd := range p.commands {
		score := fuzzyScore(query, strings.ToLower(cmd.Name))
		// Also check description for matches
		descScore := fuzzyScore(query, strings.ToLower(cmd.Description))
		if descScore > score {
			score = descScore
		}
		if score > 0 {
			matched = append(matched, struct {
				cmd   Command
				score int
			}{cmd, score})
		}
	}

	// Sort by score descending
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].score > matched[j].score
	})

	p.filtered = make([]Command, len(matched))
	for i, m := range matched {
		p.filtered[i] = m.cmd
	}

	// Clamp selection
	if p.selected >= len(p.filtered) {
		p.selected = len(p.filtered) - 1
	}
	if p.selected < 0 {
		p.selected = 0
	}
}

// fuzzyScore returns a matching score (0 = no match).
func fuzzyScore(query, text string) int {
	if strings.Contains(text, query) {
		// Bonus for substring match
		return 100 + (100 - len(text))
	}

	// Simple fuzzy: all query chars must appear in order
	qi := 0
	score := 0
	lastMatch := -1
	for ti := 0; ti < len(text) && qi < len(query); ti++ {
		if text[ti] == query[qi] {
			qi++
			score += 10
			// Bonus for consecutive matches
			if lastMatch == ti-1 {
				score += 5
			}
			lastMatch = ti
		}
	}

	if qi == len(query) {
		return score
	}
	return 0
}

// View renders the command palette.
func (p *CommandPalette) View() string {
	if !p.visible {
		return ""
	}

	paletteWidth := p.width - 4
	if paletteWidth < 40 {
		paletteWidth = 40
	}
	if paletteWidth > 80 {
		paletteWidth = 80
	}

	// Input field
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Width(paletteWidth - 4)

	input := inputStyle.Render(p.input.View())

	// Results list
	var results []string
	if len(p.filtered) == 0 {
		noResultsStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		results = append(results, noResultsStyle.Render("  No matching commands"))
	} else {
		for i, cmd := range p.filtered {
			if i >= p.maxResults {
				remaining := len(p.filtered) - p.maxResults
				moreStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")).
					Italic(true)
				results = append(results, moreStyle.Render(
					fmt.Sprintf("  ...and %d more", remaining)))
				break
			}

			style := lipgloss.NewStyle().Width(paletteWidth - 6)
			if i == p.selected {
				style = style.
					Background(lipgloss.Color("62")).
					Foreground(lipgloss.Color("255"))
			}

			nameStyle := lipgloss.NewStyle().Bold(true)
			shortcutStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("39"))
			categoryStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true)

			line := nameStyle.Render(cmd.Name)
			if cmd.Shortcut != "" {
				line += " " + shortcutStyle.Render("["+cmd.Shortcut+"]")
			}
			line += " " + categoryStyle.Render("("+cmd.Category+")")

			results = append(results, style.Render("  "+line))
		}
	}

	resultList := strings.Join(results, "\n")

	// Container
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 1).
		Width(paletteWidth)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("Command Palette")

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("  Ctrl+K to open | Esc to close | Enter to select")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		hint,
		"",
		input,
		"",
		resultList,
	)

	return containerStyle.Render(content)
}

// CommandPaletteClosedMsg is sent when the command palette is closed.
type CommandPaletteClosedMsg struct{}
