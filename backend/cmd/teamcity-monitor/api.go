package main

import (
	"encoding/json"
	"net/http"
	"time"

	"teamcity-monitor/internal/monitor"
)

// statusResponse is the /api/status payload. Ready is false while the
// poller hasn't completed its first refresh yet; Environments/GeneratedAt
// are omitted in that case rather than sent as empty/zero values.
type statusResponse struct {
	Ready        bool                        `json:"ready"`
	GeneratedAt  *time.Time                  `json:"generatedAt,omitempty"`
	Environments []monitor.EnvironmentStatus `json:"environments,omitempty"`
}

func statusHandler(poller *monitor.Poller) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		snapshot, ready := poller.Snapshot()

		resp := statusResponse{Ready: ready}
		if ready {
			resp.GeneratedAt = &snapshot.GeneratedAt
			resp.Environments = snapshot.Environments
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
