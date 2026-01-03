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
func CreateTestCollection(t *testing.T, app *pocketbase.PocketBase, name string, fields ...core.Field) *core.Collection {
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
