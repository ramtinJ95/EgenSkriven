package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Get existing tasks collection
		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		// Add sequence number field
		// This is the numeric part of the display ID (e.g., 123 in WRK-123)
		// Auto-incremented per board when creating tasks
		minVal := float64(1)
		collection.Fields.Add(&core.NumberField{
			Name: "seq",
			Min:  &minVal,
		})

		// Add compound index for efficient sequence queries
		// Used to find max sequence for a board when creating new tasks
		collection.Indexes = append(collection.Indexes,
			"CREATE INDEX idx_tasks_board_seq ON tasks(board, seq)",
		)

		return app.Save(collection)
	}, func(app core.App) error {
		// Rollback: remove seq field
		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}
		collection.Fields.RemoveByName("seq")
		return app.Save(collection)
	})
}
