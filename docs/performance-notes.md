# Performance Notes

Documentation of performance characteristics for the EgenSkriven collaborative
workflow feature.

---

## Task 1: Implement Performance Benchmarks

**Goal**: Create benchmark tests to verify all performance targets are met.

**Status**: Not Started

### Subtask 1.1: Create Benchmark Test Infrastructure

- [x] Create `internal/commands/benchmark_test.go` for command benchmarks
- [x] Create `internal/resume/benchmark_test.go` for context building benchmarks
- [x] Create `internal/autoresume/benchmark_test.go` for auto-resume benchmarks
- [x] Add benchmark helper utilities for database seeding

**Files to create/modify**:
- `internal/commands/benchmark_test.go` (new)
- `internal/resume/benchmark_test.go` (new)
- `internal/autoresume/benchmark_test.go` (new)

### Subtask 1.2: Implement Operation Benchmarks

Create benchmarks for each operation with target verification:

| Operation | Target | Benchmark Function |
|-----------|--------|-------------------|
| Block task | < 100ms | `BenchmarkBlockTask` |
| Add comment | < 100ms | `BenchmarkAddComment` |
| List comments | < 200ms | `BenchmarkListComments` |
| Resume generation | < 50ms | `BenchmarkResumeGeneration` |
| Auto-resume trigger | < 1s | `BenchmarkAutoResumeTrigger` |
| Session link | < 50ms | `BenchmarkSessionLink` |
| List --need-input | < 100ms | `BenchmarkListNeedInput` |

- [x] `BenchmarkBlockTask` - Test atomic transaction (move + comment)
  - Location: `internal/commands/block.go:103-138`
- [x] `BenchmarkAddComment` - Test with mention extraction
  - Location: `internal/commands/comment.go:102-113`
- [x] `BenchmarkListComments` - Test up to 100 comments per task
  - Location: `internal/commands/comments.go:73-83`
- [x] `BenchmarkResumeGeneration` - Test context building (as `BenchmarkBuildContextPrompt`)
  - Location: `internal/resume/context.go:20-72`
- [x] `BenchmarkAutoResumeTrigger` - Test goroutine execution (as `BenchmarkCheckAndResume`)
  - Location: `internal/autoresume/service.go:34-90`
- [x] `BenchmarkSessionLink` - Test single record update
  - Location: `internal/commands/session.go:124-176`
- [x] `BenchmarkListNeedInput` - Test indexed query
  - Location: `internal/commands/list.go:118-120`

### Subtask 1.3: Implement Scaling Benchmarks

- [ ] `BenchmarkCommentsScaling` - Test 1000+ comments per task
- [ ] `BenchmarkTasksScaling` - Test 10000+ tasks per board
- [ ] `BenchmarkConcurrentSessions` - Test 100+ concurrent sessions
- [ ] `BenchmarkLargeCommentContent` - Test 50KB comment content

**Acceptance Criteria**:
```bash
go test -bench=. ./internal/commands/ ./internal/resume/ ./internal/autoresume/
# All benchmarks should pass with results within target thresholds
```

---

## Task 2: Verify Database Indexes

**Goal**: Ensure all documented indexes exist and are being used effectively.

**Status**: Not Started

### Subtask 2.1: Audit Existing Index Definitions

Verify indexes defined in migrations:

- [ ] `idx_comments_task` - `migrations/1700000014_comments_collection.go:83`
- [ ] `idx_comments_created` - `migrations/1700000014_comments_collection.go:84`
- [ ] `idx_sessions_task` - `migrations/1700000015_sessions_collection.go:91`
- [ ] `idx_sessions_external_ref` - `migrations/1700000015_sessions_collection.go:93`
- [ ] `idx_sessions_status` - `migrations/1700000015_sessions_collection.go:92`

### Subtask 2.2: Create Index Verification Tests

- [ ] Create `internal/db/index_test.go` with tests that verify:
  - Indexes exist after migration
  - Queries use expected indexes (EXPLAIN QUERY PLAN)
  - Index performance vs table scan comparison

