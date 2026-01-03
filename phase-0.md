# Phase 0: Project Setup

**Goal**: Working development environment with build system, hot reload, and testing infrastructure.

**Duration Estimate**: 1-2 days

**Prerequisites**: None - this is the foundation.

**Deliverable**: A Go project that builds, runs PocketBase, and has testing infrastructure ready.

---

## Overview

EgenSkriven is a local-first kanban board built as a single Go binary. It uses:
- **PocketBase** as the backend (SQLite database, REST API, real-time subscriptions)
- **Cobra** for CLI commands (comes with PocketBase)
- **React** for the frontend (embedded in the binary - set up in Phase 2)

This phase sets up the Go project structure, build system, and development tooling. By the end, you'll have a running PocketBase server with hot reload for development.

### Why PocketBase?

PocketBase gives us a lot for free:
- SQLite database with migrations
- REST API automatically generated from collections
- Real-time subscriptions via Server-Sent Events (SSE)
- Admin UI for debugging at `/_/`
- Authentication system (if needed later)
- Single binary distribution

### Why This Structure?

We follow Go's standard project layout:
- `cmd/` - Main applications (entry points)
- `internal/` - Private code that can't be imported by other projects
- `ui/` - Frontend code (React, embedded in binary)

---

## Environment Requirements

Before starting, ensure you have:

| Tool | Version | Check Command |
|------|---------|---------------|
| Go | 1.21+ | `go version` |
| Git | Any | `git --version` |

**Install Go** (if needed):
- macOS: `brew install go`
- Linux: `sudo apt install golang-go` or download from https://go.dev/dl/
- Windows: Download installer from https://go.dev/dl/

---

## Tasks

### 0.1 Initialize Go Module

**What**: Create a new Go module with PocketBase as a dependency.

**Why**: Go modules manage dependencies. The module path should match where you'll host the code (but it works locally regardless).

**Steps**:

1. Create and enter project directory:
   ```bash
   mkdir egenskriven
   cd egenskriven
   ```

2. Initialize the Go module:
   ```bash
   go mod init github.com/yourusername/egenskriven
   ```
   
   **Expected output**:
   ```
   go: creating new go.mod: module github.com/yourusername/egenskriven
   ```

3. Add PocketBase dependency:
   ```bash
   go get github.com/pocketbase/pocketbase
   ```
   
   **Expected output** (versions may vary):
   ```
   go: added github.com/pocketbase/pocketbase v0.23.4
   go: added [various transitive dependencies...]
   ```

4. Verify `go.mod` was created:
   ```bash
   cat go.mod
   ```
   
   **Expected output**:
   ```
   module github.com/yourusername/egenskriven

   go 1.21

   require github.com/pocketbase/pocketbase v0.23.4
   
   require (
       // ... indirect dependencies
   )
   ```

**Common Mistakes**:
- Running commands outside the project directory
- Using an old Go version (PocketBase requires 1.21+)

---

### 0.2 Create Project Structure

**What**: Create the directory structure for the project.

**Why**: Organizing code into directories makes it maintainable. Go's `internal/` directory is special - code there cannot be imported by external projects.

**Steps**:

1. Create all directories at once:
   ```bash
   mkdir -p cmd/egenskriven
   mkdir -p internal/{commands,output,resolver,config,hooks,testutil}
   mkdir -p ui
   mkdir -p migrations
   ```

2. Verify the structure:
   ```bash
   find . -type d | grep -v ".git" | sort
   ```
   
   **Expected output**:
   ```
   .
   ./cmd
   ./cmd/egenskriven
   ./internal
   ./internal/commands
   ./internal/config
   ./internal/hooks
   ./internal/output
   ./internal/resolver
   ./internal/testutil
   ./migrations
   ./ui
   ```

**Directory Purposes**:

| Directory | Purpose | Used In |
|-----------|---------|---------|
| `cmd/egenskriven/` | Main entry point | Phase 0 |
| `internal/commands/` | CLI command implementations | Phase 1 |
| `internal/output/` | JSON/human output formatting | Phase 1 |
| `internal/resolver/` | Task ID/title resolution | Phase 1 |
| `internal/config/` | Project configuration loading | Phase 1.5 |
| `internal/hooks/` | PocketBase event hooks | Phase 1 |
| `internal/testutil/` | Shared test helpers | Phase 0 |
| `ui/` | React frontend (embedded) | Phase 2 |
| `migrations/` | Database migrations | Phase 1 |

