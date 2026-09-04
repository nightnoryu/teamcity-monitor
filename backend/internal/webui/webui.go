// Package webui exposes the compiled frontend bundle as an embedded filesystem
// so the backend can serve it directly from a single binary.
package webui

import (
	"embed"
	"io/fs"

	"github.com/go-faster/errors"
)

// dist holds the built frontend bundle. In local development it only contains a
// placeholder index.html; the real bundle is produced by `vite build` and copied
// in during the container image build.
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
