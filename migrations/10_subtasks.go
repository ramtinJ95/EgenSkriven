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

		// Add parent field - self-referential relation
		// A task can have one parent task (for sub-tasks)
		tasks.Fields.Add(&core.RelationField{
			Name:          "parent",
			Required:      false,
			CollectionId:  tasks.Id, // Self-reference
			MaxSelect:     1,
			CascadeDelete: false, // Keep orphaned sub-tasks if parent deleted
		})

		return app.Save(tasks)
	}, func(app core.App) error {
		// Rollback: remove parent field from tasks
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		tasks.Fields.RemoveByName("parent")

		return app.Save(tasks)
	})
}
