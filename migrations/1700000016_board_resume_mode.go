package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		boards, err := app.FindCollectionByNameOrId("boards")
		if err != nil {
			return err
		}

		// Check if field already exists (idempotency)
		if boards.Fields.GetByName("resume_mode") != nil {
			return nil
		}

		// Add resume_mode select field for configuring how blocked tasks are resumed
		// - "manual": Print command for user to copy/paste
		// - "command": User runs `egenskriven resume <task>` to spawn tool
		// - "auto": Auto-resume when @agent is mentioned in comment
		boards.Fields.Add(&core.SelectField{
			Name:     "resume_mode",
			Required: false,
			Values:   []string{"manual", "command", "auto"},
		})

		return app.Save(boards)
	}, func(app core.App) error {
		// Rollback: remove resume_mode field
		boards, err := app.FindCollectionByNameOrId("boards")
		if err != nil {
			return err
		}

		field := boards.Fields.GetByName("resume_mode")
		if field == nil {
			return nil // Field doesn't exist, nothing to rollback
		}

		boards.Fields.RemoveByName("resume_mode")
		return app.Save(boards)
	})
}
