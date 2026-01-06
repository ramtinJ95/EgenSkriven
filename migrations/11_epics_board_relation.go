package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Find epics collection
		epics, err := app.FindCollectionByNameOrId("epics")
		if err != nil {
			return err
		}

		// Find boards collection
		boards, err := app.FindCollectionByNameOrId("boards")
		if err != nil {
			return err
		}

		// Delete all existing epics (they are global, not board-scoped)
		existingEpics, err := app.FindAllRecords("epics")
		if err == nil {
			for _, epic := range existingEpics {
				if err := app.Delete(epic); err != nil {
					return err
				}
			}
		}

		// Add board relation field to epics
		// Each epic belongs to exactly one board
		epics.Fields.Add(&core.RelationField{
			Name:          "board",
			CollectionId:  boards.Id,
			MaxSelect:     1,
			Required:      true,
			CascadeDelete: true, // Delete epics when board is deleted
		})

		return app.Save(epics)
	}, func(app core.App) error {
		// Rollback: remove board field from epics
		epics, err := app.FindCollectionByNameOrId("epics")
		if err != nil {
			return err
		}

		epics.Fields.RemoveByName("board")

		return app.Save(epics)
	})
}
