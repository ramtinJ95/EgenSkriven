package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Check if collection already exists (idempotency)
		existing, _ := app.FindCollectionByNameOrId("comments")
		if existing != nil {
			return nil
		}

		// Find tasks collection for relation
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return fmt.Errorf("tasks collection not found: %w", err)
		}

		// Create comments collection
		collection := core.NewBaseCollection("comments")

		// Task relation (required, cascade delete when task is deleted)
		collection.Fields.Add(&core.RelationField{
			Name:          "task",
			CollectionId:  tasks.Id,
			MaxSelect:     1,
			Required:      true,
			CascadeDelete: true,
		})

		// Comment content (required, max 50KB for longer discussions)
		collection.Fields.Add(&core.TextField{
			Name:     "content",
			Required: true,
			Max:      50000,
		})

		// Author type (required) - distinguishes human from agent comments
		collection.Fields.Add(&core.SelectField{
			Name:     "author_type",
			Required: true,
			Values:   []string{"human", "agent"},
		})

		// Author identifier (optional) - username, agent name, etc.
		collection.Fields.Add(&core.TextField{
			Name:     "author_id",
			Required: false,
			Max:      200,
		})

		// Metadata JSON (optional) - for mentions, session refs, etc.
		// Schema example:
		// {
		//     "mentions": ["@agent"],
		//     "session_ref": "550e8400-..."
		// }
		collection.Fields.Add(&core.JSONField{
			Name:    "metadata",
			MaxSize: 50000,
		})

		// Auto-timestamp on creation
		collection.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})

		// Indexes for common queries
		collection.Indexes = []string{
			"CREATE INDEX idx_comments_task ON comments (task)",
			"CREATE INDEX idx_comments_created ON comments (created)",
		}

		// API Rules - allow public access (local-first tool, no auth needed)
		collection.ListRule = func() *string { s := ""; return &s }()
		collection.ViewRule = func() *string { s := ""; return &s }()
		collection.CreateRule = func() *string { s := ""; return &s }()
		collection.UpdateRule = func() *string { s := ""; return &s }()
		collection.DeleteRule = func() *string { s := ""; return &s }()

		return app.Save(collection)
	}, func(app core.App) error {
		// Rollback: delete comments collection
		collection, err := app.FindCollectionByNameOrId("comments")
		if err != nil {
			return nil // Collection doesn't exist, nothing to rollback
		}
		return app.Delete(collection)
	})
}
