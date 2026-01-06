package commands

import (
	"fmt"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
	"github.com/ramtinJ95/EgenSkriven/internal/config"
)

func newListCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		columns    []string
		types      []string
		priorities []string
		search     string
		createdBy  string
		agentName  string
		ready      bool
		isBlocked  bool
		notBlocked bool
		fields     string
		epicFilter string
		labels     []string
		limit      int
		sort       string
		boardRef   string
		allBoards  bool
		dueBefore  string
		dueAfter   string
		hasDue     bool
		noDue      bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Long: `List and filter tasks on the kanban board.

By default, shows all tasks grouped by column. Use flags to filter.

Examples:
  egenskriven list
  egenskriven list --column todo
  egenskriven list --type bug --priority urgent
  egenskriven list --ready
  egenskriven list --is-blocked
  egenskriven list --json --fields id,title,column
  egenskriven list --epic "Q1 Launch"
  egenskriven list --label frontend --label ui
  egenskriven list --limit 10
  egenskriven list --sort "-priority,position"`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Build filter expressions
			var filters []dbx.Expression

			// Validate mutually exclusive flags
			if isBlocked && notBlocked {
				return out.Error(ExitValidation,
					"--is-blocked and --not-blocked are mutually exclusive", nil)
			}

			// Board filter (unless --all-boards is set)
			var boardsMap map[string]*core.Record
			if !allBoards {
				// Determine which board to filter by
				boardRefToUse := boardRef
				if boardRefToUse == "" {
					// Check config for default board
					cfg, _ := config.LoadProjectConfig()
					if cfg != nil && cfg.DefaultBoard != "" {
						boardRefToUse = cfg.DefaultBoard
					}
				}

				if boardRefToUse != "" {
					boardRecord, err := board.GetByNameOrPrefix(app, boardRefToUse)
					if err != nil {
						return out.Error(ExitValidation, fmt.Sprintf("invalid board: %v", err), nil)
					}
					filters = append(filters, dbx.NewExp(
						"board = {:board}",
						dbx.Params{"board": boardRecord.Id},
					))
				}
			}

			// Load boards for display ID mapping
			allBoardRecords, _ := board.GetAll(app)
			boardsMap = make(map[string]*core.Record)
			for _, b := range allBoardRecords {
				boardsMap[b.Id] = b
			}

			// Ready filter: unblocked tasks in todo/backlog
			if ready {
				columns = []string{"todo", "backlog"}
				notBlocked = true
			}

			// Column filter
			if len(columns) > 0 {
				for _, col := range columns {
					if !isValidColumn(col) {
						return out.Error(ExitValidation,
							fmt.Sprintf("invalid column '%s', must be one of: %v", col, ValidColumns), nil)
					}
				}
				filters = append(filters, buildInFilter("column", columns))
			}

			// Type filter
			if len(types) > 0 {
				for _, t := range types {
					if !isValidType(t) {
						return out.Error(ExitValidation,
							fmt.Sprintf("invalid type '%s', must be one of: %v", t, ValidTypes), nil)
					}
				}
				filters = append(filters, buildInFilter("type", types))
			}

			// Priority filter
			if len(priorities) > 0 {
				for _, p := range priorities {
					if !isValidPriority(p) {
						return out.Error(ExitValidation,
							fmt.Sprintf("invalid priority '%s', must be one of: %v", p, ValidPriorities), nil)
					}
				}
				filters = append(filters, buildInFilter("priority", priorities))
			}

			// Search filter
			if search != "" {
				filters = append(filters, dbx.NewExp(
					"LOWER(title) LIKE {:search} ESCAPE '\\'",
					dbx.Params{"search": "%" + escapeLikePattern(strings.ToLower(search)) + "%"},
				))
			}

			// Created by filter
			if createdBy != "" {
				filters = append(filters, dbx.NewExp(
					"created_by = {:created_by}",
					dbx.Params{"created_by": createdBy},
				))
			}

			// Agent name filter
			if agentName != "" {
				filters = append(filters, dbx.NewExp(
					"created_by_agent = {:agent}",
					dbx.Params{"agent": agentName},
				))
			}

			// Is blocked filter
			if isBlocked {
				filters = append(filters, dbx.NewExp(
					"json_array_length(blocked_by) > 0",
				))
			}

			// Not blocked filter
			if notBlocked {
				filters = append(filters, dbx.Or(
					dbx.NewExp("blocked_by IS NULL"),
					dbx.NewExp("blocked_by = '[]'"),
					dbx.NewExp("json_array_length(blocked_by) = 0"),
				))
			}

			// Epic filter
			if epicFilter != "" {
				epicRecord, err := resolveEpic(app, epicFilter)
				if err != nil {
					return out.Error(ExitValidation, fmt.Sprintf("invalid epic filter: %v", err), nil)
				}
				filters = append(filters, dbx.NewExp(
					"epic = {:epic}",
					dbx.Params{"epic": epicRecord.Id},
				))
			}

			// Label filter
			if len(labels) > 0 {
				for _, label := range labels {
					filters = append(filters, dbx.NewExp(
						"labels LIKE {:label}",
						dbx.Params{"label": "%" + label + "%"},
					))
				}
			}

			// Validate mutually exclusive due date flags
			if hasDue && noDue {
				return out.Error(ExitValidation,
					"--has-due and --no-due are mutually exclusive", nil)
			}

			// Due date filters
			if dueBefore != "" {
				date, err := parseDate(dueBefore)
				if err != nil {
					return out.Error(ExitValidation, fmt.Sprintf("invalid --due-before date: %v", err), nil)
				}
				filters = append(filters, dbx.NewExp(
					"due_date <= {:due_before}",
					dbx.Params{"due_before": date},
				))
			}

			if dueAfter != "" {
				date, err := parseDate(dueAfter)
				if err != nil {
					return out.Error(ExitValidation, fmt.Sprintf("invalid --due-after date: %v", err), nil)
				}
				filters = append(filters, dbx.NewExp(
					"due_date >= {:due_after}",
					dbx.Params{"due_after": date},
				))
			}

			if hasDue {
				filters = append(filters, dbx.NewExp("due_date != '' AND due_date IS NOT NULL"))
			}

			if noDue {
				filters = append(filters, dbx.Or(
					dbx.NewExp("due_date = ''"),
					dbx.NewExp("due_date IS NULL"),
				))
			}

			// Execute query using RecordQuery for limit/sort support
			var tasks []*core.Record

			collection, err := app.FindCollectionByNameOrId("tasks")
			if err != nil {
				return out.Error(ExitGeneralError, "tasks collection not found", nil)
			}

			query := app.RecordQuery(collection)

			if len(filters) > 0 {
				combined := dbx.And(filters...)
				query = query.AndWhere(combined)
			}

			// Apply custom sort if specified
			if sort != "" {
				// Valid sortable fields
				validSortFields := map[string]bool{
					"id": true, "title": true, "type": true, "priority": true,
					"column": true, "position": true, "created": true, "updated": true,
					"created_by": true,
				}

				// Parse sort string (e.g., "-priority,position")
				sortFields := strings.Split(sort, ",")
				for _, field := range sortFields {
					field = strings.TrimSpace(field)
					if field == "" {
						continue
					}
					fieldName := field
					if strings.HasPrefix(field, "-") {
						fieldName = field[1:]
					}
					if !validSortFields[fieldName] {
						return out.ErrorWithSuggestion(ExitValidation,
							fmt.Sprintf("invalid sort field '%s'", fieldName),
							fmt.Sprintf("Valid sort fields: id, title, type, priority, column, position, created, updated, created_by"),
							nil)
					}
					if strings.HasPrefix(field, "-") {
						query = query.OrderBy(fieldName + " DESC")
					} else {
						query = query.OrderBy(field + " ASC")
					}
				}
			} else {
				query = query.OrderBy("column ASC", "position ASC")
			}

			// Apply limit
			if limit > 0 {
				query = query.Limit(int64(limit))
			}

			err = query.All(&tasks)
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to list tasks: %v", err), nil)
			}

			// Sort by position within each column (only if no custom sort)
			if sort == "" {
				sortTasksByPosition(tasks)
			}

			// Handle field selection for JSON output
			if jsonOutput && fields != "" {
				out.TasksWithFieldsAndBoards(tasks, strings.Split(fields, ","), boardsMap)
			} else {
				out.TasksWithBoards(tasks, boardsMap)
			}

			return nil
		},
	}

	// Define flags
	cmd.Flags().StringSliceVarP(&columns, "column", "c", nil,
		"Filter by column (repeatable)")
	cmd.Flags().StringSliceVarP(&types, "type", "t", nil,
		"Filter by type (repeatable)")
	cmd.Flags().StringSliceVarP(&priorities, "priority", "p", nil,
		"Filter by priority (repeatable)")
	cmd.Flags().StringVarP(&search, "search", "s", "",
		"Search title (case-insensitive)")
	cmd.Flags().StringVar(&createdBy, "created-by", "",
		"Filter by creator (user, agent, cli)")
	cmd.Flags().StringVar(&agentName, "agent", "",
		"Filter by agent name")
	cmd.Flags().BoolVar(&ready, "ready", false,
		"Show unblocked tasks in todo/backlog (agent-friendly)")
	cmd.Flags().BoolVar(&isBlocked, "is-blocked", false,
		"Show only tasks blocked by others")
	cmd.Flags().BoolVar(&notBlocked, "not-blocked", false,
		"Show only tasks not blocked by others")
	cmd.Flags().StringVar(&fields, "fields", "",
		"Comma-separated fields to include in JSON output")
	cmd.Flags().StringVarP(&epicFilter, "epic", "e", "",
		"Filter by epic (ID or title)")
	cmd.Flags().StringSliceVarP(&labels, "label", "l", nil,
		"Filter by label (repeatable)")
	cmd.Flags().IntVar(&limit, "limit", 0,
		"Maximum number of results (0 = no limit)")
	cmd.Flags().StringVar(&sort, "sort", "",
		"Sort order (e.g., '-priority,position')")
	cmd.Flags().StringVarP(&boardRef, "board", "b", "",
		"Filter by board (name or prefix)")
	cmd.Flags().BoolVar(&allBoards, "all-boards", false,
		"Show tasks from all boards")
	cmd.Flags().StringVar(&dueBefore, "due-before", "",
		"Tasks due before date (inclusive)")
	cmd.Flags().StringVar(&dueAfter, "due-after", "",
		"Tasks due after date (inclusive)")
	cmd.Flags().BoolVar(&hasDue, "has-due", false,
		"Only tasks with due date set")
	cmd.Flags().BoolVar(&noDue, "no-due", false,
		"Only tasks without due date")

	return cmd
}

// buildInFilter creates a SQL IN expression for a list of values.
func buildInFilter(field string, values []string) dbx.Expression {
	if len(values) == 1 {
		return dbx.NewExp(
			fmt.Sprintf("%s = {:val}", field),
			dbx.Params{"val": values[0]},
		)
	}

	// Build IN clause
	placeholders := make([]string, len(values))
	params := dbx.Params{}
	for i, v := range values {
		key := fmt.Sprintf("val%d", i)
		placeholders[i] = "{:" + key + "}"
		params[key] = v
	}

	return dbx.NewExp(
		fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", ")),
		params,
	)
}
