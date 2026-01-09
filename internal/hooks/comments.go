// Package hooks provides PocketBase hooks for application-level event handling.
package hooks

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/autoresume"
)

// RegisterCommentHooks registers hooks for the comments collection.
// This enables auto-resume functionality when comments are created.
func RegisterCommentHooks(app *pocketbase.PocketBase) {
	autoResumeService := autoresume.NewService(app)

	// After comment is created successfully
	app.OnRecordAfterCreateSuccess("comments").BindFunc(func(e *core.RecordEvent) error {
		// Check if this comment should trigger auto-resume
		// Run in goroutine to not block the create response
		go func() {
			if err := autoResumeService.CheckAndResume(e.Record); err != nil {
				// Log error but don't fail the request
				app.Logger().Error("auto-resume check failed",
					"comment", e.Record.Id,
					"error", err,
				)
			}
		}()

		return e.Next()
	})
}
