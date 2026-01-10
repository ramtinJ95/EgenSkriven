# Performance Tuning Guide

This guide covers performance characteristics of EgenSkriven and recommendations for optimal operation.

## Performance Targets

| Operation | Target | Typical Performance |
|-----------|--------|---------------------|
| Block task | < 100ms | ~570μs |
| Add comment | < 100ms | ~360μs |
| List comments | < 200ms | ~10ms/1000 |
| Resume generation | < 50ms | ~7μs |
| Auto-resume trigger | < 1s | ~1μs |
| Session link | < 50ms | ~280μs |
| List --need-input | < 100ms | ~30ms/10000 |

## SQLite Configuration

EgenSkriven uses PocketBase with SQLite as the database backend. SQLite is well-suited for single-user or small-team usage patterns.

### Recommended Settings

The default SQLite settings work well for most use cases. For high-volume scenarios:

1. **Journal Mode**: WAL (Write-Ahead Logging) is enabled by default
   - Provides better concurrency for read/write operations
   - Enables non-blocking reads during writes

2. **Synchronous Mode**: Normal (default)
   - Balances durability and performance
   - Consider `FULL` for critical data integrity requirements

3. **Cache Size**: Default is adequate for typical use
   - Increase if working with very large boards (10000+ tasks)

### Data Location

By default, data is stored in `~/.egenskriven/`. Ensure this location:
- Has sufficient disk space
- Uses SSD storage for best performance
- Is backed up regularly

## Database Indexes

Critical indexes for performance (created automatically via migrations):

| Collection | Index | Purpose |
|------------|-------|---------|
| tasks | `idx_tasks_column` | Speeds up `--need-input` filter |
| comments | `idx_comments_task` | Speeds up comment listing by task |
| comments | `idx_comments_task_created` | Optimizes sorted comment queries |
| sessions | `idx_sessions_external_ref` | Fast session lookup |

### Verifying Indexes

Check index usage with EXPLAIN:

```sql
EXPLAIN QUERY PLAN SELECT * FROM tasks WHERE column = 'need_input';
-- Should show: USING INDEX idx_tasks_column
```

## Scaling Guidelines

### Comments per Task

- **Tested**: 1000+ comments per task
- **Performance**: ~10ms to list 1000 comments
- **Recommendation**: No practical limit for typical use

### Tasks per Board

- **Tested**: 10000+ tasks per board
- **Performance**: ~30ms to filter by column
- **Recommendation**: Consider archiving completed tasks if board exceeds 50000 tasks

### Concurrent Sessions

- **Tested**: 100+ concurrent operations
- **Performance**: ~800 ops/sec
- **Recommendation**: Single-user tool; concurrent access from multiple agents works well

### Comment Content Size

- **Limit**: 5000 characters per comment (schema enforced)
- **Recommendation**: Keep comments focused; use task description for longer context

## Performance Monitoring

### Key Metrics to Watch

1. **Response Time**: Operations should complete within targets above
2. **Database Size**: Monitor `~/.egenskriven/pb_data/data.db` growth
3. **Memory Usage**: Should remain stable during extended sessions

### Using Benchmarks

Run performance benchmarks to measure your system:

```bash
# Unit benchmarks
go test -bench=. -benchmem ./internal/commands/ ./internal/resume/

# E2E performance tests
go test -v -tags=performance ./tests/performance/
```

### Detecting Regressions

Compare benchmark results over time:

```bash
# Save baseline
go test -bench=. ./internal/commands/ > baseline.txt

# After changes, compare
go test -bench=. ./internal/commands/ > current.txt
```

## Troubleshooting

### Slow Operations

1. **Check index usage**: Run EXPLAIN on slow queries
2. **Verify data volume**: Large boards may need cleanup
3. **Check disk I/O**: Ensure SSD or fast storage
4. **Review concurrent access**: Lock contention is rare but possible

### High Memory Usage

1. **Check comment volumes**: Very large conversation threads use more memory
2. **Run GC**: Memory is reclaimed after operations complete
3. **Review prompt building**: Large context prompts allocate temporary memory

### Database Locked Errors

1. **Single writer at a time**: SQLite allows one writer
2. **Use WAL mode**: Should be enabled by default
3. **Check long-running transactions**: Ensure operations complete promptly

## Performance Testing

### Running Full Suite

```bash
# All benchmarks
make bench

# Performance tests only
go test -v -tags=performance -timeout 600s ./tests/performance/
```

### CI Integration

Performance tests run automatically on PRs via GitHub Actions. Results are uploaded as artifacts for historical comparison.

## Reference

For detailed performance notes and test results, see [docs/performance-notes.md](performance-notes.md).
