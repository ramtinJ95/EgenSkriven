//go:build performance
// +build performance

// Package performance contains end-to-end performance tests for EgenSkriven.
// Run with: go test -v -tags=performance ./tests/performance/
package performance

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/autoresume"
	"github.com/ramtinJ95/EgenSkriven/internal/resume"
)

// ========== Performance Targets ==========
const (
	TargetBlockTask        = 100 * time.Millisecond
	TargetAddComment       = 100 * time.Millisecond
	TargetListComments     = 200 * time.Millisecond
	TargetResumeGeneration = 50 * time.Millisecond
	TargetAutoResumeTrigger = 1 * time.Second
	TargetSessionLink      = 50 * time.Millisecond
	TargetListNeedInput    = 100 * time.Millisecond
)

// ========== E2E Performance Tests ==========

// TestE2E_BlockResumeWorkflow_Performance tests the complete block-resume workflow timing.
func TestE2E_BlockResumeWorkflow_Performance(t *testing.T) {
	app := setupPerfTestApp(t)
	board := createPerfBoard(t, app, "PERF", "auto")
	task := createPerfTask(t, app, board.Id, "Block-resume perf test", "in_progress")

	// Link session
	task.Set("agent_session", map[string]any{
		"tool":        "claude-code",
		"ref":         "perf-session-123",
		"ref_type":    "uuid",
		"working_dir": "/tmp",
		"linked_at":   time.Now().UTC().Format(time.RFC3339),
	})
	if err := app.Save(task); err != nil {
		t.Fatalf("failed to link session: %v", err)
	}

	results := &WorkflowPerformanceResults{}

	// 1. Block task with question
	start := time.Now()
	_ = createPerfComment(t, app, task.Id, "I have a question: What should I implement next?", "agent")
	task.Set("column", "need_input")
	addHistory(task, "blocked", "agent")
	if err := app.Save(task); err != nil {
		t.Fatalf("failed to block task: %v", err)
	}
	results.BlockTask = time.Since(start)

	// 2. Human adds response with @agent mention
	start = time.Now()
	humanComment := createPerfComment(t, app, task.Id, "@agent Please implement the authentication flow", "human")
	results.AddComment = time.Since(start)

	// 3. Auto-resume trigger
	start = time.Now()
	service := autoresume.NewService(app)
	if err := service.CheckAndResume(humanComment); err != nil {
		t.Fatalf("auto-resume failed: %v", err)
	}
	results.AutoResumeTrigger = time.Since(start)

	// 4. Generate resume context
	records, _ := app.FindRecordsByFilter(
		"comments",
		"task = {:taskId}",
		"+created",
		0,
		0,
		map[string]any{"taskId": task.Id},
	)
	comments := make([]resume.Comment, len(records))
	for i, r := range records {
		comments[i] = resume.Comment{
			Content:    r.GetString("content"),
			AuthorType: r.GetString("author_type"),
			Created:    r.GetDateTime("created").Time(),
		}
	}

	refreshedTask, _ := app.FindRecordById("tasks", task.Id)
	start = time.Now()
	prompt := resume.BuildContextPrompt(refreshedTask, "T-1", comments)
	results.ResumeGeneration = time.Since(start)
	_ = prompt

	// Report results
	t.Logf("Block-Resume Workflow Performance:")
	t.Logf("  Block task: %v (target: %v)", results.BlockTask, TargetBlockTask)
	t.Logf("  Add comment: %v (target: %v)", results.AddComment, TargetAddComment)
	t.Logf("  Auto-resume trigger: %v (target: %v)", results.AutoResumeTrigger, TargetAutoResumeTrigger)
	t.Logf("  Resume generation: %v (target: %v)", results.ResumeGeneration, TargetResumeGeneration)

	// Assert targets
	assertPerformanceTarget(t, "Block task", results.BlockTask, TargetBlockTask)
	assertPerformanceTarget(t, "Add comment", results.AddComment, TargetAddComment)
	assertPerformanceTarget(t, "Auto-resume trigger", results.AutoResumeTrigger, TargetAutoResumeTrigger)
	assertPerformanceTarget(t, "Resume generation", results.ResumeGeneration, TargetResumeGeneration)
}