### Subtask 2.3: Add Missing Indexes

- [ ] Verify task column index exists for `need_input` filter
- [ ] Add composite index if needed: `idx_comments_task_created`
- [ ] Document any new indexes in migration files

**Acceptance Criteria**:
```sql
-- All these should show "USING INDEX" in query plan
EXPLAIN QUERY PLAN SELECT * FROM comments WHERE task = ? ORDER BY created ASC;
EXPLAIN QUERY PLAN SELECT * FROM tasks WHERE column = 'need_input';
EXPLAIN QUERY PLAN SELECT * FROM sessions WHERE external_ref = ?;
```

---

## Task 3: Implement Template Caching

**Goal**: Add template caching for context prompt building to improve performance.

**Status**: Not Started (Currently not implemented)

### Subtask 3.1: Design Template Cache

- [ ] Define template structure for context prompts
- [ ] Choose caching strategy (sync.Once, sync.Pool, or custom cache)
- [ ] Document cache invalidation policy

### Subtask 3.2: Implement Template Cache

- [ ] Create `internal/resume/templates.go` with cached templates
- [ ] Refactor `BuildContextPrompt()` in `internal/resume/context.go:20-72`
- [ ] Refactor `BuildMinimalPrompt()` in `internal/resume/context.go:74-109`
- [ ] Add template parsing at package init time

### Subtask 3.3: Test Template Cache Performance

- [ ] Create `BenchmarkContextPromptWithCache` benchmark
- [ ] Create `BenchmarkContextPromptWithoutCache` for comparison
- [ ] Verify memory allocation reduction with `go test -benchmem`

**Acceptance Criteria**:
- Context prompt building should show measurable improvement
- No increase in memory leaks (verify with pprof)

---

## Task 4: Implement Real-time Updates (SSE)

**Goal**: Add Server-Sent Events support for real-time task updates.

**Status**: Not Started (Currently not implemented)

### Subtask 4.1: Design SSE Architecture

- [ ] Define SSE event types (task_updated, comment_added, session_changed)
- [ ] Design subscription model (per-board, per-task)
- [ ] Document payload structure for minimal data transfer

### Subtask 4.2: Implement SSE Server

- [ ] Create `internal/realtime/sse.go` with SSE handler
- [ ] Integrate with PocketBase hooks for event triggers
- [ ] Add connection management (heartbeat, cleanup)

**Hooks to integrate** (register in `internal/hooks/`):
- Task updates: `internal/hooks/tasks.go`
- Comment creation: `internal/hooks/comments.go:12-32`
- Session changes: New hook needed

### Subtask 4.3: Implement SSE Client Support

- [ ] Add `--watch` flag to relevant commands (`list`, `comments`)
- [ ] Create SSE subscription helper in `internal/client/sse.go`
- [ ] Handle reconnection and error recovery

### Subtask 4.4: Test SSE Performance

- [ ] `BenchmarkSSEBroadcast` - Test broadcast to 100+ clients
- [ ] `BenchmarkSSEPayloadSize` - Verify minimal payload
- [ ] Integration test for end-to-end SSE flow

**Acceptance Criteria**:
- SSE updates should arrive within 100ms of change
- Payload size should be < 1KB for typical updates
- Handle 100+ concurrent subscribers without degradation

---

## Task 5: Query Performance Verification

**Goal**: Verify all documented query patterns perform as expected.

**Status**: Not Started

### Subtask 5.1: Create Query Performance Tests

Test each documented query pattern:

- [ ] Test: Comments by task lookup
  - Query: `SELECT * FROM comments WHERE task = ? ORDER BY created ASC`
  - Location: `internal/commands/comments.go:73-80`
  - Expected: O(n) with index

- [ ] Test: Tasks needing input
  - Query: `SELECT * FROM tasks WHERE column = 'need_input' ORDER BY updated DESC`
  - Location: `internal/commands/list.go:119`
  - Expected: O(n) with index

