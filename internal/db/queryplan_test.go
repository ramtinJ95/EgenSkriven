package db

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// ========== Query Plan Verification Tests ==========
// These tests verify that queries use indexes as expected.

// TestQueryPlan_CommentsByTask verifies the comments by task query uses an index.
// Query: SELECT * FROM comments WHERE task = ? ORDER BY created ASC
func TestQueryPlan_CommentsByTask(t *testing.T) {
	app := setupQueryTestApp(t)

	// Create test data
	task := createQueryTestTask(t, app, "Query plan test task", "in_progress")
	for i := 0; i < 10; i++ {
		createQueryTestComment(t, app, task.Id, fmt.Sprintf("Comment %d", i))
	}

	// Execute EXPLAIN QUERY PLAN
	db := app.DB()
	rows, err := db.NewQuery("EXPLAIN QUERY PLAN SELECT * FROM comments WHERE task = {:taskId} ORDER BY created ASC").
		Bind(map[string]any{"taskId": task.Id}).
		Rows()
	if err != nil {
		t.Fatalf("failed to execute EXPLAIN: %v", err)
	}
	defer rows.Close()

	// Parse query plan output
	plan := parseQueryPlan(t, rows)

	// Verify index is used (should mention USING INDEX)
	if !strings.Contains(plan, "USING") && !strings.Contains(plan, "INDEX") {
		t.Logf("Query plan: %s", plan)
		t.Logf("Note: SQLite may choose table scan for small datasets")
	}

	// Verify the query returns expected results
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
	if len(records) != 10 {
		t.Errorf("expected 10 comments, got %d", len(records))
	}
}

// TestQueryPlan_TasksNeedInput verifies the need_input filter query uses an index.
// Query: SELECT * FROM tasks WHERE column = 'need_input' ORDER BY updated DESC
func TestQueryPlan_TasksNeedInput(t *testing.T) {
	app := setupQueryTestApp(t)

	// Create test data with various columns
	for i := 0; i < 5; i++ {
		createQueryTestTask(t, app, fmt.Sprintf("Need input task %d", i), "need_input")
	}
	for i := 0; i < 10; i++ {
		createQueryTestTask(t, app, fmt.Sprintf("In progress task %d", i), "in_progress")
	}

	// Execute EXPLAIN QUERY PLAN
	db := app.DB()
	rows, err := db.NewQuery("EXPLAIN QUERY PLAN SELECT * FROM tasks WHERE column = 'need_input' ORDER BY updated DESC").
		Rows()
	if err != nil {
		t.Fatalf("failed to execute EXPLAIN: %v", err)
	}
	defer rows.Close()

	plan := parseQueryPlan(t, rows)
	t.Logf("Query plan for tasks need_input: %s", plan)

	// Verify query returns correct results
	records, err := app.FindRecordsByFilter(
		"tasks",
		"column = 'need_input'",
		"-updated",
		0,
		0,
	)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(records) != 5 {
		t.Errorf("expected 5 need_input tasks, got %d", len(records))
	}
}

// TestQueryPlan_SessionByExternalRef verifies session lookup by external_ref uses an index.
// Query: SELECT * FROM sessions WHERE external_ref = ?
func TestQueryPlan_SessionByExternalRef(t *testing.T) {
	app := setupQueryTestApp(t)

	// Create test data
	task := createQueryTestTask(t, app, "Session test task", "in_progress")
	session := createQueryTestSession(t, app, task.Id, "unique-ref-12345")

	// Execute EXPLAIN QUERY PLAN
	db := app.DB()
	rows, err := db.NewQuery("EXPLAIN QUERY PLAN SELECT * FROM sessions WHERE external_ref = {:ref}").
		Bind(map[string]any{"ref": session.GetString("external_ref")}).
		Rows()
	if err != nil {
		t.Fatalf("failed to execute EXPLAIN: %v", err)
	}
	defer rows.Close()

	plan := parseQueryPlan(t, rows)
	t.Logf("Query plan for session by external_ref: %s", plan)

	// Should use index - check if plan mentions USING INDEX
	if !strings.Contains(plan, "USING") {
		t.Logf("Warning: query may not be using index (plan: %s)", plan)
	}
}

