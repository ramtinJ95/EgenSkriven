package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Fix: Add API rules to boards collection
		// This was missing in the original migration
		collection, err := app.FindCollectionByNameOrId("boards")
		if err != nil {
			return err
		}

		// API Rules - allow public access (local-first tool, no auth needed)
		// Empty string means "allow all"
		collection.ListRule = func() *string { s := ""; return &s }()
		collection.ViewRule = func() *string { s := ""; return &s }()
		collection.CreateRule = func() *string { s := ""; return &s }()
		collection.UpdateRule = func() *string { s := ""; return &s }()
		collection.DeleteRule = func() *string { s := ""; return &s }()

		return app.Save(collection)
	}, func(app core.App) error {
		// Rollback: remove API rules (set to nil = superuser only)
		collection, err := app.FindCollectionByNameOrId("boards")
		if err != nil {
			return err
		}

		collection.ListRule = nil
		collection.ViewRule = nil
		collection.CreateRule = nil
		collection.UpdateRule = nil
		collection.DeleteRule = nil

		return app.Save(collection)
	})
}
