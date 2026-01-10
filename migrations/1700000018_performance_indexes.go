package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// This migration adds performance-related indexes identified during performance testing.
// These indexes optimize common query patterns:
// - Tasks filtered by column (need_input filter)
// - Comments ordered by creation time within a task (composite index)

func init() {
	m.Register(func(app core.App) error {
		// Add column index to tasks collection for need_input filter optimization
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		// Check if index already exists
		hasColumnIndex := false
		for _, idx := range tasks.Indexes {
			if idx == "CREATE INDEX idx_tasks_column ON tasks (column)" {
				hasColumnIndex = true
				break
			}
		}

		if !hasColumnIndex {
			tasks.Indexes = append(tasks.Indexes, "CREATE INDEX idx_tasks_column ON tasks (column)")
			if err := app.Save(tasks); err != nil {
				return err
			}
		}

		// Add composite index to comments collection for task + created queries
		// This optimizes: SELECT * FROM comments WHERE task = ? ORDER BY created ASC
		comments, err := app.FindCollectionByNameOrId("comments")
		if err != nil {
			return err
		}

		hasCompositeIndex := false
		for _, idx := range comments.Indexes {
			if idx == "CREATE INDEX idx_comments_task_created ON comments (task, created)" {
				hasCompositeIndex = true
				break
			}
		}

		if !hasCompositeIndex {
			comments.Indexes = append(comments.Indexes, "CREATE INDEX idx_comments_task_created ON comments (task, created)")
			if err := app.Save(comments); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		// Rollback: remove the indexes
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err == nil {
			newIndexes := []string{}
			for _, idx := range tasks.Indexes {
				if idx != "CREATE INDEX idx_tasks_column ON tasks (column)" {
					newIndexes = append(newIndexes, idx)
				}
			}
			tasks.Indexes = newIndexes
			app.Save(tasks)
		}

		comments, err := app.FindCollectionByNameOrId("comments")
		if err == nil {
			newIndexes := []string{}
			for _, idx := range comments.Indexes {
				if idx != "CREATE INDEX idx_comments_task_created ON comments (task, created)" {
					newIndexes = append(newIndexes, idx)
				}
			}
			comments.Indexes = newIndexes
			app.Save(comments)
		}

		return nil
	})
}
