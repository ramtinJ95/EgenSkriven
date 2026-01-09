# Performance Notes

Documentation of performance characteristics for the EgenSkriven collaborative
workflow feature.

## Collaborative Workflow Performance

### Expected Performance

| Operation | Target | Notes |
|-----------|--------|-------|
| Block task | < 100ms | Atomic transaction (move + comment) |
| Add comment | < 100ms | With mention extraction |
| List comments | < 200ms | Up to 100 comments per task |
| Resume generation | < 50ms | Context building in memory |
| Auto-resume trigger | < 1s | Background goroutine execution |
| Session link | < 50ms | Single record update |
| List --need-input | < 100ms | Indexed query |

### Scaling Considerations

| Metric | Expected Limit | Notes |
|--------|----------------|-------|
| Comments per task | 1000+ | Pagination available |
| Tasks per board | 10000+ | Indexed queries |
| Concurrent sessions | 100+ | SQLite handles well |
| Comment content size | 50KB | Reasonable for context |

## Database Indexes

The following indexes are created by migrations to ensure fast queries:

### Comments Collection

```sql
-- Fast comment lookup by task
CREATE INDEX idx_comments_task ON comments (task);

-- Chronological ordering for thread display
CREATE INDEX idx_comments_created ON comments (created);
```

### Sessions Collection

```sql
-- Session lookup by task
CREATE INDEX idx_sessions_task ON sessions (task);

-- Session lookup by external reference (for resume)
CREATE INDEX idx_sessions_external_ref ON sessions (external_ref);

-- Status-based queries (active sessions)
CREATE INDEX idx_sessions_status ON sessions (status);
```

### Tasks Collection

Existing indexes plus:

```sql
-- Fast lookup of tasks needing input
-- (Covered by existing column index)
```

## Optimizations

### Comment Loading

1. **Ascending sort**: Comments are fetched oldest-first for chronological
   display in the conversation thread
2. **Task filter**: Always filtered by task ID using indexed query
3. **Pagination**: Large comment threads can be paginated with `--limit`

### Context Prompt Building

1. **In-memory construction**: No additional DB queries after initial fetch
2. **String builder**: Efficient string concatenation for prompt
3. **Template caching**: Prompt templates are parsed once

### Auto-Resume

1. **Goroutine execution**: Auto-resume runs in background goroutine
2. **Non-blocking**: Comment creation returns immediately
3. **Single trigger**: Only first `@agent` mention triggers resume

### Real-time Updates

1. **PocketBase SSE**: Uses built-in Server-Sent Events for real-time
2. **Targeted subscriptions**: UI subscribes only to relevant collections
3. **Minimal payload**: Only changed fields sent in updates

## Query Patterns

### Most Common Queries

```sql
-- List comments for a task (most common)
SELECT * FROM comments WHERE task = ? ORDER BY created ASC;

-- List tasks needing input
SELECT * FROM tasks WHERE column = 'need_input' ORDER BY updated DESC;

-- Get session for task
SELECT agent_session FROM tasks WHERE id = ?;

-- Check for @agent mention
SELECT * FROM comments
WHERE task = ? AND metadata LIKE '%@agent%'
ORDER BY created DESC LIMIT 1;
```

### Query Complexity

| Query | Complexity | Index Used |
|-------|------------|------------|
| Comments by task | O(n) | idx_comments_task |
| Tasks by column | O(n) | existing column index |
| Session by ref | O(1) | idx_sessions_external_ref |
| Latest comment | O(1) | idx_comments_created |

## Memory Usage

| Component | Memory | Notes |
|-----------|--------|-------|
| Comment record | ~1KB | Typical size |
| Context prompt | ~10KB | Includes all comments |
| Session record | ~500B | Minimal metadata |
| Task with session | +500B | agent_session JSON field |

## Benchmarking

To benchmark the collaborative workflow:

```bash
# Build with profiling
go build -o egenskriven-bench ./cmd/egenskriven

# Run with timing
time egenskriven block TASK-1 "test question"
time egenskriven comments TASK-1 --json
time egenskriven resume TASK-1 --json

# Profile specific operations
go test -bench=BenchmarkBlock ./internal/commands/
go test -bench=BenchmarkResume ./internal/commands/
```

## Recommendations

1. **Keep comment threads reasonable**: Very long threads (500+ comments)
   may slow down context building

2. **Use --limit for large histories**: When listing many comments, use
   pagination to reduce memory usage

3. **Monitor SQLite file size**: The database file grows with comments;
   consider archival for very long-running projects

4. **Session cleanup**: Old/abandoned sessions should be periodically
   cleaned up to maintain query performance
