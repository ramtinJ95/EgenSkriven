package migrations

import (
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

		// Add due_date field to tasks
		// DateField stores dates as ISO 8601 strings (YYYY-MM-DD)
		tasks.Fields.Add(&core.DateField{
			Name:     "due_date",
			Required: false,
		})

		return app.Save(tasks)
	}, func(app core.App) error {
		// Rollback: remove due_date field from tasks
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		tasks.Fields.RemoveByName("due_date")

		return app.Save(tasks)
	})
}