// TestE2E_HighVolumeComments_Performance tests performance with 1000+ comments.
func TestE2E_HighVolumeComments_Performance(t *testing.T) {
	app := setupPerfTestApp(t)
	board := createPerfBoard(t, app, "VOL", "manual")
	task := createPerfTask(t, app, board.Id, "High volume comments test", "in_progress")

	// Create 1000 comments
	numComments := 1000
	t.Logf("Creating %d comments...", numComments)
	createStart := time.Now()
	for i := 0; i < numComments; i++ {
		authorType := "human"
		if i%2 == 0 {
			authorType = "agent"
		}
		createPerfComment(t, app, task.Id, fmt.Sprintf("Comment %d: This is test content for performance testing purposes.", i), authorType)
	}
	createDuration := time.Since(createStart)
	t.Logf("Created %d comments in %v (%.2f comments/sec)", numComments, createDuration, float64(numComments)/createDuration.Seconds())

	// Test listing comments
	listStart := time.Now()
	records, err := app.FindRecordsByFilter(
		"comments",
		"task = {:taskId}",
		"+created",
		0,
		0,
		map[string]any{"taskId": task.Id},
	)
	listDuration := time.Since(listStart)

	if err != nil {
		t.Fatalf("failed to list comments: %v", err)
	}
	if len(records) != numComments {
		t.Errorf("expected %d comments, got %d", numComments, len(records))
	}

	t.Logf("List %d comments: %v (target: %v)", numComments, listDuration, TargetListComments)
	assertPerformanceTarget(t, "List comments", listDuration, TargetListComments)

	// Test context prompt building
	comments := make([]resume.Comment, len(records))
	for i, r := range records {
		comments[i] = resume.Comment{
			Content:    r.GetString("content"),
			AuthorType: r.GetString("author_type"),
			Created:    r.GetDateTime("created").Time(),
		}
	}

	promptStart := time.Now()
	prompt := resume.BuildContextPrompt(task, "T-1", comments)
	promptDuration := time.Since(promptStart)

	t.Logf("Build context prompt with %d comments: %v (target: %v)", numComments, promptDuration, TargetResumeGeneration)
	t.Logf("Prompt size: %d bytes", len(prompt))

	// Allow extra time for high volume
	highVolumeTarget := TargetResumeGeneration * 10 // 500ms for 1000 comments
	assertPerformanceTarget(t, "Build context prompt (high volume)", promptDuration, highVolumeTarget)
}

// TestE2E_ConcurrentOperations_Performance tests parallel operations.
func TestE2E_ConcurrentOperations_Performance(t *testing.T) {
	app := setupPerfTestApp(t)
	board := createPerfBoard(t, app, "CONC", "manual")

	// Create 50 tasks
	numTasks := 50
	tasks := make([]*core.Record, numTasks)
	for i := 0; i < numTasks; i++ {
		tasks[i] = createPerfTask(t, app, board.Id, fmt.Sprintf("Concurrent test task %d", i), "in_progress")
	}

	// Add some comments to each task
	for _, task := range tasks {
		for j := 0; j < 10; j++ {
			createPerfComment(t, app, task.Id, fmt.Sprintf("Initial comment %d", j), "human")
		}
	}

	// Test concurrent comment creation
	numClients := 100
	var wg sync.WaitGroup
	errors := make(chan error, numClients)
	durations := make(chan time.Duration, numClients)

	start := time.Now()
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			taskIdx := idx % numTasks
			opStart := time.Now()
			createPerfComment(t, app, tasks[taskIdx].Id, fmt.Sprintf("Concurrent comment from client %d", idx), "human")
			durations <- time.Since(opStart)
		}(i)
	}
	wg.Wait()
	close(errors)
	close(durations)

	totalDuration := time.Since(start)

	// Calculate stats
	var maxDuration, totalOps time.Duration
	count := 0
	for d := range durations {
		totalOps += d
		if d > maxDuration {
			maxDuration = d
		}
		count++
	}
	avgDuration := totalOps / time.Duration(count)

	t.Logf("Concurrent Operations Performance (%d clients):", numClients)
	t.Logf("  Total wall time: %v", totalDuration)
	t.Logf("  Avg operation time: %v", avgDuration)
	t.Logf("  Max operation time: %v", maxDuration)
	t.Logf("  Operations/sec: %.2f", float64(numClients)/totalDuration.Seconds())

	// Concurrent operations should complete reasonably fast
	targetTotal := 2 * time.Second
	if totalDuration > targetTotal {
		t.Errorf("concurrent operations took %v, expected < %v", totalDuration, targetTotal)
	}
}

