package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		// Check if field already exists (idempotency)
		if tasks.Fields.GetByName("agent_session") != nil {
			return nil
		}

		// Add agent_session JSON field for storing current linked agent session.
		// Schema:
		// {
		//     "tool": "opencode" | "claude-code" | "codex",
		//     "ref": string,           // Session/thread ID (UUID or path)
		//     "ref_type": "uuid" | "path",
		//     "working_dir": string,   // Absolute path to project directory
		//     "linked_at": string      // ISO 8601 timestamp
		// }
		tasks.Fields.Add(&core.JSONField{
			Name:    "agent_session",
			MaxSize: 10000, // 10KB should be plenty for session metadata
		})

		return app.Save(tasks)
	}, func(app core.App) error {
		// Rollback: remove agent_session field
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		field := tasks.Fields.GetByName("agent_session")
		if field == nil {
			return nil // Field doesn't exist, nothing to rollback
		}

		tasks.Fields.RemoveByName("agent_session")
		return app.Save(tasks)
	})
}
