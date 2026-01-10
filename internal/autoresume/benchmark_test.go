package autoresume

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// ========== Benchmark Helper Functions ==========

// setupBenchmarkApp creates a test app with collections for benchmarking.
func setupBenchmarkApp(b *testing.B) *pocketbase.PocketBase {
	b.Helper()

	t := &testing.T{}
	app := testutil.NewTestApp(t)

	// Create boards collection
	if _, err := app.FindCollectionByNameOrId("boards"); err != nil {
		boards := core.NewBaseCollection("boards")
		boards.Fields.Add(&core.TextField{Name: "name", Required: true})
		boards.Fields.Add(&core.TextField{Name: "prefix", Required: true})
		boards.Fields.Add(&core.TextField{Name: "resume_mode"})
		boards.Fields.Add(&core.JSONField{Name: "columns"})
		boards.Fields.Add(&core.NumberField{Name: "next_seq"})
		if err := app.Save(boards); err != nil {
			b.Fatalf("failed to create boards collection: %v", err)
		}
	}

	// Create tasks collection
	if _, err := app.FindCollectionByNameOrId("tasks"); err != nil {
		tasks := core.NewBaseCollection("tasks")
		tasks.Fields.Add(&core.TextField{Name: "title", Required: true})
		tasks.Fields.Add(&core.TextField{Name: "column"})
		tasks.Fields.Add(&core.TextField{Name: "board"})
		tasks.Fields.Add(&core.JSONField{Name: "agent_session"})
		tasks.Fields.Add(&core.JSONField{Name: "history"})
		tasks.Fields.Add(&core.NumberField{Name: "seq"})
		tasks.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})
		if err := app.Save(tasks); err != nil {
			b.Fatalf("failed to create tasks collection: %v", err)
		}
	}

	// Create comments collection
	if _, err := app.FindCollectionByNameOrId("comments"); err != nil {
		comments := core.NewBaseCollection("comments")
		comments.Fields.Add(&core.TextField{Name: "task", Required: true})
		comments.Fields.Add(&core.TextField{Name: "content", Required: true})
		comments.Fields.Add(&core.TextField{Name: "author_type", Required: true})
		comments.Fields.Add(&core.TextField{Name: "author_id"})
		comments.Fields.Add(&core.JSONField{Name: "metadata"})
		comments.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})
		if err := app.Save(comments); err != nil {
			b.Fatalf("failed to create comments collection: %v", err)
		}
	}

	return app
}

