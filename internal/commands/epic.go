package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

func newEpicCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epic",
		Short: "Manage epics",
		Long: `Manage epics for grouping related tasks.

Epics are larger initiatives that contain multiple tasks.
Examples: "Q1 Launch", "Auth Refactor", "Performance Sprint"`,
	}

	// Add subcommands
	cmd.AddCommand(newEpicListCmd(app))
	cmd.AddCommand(newEpicAddCmd(app))
	cmd.AddCommand(newEpicShowCmd(app))
	cmd.AddCommand(newEpicDeleteCmd(app))

	return cmd
}

// ========== Epic List ==========

func newEpicListCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all epics",
		Long:  `List all epics with their task counts.`,
		Example: `  egenskriven epic list
  egenskriven epic list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Find all epics
			records, err := app.FindAllRecords("epics")
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to list epics: %v", err), nil)
			}

			// Output
			if out.JSON {
				epics := make([]map[string]any, 0, len(records))
				for _, record := range records {
					taskCount := getEpicTaskCount(app, record.Id)
					epics = append(epics, map[string]any{
						"id":          record.Id,
						"title":       record.GetString("title"),
						"description": record.GetString("description"),
						"color":       record.GetString("color"),
						"task_count":  taskCount,
						"created":     record.GetDateTime("created").String(),
						"updated":     record.GetDateTime("updated").String(),
					})
				}
				return json.NewEncoder(os.Stdout).Encode(map[string]any{
					"epics": epics,
					"count": len(epics),
				})
			}

			// Human-readable output
			if len(records) == 0 {
				fmt.Println("No epics found. Create one with: egenskriven epic add \"Epic title\"")
				return nil
			}

			fmt.Println("EPICS")
			fmt.Println(strings.Repeat("-", 40))
			for _, record := range records {
				taskCount := getEpicTaskCount(app, record.Id)
				colorIndicator := ""
				if color := record.GetString("color"); color != "" {
					colorIndicator = fmt.Sprintf(" %s", color)
				}
				fmt.Printf("  [%s] %s%s (%d tasks)\n",
					shortID(record.Id), record.GetString("title"), colorIndicator, taskCount)
			}
			fmt.Printf("\nTotal: %d epics\n", len(records))

			return nil
		},
	}

	return cmd
}

// ========== Epic Add ==========

func newEpicAddCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		color       string
		description string
	)

	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Create a new epic",
		Long: `Create a new epic to group related tasks.

Epics can have a color for visual identification in the UI.
Color must be a valid hex code (e.g., #3B82F6).`,
		Example: `  egenskriven epic add "Q1 Launch"
  egenskriven epic add "Auth Refactor" --color "#22C55E"
  egenskriven epic add "Tech Debt" --description "Clean up legacy code"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			title := args[0]

			// Validate color format if provided
			if color != "" && !isValidHexColor(color) {
				return out.Error(ExitValidation,
					fmt.Sprintf("invalid color format '%s' (must be hex like #3B82F6)", color), nil)
			}

			// Find epics collection
			collection, err := app.FindCollectionByNameOrId("epics")
			if err != nil {
				return out.Error(ExitGeneralError, "epics collection not found - run migrations first", nil)
			}

			// Create record
			record := core.NewRecord(collection)
			record.Set("title", title)
			if description != "" {
				record.Set("description", description)
			}
			if color != "" {
				record.Set("color", color)
			}

			if err := app.Save(record); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to create epic: %v", err), nil)
			}

			// Output
			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]any{
					"id":          record.Id,
					"title":       record.GetString("title"),
					"description": record.GetString("description"),
					"color":       record.GetString("color"),
					"task_count":  0,
					"created":     record.GetDateTime("created").String(),
				})
			}

			colorDisplay := ""
			if color != "" {
				colorDisplay = fmt.Sprintf(" %s", color)
			}
			fmt.Printf("Created epic: %s%s [%s]\n", title, colorDisplay, shortID(record.Id))

			return nil
		},
	}

	cmd.Flags().StringVarP(&color, "color", "c", "", "Epic color (hex, e.g., #3B82F6)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Epic description")

	return cmd
}

// ========== Epic Show ==========

func newEpicShowCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <epic>",
		Short: "Show epic details",
		Long: `Show detailed information about an epic including linked tasks.

You can reference an epic by:
- Full ID: abc123def456
- Partial ID: abc123
- Title (case-insensitive): "q1 launch"`,
		Example: `  egenskriven epic show abc123
  egenskriven epic show "Q1 Launch"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			ref := args[0]

			// Resolve epic reference
			record, err := resolveEpic(app, ref)
			if err != nil {
				return out.Error(ExitNotFound, err.Error(), nil)
			}

			// Get linked tasks
			linkedTasks, err := app.FindAllRecords("tasks",
				dbx.NewExp("epic = {:epicId}", dbx.Params{"epicId": record.Id}),
			)
			if err != nil {
				linkedTasks = []*core.Record{}
			}

			// Output
			if out.JSON {
				tasks := make([]map[string]any, 0, len(linkedTasks))
				for _, task := range linkedTasks {
					tasks = append(tasks, map[string]any{
						"id":       task.Id,
						"title":    task.GetString("title"),
						"column":   task.GetString("column"),
						"priority": task.GetString("priority"),
					})
				}

				return json.NewEncoder(os.Stdout).Encode(map[string]any{
					"epic": map[string]any{
						"id":          record.Id,
						"title":       record.GetString("title"),
						"description": record.GetString("description"),
						"color":       record.GetString("color"),
						"task_count":  len(linkedTasks),
						"created":     record.GetDateTime("created").String(),
						"updated":     record.GetDateTime("updated").String(),
					},
					"tasks": tasks,
				})
			}

			// Human-readable output
			fmt.Printf("Epic: %s\n", record.Id)
			fmt.Printf("Title:       %s\n", record.GetString("title"))
			if color := record.GetString("color"); color != "" {
				fmt.Printf("Color:       %s\n", color)
			}
			if desc := record.GetString("description"); desc != "" {
				fmt.Printf("Description: %s\n", desc)
			}
			fmt.Printf("Created:     %s\n", record.GetDateTime("created").String())
			fmt.Printf("Updated:     %s\n", record.GetDateTime("updated").String())
			fmt.Printf("\nLinked Tasks (%d):\n", len(linkedTasks))

			if len(linkedTasks) == 0 {
				fmt.Println("  (no tasks)")
			} else {
				for _, task := range linkedTasks {
					fmt.Printf("  [%s] %s (%s, %s)\n",
						shortID(task.Id),
						task.GetString("title"),
						task.GetString("column"),
						task.GetString("priority"),
					)
				}
			}

			return nil
		},
	}

	return cmd
}

