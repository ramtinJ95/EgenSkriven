package commands

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/config"
	"github.com/ramtinJ95/EgenSkriven/internal/tui"
)

// newTuiCmd creates the 'tui' command for launching the terminal UI.
func newTuiCmd(app *pocketbase.PocketBase) *cobra.Command {
	var boardRef string

	cmd := &cobra.Command{
		Use:     "tui",
		Aliases: []string{"ui", "board"},
		Short:   "Open interactive kanban board",
		Long: `Launch the terminal user interface for managing tasks in a kanban board view.

The TUI provides a full kanban board experience with:
- Column-based task view (backlog, todo, in_progress, need_input, review, done)
- Vim-style navigation (h/j/k/l)
- Task details view
- Real-time updates when server is running

Examples:
  egenskriven tui
  egenskriven tui --board work
  egenskriven tui -b WRK

Navigation:
  h/l       Move between columns
  j/k       Move between tasks within a column
  Enter     View task details
  q         Quit`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Bootstrap the app to ensure database is ready
			if err := app.Bootstrap(); err != nil {
				return fmt.Errorf("failed to bootstrap: %w", err)
			}

			// Load config for default board if not specified
			if boardRef == "" {
				cfg, _ := config.LoadProjectConfig()
				if cfg != nil && cfg.DefaultBoard != "" {
					boardRef = cfg.DefaultBoard
				}
			}

			// Run the TUI
			return tui.Run(app, boardRef)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&boardRef, "board", "b", "",
		"Board to open (name or prefix)")

	return cmd
}