// createBenchmarkBoard creates a board for benchmarking.
func createBenchmarkBoard(b *testing.B, app *pocketbase.PocketBase, prefix, resumeMode string) *core.Record {
	b.Helper()

	collection, err := app.FindCollectionByNameOrId("boards")
	if err != nil {
		b.Fatalf("boards collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("name", "Benchmark Board")
	record.Set("prefix", prefix)
	record.Set("resume_mode", resumeMode)
	record.Set("columns", []string{"backlog", "todo", "in_progress", "need_input", "done"})
	record.Set("next_seq", 1)

	if err := app.Save(record); err != nil {
		b.Fatalf("failed to create benchmark board: %v", err)
	}

	return record
}

// createBenchmarkTask creates a task for benchmarking.
func createBenchmarkTask(b *testing.B, app *pocketbase.PocketBase, boardId, column string, withSession bool) *core.Record {
	b.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		b.Fatalf("tasks collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("title", "Benchmark Task")
	record.Set("board", boardId)
	record.Set("column", column)
	record.Set("seq", 1)
	record.Set("history", []map[string]any{})

	if withSession {
		record.Set("agent_session", map[string]any{
			"tool":        "claude",
			"ref":         "benchmark-session-123",
			"ref_type":    "uuid",
			"working_dir": "/tmp",
		})
	}

	if err := app.Save(record); err != nil {
		b.Fatalf("failed to create benchmark task: %v", err)
	}

	return record
}

// createBenchmarkComment creates a comment for benchmarking.
func createBenchmarkComment(b *testing.B, app *pocketbase.PocketBase, taskId, content, authorType string, mentions []string) *core.Record {
	b.Helper()

	collection, err := app.FindCollectionByNameOrId("comments")
	if err != nil {
		b.Fatalf("comments collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("task", taskId)
	record.Set("content", content)
	record.Set("author_type", authorType)
	record.Set("author_id", "benchmark-user")
	record.Set("metadata", map[string]any{"mentions": mentions})

	if err := app.Save(record); err != nil {
		b.Fatalf("failed to create benchmark comment: %v", err)
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
		authorType := "human"
		if i%2 == 0 {
			authorType = "agent"
		}

		record := core.NewRecord(collection)
		record.Set("task", taskId)
		record.Set("content", fmt.Sprintf("Benchmark comment %d with content", i))
		record.Set("author_type", authorType)
		record.Set("author_id", "benchmark-user")
		record.Set("metadata", map[string]any{})

		if err := app.Save(record); err != nil {
			b.Fatalf("failed to create benchmark comment %d: %v", i, err)
		}
	}
}

// ========== Core Operation Benchmarks ==========

// BenchmarkCheckAndResume measures the full check path performance.
// Target: < 1s for auto-resume trigger
// Tests: complete CheckAndResume flow
func BenchmarkCheckAndResume(b *testing.B) {
	app := setupBenchmarkApp(b)
	board := createBenchmarkBoard(b, app, "TEST", "auto")
	task := createBenchmarkTask(b, app, board.Id, "need_input", true)

	service := NewService(app)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Create a fresh comment for each iteration
		comment := createBenchmarkComment(b, app, task.Id, "@agent continue", "human", []string{"@agent"})
		// Reset task to need_input after auto-resume may have moved it
		task.Set("column", "need_input")
		app.Save(task)
		b.StartTimer()

		// Note: This won't actually execute the resume command, just the check path
		_ = service.CheckAndResume(comment)
	}
}

// BenchmarkCheckAndResume_EarlyExit measures early exit paths.
// Tests: scenarios where auto-resume should NOT trigger
func BenchmarkCheckAndResume_EarlyExit(b *testing.B) {
	scenarios := []struct {
		name       string
		authorType string
		column     string
		resumeMode string
		withMention bool
		withSession bool
	}{
		{"agent_comment", "agent", "need_input", "auto", true, true},
		{"no_mention", "human", "need_input", "auto", false, true},
		{"wrong_column", "human", "in_progress", "auto", true, true},
		{"manual_mode", "human", "need_input", "manual", true, true},
		{"no_session", "human", "need_input", "auto", true, false},
	}

	for _, sc := range scenarios {
		b.Run(sc.name, func(b *testing.B) {
			app := setupBenchmarkApp(b)
			board := createBenchmarkBoard(b, app, "TEST", sc.resumeMode)
			task := createBenchmarkTask(b, app, board.Id, sc.column, sc.withSession)

			var mentions []string
			if sc.withMention {
				mentions = []string{"@agent"}
			}
			comment := createBenchmarkComment(b, app, task.Id, "@agent test", sc.authorType, mentions)

			service := NewService(app)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = service.CheckAndResume(comment)
			}
		})
	}
}

// BenchmarkHasAgentMention measures mention detection performance.
func BenchmarkHasAgentMention(b *testing.B) {
	app := setupBenchmarkApp(b)
	collection, _ := app.FindCollectionByNameOrId("comments")
	taskCollection, _ := app.FindCollectionByNameOrId("tasks")

	// Create a task
	task := core.NewRecord(taskCollection)
	task.Set("title", "Test")
	task.Set("column", "todo")
	app.Save(task)

	testCases := []struct {
		name     string
		metadata map[string]any
	}{
		{"nil_metadata", nil},
		{"empty_metadata", map[string]any{}},
		{"no_mentions", map[string]any{"mentions": []any{}}},
		{"single_agent", map[string]any{"mentions": []any{"@agent"}}},
		{"multiple_mentions", map[string]any{"mentions": []any{"@user", "@agent", "@admin"}}},
		{"many_mentions", map[string]any{"mentions": []any{"@a", "@b", "@c", "@d", "@e", "@f", "@g", "@h", "@i", "@agent"}}},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			record := core.NewRecord(collection)
			record.Set("task", task.Id)
			record.Set("content", "test")
			record.Set("author_type", "human")
			record.Set("metadata", tc.metadata)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				hasAgentMention(record)
			}
		})
	}
}

// BenchmarkHasAgentMention_JSONParsing measures JSON parsing overhead.
func BenchmarkHasAgentMention_JSONParsing(b *testing.B) {
	// Simulate the JSON parsing path that happens with types.JSONRaw
	metadataJSON := `{"mentions":["@user","@agent","@admin"]}`

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var metaMap map[string]any
		json.Unmarshal([]byte(metadataJSON), &metaMap)

		mentions, _ := metaMap["mentions"].([]any)
		for _, m := range mentions {
			if mention, ok := m.(string); ok && mention == "@agent" {
				break
			}
		}
	}
}

// BenchmarkEnsureHistorySlice measures history slice conversion.
func BenchmarkEnsureHistorySlice(b *testing.B) {
	testCases := []struct {
		name  string
		input any
	}{
		{"nil", nil},
		{"empty", []any{}},
		{"single_entry", []any{map[string]any{"action": "test"}}},
		{"ten_entries", func() []any {
			result := make([]any, 10)
			for i := range result {
				result[i] = map[string]any{"action": fmt.Sprintf("action_%d", i)}
			}
			return result
		}()},
		{"fifty_entries", func() []any {
			result := make([]any, 50)
			for i := range result {
				result[i] = map[string]any{"action": fmt.Sprintf("action_%d", i)}
			}
			return result
		}()},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ensureHistorySlice(tc.input)
			}
		})
	}
}

