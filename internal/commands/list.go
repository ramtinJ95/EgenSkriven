package commands

import (
	"fmt"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
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
  egenskriven list --json --fields id,title,column`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Build filter expressions
			var filters []dbx.Expression

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

			// Execute query
			var tasks []*core.Record
			var err error

			if len(filters) > 0 {
				combined := dbx.And(filters...)
				tasks, err = app.FindAllRecords("tasks", combined)
			} else {
				tasks, err = app.FindAllRecords("tasks")
			}

			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to list tasks: %v", err), nil)
			}

			// Sort by position within each column
			sortTasksByPosition(tasks)

			// Handle field selection for JSON output
			if jsonOutput && fields != "" {
				out.TasksWithFields(tasks, strings.Split(fields, ","))
			} else {
				out.Tasks(tasks)
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
