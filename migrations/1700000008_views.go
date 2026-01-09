package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Create views collection for saved filter views
		collection := core.NewBaseCollection("views")

		// Name: Human-readable view name
		// Examples: "High Priority Bugs", "My Tasks", "Due This Week"
		collection.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
			Min:      1,
			Max:      100,
		})

		// Board: Reference to the board this view belongs to
		boardsCollection, err := app.FindCollectionByNameOrId("boards")
		if err != nil {
			return err
		}
		collection.Fields.Add(&core.RelationField{
			Name:         "board",
			CollectionId: boardsCollection.Id,
			Required:     true,
			MaxSelect:    1,
		})

		// Filters: JSON array of filter definitions
		// Each filter has: id, field, operator, value
		// Examples: [{"id":"abc","field":"priority","operator":"is","value":"high"}]
		collection.Fields.Add(&core.JSONField{
			Name:    "filters",
			MaxSize: 10000,
		})

		// Display: JSON object with display options
		// Contains: viewMode, density, visibleFields, groupBy
		collection.Fields.Add(&core.JSONField{
			Name:    "display",
			MaxSize: 5000,
		})

		// IsFavorite: Whether this view is marked as favorite
		// Favorite views appear at the top of the views list
		collection.Fields.Add(&core.BoolField{
			Name: "is_favorite",
		})

		// MatchMode: How multiple filters are combined
		// "all" = AND (all filters must match)
		// "any" = OR (at least one filter must match)
		// Required is true to ensure a value is always set; frontend defaults to "all"
		collection.Fields.Add(&core.SelectField{
			Name:     "match_mode",
			Values:   []string{"all", "any"},
			Required: true,
		})

		// API Rules - allow public access (local-first tool, no auth needed)
		collection.ListRule = func() *string { s := ""; return &s }()
		collection.ViewRule = func() *string { s := ""; return &s }()
		collection.CreateRule = func() *string { s := ""; return &s }()
		collection.UpdateRule = func() *string { s := ""; return &s }()
		collection.DeleteRule = func() *string { s := ""; return &s }()

		if err := app.Save(collection); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: drop the views collection
		collection, err := app.FindCollectionByNameOrId("views")
		if err != nil {
			return nil // Collection doesn't exist, nothing to rollback
		}
		return app.Delete(collection)
	})
}
