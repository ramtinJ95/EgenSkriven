package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Create boards collection
		collection := core.NewBaseCollection("boards")

		// Name: Human-readable board name
		// Examples: "Work", "Personal", "Side Projects"
		collection.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
			Min:      1,
			Max:      100,
		})

		// Prefix: Short uppercase identifier for task IDs
		// Examples: "WRK", "PER", "SIDE"
		// Must be unique across all boards
		collection.Fields.Add(&core.TextField{
			Name:     "prefix",
			Required: true,
			Min:      1,
			Max:      10,
			// Validated as uppercase in application code
		})

		// Columns: JSON array of column definitions
		// Default: ["backlog", "todo", "in_progress", "review", "done"]
		// Allows boards to have custom workflows
		collection.Fields.Add(&core.JSONField{
			Name:    "columns",
			MaxSize: 10000,
		})

		// Color: Hex color for board accent
		// Used in UI for board identification
		// Examples: "#3B82F6" (blue), "#22C55E" (green)
		collection.Fields.Add(&core.TextField{
			Name: "color",
			Max:  7, // #RRGGBB format
		})

		// Add unique index on prefix
		collection.Indexes = []string{
			"CREATE UNIQUE INDEX idx_boards_prefix ON boards(prefix)",
		}

		if err := app.Save(collection); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: drop the boards collection
		collection, err := app.FindCollectionByNameOrId("boards")
		if err != nil {
			return err
		}
		return app.Delete(collection)
	})
}
