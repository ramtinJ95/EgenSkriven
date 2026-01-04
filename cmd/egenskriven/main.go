package main

import (
	"log"

	"github.com/pocketbase/pocketbase"

	"github.com/ramtinJ95/EgenSkriven/internal/commands"
	_ "github.com/ramtinJ95/EgenSkriven/migrations" // Auto-register migrations
)

func main() {
	app := pocketbase.New()

	// Register custom CLI commands
	commands.Register(app)

	// Start the application
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