// TestE2E_LargeBoard_Performance tests performance with 10000+ tasks.
func TestE2E_LargeBoard_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large board test in short mode")
	}

	app := setupPerfTestApp(t)
	board := createPerfBoard(t, app, "LARGE", "manual")

	// Create 10000 tasks with various columns
	numTasks := 10000
	t.Logf("Creating %d tasks...", numTasks)
	createStart := time.Now()

	columns := []string{"backlog", "todo", "in_progress", "need_input", "review", "done"}
	needInputCount := 0
	for i := 0; i < numTasks; i++ {
		col := columns[i%len(columns)]
		if col == "need_input" {
			needInputCount++
		}
		createPerfTask(t, app, board.Id, fmt.Sprintf("Large board task %d", i), col)
	}
	createDuration := time.Since(createStart)
	t.Logf("Created %d tasks in %v (%.2f tasks/sec)", numTasks, createDuration, float64(numTasks)/createDuration.Seconds())

	// Test listing all tasks
	listAllStart := time.Now()
	allTasks, err := app.FindRecordsByFilter(
		"tasks",
		"board = {:boardId}",
		"-updated",
		0,
		0,
		map[string]any{"boardId": board.Id},
	)
	listAllDuration := time.Since(listAllStart)

	if err != nil {
		t.Fatalf("failed to list all tasks: %v", err)
	}
	if len(allTasks) != numTasks {
		t.Errorf("expected %d tasks, got %d", numTasks, len(allTasks))
	}
	t.Logf("List all %d tasks: %v", numTasks, listAllDuration)

	// Test listing need_input tasks
	listNeedInputStart := time.Now()
	needInputTasks, err := app.FindRecordsByFilter(
		"tasks",
		"board = {:boardId} && column = 'need_input'",
		"-updated",
		0,
		0,
		map[string]any{"boardId": board.Id},
	)
	listNeedInputDuration := time.Since(listNeedInputStart)

	if err != nil {
		t.Fatalf("failed to list need_input tasks: %v", err)
	}
	t.Logf("List %d need_input tasks: %v (target: %v)", len(needInputTasks), listNeedInputDuration, TargetListNeedInput)

	assertPerformanceTarget(t, "List need_input tasks", listNeedInputDuration, TargetListNeedInput)

	// Verify count
	if len(needInputTasks) != needInputCount {
		t.Errorf("expected %d need_input tasks, got %d", needInputCount, len(needInputTasks))
	}
}

// TestE2E_SessionLinking_Performance tests session operations at scale.
func TestE2E_SessionLinking_Performance(t *testing.T) {
	app := setupPerfTestApp(t)
	board := createPerfBoard(t, app, "SESS", "manual")

	// Create 100 tasks with sessions
	numTasks := 100
	tasks := make([]*core.Record, numTasks)
	for i := 0; i < numTasks; i++ {
		tasks[i] = createPerfTask(t, app, board.Id, fmt.Sprintf("Session test task %d", i), "in_progress")
	}

	// Time session linking
	linkStart := time.Now()
	for i, task := range tasks {
		task.Set("agent_session", map[string]any{
			"tool":        "claude-code",
			"ref":         fmt.Sprintf("session-%d", i),
			"ref_type":    "uuid",
			"working_dir": "/tmp/project",
			"linked_at":   time.Now().UTC().Format(time.RFC3339),
		})
		if err := app.Save(task); err != nil {
			t.Fatalf("failed to link session: %v", err)
		}
	}
	linkDuration := time.Since(linkStart)
	avgLinkTime := linkDuration / time.Duration(numTasks)

	t.Logf("Session Linking Performance:")
	t.Logf("  Total time for %d sessions: %v", numTasks, linkDuration)
	t.Logf("  Average time per session: %v (target: %v)", avgLinkTime, TargetSessionLink)

	assertPerformanceTarget(t, "Session link (avg)", avgLinkTime, TargetSessionLink)
}

// ========== Helper Types ==========

type WorkflowPerformanceResults struct {
	BlockTask        time.Duration
	AddComment       time.Duration
	AutoResumeTrigger time.Duration
	ResumeGeneration time.Duration
	ListComments     time.Duration
}

// ========== Assertion Helpers ==========

func assertPerformanceTarget(t *testing.T, name string, actual, target time.Duration) {
	t.Helper()
	if actual > target {
		t.Errorf("%s: %v exceeded target of %v", name, actual, target)
	}
}

// ========== Setup Helpers ==========

func setupPerfTestApp(t *testing.T) *pocketbase.PocketBase {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "egenskriven-perf-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	// Create app
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: tmpDir,
	})
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("failed to bootstrap app: %v", err)
	}

	// Setup collections
	setupPerfCollections(t, app)

	return app
}