// ========== Epic Delete ==========

func newEpicDeleteCmd(app *pocketbase.PocketBase) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <epic>",
		Short: "Delete an epic",
		Long: `Delete an epic.

Tasks linked to the epic will remain but will no longer be associated with the epic.
Use --force to skip the confirmation prompt.`,
		Example: `  egenskriven epic delete abc123
  egenskriven epic delete "Q1 Launch" --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			ref := args[0]

			// Resolve epic
			record, err := resolveEpic(app, ref)
			if err != nil {
				return out.Error(ExitNotFound, err.Error(), nil)
			}

			// Count linked tasks
			taskCount := getEpicTaskCount(app, record.Id)

			// Confirm deletion
			if !force && !out.Quiet && !out.JSON {
				fmt.Printf("Delete epic '%s'?", record.GetString("title"))
				if taskCount > 0 {
					fmt.Printf(" (%d tasks will be unlinked)", taskCount)
				}
				fmt.Print("\nType 'yes' to confirm: ")

				var response string
				fmt.Scanln(&response)
				if response != "yes" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			// Delete the epic
			if err := app.Delete(record); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to delete epic: %v", err), nil)
			}

			// Output
			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]any{
					"deleted":        record.Id,
					"title":          record.GetString("title"),
					"tasks_unlinked": taskCount,
				})
			}

			fmt.Printf("Deleted epic: %s [%s]\n", record.GetString("title"), shortID(record.Id))
			if taskCount > 0 {
				fmt.Printf("  %d tasks unlinked\n", taskCount)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

// ========== Helper Functions ==========

// resolveEpic finds an epic by ID, ID prefix, or title
func resolveEpic(app *pocketbase.PocketBase, ref string) (*core.Record, error) {
	// Try exact ID match
	record, err := app.FindRecordById("epics", ref)
	if err == nil {
		return record, nil
	}

	// Try ID prefix match
	records, err := app.FindAllRecords("epics",
		dbx.NewExp("id LIKE {:prefix}", dbx.Params{"prefix": ref + "%"}),
	)
	if err == nil {
		switch len(records) {
		case 1:
			return records[0], nil
		case 0:
			// No ID prefix matches, continue to title search
		default:
			// Multiple ID prefix matches - ambiguous
			var matches []string
			for _, r := range records {
				matches = append(matches, fmt.Sprintf("[%s] %s", shortID(r.Id), r.GetString("title")))
			}
			return nil, fmt.Errorf("ambiguous epic ID prefix '%s' matches multiple epics:\n  %s",
				ref, strings.Join(matches, "\n  "))
		}
	}

	// Try title match (case-insensitive)
	records, err = app.FindAllRecords("epics",
		dbx.NewExp("LOWER(title) LIKE {:title}",
			dbx.Params{"title": "%" + strings.ToLower(ref) + "%"}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search epics: %w", err)
	}

	switch len(records) {
	case 0:
		return nil, fmt.Errorf("no epic found matching: %s", ref)
	case 1:
		return records[0], nil
	default:
		// Multiple matches - ambiguous
		var matches []string
		for _, r := range records {
			matches = append(matches, fmt.Sprintf("[%s] %s", shortID(r.Id), r.GetString("title")))
		}
		return nil, fmt.Errorf("ambiguous epic reference '%s' matches multiple epics:\n  %s",
			ref, strings.Join(matches, "\n  "))
	}
}

// getEpicTaskCount returns the number of tasks linked to an epic
func getEpicTaskCount(app *pocketbase.PocketBase, epicID string) int {
	tasks, err := app.FindAllRecords("tasks",
		dbx.NewExp("epic = {:epicId}", dbx.Params{"epicId": epicID}),
	)
	if err != nil {
		return 0
	}
	return len(tasks)
}

// isValidHexColor validates a hex color string (#RRGGBB format)
func isValidHexColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		c := color[i]
		valid := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
		if !valid {
			return false
		}
	}
	return true
}
