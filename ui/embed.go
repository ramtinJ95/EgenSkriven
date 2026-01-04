package ui

import (
	"embed"
	"io/fs"
)

// DistFS holds the embedded React build output.
//
// During development (when dist/ doesn't exist), this will cause a build error.
// Run `cd ui && npm run build` first to create the dist/ directory.
//
// The "all:" prefix includes files starting with "." or "_" which Vite may create.

//go:embed all:dist
var distDir embed.FS

// DistFS is the filesystem containing the built React application.
// It's a sub-filesystem rooted at "dist" for cleaner path handling.
var DistFS, _ = fs.Sub(distDir, "dist")
