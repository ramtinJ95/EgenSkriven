package commands

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/output"
)

// ExportData represents the full export structure
type ExportData struct {
	Version  string        `json:"version"`
	Exported string        `json:"exported"`
	Boards   []ExportBoard `json:"boards"`
	Epics    []ExportEpic  `json:"epics"`
	Tasks    []ExportTask  `json:"tasks"`
}

// ExportBoard represents a board in export format
type ExportBoard struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Prefix  string   `json:"prefix"`
	Columns []string `json:"columns,omitempty"`
	Color   string   `json:"color,omitempty"`
}

// ExportEpic represents an epic in export format
type ExportEpic struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

// ExportTask represents a task in export format
type ExportTask struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type"`
	Priority    string   `json:"priority"`
	Column      string   `json:"column"`
	Position    float64  `json:"position"`
	Board       string   `json:"board,omitempty"`
	Epic        string   `json:"epic,omitempty"`
	Parent      string   `json:"parent,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	BlockedBy   []string `json:"blocked_by,omitempty"`
	DueDate     string   `json:"due_date,omitempty"`
	Created     string   `json:"created"`
	Updated     string   `json:"updated"`
}

// newExportCmd creates the export command
func newExportCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		format     string
		boardName  string
		outputFile string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export tasks and boards to a file",
		Long: `Export all data as JSON or CSV for backup or migration.

The JSON format includes all boards, epics, and tasks with full metadata.
The CSV format exports only tasks in a flat table format.

Examples:
  egenskriven export                           # JSON to stdout
  egenskriven export --format json > backup.json
  egenskriven export --format csv > tasks.csv
  egenskriven export --board work --format json
  egenskriven export -o backup.json            # Write to file`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Validate format
			format = strings.ToLower(format)
			if format != "json" && format != "csv" {
				return out.Error(ExitValidation, fmt.Sprintf("unsupported format: %s (use 'json' or 'csv')", format), nil)
			}

			// Determine output destination
			var writer *os.File
			if outputFile != "" && outputFile != "-" {
				f, err := os.Create(outputFile)
				if err != nil {
					return out.Error(ExitGeneralError, fmt.Sprintf("failed to create file: %v", err), nil)
				}
				defer f.Close()
				writer = f
			} else {
				writer = os.Stdout
			}

			switch format {
			case "json":
				return exportJSON(app, boardName, writer, out)
			case "csv":
				return exportCSV(app, boardName, writer, out)
			default:
				return out.Error(ExitValidation, fmt.Sprintf("unsupported format: %s", format), nil)
			}
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "json", "Output format: json, csv")
	cmd.Flags().StringVarP(&boardName, "board", "b", "", "Export specific board only")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")

	return cmd
}

// exportJSON exports all data in JSON format
func exportJSON(app *pocketbase.PocketBase, boardFilter string, writer *os.File, out *output.Formatter) error {
	data := ExportData{
		Version:  "1.0",
		Exported: time.Now().UTC().Format(time.RFC3339),
		Boards:   []ExportBoard{},
		Epics:    []ExportEpic{},
		Tasks:    []ExportTask{},
	}

	// Export boards
	boards, err := app.FindAllRecords("boards")
	if err == nil {
		for _, b := range boards {
			columns := getExportStringSlice(b.Get("columns"))
			data.Boards = append(data.Boards, ExportBoard{
				ID:      b.Id,
				Name:    b.GetString("name"),
				Prefix:  b.GetString("prefix"),
				Columns: columns,
				Color:   b.GetString("color"),
			})
		}
	}

	// Export epics
	epics, err := app.FindAllRecords("epics")
	if err == nil {
		for _, e := range epics {
			data.Epics = append(data.Epics, ExportEpic{
				ID:          e.Id,
				Title:       e.GetString("title"),
				Description: e.GetString("description"),
				Color:       e.GetString("color"),
			})
		}
	}

	// Export tasks (optionally filtered by board)
	var boardID string
	if boardFilter != "" {
		board, err := findExportBoardByNameOrPrefix(app, boardFilter)
		if err != nil {
			return out.Error(ExitNotFound, fmt.Sprintf("board not found: %s", boardFilter), nil)
		}
		boardID = board.Id
	}

	var tasks []*core.Record
	if boardID != "" {
		tasks, err = app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
	} else {
		tasks, err = app.FindAllRecords("tasks")
	}
	if err != nil {
		return out.Error(ExitGeneralError, fmt.Sprintf("failed to fetch tasks: %v", err), nil)
	}

	for _, t := range tasks {
		data.Tasks = append(data.Tasks, ExportTask{
			ID:          t.Id,
			Title:       t.GetString("title"),
			Description: t.GetString("description"),
			Type:        t.GetString("type"),
			Priority:    t.GetString("priority"),
			Column:      t.GetString("column"),
			Position:    t.GetFloat("position"),
			Board:       t.GetString("board"),
			Epic:        t.GetString("epic"),
			Parent:      t.GetString("parent"),
			Labels:      getExportStringSlice(t.Get("labels")),
			BlockedBy:   getExportStringSlice(t.Get("blocked_by")),
			DueDate:     t.GetString("due_date"),
			Created:     t.GetDateTime("created").Time().Format(time.RFC3339),
			Updated:     t.GetDateTime("updated").Time().Format(time.RFC3339),
		})
	}

	// Output JSON
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return out.Error(ExitGeneralError, fmt.Sprintf("failed to encode JSON: %v", err), nil)
	}

	// Print summary to stderr if not quiet
	if !quietMode && writer != os.Stdout {
		fmt.Fprintf(os.Stderr, "Exported %d boards, %d epics, %d tasks\n",
			len(data.Boards), len(data.Epics), len(data.Tasks))
	}

	return nil
}