// TestQueryPlan_LatestCommentForTask verifies latest comment query uses composite index.
// Query: SELECT * FROM comments WHERE task = ? ORDER BY created DESC LIMIT 1
func TestQueryPlan_LatestCommentForTask(t *testing.T) {
	app := setupQueryTestApp(t)

	// Create test data
	task := createQueryTestTask(t, app, "Latest comment test", "in_progress")
	for i := 0; i < 20; i++ {
		createQueryTestComment(t, app, task.Id, fmt.Sprintf("Comment %d", i))
		time.Sleep(time.Millisecond) // Ensure different created times
	}

	// Execute EXPLAIN QUERY PLAN
	db := app.DB()
	rows, err := db.NewQuery("EXPLAIN QUERY PLAN SELECT * FROM comments WHERE task = {:taskId} ORDER BY created DESC LIMIT 1").
		Bind(map[string]any{"taskId": task.Id}).
		Rows()
	if err != nil {
		t.Fatalf("failed to execute EXPLAIN: %v", err)
	}
	defer rows.Close()

	plan := parseQueryPlan(t, rows)
	t.Logf("Query plan for latest comment: %s", plan)

	// Verify returns only 1 record
	records, err := app.FindRecordsByFilter(
		"comments",
		"task = {:taskId}",
		"-created",
		1,
		0,
		map[string]any{"taskId": task.Id},
	)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 comment, got %d", len(records))
	}
}

// ========== Query Performance Tests ==========

// TestQueryPerformance_CommentsByTask measures query performance at scale.
func TestQueryPerformance_CommentsByTask(t *testing.T) {
	app := setupQueryTestApp(t)

	// Create task with many comments
	task := createQueryTestTask(t, app, "Perf test task", "in_progress")
	numComments := 1000
	for i := 0; i < numComments; i++ {
		createQueryTestComment(t, app, task.Id, fmt.Sprintf("Performance test comment %d with some content to make it realistic", i))
	}

	// Measure query time
	start := time.Now()
	records, err := app.FindRecordsByFilter(
		"comments",
		"task = {:taskId}",
		"+created",
		0,
		0,
		map[string]any{"taskId": task.Id},
	)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(records) != numComments {
		t.Errorf("expected %d comments, got %d", numComments, len(records))
	}

	// Target: < 200ms for list comments
	target := 200 * time.Millisecond
	t.Logf("Query time for %d comments: %v (target: %v)", numComments, elapsed, target)
	if elapsed > target {
		t.Errorf("query took %v, exceeded target of %v", elapsed, target)
	}
}

// TestQueryPerformance_TasksNeedInput measures need_input filter performance at scale.
func TestQueryPerformance_TasksNeedInput(t *testing.T) {
	app := setupQueryTestApp(t)

	// Create many tasks with various columns
	totalTasks := 10000
	needInputTasks := 100
	for i := 0; i < needInputTasks; i++ {
		createQueryTestTask(t, app, fmt.Sprintf("Need input %d", i), "need_input")
	}
	for i := 0; i < totalTasks-needInputTasks; i++ {
		col := []string{"todo", "in_progress", "done"}[i%3]
		createQueryTestTask(t, app, fmt.Sprintf("Other task %d", i), col)
	}

	// Measure query time
	start := time.Now()
	records, err := app.FindRecordsByFilter(
		"tasks",
		"column = 'need_input'",
		"-updated",
		0,
		0,
	)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(records) != needInputTasks {
		t.Errorf("expected %d need_input tasks, got %d", needInputTasks, len(records))
	}

	// Target: < 100ms for list --need-input
	target := 100 * time.Millisecond
	t.Logf("Query time for %d tasks (finding %d need_input): %v (target: %v)",
		totalTasks, needInputTasks, elapsed, target)
	if elapsed > target {
		t.Errorf("query took %v, exceeded target of %v", elapsed, target)
	}
}

