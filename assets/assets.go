package assets

import "embed"

//go:embed style/*.css
var files embed.FS

// Files returns a filesystem with static files.
func Files() embed.FS {
	return files
}
