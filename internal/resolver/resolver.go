package resolver

import (
	"fmt"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// Resolution represents the result of resolving a task reference.
type Resolution struct {
	// Task is the resolved task (nil if not found or ambiguous)
	Task *core.Record
	// Matches contains all matching tasks (populated if ambiguous)
	Matches []*core.Record
}

// IsAmbiguous returns true if the resolution matched multiple tasks.
func (r *Resolution) IsAmbiguous() bool {
	return len(r.Matches) > 1
}

// IsNotFound returns true if no tasks matched.
func (r *Resolution) IsNotFound() bool {
	return r.Task == nil && len(r.Matches) == 0
}

// ResolveTask attempts to find a task by reference.
// Resolution order:
// 1. Exact ID match
// 2. ID prefix match (must be unique)
// 3. Title substring match (case-insensitive, must be unique)
func ResolveTask(app *pocketbase.PocketBase, ref string) (*Resolution, error) {
	// 1. Try exact ID match
	task, err := app.FindRecordById("tasks", ref)
	if err == nil {
		return &Resolution{Task: task}, nil
	}

	// 2. Try ID prefix match
	tasks, err := app.FindAllRecords("tasks",
		dbx.NewExp("id LIKE {:prefix}", dbx.Params{"prefix": ref + "%"}),
	)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if len(tasks) == 1 {
		return &Resolution{Task: tasks[0]}, nil
	}

	if len(tasks) > 1 {
		return &Resolution{Matches: tasks}, nil
	}

	// 3. Try title match (case-insensitive substring)
	tasks, err = app.FindAllRecords("tasks",
		dbx.NewExp("LOWER(title) LIKE {:title}",
			dbx.Params{"title": "%" + strings.ToLower(ref) + "%"}),
	)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	switch len(tasks) {
	case 0:
		return &Resolution{}, nil // not found
	case 1:
		return &Resolution{Task: tasks[0]}, nil
	default:
		return &Resolution{Matches: tasks}, nil // ambiguous
	}
}

// MustResolve resolves a task and returns an error if not found or ambiguous.
// This is a convenience wrapper for commands that need exactly one task.
func MustResolve(app *pocketbase.PocketBase, ref string) (*core.Record, error) {
	resolution, err := ResolveTask(app, ref)
	if err != nil {
		return nil, err
	}

	if resolution.IsNotFound() {
		return nil, fmt.Errorf("no task found matching: %s", ref)
	}

	if resolution.IsAmbiguous() {
		return nil, &AmbiguousError{
			Reference: ref,
			Matches:   resolution.Matches,
		}
	}

	return resolution.Task, nil
}

// AmbiguousError is returned when a reference matches multiple tasks.
type AmbiguousError struct {
	Reference string
	Matches   []*core.Record
}

func (e *AmbiguousError) Error() string {
	return fmt.Sprintf("ambiguous task reference: '%s' matches %d tasks", e.Reference, len(e.Matches))
}