// exportCSV exports tasks in CSV format
func exportCSV(app *pocketbase.PocketBase, boardFilter string, writer *os.File, out *output.Formatter) error {
	// Get tasks
	var boardID string
	if boardFilter != "" {
		board, err := findExportBoardByNameOrPrefix(app, boardFilter)
		if err != nil {
			return out.Error(ExitNotFound, fmt.Sprintf("board not found: %s", boardFilter), nil)
		}
		boardID = board.Id
	}

	var tasks []*core.Record
	var err error
	if boardID != "" {
		tasks, err = app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
	} else {
		tasks, err = app.FindAllRecords("tasks")
	}
	if err != nil {
		return out.Error(ExitGeneralError, fmt.Sprintf("failed to fetch tasks: %v", err), nil)
	}

	// Write CSV
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Header
	header := []string{
		"id", "title", "description", "type", "priority", "column",
		"position", "board", "epic", "parent", "labels", "blocked_by",
		"due_date", "created", "updated",
	}
	if err := csvWriter.Write(header); err != nil {
		return out.Error(ExitGeneralError, fmt.Sprintf("failed to write CSV header: %v", err), nil)
	}

	// Rows
	for _, t := range tasks {
		row := []string{
			t.Id,
			t.GetString("title"),
			t.GetString("description"),
			t.GetString("type"),
			t.GetString("priority"),
			t.GetString("column"),
			fmt.Sprintf("%f", t.GetFloat("position")),
			t.GetString("board"),
			t.GetString("epic"),
			t.GetString("parent"),
			strings.Join(getExportStringSlice(t.Get("labels")), ";"),
			strings.Join(getExportStringSlice(t.Get("blocked_by")), ";"),
			t.GetString("due_date"),
			t.GetDateTime("created").Time().Format(time.RFC3339),
			t.GetDateTime("updated").Time().Format(time.RFC3339),
		}
		if err := csvWriter.Write(row); err != nil {
			return out.Error(ExitGeneralError, fmt.Sprintf("failed to write CSV row: %v", err), nil)
		}
	}

	// Print summary to stderr if not quiet
	if !quietMode && writer != os.Stdout {
		fmt.Fprintf(os.Stderr, "Exported %d tasks\n", len(tasks))
	}

	return nil
}

// getExportStringSlice safely converts an interface to a string slice
func getExportStringSlice(v any) []string {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case []string:
		return val
	case []any:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	default:
		return nil
	}
}

// findExportBoardByNameOrPrefix finds a board by name or prefix
func findExportBoardByNameOrPrefix(app *pocketbase.PocketBase, query string) (*core.Record, error) {
	// Try by name first
	boards, err := app.FindAllRecords("boards",
		dbx.NewExp("name = {:query} OR prefix = {:query}",
			dbx.Params{"query": query}),
	)
	if err != nil {
		return nil, err
	}

	if len(boards) == 0 {
		return nil, fmt.Errorf("board not found: %s", query)
	}

	if len(boards) > 1 {
		return nil, fmt.Errorf("ambiguous board reference: %s", query)
	}

	return boards[0], nil
}
