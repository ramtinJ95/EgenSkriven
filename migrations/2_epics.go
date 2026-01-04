package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection := core.NewBaseCollection("epics")

		// Title - required, the epic name
		collection.Fields.Add(&core.TextField{
			Name:     "title",
			Required: true,
			Max:      200,
		})

		// Description - optional longer description of the epic
		collection.Fields.Add(&core.TextField{
			Name: "description",
			Max:  5000,
		})

		// Color - hex color for visual grouping (e.g., "#3B82F6")
		collection.Fields.Add(&core.TextField{
			Name:    "color",
			Pattern: `^#[0-9A-Fa-f]{6}$`,
			Max:     7,
		})

		// Autodate fields for created and updated timestamps
		collection.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
			Hidden:   false,
		})

		collection.Fields.Add(&core.AutodateField{
			Name:     "updated",
			OnCreate: true,
			OnUpdate: true,
			Hidden:   false,
		})

		// API Rules - allow public access (local-first tool, no auth needed)
		collection.ListRule = func() *string { s := ""; return &s }()
		collection.ViewRule = func() *string { s := ""; return &s }()
		collection.CreateRule = func() *string { s := ""; return &s }()
		collection.UpdateRule = func() *string { s := ""; return &s }()
		collection.DeleteRule = func() *string { s := ""; return &s }()

		return app.Save(collection)
	}, func(app core.App) error {
		// Rollback: delete the collection
		collection, err := app.FindCollectionByNameOrId("epics")
		if err != nil {
			return err
		}
		return app.Delete(collection)
	})
}
