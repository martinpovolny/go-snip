package main

import (
	"embed"
	"io/fs"
)

//go:embed web
var embeddedWeb embed.FS

// webFS returns the "web/" subtree as a filesystem root.
func webFS() fs.FS {
	sub, _ := fs.Sub(embeddedWeb, "web")
	return sub
}
