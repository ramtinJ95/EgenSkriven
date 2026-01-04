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

		// Find epics collection
		epics, err := app.FindCollectionByNameOrId("epics")
		if err != nil {
			return err
		}

		// Add epic relation field to tasks
		tasks.Fields.Add(&core.RelationField{
			Name:          "epic",
			CollectionId:  epics.Id,
			MaxSelect:     1,
			CascadeDelete: false, // Tasks remain when epic is deleted
		})

		return app.Save(tasks)
	}, func(app core.App) error {
		// Rollback: remove epic field from tasks
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		tasks.Fields.RemoveByName("epic")

		return app.Save(tasks)
	})
}
