package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// keyMap defines all keybindings for the TUI.
// Each binding includes the keys and help text.
type keyMap struct {
	// Navigation - moving between columns and tasks
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding

	// Actions - interacting with tasks
	Enter key.Binding

	// Board operations
	Board key.Binding

	// Filtering - search and filter operations
	// Note: fp, ft, fl, fe, fc are two-key sequences handled by pendingFilterKey
	Search         key.Binding
	FilterPriority key.Binding
	FilterType     key.Binding
	FilterLabel    key.Binding
	FilterEpic     key.Binding
	ClearFilters   key.Binding

	// Global - application-level controls
	Quit   key.Binding
	Help   key.Binding
	Escape key.Binding
}

// defaultKeyMap returns the default keybindings.
// Uses vim-style navigation (h/j/k/l) with arrow key alternatives.
func defaultKeyMap() keyMap {
	return keyMap{
		// Up moves selection up within a column (previous task)
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k/up", "up"),
		),
		// Down moves selection down within a column (next task)
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j/down", "down"),
		),
		// Left moves focus to the previous column
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("h/left", "prev column"),
		),
		// Right moves focus to the next column
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("l/right", "next column"),
		),
		// Enter opens task details (Phase 2)
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view details"),
		),
		// Board opens the board selector (Phase 3)
		Board: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "switch board"),
		),
		// Search opens the search overlay
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		// FilterPriority opens priority filter (two-key: f then p)
		FilterPriority: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("fp", "filter priority"),
		),
		// FilterType opens type filter (two-key: f then t)
		FilterType: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("ft", "filter type"),
		),
		// FilterLabel opens label filter (two-key: f then l)
		FilterLabel: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("fl", "filter label"),
		),
		// FilterEpic opens epic filter (two-key: f then e)
		FilterEpic: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("fe", "filter epic"),
		),
		// ClearFilters clears all active filters (two-key: f then c)
		ClearFilters: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("fc", "clear filters"),
		),
		// Quit exits the TUI
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		// Help toggles the help overlay (Phase 2)
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		// Escape closes overlays or cancels operations
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// ShortHelp returns the keybindings to show in the short help view.
// These are displayed in the status bar.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.Up, k.Down, k.Enter, k.Search, k.Board, k.Quit, k.Help}
}

// FullHelp returns all keybindings grouped for the full help view.
// Used when the user presses '?' to see all available keys.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},                                              // Navigation
		{k.Enter, k.Board, k.Quit, k.Help},                                           // Actions
		{k.Search, k.FilterPriority, k.FilterType, k.FilterLabel, k.ClearFilters},    // Filtering
		{k.Escape},                                                                   // Global
	}
}

// FilterHelp returns the filter-related keybindings.
// Used for displaying filter shortcuts in the status bar when appropriate.
func (k keyMap) FilterHelp() []key.Binding {
	return []key.Binding{
		k.Search,
		k.FilterPriority,
		k.FilterType,
		k.FilterLabel,
		k.FilterEpic,
		k.ClearFilters,
	}
}