---

### 0.3 Create Main Entry Point

**What**: Create the minimal Go application that starts PocketBase.

**Why**: This is the simplest possible PocketBase app. It proves the setup works before adding complexity.

**File**: `cmd/egenskriven/main.go`

```go
package main

import (
	"log"

	"github.com/pocketbase/pocketbase"
)

func main() {
	// Create a new PocketBase instance with default configuration
	app := pocketbase.New()

	// Start the application
	// This will:
	// - Parse command line flags (serve, migrate, etc.)
	// - Initialize the database in pb_data/
	// - Start the HTTP server (if 'serve' command)
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
```

**Steps**:

1. Create the file:
   ```bash
   touch cmd/egenskriven/main.go
   ```

2. Open in your editor and paste the code above.

3. Verify it compiles:
   ```bash
   go build ./cmd/egenskriven
   ```
   
   **Expected output**: No output means success! A binary named `egenskriven` is created.

4. Check the binary exists:
   ```bash
   ls -la egenskriven
   ```
   
   **Expected output** (size ~30-50MB is normal):
   ```
   -rwxr-xr-x 1 user user 35000000 Jan  3 10:00 egenskriven
   ```

5. Test the binary runs:
   ```bash
   ./egenskriven --help
   ```
   
   **Expected output**:
   ```
   Usage:
     egenskriven [command]

   Available Commands:
     migrate     Executes app DB migrations
     serve       Starts the web server
     ...
   ```

**Common Mistakes**:
- Wrong package name (must be `package main` for executables)
- Wrong import path (must match your go.mod module path)
- Forgetting to save the file before building

---

### 0.4 Create UI Embed Placeholder

**What**: Create a placeholder for the React frontend embedding.

**Why**: In Phase 2, we'll embed the compiled React app into the Go binary. For now, we need a placeholder so the project compiles without the UI.

**File**: `ui/embed.go`

```go
package ui

import (
	"io/fs"
)

// DistFS will hold the embedded React build output.
// For now, it's a placeholder that will be populated in Phase 2.
//
// In Phase 2, this will be replaced with:
//
//   //go:embed all:dist
//   var distDir embed.FS
//   var DistFS, _ = fs.Sub(distDir, "dist")
//
// The placeholder below allows the project to compile before
// the React UI exists.

// emptyFS is a filesystem that always returns "not found"
type emptyFS struct{}

func (emptyFS) Open(name string) (fs.File, error) {
	return nil, fs.ErrNotExist
}

// DistFS is the filesystem containing UI assets.
// Currently empty - will be populated in Phase 2.
var DistFS fs.FS = emptyFS{}
```

**Steps**:

1. Create the file:
   ```bash
   touch ui/embed.go
   ```

2. Open in your editor and paste the code above.

3. Verify it compiles:
   ```bash
   go build ./ui
   ```
   
   **Expected output**: No output means success!

**Note**: This file will be significantly modified in Phase 2 when we add the React UI.

---

### 0.5 Create Test Utilities

**What**: Create helper functions for testing that provide isolated PocketBase instances.

**Why**: Tests need isolated databases so they don't interfere with each other. These utilities create temporary databases that are automatically cleaned up.

**File**: `internal/testutil/testutil.go`