func setupPerfCollections(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	// Boards collection
	if _, err := app.FindCollectionByNameOrId("boards"); err != nil {
		boards := core.NewBaseCollection("boards")
		boards.Fields.Add(&core.TextField{Name: "name", Required: true})
		boards.Fields.Add(&core.TextField{Name: "prefix", Required: true})
		boards.Fields.Add(&core.SelectField{
			Name:     "resume_mode",
			Required: true,
			Values:   []string{"auto", "manual"},
		})
		boards.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
		if err := app.Save(boards); err != nil {
			t.Fatalf("failed to create boards collection: %v", err)
		}
	}

	// Tasks collection with indexes
	if _, err := app.FindCollectionByNameOrId("tasks"); err != nil {
		boards, _ := app.FindCollectionByNameOrId("boards")
		tasks := core.NewBaseCollection("tasks")
		tasks.Fields.Add(&core.RelationField{
			Name:         "board",
			CollectionId: boards.Id,
			MaxSelect:    1,
			Required:     true,
		})
		tasks.Fields.Add(&core.TextField{Name: "title", Required: true})
		tasks.Fields.Add(&core.TextField{Name: "description"})
		tasks.Fields.Add(&core.SelectField{
			Name:     "type",
			Required: true,
			Values:   []string{"bug", "feature", "chore"},
		})
		tasks.Fields.Add(&core.SelectField{
			Name:     "priority",
			Required: true,
			Values:   []string{"low", "medium", "high", "urgent"},
		})
		tasks.Fields.Add(&core.SelectField{
			Name:     "column",
			Required: true,
			Values:   []string{"backlog", "todo", "in_progress", "need_input", "review", "done"},
		})
		tasks.Fields.Add(&core.NumberField{Name: "position"})
		tasks.Fields.Add(&core.JSONField{Name: "history"})
		tasks.Fields.Add(&core.JSONField{Name: "agent_session"})
		tasks.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
		tasks.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
		// Add indexes
		tasks.Indexes = []string{
			"CREATE INDEX idx_tasks_board ON tasks (board)",
			"CREATE INDEX idx_tasks_column ON tasks (column)",
		}
		if err := app.Save(tasks); err != nil {
			t.Fatalf("failed to create tasks collection: %v", err)
		}
	}

	// Comments collection with indexes
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
		// Add indexes
		comments.Indexes = []string{
			"CREATE INDEX idx_comments_task ON comments (task)",
			"CREATE INDEX idx_comments_created ON comments (created)",
			"CREATE INDEX idx_comments_task_created ON comments (task, created)",
		}
		if err := app.Save(comments); err != nil {
			t.Fatalf("failed to create comments collection: %v", err)
		}
	}
}

func createPerfBoard(t *testing.T, app *pocketbase.PocketBase, prefix, resumeMode string) *core.Record {
	t.Helper()
	collection, _ := app.FindCollectionByNameOrId("boards")
	record := core.NewRecord(collection)
	record.Set("name", prefix+" Board")
	record.Set("prefix", prefix)
	record.Set("resume_mode", resumeMode)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create board: %v", err)
	}
	return record
}

func createPerfTask(t *testing.T, app *pocketbase.PocketBase, boardId, title, column string) *core.Record {
	t.Helper()
	collection, _ := app.FindCollectionByNameOrId("tasks")
	record := core.NewRecord(collection)
	record.Set("board", boardId)
	record.Set("title", title)
	record.Set("type", "feature")
	record.Set("priority", "medium")
	record.Set("column", column)
	record.Set("position", 0)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	return record
}

func createPerfComment(t *testing.T, app *pocketbase.PocketBase, taskId, content, authorType string) *core.Record {
	t.Helper()
	collection, _ := app.FindCollectionByNameOrId("comments")
	record := core.NewRecord(collection)
	record.Set("task", taskId)
	record.Set("content", content)
	record.Set("author_type", authorType)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}
	return record
}

func addHistory(task *core.Record, action, actor string) {
	history := task.Get("history")
	var historySlice []map[string]any
	if history != nil {
		if data, err := json.Marshal(history); err == nil {
			json.Unmarshal(data, &historySlice)
		}
	}
	if historySlice == nil {
		historySlice = []map[string]any{}
	}
	historySlice = append(historySlice, map[string]any{
		"action":     action,
		"actor":      actor,
		"from":       "in_progress",
		"to":         "need_input",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
	task.Set("history", historySlice)
}
