package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Find tasks collection
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		// Get the column field and update its values
		columnField := tasks.Fields.GetByName("column")
		if columnField == nil {
			return fmt.Errorf("column field not found on tasks collection")
		}

		selectField, ok := columnField.(*core.SelectField)
		if !ok {
			return fmt.Errorf("column field is not a select field")
		}

		// Check if need_input is already present (idempotency)
		for _, v := range selectField.Values {
			if v == "need_input" {
				return nil // Already migrated
			}
		}

		// Add need_input to valid values (after in_progress, before review)
		selectField.Values = []string{
			"backlog",
			"todo",
			"in_progress",
			"need_input", // NEW: Agent is blocked, waiting for human input
			"review",
			"done",
		}

		return app.Save(tasks)
	}, func(app core.App) error {
		// Rollback: remove need_input from valid values
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		columnField := tasks.Fields.GetByName("column")
		if columnField == nil {
			return nil // Field doesn't exist, nothing to rollback
		}

		selectField, ok := columnField.(*core.SelectField)
		if !ok {
			return nil
		}

		// Restore original values (without need_input)
		// Note: If any tasks are in need_input state, this rollback will fail
		// That's intentional - don't rollback if data exists
		selectField.Values = []string{
			"backlog",
			"todo",
			"in_progress",
			"review",
			"done",
		}

		return app.Save(tasks)
	})
}
