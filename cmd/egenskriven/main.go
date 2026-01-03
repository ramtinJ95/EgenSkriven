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