- [ ] Test: Session by external ref
  - Query: `SELECT * FROM sessions WHERE external_ref = ?`
  - Location: `internal/commands/session.go:562-568`
  - Expected: O(1) with unique index

- [ ] Test: Latest comment for task
  - Query: `SELECT * FROM comments WHERE task = ? ORDER BY created DESC LIMIT 1`
  - Expected: O(1) with composite index

### Subtask 5.2: Add Query Plan Assertions

- [ ] Create `internal/db/queryplan_test.go`
- [ ] Add assertions that verify index usage via EXPLAIN
- [ ] Fail tests if queries fall back to table scan

### Subtask 5.3: Benchmark Query Patterns at Scale

- [ ] Test with 1000 comments per task
- [ ] Test with 10000 tasks per board
- [ ] Test with 100 concurrent query clients
- [ ] Document actual performance vs targets

**Acceptance Criteria**:
| Query | Target Complexity | Verified |
|-------|-------------------|----------|
| Comments by task | O(n) indexed | [ ] |
| Tasks by column | O(n) indexed | [ ] |
| Session by ref | O(1) | [ ] |
| Latest comment | O(1) | [ ] |

---

## Task 6: Memory Usage Verification

**Goal**: Verify memory usage matches documented expectations.

**Status**: Not Started

### Subtask 6.1: Create Memory Profiling Tests

- [ ] Create `internal/memory/profile_test.go`
- [ ] Add memory allocation benchmarks with `-benchmem`

### Subtask 6.2: Verify Component Memory Usage

| Component | Expected | Test Function |
|-----------|----------|---------------|
| Comment record | ~1KB | `TestCommentRecordMemory` |
| Context prompt | ~10KB | `TestContextPromptMemory` |
| Session record | ~500B | `TestSessionRecordMemory` |
| Task with session | +500B | `TestTaskWithSessionMemory` |

- [ ] `TestCommentRecordMemory` - Measure typical comment size
- [ ] `TestContextPromptMemory` - Measure with 100 comments
- [ ] `TestSessionRecordMemory` - Measure session metadata
- [ ] `TestTaskWithSessionMemory` - Measure agent_session JSON overhead

### Subtask 6.3: Add Memory Leak Detection

- [ ] Add long-running test with repeated operations
- [ ] Use `runtime.ReadMemStats` to track allocations
- [ ] Verify garbage collection reclaims memory

**Acceptance Criteria**:
```bash
go test -benchmem -bench=Memory ./internal/memory/
# All memory benchmarks should be within 20% of documented values
```

---

## Task 7: Optimize Auto-Resume Performance

**Goal**: Verify and optimize auto-resume goroutine execution.

**Status**: Partially Implemented

### Subtask 7.1: Verify Current Implementation

- [x] Goroutine execution: `internal/hooks/comments.go:19-27`
- [x] Background resume: `internal/autoresume/service.go:145`
- [ ] Single trigger guard: Verify in `internal/autoresume/service.go:254-286`

### Subtask 7.2: Benchmark Auto-Resume Path

- [ ] `BenchmarkCheckAndResume` - Full check path
- [ ] `BenchmarkHasAgentMention` - Mention detection
- [ ] `BenchmarkTriggerResume` - Resume execution setup

### Subtask 7.3: Optimize Hot Paths

- [ ] Profile `hasAgentMention()` for string scanning efficiency
- [ ] Consider caching recent mention checks
- [ ] Optimize comment fetch query if needed

**Acceptance Criteria**:
- Auto-resume trigger should complete in < 1s
- Comment creation should return immediately (non-blocking verified)
- No duplicate resume triggers for same comment

---

## Task 8: End-to-End Performance Test Suite

**Goal**: Create comprehensive E2E performance tests.

**Status**: Not Started

### Subtask 8.1: Create Performance Test Harness

