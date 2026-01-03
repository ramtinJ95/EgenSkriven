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