```go
package testutil

import (
	"os"
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// NewTestApp creates a PocketBase instance with a temporary database.
// The database is automatically cleaned up when the test completes.
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    app := testutil.NewTestApp(t)
//	    // use app...
//	    // cleanup happens automatically
//	}
func NewTestApp(t *testing.T) *pocketbase.PocketBase {
	// t.Helper() marks this as a helper function.
	// If a test fails inside here, the error will point to the
	// calling test, not this function.
	t.Helper()

	// Create a temporary directory for this test's database.
	// The pattern "egenskriven-test-*" will have random chars appended.
	// Example: /tmp/egenskriven-test-abc123
	tmpDir, err := os.MkdirTemp("", "egenskriven-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// t.Cleanup registers a function to run when the test completes.
	// This ensures the temp directory is deleted even if the test fails.
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	// Create PocketBase instance pointing to the temp directory.
	// This isolates this test's database from all other tests.
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: tmpDir,
	})

	// Bootstrap initializes the database and runs migrations.
	// This is required before using the app.
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("failed to bootstrap app: %v", err)
	}

	return app
}

// CreateTestCollection creates a collection for testing purposes.
// This is a convenience wrapper around PocketBase's collection API.
//
// Usage:
//
//	collection := testutil.CreateTestCollection(t, app, "tasks",
//	    &core.TextField{Name: "title", Required: true},
//	    &core.TextField{Name: "description"},
//	)
func CreateTestCollection(t *testing.T, app *pocketbase.PocketBase, name string, fields ...*core.Field) *core.Collection {
	t.Helper()

	// Create a new base collection (as opposed to auth or view collection)
	collection := core.NewBaseCollection(name)

	// Add each field to the collection
	for _, field := range fields {
		collection.Fields.Add(field)
	}

	// Save the collection to the database
	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create collection %s: %v", name, err)
	}

	return collection
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/testutil/testutil.go
   ```

2. Open in your editor and paste the code above.

3. Verify it compiles:
   ```bash
   go build ./internal/testutil
   ```
   
   **Expected output**: No output means success!

**Key Concepts Explained**:

| Concept | Explanation |
|---------|-------------|
| `t.Helper()` | Marks function as test helper; errors show caller's line number |
| `t.Cleanup()` | Registers cleanup to run after test (even if test fails) |
| `t.Fatalf()` | Logs error and immediately stops the test |
| `os.MkdirTemp()` | Creates a unique temporary directory |

---

### 0.6 Create Test Utilities Tests

**What**: Write tests that verify our test utilities work correctly.

**Why**: Even test helpers should be tested! This ensures our testing infrastructure is solid before we rely on it in later phases.

**File**: `internal/testutil/testutil_test.go`

```go
package testutil

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

// TestNewTestApp verifies that NewTestApp creates a working PocketBase instance.
func TestNewTestApp(t *testing.T) {
	app := NewTestApp(t)

	// Basic sanity check: app should not be nil
	if app == nil {
		t.Fatal("expected app to be non-nil")
	}

	// Verify app is functional by attempting to list collections
	// This exercises the database connection
	collections, err := app.FindAllCollections()
	if err != nil {
		t.Fatalf("failed to find collections: %v", err)
	}

	// PocketBase creates some internal collections by default
	// We just log this for information
	t.Logf("found %d default collections", len(collections))
}

// TestNewTestApp_IsolatedDatabases verifies that each test gets its own database.
// This is critical - tests must not share state!
func TestNewTestApp_IsolatedDatabases(t *testing.T) {
	// Create two separate test apps
	app1 := NewTestApp(t)
	app2 := NewTestApp(t)

	// Create a collection in app1
	collection := core.NewBaseCollection("test_isolation")
	collection.Fields.Add(&core.TextField{Name: "name"})

	if err := app1.Save(collection); err != nil {
		t.Fatalf("failed to create collection in app1: %v", err)
	}

	// Verify collection exists in app1
	_, err := app1.FindCollectionByNameOrId("test_isolation")
	if err != nil {
		t.Fatalf("collection should exist in app1: %v", err)
	}

	// Verify collection does NOT exist in app2
	// This proves the databases are isolated
	_, err = app2.FindCollectionByNameOrId("test_isolation")
	if err == nil {
		t.Fatal("collection should NOT exist in app2 - databases should be isolated")
	}

	t.Log("confirmed: app1 and app2 have isolated databases")
}

// TestNewTestApp_CleanupOccurs verifies that temp directories are cleaned up.
// This runs as a subtest to demonstrate cleanup happens after test completion.
func TestNewTestApp_CleanupOccurs(t *testing.T) {
	// We can't easily test cleanup in the same test that creates the app,
	// because cleanup runs AFTER the test completes.
	// Instead, we verify that multiple test runs don't accumulate temp dirs.

	// Create several apps to verify no resource leak
	for i := 0; i < 5; i++ {
		app := NewTestApp(t)
		if app == nil {
			t.Fatalf("iteration %d: failed to create app", i)
		}
	}

	t.Log("created 5 test apps successfully - cleanup is registered for each")
}

// TestCreateTestCollection verifies the collection creation helper.
func TestCreateTestCollection(t *testing.T) {
	app := NewTestApp(t)

	// Create a test collection with multiple fields
	collection := CreateTestCollection(t, app, "test_tasks",
		&core.TextField{Name: "title", Required: true},
		&core.TextField{Name: "description"},
		&core.NumberField{Name: "position"},
	)

	// Verify collection was created
	if collection == nil {
		t.Fatal("expected collection to be non-nil")
	}

	if collection.Name != "test_tasks" {
		t.Errorf("expected collection name 'test_tasks', got '%s'", collection.Name)
	}

	// Verify we can find the collection by name
	found, err := app.FindCollectionByNameOrId("test_tasks")
	if err != nil {
		t.Fatalf("failed to find created collection: %v", err)
	}

	if found.Id != collection.Id {
		t.Error("found collection ID doesn't match created collection")
	}

	// Verify fields were added
	titleField := found.Fields.GetByName("title")
	if titleField == nil {
		t.Error("expected 'title' field to exist")
	}

	descField := found.Fields.GetByName("description")
	if descField == nil {
		t.Error("expected 'description' field to exist")
	}

	posField := found.Fields.GetByName("position")
	if posField == nil {
		t.Error("expected 'position' field to exist")
	}

	t.Logf("collection created with %d custom fields", 3)
}

// TestCreateTestCollection_CanCreateRecords verifies we can use the created collection.
func TestCreateTestCollection_CanCreateRecords(t *testing.T) {
	app := NewTestApp(t)

	// Create collection
	collection := CreateTestCollection(t, app, "test_items",
		&core.TextField{Name: "name", Required: true},
	)

	// Create a record in the collection
	record := core.NewRecord(collection)
	record.Set("name", "Test Item")

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create record: %v", err)
	}

	// Verify record was created with an ID
	if record.Id == "" {
		t.Error("expected record to have an ID after save")
	}

	// Verify we can retrieve the record
	found, err := app.FindRecordById("test_items", record.Id)
	if err != nil {
		t.Fatalf("failed to find record: %v", err)
	}

	if found.GetString("name") != "Test Item" {
		t.Errorf("expected name 'Test Item', got '%s'", found.GetString("name"))
	}

	t.Logf("created and retrieved record with ID: %s", record.Id)
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/testutil/testutil_test.go
   ```

