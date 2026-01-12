package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
)

// BoardOption represents a board in the selector list
type BoardOption struct {
	ID         string
	Name       string
	Prefix     string
	Color      string
	TaskCount  int
	Columns    []string
	IsSelected bool // Currently active board
}

// Implement list.Item interface for BoardOption

// FilterValue returns the string used for filtering/searching
func (b BoardOption) FilterValue() string {
	return b.Name + " " + b.Prefix
}

// Title returns the display title for the list item
func (b BoardOption) Title() string {
	prefix := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Render(b.Prefix)

	name := b.Name
	if b.IsSelected {
		name = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Render(name + " (current)")
	}

	return fmt.Sprintf("%s - %s", prefix, name)
}

// Description returns additional info shown below the title
func (b BoardOption) Description() string {
	if b.TaskCount == 0 {
		return "No tasks"
	}
	if b.TaskCount == 1 {
		return "1 task"
	}
	return fmt.Sprintf("%d tasks", b.TaskCount)
}

// BoardSelector is the modal component for selecting boards
type BoardSelector struct {
	list         list.Model
	boards       []BoardOption
	currentBoard string // ID of currently selected board
	width        int
	height       int
	keys         boardSelectorKeyMap
}

// boardSelectorKeyMap defines keybindings for the board selector
type boardSelectorKeyMap struct {
	Select key.Binding
	Cancel key.Binding
}

// defaultBoardSelectorKeys returns the default keybindings
func defaultBoardSelectorKeys() boardSelectorKeyMap {
	return boardSelectorKeyMap{
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select board"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// NewBoardSelector creates a new board selector component
func NewBoardSelector(boards []BoardOption, currentBoardID string) *BoardSelector {
	// Create list items from board options
	items := make([]list.Item, len(boards))
	for i, b := range boards {
		b.IsSelected = (b.ID == currentBoardID)
		boards[i] = b
		items[i] = b
	}

	// Create delegate with custom styling
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetHeight(2)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("62")).
		BorderLeftForeground(lipgloss.Color("62"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("240"))

	// Create list model
	l := list.New(items, delegate, 0, 0)
	l.Title = "Switch Board"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	// Find and select current board in list
	for i, b := range boards {
		if b.ID == currentBoardID {
			l.Select(i)
			break
		}
	}

	return &BoardSelector{
		list:         l,
		boards:       boards,
		currentBoard: currentBoardID,
		keys:         defaultBoardSelectorKeys(),
	}
}

// Init initializes the board selector (required by tea.Model interface pattern)
func (s *BoardSelector) Init() tea.Cmd {
	return nil
}

// Update handles messages for the board selector
func (s *BoardSelector) Update(msg tea.Msg) (*BoardSelector, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle selection
		if key.Matches(msg, s.keys.Select) {
			if item, ok := s.list.SelectedItem().(BoardOption); ok {
				return s, func() tea.Msg {
					return boardSwitchedMsg{boardID: item.ID}
				}
			}
		}

		// Handle cancellation
		if key.Matches(msg, s.keys.Cancel) {
			// Return nil command to signal cancellation
			return s, nil
		}
	}

	// Delegate to list for navigation and filtering
	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

// View renders the board selector
func (s *BoardSelector) View() string {
	// Modal container style
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(s.width).
		Height(s.height)

	return modalStyle.Render(s.list.View())
}

// SetSize updates the dimensions of the board selector
func (s *BoardSelector) SetSize(width, height int) {
	s.width = width
	s.height = height

	// Account for modal padding and border
	listWidth := width - 6
	listHeight := height - 4

	s.list.SetSize(listWidth, listHeight)
}

// SelectedBoard returns the currently highlighted board option
func (s *BoardSelector) SelectedBoard() (BoardOption, bool) {
	if item, ok := s.list.SelectedItem().(BoardOption); ok {
		return item, true
	}
	return BoardOption{}, false
}

// BoardOptionsFromRecords converts PocketBase records to BoardOption slice
func BoardOptionsFromRecords(records []*core.Record, taskCounts map[string]int) []BoardOption {
	options := make([]BoardOption, len(records))
	for i, record := range records {
		b := board.RecordToBoard(record)
		count := 0
		if taskCounts != nil {
			count = taskCounts[record.Id]
		}
		options[i] = BoardOption{
			ID:        record.Id,
			Name:      b.Name,
			Prefix:    b.Prefix,
			Color:     b.Color,
			TaskCount: count,
			Columns:   b.Columns,
		}
	}
	return options
}
