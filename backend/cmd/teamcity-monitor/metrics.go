package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/nightnoryu/go-kita/log"

	"teamcity-monitor/internal/observability"
)

func startMetrics(ctx context.Context, address string, metrics *observability.Metrics, logger log.Logger) (func(), error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", metrics.Handler())
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			logger.Error(serveErr, "metrics server")
		}
	}()
	return func() {
		shutdown, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if shutdownErr := server.Shutdown(shutdown); shutdownErr != nil {
			logger.Error(shutdownErr, "shutdown metrics server")
			_ = server.Close()
		}
	}, nil
}
