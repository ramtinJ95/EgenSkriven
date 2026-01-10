package commands

import (
	"fmt"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// ========== Benchmark Helper Functions ==========

// setupBenchmarkEnv creates a test app with tasks and comments collections.
// Returns the app and a cleanup function.
func setupBenchmarkEnv(b *testing.B) *pocketbase.PocketBase {
	b.Helper()

	// Use testing.T wrapper for testutil compatibility
	t := &testing.T{}
	app := testutil.NewTestApp(t)

	// Set up collections
	setupBenchmarkTasksCollection(b, app)
	setupBenchmarkCommentsCollection(b, app)

	return app
}

// setupBenchmarkTasksCollection creates the tasks collection for benchmarks.
func setupBenchmarkTasksCollection(b *testing.B, app *pocketbase.PocketBase) {
	b.Helper()

	_, err := app.FindCollectionByNameOrId("tasks")
	if err == nil {
		return
	}

	collection := core.NewBaseCollection("tasks")
	collection.Fields.Add(&core.TextField{Name: "title", Required: true})
	collection.Fields.Add(&core.TextField{Name: "description"})
	collection.Fields.Add(&core.SelectField{
		Name:     "type",
		Required: true,
		Values:   []string{"bug", "feature", "chore"},
	})
	collection.Fields.Add(&core.SelectField{
		Name:     "priority",
		Required: true,
		Values:   []string{"low", "medium", "high", "urgent"},
	})
	collection.Fields.Add(&core.SelectField{
		Name:     "column",
		Required: true,
		Values:   []string{"backlog", "todo", "in_progress", "need_input", "review", "done"},
	})
	collection.Fields.Add(&core.NumberField{Name: "position", Required: true})
	collection.Fields.Add(&core.JSONField{Name: "labels"})
	collection.Fields.Add(&core.JSONField{Name: "blocked_by"})
	collection.Fields.Add(&core.SelectField{
		Name:     "created_by",
		Required: true,
		Values:   []string{"user", "agent", "cli"},
	})
	collection.Fields.Add(&core.TextField{Name: "created_by_agent"})
	collection.Fields.Add(&core.JSONField{Name: "history"})
	collection.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	if err := app.Save(collection); err != nil {
		b.Fatalf("failed to create tasks collection: %v", err)
	}
}

// setupBenchmarkCommentsCollection creates the comments collection for benchmarks.
func setupBenchmarkCommentsCollection(b *testing.B, app *pocketbase.PocketBase) {
	b.Helper()

	_, err := app.FindCollectionByNameOrId("comments")
	if err == nil {
		return
	}

	collection := core.NewBaseCollection("comments")
	collection.Fields.Add(&core.TextField{Name: "task", Required: true})
	collection.Fields.Add(&core.TextField{Name: "content", Required: true})
	collection.Fields.Add(&core.SelectField{
		Name:     "author_type",
		Required: true,
		Values:   []string{"human", "agent"},
	})
	collection.Fields.Add(&core.TextField{Name: "author_id"})
	collection.Fields.Add(&core.JSONField{Name: "metadata"})
	collection.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})

	if err := app.Save(collection); err != nil {
		b.Fatalf("failed to create comments collection: %v", err)
	}
}

