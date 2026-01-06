package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/output"
)

// newImportCmd creates the import command
func newImportCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		strategy string
		dryRun   bool
	)

	cmd := &cobra.Command{
		Use:   "import [file]",
		Short: "Import tasks and boards from a file",
		Long: `Import data from a JSON backup file.

Strategies:
  merge   - Skip existing records, add new ones (default)
  replace - Overwrite existing records with same ID

Examples:
  egenskriven import backup.json
  egenskriven import backup.json --strategy replace
  egenskriven import backup.json --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Validate strategy
			if strategy != "merge" && strategy != "replace" {
				return out.Error(ExitValidation, fmt.Sprintf("invalid strategy: %s (use 'merge' or 'replace')", strategy), nil)
			}

			filename := args[0]
			return runImport(app, filename, strategy, dryRun, out)
		},
	}

	cmd.Flags().StringVar(&strategy, "strategy", "merge", "Import strategy: merge, replace")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying")

	return cmd
}

// ImportStats tracks import statistics
type ImportStats struct {
	BoardsCreated int
	BoardsUpdated int
	BoardsSkipped int
	EpicsCreated  int
	EpicsUpdated  int
	EpicsSkipped  int
	TasksCreated  int
	TasksUpdated  int
	TasksSkipped  int
}

// runImport performs the actual import
func runImport(app *pocketbase.PocketBase, filename, strategy string, dryRun bool, out *output.Formatter) error {
	// Read file
	file, err := os.Open(filename)
	if err != nil {
		return out.Error(ExitGeneralError, fmt.Sprintf("failed to open file: %v", err), nil)
	}
	defer file.Close()

	// Parse JSON
	var data ExportData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return out.Error(ExitGeneralError, fmt.Sprintf("failed to parse JSON: %v", err), nil)
	}

	if !quietMode {
		fmt.Fprintf(os.Stderr, "Importing from %s (version %s, exported %s)\n",
			filename, data.Version, data.Exported)
		fmt.Fprintf(os.Stderr, "Found: %d boards, %d epics, %d tasks\n",
			len(data.Boards), len(data.Epics), len(data.Tasks))

		if dryRun {
			fmt.Fprintln(os.Stderr, "\n[DRY RUN - no changes will be made]")
		}
		fmt.Fprintln(os.Stderr)
	}

	stats := ImportStats{}

	// Import boards
	if err := importBoards(app, data.Boards, strategy, dryRun, &stats); err != nil {
		return out.Error(ExitGeneralError, fmt.Sprintf("failed to import boards: %v", err), nil)
	}

	// Import epics
	if err := importEpics(app, data.Epics, strategy, dryRun, &stats); err != nil {
		return out.Error(ExitGeneralError, fmt.Sprintf("failed to import epics: %v", err), nil)
	}

	// Import tasks
	if err := importTasks(app, data.Tasks, strategy, dryRun, &stats); err != nil {
		return out.Error(ExitGeneralError, fmt.Sprintf("failed to import tasks: %v", err), nil)
	}

	// Output results
	if out.JSON {
		out.WriteJSON(map[string]any{
			"dry_run":  dryRun,
			"strategy": strategy,
			"stats": map[string]any{
				"boards": map[string]int{
					"created": stats.BoardsCreated,
					"updated": stats.BoardsUpdated,
					"skipped": stats.BoardsSkipped,
				},
				"epics": map[string]int{
					"created": stats.EpicsCreated,
					"updated": stats.EpicsUpdated,
					"skipped": stats.EpicsSkipped,
				},
				"tasks": map[string]int{
					"created": stats.TasksCreated,
					"updated": stats.TasksUpdated,
					"skipped": stats.TasksSkipped,
				},
			},
		})
	} else if !quietMode {
		fmt.Fprintln(os.Stderr, "\nImport Summary:")
		fmt.Fprintf(os.Stderr, "  Boards: %d created, %d updated, %d skipped\n",
			stats.BoardsCreated, stats.BoardsUpdated, stats.BoardsSkipped)
		fmt.Fprintf(os.Stderr, "  Epics:  %d created, %d updated, %d skipped\n",
			stats.EpicsCreated, stats.EpicsUpdated, stats.EpicsSkipped)
		fmt.Fprintf(os.Stderr, "  Tasks:  %d created, %d updated, %d skipped\n",
			stats.TasksCreated, stats.TasksUpdated, stats.TasksSkipped)

		if dryRun {
			fmt.Fprintln(os.Stderr, "\n[DRY RUN - no changes were made]")
		}
	}

	return nil
}

// importBoards imports board records
func importBoards(app *pocketbase.PocketBase, boards []ExportBoard, strategy string, dryRun bool, stats *ImportStats) error {
	collection, err := app.FindCollectionByNameOrId("boards")
	if err != nil {
		return err
	}

	for _, b := range boards {
		existing, err := app.FindRecordById("boards", b.ID)

		if err == nil && existing != nil {
			// Record exists
			if strategy == "replace" {
				if !dryRun {
					existing.Set("name", b.Name)
					existing.Set("prefix", b.Prefix)
					if len(b.Columns) > 0 {
						existing.Set("columns", b.Columns)
					}
					if b.Color != "" {
						existing.Set("color", b.Color)
					}
					if err := app.Save(existing); err != nil {
						return fmt.Errorf("failed to update board %s: %w", b.Name, err)
					}
				}
				stats.BoardsUpdated++
			} else {
				stats.BoardsSkipped++
			}
			continue
		}

		// Create new record
		if !dryRun {
			record := core.NewRecord(collection)
			record.Id = b.ID
			record.Set("name", b.Name)
			record.Set("prefix", b.Prefix)
			if len(b.Columns) > 0 {
				record.Set("columns", b.Columns)
			} else {
				record.Set("columns", []string{"backlog", "todo", "in_progress", "review", "done"})
			}
			if b.Color != "" {
				record.Set("color", b.Color)
			}
			if err := app.Save(record); err != nil {
				return fmt.Errorf("failed to import board %s: %w", b.Name, err)
			}
		}
		stats.BoardsCreated++
	}

	return nil
}

// importEpics imports epic records
func importEpics(app *pocketbase.PocketBase, epics []ExportEpic, strategy string, dryRun bool, stats *ImportStats) error {
	collection, err := app.FindCollectionByNameOrId("epics")
	if err != nil {
		return err
	}

	for _, e := range epics {
		existing, err := app.FindRecordById("epics", e.ID)

		if err == nil && existing != nil {
			// Record exists
			if strategy == "replace" {
				if !dryRun {
					existing.Set("title", e.Title)
					existing.Set("description", e.Description)
					if e.Color != "" {
						existing.Set("color", e.Color)
					}
					if err := app.Save(existing); err != nil {
						return fmt.Errorf("failed to update epic %s: %w", e.Title, err)
					}
				}
				stats.EpicsUpdated++
			} else {
				stats.EpicsSkipped++
			}
			continue
		}

		// Create new record
		if !dryRun {
			record := core.NewRecord(collection)
			record.Id = e.ID
			record.Set("title", e.Title)
			record.Set("description", e.Description)
			if e.Color != "" {
				record.Set("color", e.Color)
			}
			if err := app.Save(record); err != nil {
				return fmt.Errorf("failed to import epic %s: %w", e.Title, err)
			}
		}
		stats.EpicsCreated++
	}

	return nil
}

// importTasks imports task records
func importTasks(app *pocketbase.PocketBase, tasks []ExportTask, strategy string, dryRun bool, stats *ImportStats) error {
	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		return err
	}

	for _, t := range tasks {
		existing, err := app.FindRecordById("tasks", t.ID)

		if err == nil && existing != nil {
			// Record exists
			if strategy == "replace" {
				if !dryRun {
					// Replace strategy: set all fields from import data
					// Empty values in import data will clear the fields
					existing.Set("title", t.Title)
					existing.Set("description", t.Description)
					existing.Set("type", t.Type)
					existing.Set("priority", t.Priority)
					existing.Set("column", t.Column)
					existing.Set("position", t.Position)
					existing.Set("board", t.Board)          // Clear if empty
					existing.Set("epic", t.Epic)            // Clear if empty
					existing.Set("parent", t.Parent)        // Clear if empty
					existing.Set("labels", t.Labels)        // Clear if nil/empty
					existing.Set("blocked_by", t.BlockedBy) // Clear if nil/empty
					existing.Set("due_date", t.DueDate)     // Clear if empty
					if err := app.Save(existing); err != nil {
						return fmt.Errorf("failed to update task %s: %w", t.Title, err)
					}
				}
				stats.TasksUpdated++
			} else {
				stats.TasksSkipped++
			}
			continue
		}

		// Create new record
		if !dryRun {
			record := core.NewRecord(collection)
			record.Id = t.ID
			record.Set("title", t.Title)
			record.Set("description", t.Description)
			record.Set("type", t.Type)
			record.Set("priority", t.Priority)
			record.Set("column", t.Column)
			record.Set("position", t.Position)
			// Preserve original created_by if available, otherwise default to "cli"
			if t.CreatedBy != "" {
				record.Set("created_by", t.CreatedBy)
			} else {
				record.Set("created_by", "cli")
			}
			if t.Board != "" {
				record.Set("board", t.Board)
			}
			if t.Epic != "" {
				record.Set("epic", t.Epic)
			}
			if t.Parent != "" {
				record.Set("parent", t.Parent)
			}
			if len(t.Labels) > 0 {
				record.Set("labels", t.Labels)
			}
			if len(t.BlockedBy) > 0 {
				record.Set("blocked_by", t.BlockedBy)
			}
			if t.DueDate != "" {
				record.Set("due_date", t.DueDate)
			}
			if err := app.Save(record); err != nil {
				return fmt.Errorf("failed to import task %s: %w", t.Title, err)
			}
		}
		stats.TasksCreated++
	}

	return nil
}