- [ ] Create `tests/performance/` directory
- [ ] Add test fixtures for large datasets
- [ ] Create timing assertion helpers

### Subtask 8.2: Implement E2E Performance Scenarios

- [ ] `TestE2E_BlockResumeWorkflow_Performance` - Full workflow timing
- [ ] `TestE2E_HighVolumeComments_Performance` - 1000+ comments
- [ ] `TestE2E_ConcurrentOperations_Performance` - Parallel operations
- [ ] `TestE2E_LargeBoard_Performance` - 10000+ tasks

### Subtask 8.3: Add CI Performance Regression Tests

- [ ] Create GitHub Action for performance tests
- [ ] Store baseline metrics
- [ ] Alert on performance regression > 20%

**Acceptance Criteria**:
```bash
go test -v -tags=performance ./tests/performance/
# All E2E performance tests pass
# Results logged for historical tracking
```

---

## Task 9: Documentation and Monitoring

**Goal**: Add performance monitoring and documentation.

**Status**: Not Started

### Subtask 9.1: Add Performance Metrics Collection

- [ ] Create `internal/metrics/performance.go`
- [ ] Add timing instrumentation to key operations
- [ ] Support optional metrics export (prometheus/statsd)

### Subtask 9.2: Create Performance Monitoring Dashboard

- [ ] Document key metrics to monitor
- [ ] Create sample Grafana dashboard (if applicable)
- [ ] Add alerting thresholds

### Subtask 9.3: Update Documentation

- [ ] Add performance tuning guide to docs/
- [ ] Document SQLite configuration recommendations
- [ ] Add troubleshooting guide for performance issues

**Acceptance Criteria**:
- All performance targets documented with measurement methodology
- Monitoring can detect < 100ms to > 500ms degradation
- Runbook for performance troubleshooting

---

## Reference: Current Implementation Locations

### Collaborative Workflow Operations

| Operation | File | Key Lines |
|-----------|------|-----------|
| Block task | `internal/commands/block.go` | 103-138 |
| Add comment | `internal/commands/comment.go` | 102-113 |
| List comments | `internal/commands/comments.go` | 73-83 |
| Resume generation | `internal/resume/context.go` | 20-72 |
| Auto-resume | `internal/autoresume/service.go` | 34-90 |
| Session link | `internal/commands/session.go` | 124-176 |
| List --need-input | `internal/commands/list.go` | 118-120 |

### Database Migrations

| Collection | Migration File |
|------------|----------------|
| Tasks | `migrations/1700000001_initial.go` |
| need_input column | `migrations/1700000012_need_input_column.go` |
| agent_session | `migrations/1700000013_agent_session_field.go` |
| Comments + indexes | `migrations/1700000014_comments_collection.go` |
| Sessions + indexes | `migrations/1700000015_sessions_collection.go` |
| Board resume_mode | `migrations/1700000016_board_resume_mode.go` |

### Existing Tests

| Feature | Test File |
|---------|-----------|
| Block | `internal/commands/block_test.go` |
| Comment | `internal/commands/comment_test.go` |
| Resume | `internal/commands/resume_test.go` |
| Session | `internal/commands/session_test.go` |
| List | `internal/commands/list_test.go` |
| Auto-resume | `internal/autoresume/service_test.go` |
| Context | `internal/resume/context_test.go` |
| E2E | `tests/e2e/autoresume_test.go` |

---

## Performance Targets Summary

| Operation | Target | Priority |
|-----------|--------|----------|
| Block task | < 100ms | High |
| Add comment | < 100ms | High |
| List comments | < 200ms | High |
| Resume generation | < 50ms | High |
| Auto-resume trigger | < 1s | Medium |
| Session link | < 50ms | Medium |
| List --need-input | < 100ms | High |

| Scaling Metric | Target | Priority |
|----------------|--------|----------|
| Comments per task | 1000+ | Medium |
| Tasks per board | 10000+ | Medium |
| Concurrent sessions | 100+ | Low |
| Comment content size | 50KB | Low |
