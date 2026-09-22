// Package webui exposes the compiled frontend bundle as an embedded filesystem
// so the backend can serve it directly from a single binary.
package webui

import (
	"embed"
	"io/fs"

	"github.com/go-faster/errors"
)

// dist holds the built frontend bundle. The production build task stages the
// Vite output here before compiling Go; development builds use the placeholder.
//
//go:embed all:dist
var dist embed.FS

// Assets returns the frontend bundle rooted so that index.html sits at the FS root.
func Assets() (fs.FS, error) {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		return nil, errors.Wrap(err, "sub dist fs")
	}
	return sub, nil
}
