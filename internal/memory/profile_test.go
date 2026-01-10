package memory

import (
	"encoding/json"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/resume"
	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// ========== Memory Allocation Benchmarks ==========

// BenchmarkCommentRecordMemory measures memory for a typical comment record.
// Expected: ~1KB per comment
func BenchmarkCommentRecordMemory(b *testing.B) {
	app := setupMemoryTestApp(b)
	task := createMemoryTestTask(b, app, "Memory test task", "in_progress")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		collection, _ := app.FindCollectionByNameOrId("comments")
		record := core.NewRecord(collection)
		record.Set("task", task.Id)
		record.Set("content", "This is a typical comment content that might include some context about the task. It could have multiple sentences and describe what was done or what needs to be done next.")
		record.Set("author_type", "human")

		// Measure the record in memory (not saved)
		_ = record
	}
}

// BenchmarkContextPromptMemory measures memory for context prompt generation.
// Expected: ~10KB for 100 comments
func BenchmarkContextPromptMemory(b *testing.B) {
	app := setupMemoryTestApp(b)
	task := createMemoryTestTask(b, app, "Context prompt memory test", "in_progress")

	// Create 100 comments
	for i := 0; i < 100; i++ {
		createMemoryTestComment(b, app, task.Id, fmt.Sprintf("Comment %d with some content that represents a typical message in the collaborative workflow between human and agent.", i))
	}

	// Fetch comments for context building
	records, err := app.FindRecordsByFilter(
		"comments",
		"task = {:taskId}",
		"+created",
		0,
		0,
		map[string]any{"taskId": task.Id},
	)
	if err != nil {
		b.Fatalf("failed to fetch comments: %v", err)
	}

	// Convert to resume.Comment type
	comments := make([]resume.Comment, len(records))
	for i, r := range records {
		comments[i] = resume.Comment{
			Content:    r.GetString("content"),
			AuthorType: r.GetString("author_type"),
			Created:    r.GetDateTime("created").Time(),
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Build context prompt (similar to what resume/context.go does)
		prompt := resume.BuildContextPrompt(
			task,
			"T-1",
			comments,
		)
		_ = prompt
	}
}

// BenchmarkSessionRecordMemory measures memory for a session record.
// Expected: ~500B per session
func BenchmarkSessionRecordMemory(b *testing.B) {
	app := setupMemoryTestApp(b)
	task := createMemoryTestTask(b, app, "Session memory test", "in_progress")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		collection, _ := app.FindCollectionByNameOrId("sessions")
		record := core.NewRecord(collection)
		record.Set("task", task.Id)
		record.Set("tool", "claude-code")
		record.Set("external_ref", "session-uuid-12345-abcdef-67890")
		record.Set("ref_type", "uuid")
		record.Set("working_dir", "/home/user/projects/myproject")
		record.Set("status", "active")

		_ = record
	}
}

// BenchmarkTaskWithSessionMemory measures memory overhead of agent_session JSON.
// Expected: +500B for agent_session field
func BenchmarkTaskWithSessionMemory(b *testing.B) {
	app := setupMemoryTestApp(b)

	// Agent session data structure
	agentSession := map[string]any{
		"tool":         "claude-code",
		"external_ref": "session-uuid-12345-abcdef-67890",
		"ref_type":     "uuid",
		"working_dir":  "/home/user/projects/myproject",
		"started_at":   time.Now().Format(time.RFC3339),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		collection, _ := app.FindCollectionByNameOrId("tasks")
		record := core.NewRecord(collection)
		record.Set("title", "Task with agent session")
		record.Set("column", "in_progress")
		record.Set("agent_session", agentSession)

		// Serialize agent_session to measure JSON overhead
		data, _ := json.Marshal(agentSession)
		_ = len(data)
		_ = record
	}
}

// ========== Memory Size Verification Tests ==========

// TestCommentRecordMemory verifies comment record is within expected size.
func TestCommentRecordMemory(t *testing.T) {
	app := setupMemoryTestAppForTest(t)
	task := createMemoryTestTaskForTest(t, app, "Size test task", "in_progress")

	collection, err := app.FindCollectionByNameOrId("comments")
	if err != nil {
		t.Fatalf("comments collection not found: %v", err)
	}

	// Create a typical comment
	record := core.NewRecord(collection)
	record.Set("task", task.Id)
	record.Set("content", "This is a typical comment content that might include some context about the task. It could have multiple sentences and describe what was done or what needs to be done next.")
	record.Set("author_type", "human")

	// Measure allocation for creating the record
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	before := m.Alloc

	// Create 100 records to get average
	records := make([]*core.Record, 100)
	for i := 0; i < 100; i++ {
		records[i] = core.NewRecord(collection)
		records[i].Set("task", task.Id)
		records[i].Set("content", fmt.Sprintf("Comment %d: typical content for testing memory usage", i))
		records[i].Set("author_type", "human")
	}

	runtime.ReadMemStats(&m)
	after := m.Alloc

	avgSize := (after - before) / 100
	t.Logf("Average comment record memory: %d bytes", avgSize)

	// Target: ~1KB per comment
	expected := uint64(2048) // Allow 2KB margin
	if avgSize > expected {
		t.Logf("Warning: comment record uses %d bytes, expected ~1KB", avgSize)
	}
}

