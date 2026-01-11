package main

import (
	"log"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
	"github.com/ramtinJ95/EgenSkriven/internal/commands"
	"github.com/ramtinJ95/EgenSkriven/internal/hooks"
	_ "github.com/ramtinJ95/EgenSkriven/migrations" // Auto-register migrations
	"github.com/ramtinJ95/EgenSkriven/ui"
)

func main() {
	// Resolve data directory from environment variable or use PocketBase default
	var app *pocketbase.PocketBase
	if dataDir := os.Getenv("EGENSKRIVEN_DIR"); dataDir != "" {
		app = pocketbase.NewWithConfig(pocketbase.Config{
			DefaultDataDir: dataDir,
		})
	} else {
		app = pocketbase.New()
	}

	// Hook: Run app migrations after bootstrap
	// Bootstrap() only runs system migrations, so we need to explicitly
	// run app migrations (our custom collections like tasks, comments, etc.)
	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		// Run app migrations after system migrations complete
		return e.App.RunAppMigrations()
	})

	// Register custom CLI commands
	commands.Register(app)

	// Register comment hooks for auto-resume functionality
	hooks.RegisterCommentHooks(app)

	// Hook: Assign sequence number to tasks created via API
	// This ensures the UI doesn't need to handle sequence assignment,
	// avoiding race conditions when multiple tasks are created concurrently.
	app.OnRecordCreate("tasks").BindFunc(func(e *core.RecordEvent) error {
		record := e.Record

		// Only assign seq if task has a board but no seq yet
		boardID := record.GetString("board")
		if boardID != "" && record.GetInt("seq") == 0 {
			seq, err := board.GetAndIncrementSequence(app, boardID)
			if err != nil {
				return err
			}
			record.Set("seq", seq)
		}

		return e.Next()
	})

	// Serve embedded React UI for non-API routes
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Catch-all route for the React SPA
		// This runs AFTER PocketBase's built-in routes (/api/*, /_/*)
		e.Router.GET("/{path...}", func(re *core.RequestEvent) error {
			path := re.Request.PathValue("path")

			// Skip API and admin routes (handled by PocketBase)
			if strings.HasPrefix(path, "api/") || strings.HasPrefix(path, "_/") {
				return re.Next()
			}

			// Try to serve the exact file from embedded filesystem
			// This handles JS, CSS, images, etc.
			if f, err := ui.DistFS.Open(path); err == nil {
				f.Close()
				return re.FileFS(ui.DistFS, path)
			}

			// For all other paths, serve index.html (SPA client-side routing)
			// This enables React Router to handle /board, /task/123, etc.
			return re.FileFS(ui.DistFS, "index.html")
		})

		return e.Next()
	})

	// Start the application
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