// createBenchmarkTask creates a task for benchmarking.
func createBenchmarkTask(b *testing.B, app *pocketbase.PocketBase, title, column string) *core.Record {
	b.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		b.Fatalf("tasks collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("type", "feature")
	record.Set("priority", "medium")
	record.Set("column", column)
	record.Set("position", 1000.0)
	record.Set("labels", []string{})
	record.Set("blocked_by", []string{})
	record.Set("created_by", "cli")
	record.Set("history", []map[string]any{})

	if err := app.Save(record); err != nil {
		b.Fatalf("failed to create task: %v", err)
	}

	return record
}

// createBenchmarkComment creates a comment for benchmarking.
func createBenchmarkComment(b *testing.B, app *pocketbase.PocketBase, taskId, content, authorType, authorId string) *core.Record {
	b.Helper()

	collection, err := app.FindCollectionByNameOrId("comments")
	if err != nil {
		b.Fatalf("comments collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("task", taskId)
	record.Set("content", content)
	record.Set("author_type", authorType)
	record.Set("author_id", authorId)
	record.Set("metadata", map[string]any{})

	if err := app.Save(record); err != nil {
		b.Fatalf("failed to create comment: %v", err)
	}

	return record
}

// seedBenchmarkComments creates multiple comments for a task.
func seedBenchmarkComments(b *testing.B, app *pocketbase.PocketBase, taskId string, count int) {
	b.Helper()

	collection, err := app.FindCollectionByNameOrId("comments")
	if err != nil {
		b.Fatalf("comments collection not found: %v", err)
	}

	for i := 0; i < count; i++ {
		record := core.NewRecord(collection)
		record.Set("task", taskId)
		record.Set("content", fmt.Sprintf("Benchmark comment %d with some content to simulate real usage", i))
		record.Set("author_type", "human")
		record.Set("author_id", "benchmark-user")
		record.Set("metadata", map[string]any{})

		if err := app.Save(record); err != nil {
			b.Fatalf("failed to create comment %d: %v", i, err)
		}
	}
}

// ========== Operation Benchmarks ==========

// BenchmarkBlockTask measures the performance of the block task operation.
// Target: < 100ms
// Tests: atomic transaction (move task + create comment)
func BenchmarkBlockTask(b *testing.B) {
	app := setupBenchmarkEnv(b)

	commentsCollection, err := app.FindCollectionByNameOrId("comments")
	if err != nil {
		b.Fatalf("comments collection not found: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Create a fresh task for each iteration
		task := createBenchmarkTask(b, app, fmt.Sprintf("Block benchmark task %d", i), "todo")
		currentColumn := task.GetString("column")
		b.StartTimer()

		// Execute block operation (same as block.go:103-138)
		err := app.RunInTransaction(func(txApp core.App) error {
			task.Set("column", "need_input")

			addHistoryEntry(task, "blocked", "benchmark-agent", map[string]any{
				"column": map[string]any{
					"from": currentColumn,
					"to":   "need_input",
				},
				"reason": "Benchmark question?",
			})

			if err := txApp.Save(task); err != nil {
				return err
			}

			comment := core.NewRecord(commentsCollection)
			comment.Set("task", task.Id)
			comment.Set("content", "Benchmark question?")
			comment.Set("author_type", "agent")
			comment.Set("author_id", "benchmark-agent")
			comment.Set("metadata", map[string]any{"action": "block_question"})

			return txApp.Save(comment)
		})

		if err != nil {
			b.Fatalf("block operation failed: %v", err)
		}
	}
}

// BenchmarkAddComment measures the performance of adding a comment.
// Target: < 100ms
// Tests: comment creation with mention extraction
func BenchmarkAddComment(b *testing.B) {
	app := setupBenchmarkEnv(b)

	task := createBenchmarkTask(b, app, "Comment benchmark task", "in_progress")

	commentsCollection, err := app.FindCollectionByNameOrId("comments")
	if err != nil {
		b.Fatalf("comments collection not found: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate comment creation (same as comment.go:102-113)
		content := fmt.Sprintf("@agent This is benchmark comment %d with a mention", i)
		mentions := extractMentions(content)

		comment := core.NewRecord(commentsCollection)
		comment.Set("task", task.Id)
		comment.Set("content", content)
		comment.Set("author_type", "human")
		comment.Set("author_id", "benchmark-user")
		comment.Set("metadata", map[string]any{"mentions": mentions})

		if err := app.Save(comment); err != nil {
			b.Fatalf("comment creation failed: %v", err)
		}
	}
}

// BenchmarkListComments measures the performance of listing comments for a task.
// Target: < 200ms
// Tests: querying comments with sorting
func BenchmarkListComments(b *testing.B) {
	app := setupBenchmarkEnv(b)

	task := createBenchmarkTask(b, app, "List comments benchmark task", "in_progress")
	// Seed with 100 comments (target is handling up to 1000+)
	seedBenchmarkComments(b, app, task.Id, 100)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate comments listing (same as comments.go:73-83)
		filter := fmt.Sprintf("task = '%s'", task.Id)
		_, err := app.FindRecordsByFilter(
			"comments",
			filter,
			"+created",
			0,
			0,
		)
		if err != nil {
			b.Fatalf("list comments failed: %v", err)
		}
	}
}

// BenchmarkListNeedInput measures the performance of listing tasks needing input.
// Target: < 100ms
// Tests: indexed query on column field
func BenchmarkListNeedInput(b *testing.B) {
	app := setupBenchmarkEnv(b)

	// Create some tasks in various columns
	for i := 0; i < 50; i++ {
		column := "in_progress"
		if i%5 == 0 {
			column = "need_input"
		}
		createBenchmarkTask(b, app, fmt.Sprintf("Need input benchmark task %d", i), column)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate list --need-input (same as list.go:118-120)
		_, err := app.FindRecordsByFilter(
			"tasks",
			"column = 'need_input'",
			"-updated",
			0,
			0,
		)
		if err != nil {
			b.Fatalf("list need_input failed: %v", err)
		}
	}
}

// ========== Scaling Benchmarks ==========

// BenchmarkCommentsScaling measures performance with large comment counts.
// Target: Handle 1000+ comments per task
func BenchmarkCommentsScaling(b *testing.B) {
	commentCounts := []int{10, 100, 500, 1000}

	for _, count := range commentCounts {
		b.Run(fmt.Sprintf("comments_%d", count), func(b *testing.B) {
			app := setupBenchmarkEnv(b)

			task := createBenchmarkTask(b, app, "Scaling benchmark task", "in_progress")
			seedBenchmarkComments(b, app, task.Id, count)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				filter := fmt.Sprintf("task = '%s'", task.Id)
				_, err := app.FindRecordsByFilter(
					"comments",
					filter,
					"+created",
					0,
					0,
				)
				if err != nil {
					b.Fatalf("list comments failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkTasksScaling measures performance with large task counts.
// Target: Handle 10000+ tasks per board
func BenchmarkTasksScaling(b *testing.B) {
	taskCounts := []int{100, 1000, 5000}

	for _, count := range taskCounts {
		b.Run(fmt.Sprintf("tasks_%d", count), func(b *testing.B) {
			app := setupBenchmarkEnv(b)

			// Create tasks
			for i := 0; i < count; i++ {
				columns := []string{"backlog", "todo", "in_progress", "review", "done"}
				column := columns[i%len(columns)]
				createBenchmarkTask(b, app, fmt.Sprintf("Scaling task %d", i), column)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Query all tasks in a column (simulates list command)
				_, err := app.FindRecordsByFilter(
					"tasks",
					"column = 'in_progress'",
					"-updated",
					0,
					0,
				)
				if err != nil {
					b.Fatalf("list tasks failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkLargeCommentContent measures performance with large comment content.
// Note: The comments collection has a 5000 character limit, so we test within that limit.
// For larger content support, the schema would need to be updated.
func BenchmarkLargeCommentContent(b *testing.B) {
	contentSizes := []int{1024, 2048, 4096} // 1KB, 2KB, 4KB (within 5000 char limit)

	for _, size := range contentSizes {
		b.Run(fmt.Sprintf("content_%dKB", size/1024), func(b *testing.B) {
			app := setupBenchmarkEnv(b)

			task := createBenchmarkTask(b, app, "Large content benchmark task", "in_progress")

			commentsCollection, err := app.FindCollectionByNameOrId("comments")
			if err != nil {
				b.Fatalf("comments collection not found: %v", err)
			}

			// Generate large content
			content := make([]byte, size)
			for i := range content {
				content[i] = byte('a' + (i % 26))
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				comment := core.NewRecord(commentsCollection)
				comment.Set("task", task.Id)
				comment.Set("content", string(content))
				comment.Set("author_type", "human")
				comment.Set("author_id", "benchmark-user")
				comment.Set("metadata", map[string]any{})

				if err := app.Save(comment); err != nil {
					b.Fatalf("comment creation failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkConcurrentCommentReads measures concurrent read performance.
// Target: Support 100+ concurrent operations
func BenchmarkConcurrentCommentReads(b *testing.B) {
	app := setupBenchmarkEnv(b)

	task := createBenchmarkTask(b, app, "Concurrent read benchmark task", "in_progress")
	seedBenchmarkComments(b, app, task.Id, 100)

	filter := fmt.Sprintf("task = '%s'", task.Id)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := app.FindRecordsByFilter(
				"comments",
				filter,
				"+created",
				0,
				0,
			)
			if err != nil {
				b.Errorf("concurrent read failed: %v", err)
			}
		}
	})
}

// BenchmarkMentionExtraction measures the performance of extracting mentions.
func BenchmarkMentionExtraction(b *testing.B) {
	contents := []struct {
		name    string
		content string
	}{
		{"no_mentions", "This is a comment without any mentions"},
		{"single_mention", "@agent please review this"},
		{"multiple_mentions", "@agent and @senior-dev please check @reviewer"},
		{"complex_mentions", "Hey @agent, can you ask @product-manager about @designer's feedback?"},
	}

	for _, tc := range contents {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				extractMentions(tc.content)
			}
		})
	}
}

// BenchmarkHistoryEntry measures the performance of adding history entries.
func BenchmarkHistoryEntry(b *testing.B) {
	app := setupBenchmarkEnv(b)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		task := createBenchmarkTask(b, app, fmt.Sprintf("History benchmark task %d", i), "todo")
		b.StartTimer()

		addHistoryEntry(task, "blocked", "benchmark-agent", map[string]any{
			"column": map[string]any{
				"from": "todo",
				"to":   "need_input",
			},
			"reason": "Benchmark question?",
		})

		if err := app.Save(task); err != nil {
			b.Fatalf("failed to save task with history: %v", err)
		}
	}
}

// BenchmarkTaskLookupByID measures the performance of finding a task by ID.
func BenchmarkTaskLookupByID(b *testing.B) {
	app := setupBenchmarkEnv(b)

	// Create multiple tasks and collect IDs
	taskIds := make([]string, 100)
	for i := 0; i < 100; i++ {
		task := createBenchmarkTask(b, app, fmt.Sprintf("Lookup benchmark task %d", i), "in_progress")
		taskIds[i] = task.Id
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		taskId := taskIds[i%len(taskIds)]
		_, err := app.FindRecordById("tasks", taskId)
		if err != nil {
			b.Fatalf("task lookup failed: %v", err)
		}
	}
}

// BenchmarkCommentsByTaskWithIndex measures query performance using the task index.
func BenchmarkCommentsByTaskWithIndex(b *testing.B) {
	app := setupBenchmarkEnv(b)

	// Create multiple tasks with comments
	taskIds := make([]string, 10)
	for i := 0; i < 10; i++ {
		task := createBenchmarkTask(b, app, fmt.Sprintf("Index benchmark task %d", i), "in_progress")
		taskIds[i] = task.Id
		seedBenchmarkComments(b, app, task.Id, 50)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		taskId := taskIds[i%len(taskIds)]
		// Using parameterized query (safe from SQL injection)
		_, err := app.FindRecordsByFilter(
			"comments",
			"task = {:taskId}",
			"+created",
			0,
			0,
			dbx.Params{"taskId": taskId},
		)
		if err != nil {
			b.Fatalf("comments query failed: %v", err)
		}
	}
}
