package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection := core.NewBaseCollection("tasks")

		// Title - required, the main task identifier
		collection.Fields.Add(&core.TextField{
			Name:     "title",
			Required: true,
			Max:      500,
		})

		// Description - optional longer text
		collection.Fields.Add(&core.TextField{
			Name: "description",
			Max:  10000,
		})

		// Type - categorizes the task
		collection.Fields.Add(&core.SelectField{
			Name:     "type",
			Required: true,
			Values:   []string{"bug", "feature", "chore"},
		})

		// Priority - importance level
		collection.Fields.Add(&core.SelectField{
			Name:     "priority",
			Required: true,
			Values:   []string{"low", "medium", "high", "urgent"},
		})

		// Column - kanban board column (status)
		collection.Fields.Add(&core.SelectField{
			Name:     "column",
			Required: true,
			Values:   []string{"backlog", "todo", "in_progress", "review", "done"},
		})

		// Position - order within column (fractional for easy reordering)
		collection.Fields.Add(&core.NumberField{
			Name:     "position",
			Required: true,
			Min:      floatPtr(0),
		})

		// Labels - array of string tags
		collection.Fields.Add(&core.JSONField{
			Name:    "labels",
			MaxSize: 10000,
		})

		// Blocked by - array of task IDs that block this task
		collection.Fields.Add(&core.JSONField{
			Name:    "blocked_by",
			MaxSize: 10000,
		})

		// Created by - who created this task
		collection.Fields.Add(&core.SelectField{
			Name:     "created_by",
			Required: true,
			Values:   []string{"user", "agent", "cli"},
		})

		// Created by agent - optional agent identifier
		collection.Fields.Add(&core.TextField{
			Name: "created_by_agent",
			Max:  100,
		})

		// History - activity log as JSON array
		collection.Fields.Add(&core.JSONField{
			Name:    "history",
			MaxSize: 100000,
		})

		return app.Save(collection)
	}, func(app core.App) error {
		// Rollback: delete the collection
		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}
		return app.Delete(collection)
	})
}

// Helper function for pointer to float
func floatPtr(f float64) *float64 {
	return &f
}
