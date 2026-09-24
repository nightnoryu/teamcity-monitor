package main

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestSPAHandler_StaticAndFallbackRoutes(t *testing.T) {
	assets := fstest.MapFS{
		"index.html":    &fstest.MapFile{Data: []byte("<html>dashboard</html>")},
		"assets/app.js": &fstest.MapFile{Data: []byte("console.log('dashboard')")},
	}
	for _, test := range []struct{ path, body, cacheControl string }{
		{"/", "<html>dashboard</html>", htmlCache},
		{"/project/alpha", "<html>dashboard</html>", htmlCache},
		{"/assets/app.js", "console.log('dashboard')", immutableAssetCache},
		{"/assets/missing.js", "<html>dashboard</html>", htmlCache},
	} {
		t.Run(test.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			spaHandler(fs.FS(assets)).ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, http.NoBody))
			require.Equal(t, http.StatusOK, response.Code)
			require.Contains(t, response.Body.String(), test.body)
			require.Equal(t, test.cacheControl, response.Header().Get("Cache-Control"))
		})
	}
}
