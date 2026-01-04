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

		// We need to find the boards collection ID
		boardsCollection, err := app.FindCollectionByNameOrId("boards")
		if err != nil {
			// Boards collection doesn't exist yet - this shouldn't happen
			// if migrations run in order, but handle gracefully
			return err
		}

		// Add board relation field
		// Each task belongs to exactly one board
		collection.Fields.Add(&core.RelationField{
			Name:          "board",
			CollectionId:  boardsCollection.Id,
			Required:      false, // Initially false for migration of existing tasks
			MaxSelect:     1,
			CascadeDelete: false, // Don't delete tasks when board is deleted (handle manually)
		})

		// Add index for faster queries by board
		// This significantly improves performance when listing tasks for a specific board
		collection.Indexes = append(collection.Indexes,
			"CREATE INDEX idx_tasks_board ON tasks(board)",
		)

		return app.Save(collection)
	}, func(app core.App) error {
		// Rollback: remove board field from tasks
		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		// Remove the board field
		collection.Fields.RemoveByName("board")

		// Note: PocketBase handles index cleanup automatically when field is removed

		return app.Save(collection)
	})
}