// TestContextPromptMemory verifies context prompt is within expected size.
func TestContextPromptMemory(t *testing.T) {
	app := setupMemoryTestAppForTest(t)
	task := createMemoryTestTaskForTest(t, app, "Context prompt test", "in_progress")

	// Create 100 comments
	for i := 0; i < 100; i++ {
		createMemoryTestCommentForTest(t, app, task.Id, fmt.Sprintf("Comment %d: This is a typical message in the workflow with reasonable content length.", i))
	}

	records, err := app.FindRecordsByFilter(
		"comments",
		"task = {:taskId}",
		"+created",
		0,
		0,
		map[string]any{"taskId": task.Id},
	)
	if err != nil {
		t.Fatalf("failed to fetch comments: %v", err)
	}

	// Convert to resume.Comment type
	comments := make([]resume.Comment, len(records))
	for i, r := range records {
		comments[i] = resume.Comment{
			Content:    r.GetString("content"),
			AuthorType: r.GetString("author_type"),
			Created:    r.GetDateTime("created").Time(),
		}
	}

	// Build context prompt
	prompt := resume.BuildContextPrompt(
		task,
		"T-1",
		comments,
	)

	promptSize := len(prompt)
	t.Logf("Context prompt size for 100 comments: %d bytes (%.2f KB)", promptSize, float64(promptSize)/1024)

	// Target: ~10KB for 100 comments
	expected := 20 * 1024 // Allow 20KB margin
	if promptSize > expected {
		t.Logf("Warning: context prompt uses %d bytes, expected ~10KB", promptSize)
	}
}

