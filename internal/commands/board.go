package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
	"github.com/ramtinJ95/EgenSkriven/internal/config"
)

// newBoardCmd creates the board command and its subcommands
func newBoardCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "board",
		Short: "Manage boards",
		Long: `Manage multiple boards for organizing tasks.

Each board has:
- A unique name (e.g., "Work", "Personal")
- A prefix for task IDs (e.g., "WRK", "PER")
- Optional custom columns
- An optional accent color`,
		Example: `  egenskriven board list
  egenskriven board add "Work" --prefix WRK
  egenskriven board show work
  egenskriven board use work
  egenskriven board delete work --force`,
	}

	cmd.AddCommand(newBoardListCmd(app))
	cmd.AddCommand(newBoardAddCmd(app))
	cmd.AddCommand(newBoardShowCmd(app))
	cmd.AddCommand(newBoardUseCmd(app))
	cmd.AddCommand(newBoardDeleteCmd(app))

	return cmd
}

// newBoardListCmd creates the 'board list' subcommand
func newBoardListCmd(app *pocketbase.PocketBase) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all boards",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()
			if err := app.Bootstrap(); err != nil {
				return err
			}

			records, err := board.GetAll(app)
			if err != nil {
				return err
			}

			boards := make([]*board.Board, len(records))
			for i, r := range records {
				boards[i] = board.RecordToBoard(r)
			}

			// Get current board from config
			cfg, _ := config.LoadProjectConfig()
			currentBoard := ""
			if cfg != nil {
				currentBoard = cfg.DefaultBoard
			}

			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"boards":        boards,
					"current_board": currentBoard,
					"count":         len(boards),
				})
			}

			if len(boards) == 0 {
				fmt.Println("No boards found. Create one with: egenskriven board add \"Name\" --prefix PREFIX")
				return nil
			}

			fmt.Println("BOARDS")
			fmt.Println("------")
			for _, b := range boards {
				marker := "  "
				if b.Prefix == currentBoard || b.Name == currentBoard {
					marker = "> "
				}
				fmt.Printf("%s%s (%s)\n", marker, b.Name, b.Prefix)
			}

			return nil
		},
	}
}

// newBoardAddCmd creates the 'board add' subcommand
func newBoardAddCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		prefix  string
		color   string
		columns []string
	)

	cmd := &cobra.Command{
		Use:   "add [name]",
		Short: "Create a new board",
		Long: `Create a new board with a unique prefix.

The prefix is used in task IDs (e.g., WRK-123) and must be:
- 1-10 characters
- Alphanumeric (letters and numbers only)
- Unique across all boards`,
		Args: cobra.ExactArgs(1),
		Example: `  egenskriven board add "Work" --prefix WRK
  egenskriven board add "Personal" --prefix PER --color "#22C55E"
  egenskriven board add "Sprint" --prefix SPR --columns "backlog,ready,doing,review,done"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()
			if err := app.Bootstrap(); err != nil {
				return err
			}

			name := args[0]
			if prefix == "" {
				return fmt.Errorf("--prefix is required")
			}

			b, err := board.Create(app, board.CreateInput{
				Name:    name,
				Prefix:  prefix,
				Columns: columns,
				Color:   color,
			})
			if err != nil {
				return err
			}

			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(b)
			}

			fmt.Printf("Created board: %s (%s)\n", b.Name, b.Prefix)
			return nil
		},
	}

	cmd.Flags().StringVarP(&prefix, "prefix", "p", "", "Task ID prefix (required, e.g., WRK)")
	cmd.Flags().StringVarP(&color, "color", "c", "", "Accent color (hex, e.g., #3B82F6)")
	cmd.Flags().StringSliceVar(&columns, "columns", nil, "Custom columns (comma-separated)")
	cmd.MarkFlagRequired("prefix")

	return cmd
}

// newBoardShowCmd creates the 'board show' subcommand
func newBoardShowCmd(app *pocketbase.PocketBase) *cobra.Command {
	return &cobra.Command{
		Use:   "show [name-or-prefix]",
		Short: "Show board details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()
			if err := app.Bootstrap(); err != nil {
				return err
			}

			record, err := board.GetByNameOrPrefix(app, args[0])
			if err != nil {
				return err
			}

			b := board.RecordToBoard(record)

			// Count tasks in this board
			tasks, _ := app.FindAllRecords("tasks",
				dbx.NewExp("board = {:board}", dbx.Params{"board": record.Id}),
			)
			taskCount := len(tasks)

			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"id":         b.ID,
					"name":       b.Name,
					"prefix":     b.Prefix,
					"columns":    b.Columns,
					"color":      b.Color,
					"task_count": taskCount,
				})
			}

			fmt.Printf("Board: %s\n", b.Name)
			fmt.Printf("Prefix: %s\n", b.Prefix)
			fmt.Printf("Columns: %v\n", b.Columns)
			if b.Color != "" {
				fmt.Printf("Color: %s\n", b.Color)
			}
			fmt.Printf("Tasks: %d\n", taskCount)

			return nil
		},
	}
}

// newBoardUseCmd creates the 'board use' subcommand
func newBoardUseCmd(app *pocketbase.PocketBase) *cobra.Command {
	return &cobra.Command{
		Use:   "use [name-or-prefix]",
		Short: "Set the default board for CLI commands",
		Long: `Set the default board used by CLI commands.

When set, commands like 'add', 'list', 'show' will operate on this board
unless overridden with the --board flag.

The setting is stored in .egenskriven/config.json in the current directory.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()
			if err := app.Bootstrap(); err != nil {
				return err
			}

			record, err := board.GetByNameOrPrefix(app, args[0])
			if err != nil {
				return err
			}

			// Update project config
			cfg, err := config.LoadProjectConfig()
			if err != nil {
				cfg = config.DefaultConfig()
			}

			cfg.DefaultBoard = record.GetString("prefix")

			if err := config.SaveConfig(".", cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]string{
					"default_board": cfg.DefaultBoard,
				})
			}

			fmt.Printf("Default board set to: %s (%s)\n",
				record.GetString("name"),
				record.GetString("prefix"))

			return nil
		},
	}
}

// newBoardDeleteCmd creates the 'board delete' subcommand
func newBoardDeleteCmd(app *pocketbase.PocketBase) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [name-or-prefix]",
		Short: "Delete a board",
		Long: `Delete a board and all its tasks.

WARNING: This permanently deletes all tasks in the board!
Use --force to skip the confirmation prompt.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()
			if err := app.Bootstrap(); err != nil {
				return err
			}

			record, err := board.GetByNameOrPrefix(app, args[0])
			if err != nil {
				return err
			}

			boardName := record.GetString("name")
			boardPrefix := record.GetString("prefix")

			// Count tasks that will be deleted
			tasks, _ := app.FindAllRecords("tasks",
				dbx.NewExp("board = {:board}", dbx.Params{"board": record.Id}),
			)
			taskCount := len(tasks)

			// Confirm deletion
			if !force && taskCount > 0 {
				fmt.Printf("WARNING: Board '%s' contains %d task(s).\n", boardName, taskCount)
				fmt.Print("Delete board and all its tasks? [y/N]: ")

				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			// Delete board and tasks
			if err := board.Delete(app, record.Id, true); err != nil {
				return err
			}

			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"deleted":       true,
					"board":         boardName,
					"prefix":        boardPrefix,
					"tasks_deleted": taskCount,
				})
			}

			fmt.Printf("Deleted board '%s' and %d task(s)\n", boardName, taskCount)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}
