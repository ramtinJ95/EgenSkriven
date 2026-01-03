package ui

import (
	"io/fs"
)

// DistFS will hold the embedded React build output.
// For now, it's a placeholder that will be populated in Phase 2.
//
// In Phase 2, this will be replaced with:
//
//	//go:embed all:dist
//	var distDir embed.FS
//	var DistFS, _ = fs.Sub(distDir, "dist")
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
