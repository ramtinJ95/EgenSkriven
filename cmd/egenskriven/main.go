package main

import (
	"log"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/commands"
	_ "github.com/ramtinJ95/EgenSkriven/migrations" // Auto-register migrations
	"github.com/ramtinJ95/EgenSkriven/ui"
)

func main() {
	app := pocketbase.New()

	// Register custom CLI commands
	commands.Register(app)

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
