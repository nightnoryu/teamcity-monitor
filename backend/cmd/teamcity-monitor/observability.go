package main

import (
	"net/http"
	"strings"

	"github.com/felixge/httpsnoop"
	"github.com/nightnoryu/go-kita/log"

	"teamcity-monitor/internal/observability"
)

const (
	routeHealth = "health"
	routeStatus = "status"
)

func observeHTTP(next http.Handler, metrics *observability.Metrics, logger log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := "page"
		switch {
		case r.URL.Path == "/livez" || r.URL.Path == "/healthz":
			route = routeHealth
		case r.URL.Path == "/api/status":
			route = routeStatus
		case strings.HasPrefix(r.URL.Path, "/assets/"):
			route = "asset"
		}
		result := httpsnoop.CaptureMetrics(next, w, r)
		metrics.Request(route, result.Code, result.Duration)
		if route != routeHealth {
			logger.WithFields(log.Fields{"route": route, "status": result.Code, "duration_ms": result.Duration.Milliseconds(), "bytes": result.Written}).Info("HTTP request")
		}
	})
}