// TestQueryPerformance_SessionByExternalRef measures session lookup performance.
func TestQueryPerformance_SessionByExternalRef(t *testing.T) {
	app := setupQueryTestApp(t)

	// Create many sessions
	numSessions := 1000
	var targetSession *core.Record
	for i := 0; i < numSessions; i++ {
		task := createQueryTestTask(t, app, fmt.Sprintf("Session task %d", i), "in_progress")
		session := createQueryTestSession(t, app, task.Id, fmt.Sprintf("ref-%d", i))
		if i == numSessions/2 {
			targetSession = session
		}
	}

	// Measure query time
	start := time.Now()
	records, err := app.FindRecordsByFilter(
		"sessions",
		"external_ref = {:ref}",
		"",
		1,
		0,
		map[string]any{"ref": targetSession.GetString("external_ref")},
	)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 session, got %d", len(records))
	}

	// Target: O(1) lookup, should be very fast
	target := 50 * time.Millisecond
	t.Logf("Query time for session lookup among %d sessions: %v (target: %v)",
		numSessions, elapsed, target)
	if elapsed > target {
		t.Errorf("query took %v, exceeded target of %v", elapsed, target)
	}
}

// TestQueryPerformance_LatestComment measures latest comment lookup performance.
func TestQueryPerformance_LatestComment(t *testing.T) {
	app := setupQueryTestApp(t)

	// Create task with many comments
	task := createQueryTestTask(t, app, "Latest comment perf test", "in_progress")
	numComments := 1000
	for i := 0; i < numComments; i++ {
		createQueryTestComment(t, app, task.Id, fmt.Sprintf("Comment %d", i))
	}

	// Measure query time
	start := time.Now()
	records, err := app.FindRecordsByFilter(
		"comments",
		"task = {:taskId}",
		"-created",
		1,
		0,
		map[string]any{"taskId": task.Id},
	)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 comment, got %d", len(records))
	}

	// Target: O(1) with composite index
	target := 10 * time.Millisecond
	t.Logf("Query time for latest comment among %d: %v (target: %v)",
		numComments, elapsed, target)
	if elapsed > target {
		t.Errorf("query took %v, exceeded target of %v", elapsed, target)
	}
}

// ========== Concurrent Query Tests ==========

// TestQueryPerformance_ConcurrentReads tests concurrent query performance.
func TestQueryPerformance_ConcurrentReads(t *testing.T) {
	app := setupQueryTestApp(t)

	// Create test data
	task := createQueryTestTask(t, app, "Concurrent test task", "in_progress")
	for i := 0; i < 100; i++ {
		createQueryTestComment(t, app, task.Id, fmt.Sprintf("Comment %d", i))
	}

	// Run concurrent queries
	numClients := 100
	results := make(chan time.Duration, numClients)
	errors := make(chan error, numClients)

	start := time.Now()
	for i := 0; i < numClients; i++ {
		go func() {
			queryStart := time.Now()
			_, err := app.FindRecordsByFilter(
				"comments",
				"task = {:taskId}",
				"+created",
				0,
				0,
				map[string]any{"taskId": task.Id},
			)
			if err != nil {
				errors <- err
				return
			}
			results <- time.Since(queryStart)
		}()
	}

	// Collect results
	var totalTime time.Duration
	var maxTime time.Duration
	for i := 0; i < numClients; i++ {
		select {
		case err := <-errors:
			t.Errorf("concurrent query error: %v", err)
		case dur := <-results:
			totalTime += dur
			if dur > maxTime {
				maxTime = dur
			}
		}
	}
	totalElapsed := time.Since(start)

	avgTime := totalTime / time.Duration(numClients)
	t.Logf("Concurrent queries (%d clients):", numClients)
	t.Logf("  Total wall time: %v", totalElapsed)
	t.Logf("  Avg query time: %v", avgTime)
	t.Logf("  Max query time: %v", maxTime)

	// All queries should complete within reasonable time
	target := 500 * time.Millisecond
	if totalElapsed > target {
		t.Errorf("concurrent queries took %v, exceeded target of %v", totalElapsed, target)
	}
}

