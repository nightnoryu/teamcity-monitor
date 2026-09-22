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
	Ready            bool                        `json:"ready"`
	PollIntervalMs   int64                       `json:"pollIntervalMs"`
	GeneratedAt      *time.Time                  `json:"generatedAt,omitempty"`
	LastSuccessfulAt *time.Time                  `json:"lastSuccessfulAt,omitempty"`
	CollectionHealth monitor.CollectionHealth    `json:"collectionHealth,omitempty"`
	PollDurationMs   int64                       `json:"pollDurationMs"`
	FailedBuilds     int                         `json:"failedBuilds"`
	Environments     []monitor.EnvironmentStatus `json:"environments,omitempty"`
}

func statusHandler(poller *monitor.Poller) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		snapshot, ready := poller.Snapshot()

		resp := statusResponse{Ready: ready, PollIntervalMs: poller.Interval().Milliseconds()}
		if ready {
			resp.GeneratedAt = &snapshot.GeneratedAt
			resp.LastSuccessfulAt = snapshot.LastSuccessfulAt
			resp.CollectionHealth = snapshot.CollectionHealth
			resp.PollDurationMs = snapshot.PollDurationMs
			resp.FailedBuilds = snapshot.FailedBuilds
			resp.Environments = snapshot.Environments
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
