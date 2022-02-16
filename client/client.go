package client

import (
	"embed"
	"io/fs"
)

//go:embed build/*
var files embed.FS

// Files returns a filesystem with static files.
func Files() fs.FS {
	sub, err := fs.Sub(files, "build")
	if err != nil {
		return embed.FS{}
	}

	return sub
}