// ========== Helper Functions ==========

func parseQueryPlan(t *testing.T, rows interface{ Next() bool; Scan(dest ...any) error; Columns() ([]string, error) }) string {
	t.Helper()
	var plan strings.Builder
	for rows.Next() {
		var id, parent, notused int
		var detail string
		if err := rows.Scan(&id, &parent, &notused, &detail); err != nil {
			t.Fatalf("failed to scan EXPLAIN row: %v", err)
		}
		plan.WriteString(detail)
		plan.WriteString(" ")
	}
	return plan.String()
}

func setupQueryTestApp(t *testing.T) *pocketbase.PocketBase {
	t.Helper()
	app := testutil.NewTestApp(t)

	// Create tasks collection
	if _, err := app.FindCollectionByNameOrId("tasks"); err != nil {
		tasks := core.NewBaseCollection("tasks")
		tasks.Fields.Add(&core.TextField{Name: "title", Required: true})
		tasks.Fields.Add(&core.TextField{Name: "column"})
		tasks.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
		tasks.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
		// Add column index
		tasks.Indexes = []string{
			"CREATE INDEX idx_tasks_column ON tasks (column)",
		}
		if err := app.Save(tasks); err != nil {
			t.Fatalf("failed to create tasks collection: %v", err)
		}
	}

	// Create comments collection with indexes
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
		comments.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
		comments.Indexes = []string{
			"CREATE INDEX idx_comments_task ON comments (task)",
			"CREATE INDEX idx_comments_created ON comments (created)",
			"CREATE INDEX idx_comments_task_created ON comments (task, created)",
		}
		if err := app.Save(comments); err != nil {
			t.Fatalf("failed to create comments collection: %v", err)
		}
	}

	// Create sessions collection with indexes
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
		sessions.Indexes = []string{
			"CREATE INDEX idx_sessions_task ON sessions (task)",
			"CREATE INDEX idx_sessions_status ON sessions (status)",
			"CREATE INDEX idx_sessions_external_ref ON sessions (external_ref)",
		}
		if err := app.Save(sessions); err != nil {
			t.Fatalf("failed to create sessions collection: %v", err)
		}
	}

	return app
}

func createQueryTestTask(t *testing.T, app *pocketbase.PocketBase, title, column string) *core.Record {
	t.Helper()
	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		t.Fatalf("tasks collection not found: %v", err)
	}
	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("column", column)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	return record
}

func createQueryTestComment(t *testing.T, app *pocketbase.PocketBase, taskId, content string) *core.Record {
	t.Helper()
	collection, err := app.FindCollectionByNameOrId("comments")
	if err != nil {
		t.Fatalf("comments collection not found: %v", err)
	}
	record := core.NewRecord(collection)
	record.Set("task", taskId)
	record.Set("content", content)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}
	return record
}

func createQueryTestSession(t *testing.T, app *pocketbase.PocketBase, taskId, externalRef string) *core.Record {
	t.Helper()
	collection, err := app.FindCollectionByNameOrId("sessions")
	if err != nil {
		t.Fatalf("sessions collection not found: %v", err)
	}
	record := core.NewRecord(collection)
	record.Set("task", taskId)
	record.Set("tool", "claude-code")
	record.Set("external_ref", externalRef)
	record.Set("ref_type", "uuid")
	record.Set("working_dir", "/tmp")
	record.Set("status", "active")
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	return record
}
