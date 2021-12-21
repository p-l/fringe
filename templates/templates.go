package templates

import (
	"embed"
)

//go:embed layouts/*.gohtml default/*.gohtml user/*.gohtml
var files embed.FS

// Files returns a filesystem with static files.
func Files() embed.FS {
	return files
}