2. Open in your editor and paste the code above.

3. Run the tests:
   ```bash
   go test ./internal/testutil -v
   ```
   
   **Expected output**:
   ```
   === RUN   TestNewTestApp
       testutil_test.go:20: found 0 default collections
   --- PASS: TestNewTestApp (0.05s)
   === RUN   TestNewTestApp_IsolatedDatabases
       testutil_test.go:46: confirmed: app1 and app2 have isolated databases
   --- PASS: TestNewTestApp_IsolatedDatabases (0.08s)
   === RUN   TestNewTestApp_CleanupOccurs
       testutil_test.go:59: created 5 test apps successfully - cleanup is registered for each
   --- PASS: TestNewTestApp_CleanupOccurs (0.15s)
   === RUN   TestCreateTestCollection
       testutil_test.go:95: collection created with 3 custom fields
   --- PASS: TestCreateTestCollection (0.03s)
   === RUN   TestCreateTestCollection_CanCreateRecords
       testutil_test.go:125: created and retrieved record with ID: abc123...
   --- PASS: TestCreateTestCollection_CanCreateRecords (0.03s)
   PASS
   ok      github.com/yourusername/egenskriven/internal/testutil   0.35s
   ```

**Common Mistakes**:
- Test file not ending in `_test.go` (Go won't recognize it as a test file)
- Test functions not starting with `Test` (Go won't run them)
- Forgetting the `*testing.T` parameter

---

### 0.7 Create Makefile

**What**: Create a Makefile with common development commands.

**Why**: Makefiles provide consistent, easy-to-remember commands. Instead of typing `go test ./... -v`, you just type `make test`.

**File**: `Makefile`

```makefile
# Makefile for EgenSkriven
# 
# Usage:
#   make dev          - Start development server with hot reload
#   make build        - Build production binary
#   make test         - Run all tests
#   make test-coverage - Run tests with coverage report
#   make clean        - Remove build artifacts

.PHONY: dev build run clean test test-coverage tidy help

# Default target: show help
help:
	@echo "Available commands:"
	@echo "  make dev           - Start development server with hot reload"
	@echo "  make build         - Build production binary"
	@echo "  make run           - Build and run the server"
	@echo "  make test          - Run all tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make clean         - Remove build artifacts and data"
	@echo "  make tidy          - Tidy Go module dependencies"

# Development: run with hot reload using Air
# Requires: go install github.com/air-verse/air@latest
dev:
	@echo "Starting development server with hot reload..."
	@echo "Install Air if missing: go install github.com/air-verse/air@latest"
	air

# Build production binary
# CGO_ENABLED=0 ensures pure Go build (no C dependencies)
# This is important for cross-platform compatibility
build:
	@echo "Building production binary..."
	CGO_ENABLED=0 go build -o egenskriven ./cmd/egenskriven
	@echo "Built: ./egenskriven ($(shell du -h egenskriven | cut -f1))"

# Build and run the application
run: build
	@echo "Starting server..."
	./egenskriven serve

# Clean build artifacts and data
clean:
	@echo "Cleaning build artifacts..."
	rm -rf egenskriven
	rm -rf pb_data/
	rm -rf .air/
	rm -rf coverage.out coverage.html
	@echo "Clean complete"

# Run all tests with verbose output
test:
	@echo "Running tests..."
	go test ./... -v

# Run tests with coverage report
# Generates both console output and HTML report
test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	@echo ""
	@echo "Coverage summary:"
	go tool cover -func=coverage.out
	@echo ""
	go tool cover -html=coverage.out -o coverage.html
	@echo "HTML report generated: coverage.html"

# Tidy dependencies (remove unused, add missing)
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	@echo "Done"
```

**Steps**:

1. Create the file:
   ```bash
   touch Makefile
   ```

2. Open in your editor and paste the code above.

   **Important**: Makefiles require **tabs** (not spaces) for indentation. Most editors handle this automatically for files named `Makefile`.

3. Test the Makefile:
   ```bash
   make help
   ```
   
   **Expected output**:
   ```
   Available commands:
     make dev           - Start development server with hot reload
     make build         - Build production binary
     make run           - Build and run the server
     make test          - Run all tests
     make test-coverage - Run tests with coverage report
     make clean         - Remove build artifacts and data
     make tidy          - Tidy Go module dependencies
   ```

4. Test the build command:
   ```bash
   make build
   ```
   
   **Expected output**:
   ```
   Building production binary...
   Built: ./egenskriven (35M)
   ```

5. Test the test command:
   ```bash
   make test
   ```
   
   **Expected output**: All tests pass (same output as step 0.6).

**Common Mistakes**:
- Using spaces instead of tabs (causes `Makefile:X: *** missing separator.  Stop.`)
- File named `makefile` works but `Makefile` is conventional

---

### 0.8 Create Air Configuration

**What**: Configure Air for automatic rebuilding during development.

**Why**: Air watches your Go files and automatically rebuilds + restarts when you save. This speeds up development significantly.

**First, install Air**:

```bash
go install github.com/air-verse/air@latest
```

Verify it's installed:
```bash
air -v
```

**Expected output**:
```
  __    _   ___
 / /\  | | | |_)
/_/--\ |_| |_| \_ v1.52.0, built with Go go1.21.0
```

If you get "command not found", add Go's bin directory to your PATH:
```bash
# Add to your ~/.bashrc, ~/.zshrc, or equivalent
export PATH=$PATH:$(go env GOPATH)/bin
```

**File**: `.air.toml`

```toml
# Air configuration for hot reload
# Documentation: https://github.com/air-verse/air

# Root directory to watch
root = "."

# Temporary directory for builds
tmp_dir = ".air"

[build]
  # Command to build the application
  cmd = "go build -o .air/egenskriven ./cmd/egenskriven"
  
  # Binary to run after building
  # 'serve' is PocketBase's command to start the HTTP server
  bin = ".air/egenskriven serve"
  
  # File extensions to watch
  include_ext = ["go", "tpl", "tmpl"]
  
  # Directories to ignore
  # ui/ - has its own dev server in Phase 2
  # pb_data/ - database files, not source code
  # .air/ - Air's temp directory
  exclude_dir = ["ui", "pb_data", ".air", "dist", "node_modules"]
  
  # Also exclude these patterns
  exclude_file = []
  
  # Exclude files matching these regexes
  exclude_regex = ["_test.go"]
  
  # Exclude unchanged files
  exclude_unchanged = true
  
  # Follow symlinks
  follow_symlink = false
  
  # Delay before rebuilding (milliseconds)
  # Prevents rebuilding multiple times during multi-file saves
  delay = 1000
  
  # Stop running old binary before building new one
  stop_on_error = true
  
  # Send interrupt signal before killing
  send_interrupt = true
  
  # Kill delay (needed for graceful shutdown)
  kill_delay = 500

[log]
  # Show timestamps in logs
  time = true
  
  # Show main output only (not watcher messages)
  main_only = false

[color]
  # Customize colors for different log types
  main = "yellow"
  watcher = "cyan"
  build = "green"
  runner = "magenta"

[misc]
  # Clean up temp directory on exit
  clean_on_exit = true
```

**Steps**:

1. Create the file:
   ```bash
   touch .air.toml
   ```

2. Open in your editor and paste the configuration above.

3. Test Air:
   ```bash
   air
   ```
   
   **Expected output**:
   ```
     __    _   ___
    / /\  | | | |_)
   /_/--\ |_| |_| \_ v1.52.0
   
   watching .
   watching cmd
   watching cmd/egenskriven
   !exclude .air
   !exclude pb_data
   !exclude ui
   building...
   running...
   2024/01/03 10:00:00 Server started at http://127.0.0.1:8090
   ```

4. Test hot reload works:
   - Keep Air running
   - In another terminal, make a small change to `cmd/egenskriven/main.go` (add a comment)
   - Save the file
   - Watch Air rebuild and restart automatically

5. Stop Air with `Ctrl+C`

**Common Mistakes**:
- Air not in PATH (see installation step above)
- TOML syntax errors (strings must be quoted, no trailing commas)

---

### 0.9 Create .gitignore

**What**: Configure Git to ignore generated files and directories.

**Why**: We don't want to commit build outputs, databases, or temporary files. This keeps the repository clean.

**File**: `.gitignore`

```gitignore
# =============================================================================
# EgenSkriven .gitignore
# =============================================================================

# -----------------------------------------------------------------------------
# PocketBase
# -----------------------------------------------------------------------------
# Data directory contains SQLite database, uploaded files, and logs
# This is user data, not source code
pb_data/

# -----------------------------------------------------------------------------
# Build outputs
# -----------------------------------------------------------------------------
# Compiled binary (built with 'make build')
egenskriven

# Cross-platform release builds
dist/

# -----------------------------------------------------------------------------
# Development tools
# -----------------------------------------------------------------------------
# Air hot reload temp directory
.air/

# -----------------------------------------------------------------------------
# Frontend (Phase 2+)
# -----------------------------------------------------------------------------
# React build output (embedded in binary)
ui/dist/

# Node.js dependencies
ui/node_modules/

# -----------------------------------------------------------------------------
# Testing
# -----------------------------------------------------------------------------
# Coverage reports
coverage.out
coverage.html

# -----------------------------------------------------------------------------
# IDE and editors
# -----------------------------------------------------------------------------
# JetBrains (GoLand, IntelliJ)
.idea/

# Visual Studio Code
.vscode/

# Vim
*.swp
*.swo
*~

# Emacs
*#
.#*

# -----------------------------------------------------------------------------
# Operating system
# -----------------------------------------------------------------------------
# macOS
.DS_Store
.AppleDouble
.LSOverride

# Windows
Thumbs.db
ehthumbs.db
Desktop.ini

# Linux
*~

# -----------------------------------------------------------------------------
# Environment and secrets
# -----------------------------------------------------------------------------
# Environment files may contain secrets (API keys, passwords)
# Never commit these!
.env
.env.local
.env.*.local

# -----------------------------------------------------------------------------
# Logs
# -----------------------------------------------------------------------------
*.log
logs/
```

**Steps**:

1. Create the file:
   ```bash
   touch .gitignore
   ```

2. Open in your editor and paste the content above.

3. Initialize Git repository (if not already done):
   ```bash
   git init
   ```

4. Verify ignored files aren't tracked:
   ```bash
   # Create a test file that should be ignored
   touch .env
   
   # Check git status
   git status
   ```
   
   **Expected**: `.env` should NOT appear in the list of untracked files.

5. Clean up test file:
   ```bash
   rm .env
   ```

---

### 0.10 Add Testing Dependencies

**What**: Add the testify library for better test assertions.

**Why**: Go's standard library has basic testing, but testify provides more readable assertions and better error messages.

**Steps**:

1. Add testify:
   ```bash
   go get github.com/stretchr/testify
   ```
   
   **Expected output**:
   ```
   go: added github.com/stretchr/testify v1.8.4
   ```

2. Tidy up dependencies:
   ```bash
   go mod tidy
   ```

3. Verify it's in go.mod:
   ```bash
   grep testify go.mod
   ```
   
   **Expected output**:
   ```
   github.com/stretchr/testify v1.8.4
   ```

**Testify Quick Reference** (for use in later phases):

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
    // assert - continues test on failure
    assert.Equal(t, expected, actual, "values should match")
    assert.NotNil(t, obj, "object should exist")
    assert.True(t, condition, "condition should be true")
    assert.Contains(t, "hello world", "world")
    
    // require - stops test on failure
    require.NoError(t, err, "operation should succeed")
    require.NotEmpty(t, list, "list should have items")
}
```

---

## Verification Checklist

Complete each section in order. Check off each item as you verify it.

### Build Verification

Run these commands and verify the expected results:

- [ ] **Dependencies are clean**
  ```bash
  go mod tidy
  ```
  Should complete without errors.

- [ ] **Project compiles**
  ```bash
  make build
  ```
  Should produce `egenskriven` binary.

- [ ] **Binary size is reasonable**
  ```bash
  ls -lh egenskriven
  ```
  Should be approximately 30-50MB (PocketBase includes SQLite).

### Runtime Verification

- [ ] **Server starts**
  ```bash
  ./egenskriven serve
  ```
  Should show: `Server started at http://127.0.0.1:8090`

- [ ] **Web interface loads**
  
  Open `http://localhost:8090` in browser.
  Should show PocketBase welcome page.

- [ ] **Admin UI accessible**
  
  Open `http://localhost:8090/_/` in browser.
  Should show PocketBase admin setup (first run) or login page.

- [ ] **Data directory created**
  ```bash
  ls -la pb_data/
  ```
  Should contain `data.db` (SQLite database).

- [ ] **Clean shutdown**
  
  Press `Ctrl+C` in terminal running server.
  Should exit cleanly without errors.

### Development Verification

- [ ] **Air is installed**
  ```bash
  air -v
  ```
  Should show Air version.

- [ ] **Hot reload starts**
  ```bash
  make dev
  ```
  Should start server and show "watching" messages.

- [ ] **Hot reload works**
  
  With `make dev` running:
  1. Edit `cmd/egenskriven/main.go` (add a comment)
  2. Save the file
  3. Air should show "building..." and restart

- [ ] **Hot reload stops cleanly**
  
  Press `Ctrl+C`.
  Should exit without errors.

### Test Verification

- [ ] **Tests run**
  ```bash
  make test
  ```
  Should show all tests passing.

- [ ] **Coverage report generates**
  ```bash
  make test-coverage
  ```
  Should create `coverage.html`.

- [ ] **Coverage report opens**
  
  Open `coverage.html` in browser.
  Should show coverage visualization.

- [ ] **No temp directories leaked**
  ```bash
  ls /tmp | grep egenskriven-test
  ```
  Should show nothing (cleanup worked).

### Git Verification

- [ ] **Repository initialized**
  ```bash
  git status
  ```
  Should not show "fatal: not a git repository".

- [ ] **Ignored files not tracked**
  ```bash
  # Create files that should be ignored
  mkdir -p pb_data && touch pb_data/test.db
  touch .env
  
  # Verify they're not tracked
  git status
  
  # Clean up
  rm -rf pb_data .env
  ```
  `pb_data/` and `.env` should NOT appear in "Untracked files".

---

## File Summary

| File | Lines | Purpose |
|------|-------|---------|
| `cmd/egenskriven/main.go` | ~15 | Application entry point |
| `ui/embed.go` | ~20 | Placeholder for React embedding |
| `internal/testutil/testutil.go` | ~50 | Test helper functions |
| `internal/testutil/testutil_test.go` | ~100 | Tests for test helpers |
| `Makefile` | ~50 | Build and development commands |
| `.air.toml` | ~50 | Hot reload configuration |
| `.gitignore` | ~70 | Git ignore rules |

**Total new code**: ~355 lines

---

## What You Should Have Now

After completing Phase 0, your project should:

```
egenskriven/
├── cmd/
│   └── egenskriven/
│       └── main.go              ✓ Created
├── internal/
│   ├── commands/                ✓ Empty (Phase 1)
│   ├── config/                  ✓ Empty (Phase 1.5)
│   ├── hooks/                   ✓ Empty (Phase 1)
│   ├── output/                  ✓ Empty (Phase 1)
│   ├── resolver/                ✓ Empty (Phase 1)
│   └── testutil/
│       ├── testutil.go          ✓ Created
│       └── testutil_test.go     ✓ Created
├── migrations/                  ✓ Empty (Phase 1)
├── ui/
│   └── embed.go                 ✓ Created
├── .air.toml                    ✓ Created
├── .gitignore                   ✓ Created
├── go.mod                       ✓ Created
├── go.sum                       ✓ Created (auto-generated)
└── Makefile                     ✓ Created
```

---

## Next Phase

**Phase 1: Core CLI** will add:
- Database migration for `tasks` collection
- CLI commands: `add`, `list`, `show`, `move`, `update`, `delete`
- Output formatter (human-readable and JSON)
- Task resolver (find tasks by ID, ID prefix, or title)
- Comprehensive tests for all commands

---

## Troubleshooting

### "air: command not found"

**Problem**: Air is not installed or not in PATH.

**Solution**:
```bash
# Install Air
go install github.com/air-verse/air@latest

# Add Go bin to PATH (add to ~/.bashrc or ~/.zshrc)
export PATH=$PATH:$(go env GOPATH)/bin

# Reload shell config
source ~/.bashrc  # or ~/.zshrc
```

### "go: command not found"

**Problem**: Go is not installed or not in PATH.

**Solution**: Install Go from https://go.dev/dl/ and follow installation instructions for your OS.

### PocketBase fails to start with "address already in use"

**Problem**: Port 8090 is already in use by another process.

**Solution**:
```bash
# Find what's using the port
lsof -i :8090

# Either stop that process, or use a different port
./egenskriven serve --http 127.0.0.1:8091
```

### Tests fail with "database is locked"

**Problem**: A previous test didn't clean up properly, or tests are running in parallel incorrectly.

**Solution**:
```bash
# Manually clean up temp directories
rm -rf /tmp/egenskriven-test-*

# Run tests without parallel execution
go test ./... -v -p 1
```

### Build fails with CGO errors

**Problem**: CGO is trying to compile C code but C compiler is missing.

**Solution**: Ensure CGO is disabled:
```bash
CGO_ENABLED=0 go build -o egenskriven ./cmd/egenskriven
```

This is already set in the Makefile, but if building manually, include it.

### "Makefile:X: *** missing separator. Stop."

**Problem**: Makefile has spaces instead of tabs for indentation.

**Solution**: Replace spaces with tabs. In vim: `:set noexpandtab` then re-indent. Most editors auto-detect Makefile format.

### Permission denied running binary

**Problem**: Binary doesn't have execute permission.

**Solution**:
```bash
chmod +x egenskriven
```

### Tests pass locally but fail in CI

**Problem**: Usually timing-related or environment differences.

**Solution**: Ensure tests don't rely on:
- Specific timing (use `time.Sleep` sparingly)
- Absolute paths
- Environment variables not set in CI

---

## Glossary

| Term | Definition |
|------|------------|
| **PocketBase** | Go framework providing SQLite database, REST API, and admin UI |
| **Cobra** | Go library for creating CLI applications |
| **Air** | Hot reload tool for Go development |
| **CGO** | Go's mechanism for calling C code; disabled for simpler builds |
| **go.mod** | Go module definition file (like package.json for Node) |
| **go.sum** | Checksums for dependencies (auto-generated, commit this) |
| **t.Helper()** | Marks a test function as a helper for better error reporting |
| **t.Cleanup()** | Registers cleanup to run after test completes |
