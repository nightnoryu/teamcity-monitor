package main

import (
	"context"
	stderrors "errors"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/go-faster/errors"
	"github.com/gorilla/mux"
	"github.com/nightnoryu/go-kita/log"

	"teamcity-monitor/internal/webui"
)

const shutdownTimeout = 10 * time.Second

var errServiceStopped = stderrors.New("service stopped without errors")

func service(ctx context.Context, config *config, logger log.Logger) error {
	router := mux.NewRouter()

	router.HandleFunc("/resilience/live", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, http.StatusText(http.StatusOK))
	})

	assets, err := webui.Assets()
	if err != nil {
		return errors.Wrap(err, "load embedded web assets")
	}
	router.PathPrefix("/").Handler(spaHandler(assets))

	httpServer := &http.Server{
		Handler:           router,
		Addr:              config.ServeRESTAddress,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	// Shutdown must use a fresh context; ctx is canceled by this point - hence the nolints
	go func() { //nolint:gosec
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if shutdownErr := httpServer.Shutdown(shutdownCtx); shutdownErr != nil { //nolint:contextcheck
			logger.Error(shutdownErr, "failed to gracefully shut down HTTP server")
		}
	}()

	logger.Info("Listening and serving...")
	err = httpServer.ListenAndServe()
	return translateStopErr(err, errServiceStopped)
}

func spaHandler(assets fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(assets))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}

		if _, statErr := fs.Stat(assets, name); statErr != nil {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})
}

func translateStopErr(from, to error) error {
	switch {
	case errors.Is(from, http.ErrServerClosed):
		return to
	default:
		return from
	}
}