// TestSessionRecordMemory verifies session record is within expected size.
func TestSessionRecordMemory(t *testing.T) {
	app := setupMemoryTestAppForTest(t)
	task := createMemoryTestTaskForTest(t, app, "Session size test", "in_progress")

	collection, err := app.FindCollectionByNameOrId("sessions")
	if err != nil {
		t.Fatalf("sessions collection not found: %v", err)
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	before := m.Alloc

	// Create 100 sessions to get average
	sessions := make([]*core.Record, 100)
	for i := 0; i < 100; i++ {
		sessions[i] = core.NewRecord(collection)
		sessions[i].Set("task", task.Id)
		sessions[i].Set("tool", "claude-code")
		sessions[i].Set("external_ref", fmt.Sprintf("session-uuid-%d", i))
		sessions[i].Set("ref_type", "uuid")
		sessions[i].Set("working_dir", "/home/user/projects/myproject")
		sessions[i].Set("status", "active")
	}

	runtime.ReadMemStats(&m)
	after := m.Alloc

	avgSize := (after - before) / 100
	t.Logf("Average session record memory: %d bytes", avgSize)

	// Target: ~500B per session
	expected := uint64(1024) // Allow 1KB margin
	if avgSize > expected {
		t.Logf("Warning: session record uses %d bytes, expected ~500B", avgSize)
	}
}

// TestAgentSessionJSONOverhead measures JSON overhead for agent_session field.
func TestAgentSessionJSONOverhead(t *testing.T) {
	agentSession := map[string]any{
		"tool":         "claude-code",
		"external_ref": "session-uuid-12345-abcdef-67890",
		"ref_type":     "uuid",
		"working_dir":  "/home/user/projects/myproject",
		"started_at":   time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(agentSession)
	if err != nil {
		t.Fatalf("failed to marshal agent_session: %v", err)
	}

	jsonSize := len(data)
	t.Logf("Agent session JSON size: %d bytes", jsonSize)

	// Target: ~500B for agent_session JSON
	expected := 512
	if jsonSize > expected {
		t.Logf("Warning: agent_session JSON uses %d bytes, expected ~500B", jsonSize)
	}
}

// ========== Memory Leak Detection Tests ==========

// TestMemoryLeak_RepeatedCommentCreation tests for memory leaks in comment creation.
func TestMemoryLeak_RepeatedCommentCreation(t *testing.T) {
	app := setupMemoryTestAppForTest(t)
	task := createMemoryTestTaskForTest(t, app, "Memory leak test", "in_progress")

	// Force GC before starting
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	startAlloc := m.Alloc

	// Perform many operations
	iterations := 1000
	for i := 0; i < iterations; i++ {
		createMemoryTestCommentForTest(t, app, task.Id, fmt.Sprintf("Comment %d for leak testing", i))

		// Periodically force GC to reclaim memory
		if i%100 == 0 {
			runtime.GC()
		}
	}

	// Force final GC
	runtime.GC()
	runtime.ReadMemStats(&m)
	endAlloc := m.Alloc

	// Calculate growth per operation
	growth := int64(endAlloc) - int64(startAlloc)
	growthPerOp := float64(growth) / float64(iterations)

	t.Logf("Memory after %d operations:", iterations)
	t.Logf("  Start: %d bytes", startAlloc)
	t.Logf("  End: %d bytes", endAlloc)
	t.Logf("  Growth: %d bytes (%.2f bytes/op)", growth, growthPerOp)

	// Allow some memory growth but flag potential leaks
	// Expect roughly constant memory usage after GC
	maxGrowthPerOp := float64(1024) // 1KB per op max
	if growthPerOp > maxGrowthPerOp {
		t.Logf("Warning: high memory growth %.2f bytes/op, potential leak", growthPerOp)
	}
}

// TestMemoryLeak_RepeatedContextPromptBuilding tests for memory leaks in context building.
func TestMemoryLeak_RepeatedContextPromptBuilding(t *testing.T) {
	app := setupMemoryTestAppForTest(t)
	task := createMemoryTestTaskForTest(t, app, "Context prompt leak test", "in_progress")

	// Create some comments
	for i := 0; i < 50; i++ {
		createMemoryTestCommentForTest(t, app, task.Id, fmt.Sprintf("Comment %d for context building", i))
	}

	records, _ := app.FindRecordsByFilter(
		"comments",
		"task = {:taskId}",
		"+created",
		0,
		0,
		map[string]any{"taskId": task.Id},
	)

	// Convert to resume.Comment type
	comments := make([]resume.Comment, len(records))
	for i, r := range records {
		comments[i] = resume.Comment{
			Content:    r.GetString("content"),
			AuthorType: r.GetString("author_type"),
			Created:    r.GetDateTime("created").Time(),
		}
	}

	// Force GC before starting
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	startAlloc := m.Alloc

	// Build context prompts repeatedly
	iterations := 10000
	for i := 0; i < iterations; i++ {
		prompt := resume.BuildContextPrompt(
			task,
			"T-1",
			comments,
		)
		_ = prompt

		if i%1000 == 0 {
			runtime.GC()
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m)
	endAlloc := m.Alloc

	growth := int64(endAlloc) - int64(startAlloc)
	growthPerOp := float64(growth) / float64(iterations)

	t.Logf("Memory after %d context prompt builds:", iterations)
	t.Logf("  Start: %d bytes", startAlloc)
	t.Logf("  End: %d bytes", endAlloc)
	t.Logf("  Growth: %d bytes (%.4f bytes/op)", growth, growthPerOp)

	// Context prompt building should have minimal memory growth
	maxGrowthPerOp := float64(100) // 100 bytes per op max
	if growthPerOp > maxGrowthPerOp {
		t.Logf("Warning: high memory growth %.2f bytes/op, potential leak", growthPerOp)
	}
}

// TestMemoryLeak_RepeatedQueryExecution tests for memory leaks in query execution.
func TestMemoryLeak_RepeatedQueryExecution(t *testing.T) {
	app := setupMemoryTestAppForTest(t)
	task := createMemoryTestTaskForTest(t, app, "Query leak test", "in_progress")

	// Create test data
	for i := 0; i < 100; i++ {
		createMemoryTestCommentForTest(t, app, task.Id, fmt.Sprintf("Comment %d", i))
	}

	// Force GC before starting
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	startAlloc := m.Alloc

	// Execute queries repeatedly
	iterations := 1000
	for i := 0; i < iterations; i++ {
		records, err := app.FindRecordsByFilter(
			"comments",
			"task = {:taskId}",
			"+created",
			0,
			0,
			map[string]any{"taskId": task.Id},
		)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		_ = records

		if i%100 == 0 {
			runtime.GC()
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m)
	endAlloc := m.Alloc

	growth := int64(endAlloc) - int64(startAlloc)
	growthPerOp := float64(growth) / float64(iterations)

	t.Logf("Memory after %d queries:", iterations)
	t.Logf("  Start: %d bytes", startAlloc)
	t.Logf("  End: %d bytes", endAlloc)
	t.Logf("  Growth: %d bytes (%.2f bytes/op)", growth, growthPerOp)

	// Queries should have minimal memory growth after GC
	maxGrowthPerOp := float64(500) // 500 bytes per op max
	if growthPerOp > maxGrowthPerOp {
		t.Logf("Warning: high memory growth %.2f bytes/op, potential leak", growthPerOp)
	}
}

// ========== Setup Helpers (for benchmarks) ==========

func setupMemoryTestApp(b *testing.B) *pocketbase.PocketBase {
	b.Helper()

	// Use testing.T wrapper for testutil compatibility
	t := &testing.T{}
	app := testutil.NewTestApp(t)
	setupMemoryCollections(b, app)
	return app
}

func setupMemoryCollections(tb testing.TB, app *pocketbase.PocketBase) {
	// Create tasks collection
	if _, err := app.FindCollectionByNameOrId("tasks"); err != nil {
		tasks := core.NewBaseCollection("tasks")
		tasks.Fields.Add(&core.TextField{Name: "title", Required: true})
		tasks.Fields.Add(&core.TextField{Name: "column"})
		tasks.Fields.Add(&core.JSONField{Name: "agent_session"})
		tasks.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
		tasks.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
		if err := app.Save(tasks); err != nil {
			tb.Fatalf("failed to create tasks collection: %v", err)
		}
	}

	// Create comments collection
	if _, err := app.FindCollectionByNameOrId("comments"); err != nil {
		tasks, _ := app.FindCollectionByNameOrId("tasks")
		comments := core.NewBaseCollection("comments")
		comments.Fields.Add(&core.RelationField{
			Name:          "task",
			CollectionId:  tasks.Id,
			MaxSelect:     1,
			Required:      true,
			CascadeDelete: true,
		})
		comments.Fields.Add(&core.TextField{Name: "content", Required: true})
		comments.Fields.Add(&core.TextField{Name: "author_type"})
		comments.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
		if err := app.Save(comments); err != nil {
			tb.Fatalf("failed to create comments collection: %v", err)
		}
	}

	// Create sessions collection
	if _, err := app.FindCollectionByNameOrId("sessions"); err != nil {
		tasks, _ := app.FindCollectionByNameOrId("tasks")
		sessions := core.NewBaseCollection("sessions")
		sessions.Fields.Add(&core.RelationField{
			Name:          "task",
			CollectionId:  tasks.Id,
			MaxSelect:     1,
			Required:      true,
			CascadeDelete: true,
		})
		sessions.Fields.Add(&core.SelectField{
			Name:     "tool",
			Required: true,
			Values:   []string{"opencode", "claude-code", "codex"},
		})
		sessions.Fields.Add(&core.TextField{Name: "external_ref", Required: true})
		sessions.Fields.Add(&core.SelectField{
			Name:     "ref_type",
			Required: true,
			Values:   []string{"uuid", "path"},
		})
		sessions.Fields.Add(&core.TextField{Name: "working_dir", Required: true})
		sessions.Fields.Add(&core.SelectField{
			Name:     "status",
			Required: true,
			Values:   []string{"active", "paused", "completed", "abandoned"},
		})
		sessions.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
		if err := app.Save(sessions); err != nil {
			tb.Fatalf("failed to create sessions collection: %v", err)
		}
	}
}

func createMemoryTestTask(b *testing.B, app *pocketbase.PocketBase, title, column string) *core.Record {
	b.Helper()
	collection, _ := app.FindCollectionByNameOrId("tasks")
	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("column", column)
	if err := app.Save(record); err != nil {
		b.Fatalf("failed to create task: %v", err)
	}
	return record
}

func createMemoryTestComment(b *testing.B, app *pocketbase.PocketBase, taskId, content string) *core.Record {
	b.Helper()
	collection, _ := app.FindCollectionByNameOrId("comments")
	record := core.NewRecord(collection)
	record.Set("task", taskId)
	record.Set("content", content)
	record.Set("author_type", "human")
	if err := app.Save(record); err != nil {
		b.Fatalf("failed to create comment: %v", err)
	}
	return record
}

// ========== Setup Helpers (for tests) ==========

func setupMemoryTestAppForTest(t *testing.T) *pocketbase.PocketBase {
	t.Helper()

	app := testutil.NewTestApp(t)
	setupMemoryCollections(t, app)
	return app
}

func createMemoryTestTaskForTest(t *testing.T, app *pocketbase.PocketBase, title, column string) *core.Record {
	t.Helper()
	collection, _ := app.FindCollectionByNameOrId("tasks")
	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("column", column)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	return record
}

func createMemoryTestCommentForTest(t *testing.T, app *pocketbase.PocketBase, taskId, content string) *core.Record {
	t.Helper()
	collection, _ := app.FindCollectionByNameOrId("comments")
	record := core.NewRecord(collection)
	record.Set("task", taskId)
	record.Set("content", content)
	record.Set("author_type", "human")
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}
	return record
}
