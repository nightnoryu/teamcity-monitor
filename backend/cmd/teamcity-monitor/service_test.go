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
	for _, test := range []struct{ path, body string }{
		{"/", "<html>dashboard</html>"},
		{"/project/alpha", "<html>dashboard</html>"},
		{"/assets/app.js", "console.log('dashboard')"},
	} {
		t.Run(test.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			spaHandler(fs.FS(assets)).ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, http.NoBody))
			require.Equal(t, http.StatusOK, response.Code)
			require.Contains(t, response.Body.String(), test.body)
		})
	}
}
