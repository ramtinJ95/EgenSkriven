package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Check if collection already exists (idempotency)
		existing, _ := app.FindCollectionByNameOrId("sessions")
		if existing != nil {
			return nil
		}

		// Find tasks collection for relation
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return fmt.Errorf("tasks collection not found: %w", err)
		}

		// Create sessions collection for tracking agent session history
		collection := core.NewBaseCollection("sessions")

		// Task relation (required, cascade delete when task is deleted)
		collection.Fields.Add(&core.RelationField{
			Name:          "task",
			CollectionId:  tasks.Id,
			MaxSelect:     1,
			Required:      true,
			CascadeDelete: true,
		})

		// Tool identifier (required)
		collection.Fields.Add(&core.SelectField{
			Name:     "tool",
			Required: true,
			Values:   []string{"opencode", "claude-code", "codex"},
		})

		// External session reference (UUID or path from the AI tool)
		collection.Fields.Add(&core.TextField{
			Name:     "external_ref",
			Required: true,
			Max:      500,
		})

		// Reference type (uuid or path)
		collection.Fields.Add(&core.SelectField{
			Name:     "ref_type",
			Required: true,
			Values:   []string{"uuid", "path"},
		})

		// Working directory (project directory path)
		collection.Fields.Add(&core.TextField{
			Name:     "working_dir",
			Required: true,
			Max:      1000,
		})

		// Session status
		collection.Fields.Add(&core.SelectField{
			Name:     "status",
			Required: true,
			Values:   []string{"active", "paused", "completed", "abandoned"},
		})

		// Auto-timestamp on creation
		collection.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})

		// End timestamp (optional, set when session ends)
		collection.Fields.Add(&core.DateField{
			Name: "ended_at",
		})

		// Indexes for common queries
		collection.Indexes = []string{
			"CREATE INDEX idx_sessions_task ON sessions (task)",
			"CREATE INDEX idx_sessions_status ON sessions (status)",
			"CREATE INDEX idx_sessions_external_ref ON sessions (external_ref)",
		}

		// API Rules - allow public access (local-first tool, no auth needed)
		collection.ListRule = func() *string { s := ""; return &s }()
		collection.ViewRule = func() *string { s := ""; return &s }()
		collection.CreateRule = func() *string { s := ""; return &s }()
		collection.UpdateRule = func() *string { s := ""; return &s }()
		collection.DeleteRule = func() *string { s := ""; return &s }()

		return app.Save(collection)
	}, func(app core.App) error {
		// Rollback: delete sessions collection
		collection, err := app.FindCollectionByNameOrId("sessions")
		if err != nil {
			return nil // Collection doesn't exist, nothing to rollback
		}
		return app.Delete(collection)
	})
}