// BenchmarkFetchComments measures comment fetching performance.
func BenchmarkFetchComments(b *testing.B) {
	commentCounts := []int{1, 10, 50, 100}

	for _, count := range commentCounts {
		b.Run(fmt.Sprintf("comments_%d", count), func(b *testing.B) {
			app := setupBenchmarkApp(b)
			board := createBenchmarkBoard(b, app, "TEST", "auto")
			task := createBenchmarkTask(b, app, board.Id, "need_input", true)
			seedBenchmarkComments(b, app, task.Id, count)

			service := NewService(app)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = service.fetchComments(task.Id)
			}
		})
	}
}

// BenchmarkGetTaskDisplayID measures display ID generation.
func BenchmarkGetTaskDisplayID(b *testing.B) {
	app := setupBenchmarkApp(b)
	board := createBenchmarkBoard(b, app, "TEST", "auto")
	task := createBenchmarkTask(b, app, board.Id, "need_input", true)

	service := NewService(app)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		service.getTaskDisplayID(task)
	}
}

// ========== Mention Extraction Benchmarks ==========

// BenchmarkExtractMentions measures regex-based mention extraction.
func BenchmarkExtractMentions(b *testing.B) {
	re := regexp.MustCompile(`@\w+`)

	testCases := []struct {
		name    string
		content string
	}{
		{"no_mentions", "This is a comment without any mentions"},
		{"single_mention", "@agent please review this"},
		{"multiple_mentions", "@agent and @senior-dev please check @reviewer"},
		{"complex_content", "Hey @agent, can you ask @product-manager about @designer's feedback on the @frontend changes?"},
		{"long_content", func() string {
			content := "Start of comment. "
			for i := 0; i < 20; i++ {
				content += fmt.Sprintf("Some text before @user%d and after. ", i)
			}
			return content + "@agent please review."
		}()},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				re.FindAllString(tc.content, -1)
			}
		})
	}
}

// ========== Service Creation Benchmark ==========

// BenchmarkNewService measures service instantiation overhead.
func BenchmarkNewService(b *testing.B) {
	app := setupBenchmarkApp(b)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		NewService(app)
	}
}

// ========== JSON Parsing Benchmarks ==========

// BenchmarkSessionDataParsing measures session data parsing overhead.
func BenchmarkSessionDataParsing(b *testing.B) {
	sessionJSON := `{"tool":"claude","ref":"session-123","ref_type":"uuid","working_dir":"/home/user/project"}`

	b.Run("json_unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var sessionMap map[string]any
			json.Unmarshal([]byte(sessionJSON), &sessionMap)
		}
	})

	b.Run("map_access", func(b *testing.B) {
		sessionMap := map[string]any{
			"tool":        "claude",
			"ref":         "session-123",
			"ref_type":    "uuid",
			"working_dir": "/home/user/project",
		}

		for i := 0; i < b.N; i++ {
			_, _ = sessionMap["tool"].(string)
			_, _ = sessionMap["ref"].(string)
			_, _ = sessionMap["working_dir"].(string)
		}
	})
}

// ========== Concurrent Access Benchmarks ==========

// BenchmarkCheckAndResumeConcurrent measures concurrent check performance.
func BenchmarkCheckAndResumeConcurrent(b *testing.B) {
	app := setupBenchmarkApp(b)
	board := createBenchmarkBoard(b, app, "TEST", "manual") // Use manual to avoid state changes
	task := createBenchmarkTask(b, app, board.Id, "need_input", true)
	comment := createBenchmarkComment(b, app, task.Id, "@agent test", "human", []string{"@agent"})

	service := NewService(app)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = service.CheckAndResume(comment)
		}
	})
}

// ========== Memory Allocation Benchmarks ==========

// BenchmarkCheckAndResumeAllocs measures memory allocations.
func BenchmarkCheckAndResumeAllocs(b *testing.B) {
	app := setupBenchmarkApp(b)
	board := createBenchmarkBoard(b, app, "TEST", "manual") // Use manual to avoid triggering resume
	task := createBenchmarkTask(b, app, board.Id, "need_input", true)
	comment := createBenchmarkComment(b, app, task.Id, "@agent test", "human", []string{"@agent"})

	service := NewService(app)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = service.CheckAndResume(comment)
	}
}

// BenchmarkHasAgentMentionAllocs measures mention detection allocations.
func BenchmarkHasAgentMentionAllocs(b *testing.B) {
	app := setupBenchmarkApp(b)
	collection, _ := app.FindCollectionByNameOrId("comments")
	taskCollection, _ := app.FindCollectionByNameOrId("tasks")

	task := core.NewRecord(taskCollection)
	task.Set("title", "Test")
	task.Set("column", "todo")
	app.Save(task)

	record := core.NewRecord(collection)
	record.Set("task", task.Id)
	record.Set("content", "test")
	record.Set("author_type", "human")
	record.Set("metadata", map[string]any{"mentions": []any{"@user", "@agent", "@admin"}})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hasAgentMention(record)
	}
}
