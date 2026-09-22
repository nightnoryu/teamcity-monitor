// Package monitor aggregates TeamCity build statuses for all monitored
// builds into a cached, environment-grouped snapshot served over HTTP.
package monitor

import "time"

// BuildStatus is the dashboard-facing status of a monitored build.
type BuildStatus string

// Unknown means no build has run. Unavailable means collection failed.
const (
	BuildSuccess     BuildStatus = "success"
	BuildFailure     BuildStatus = "failure"
	BuildError       BuildStatus = "error"
	BuildRunning     BuildStatus = "running"
	BuildQueued      BuildStatus = "queued"
	BuildUnknown     BuildStatus = "unknown"
	BuildUnavailable BuildStatus = "unavailable"
)

type CollectionHealth string

const (
	CollectionHealthy CollectionHealth = "healthy"
	CollectionPartial CollectionHealth = "partial"
	CollectionFailed  CollectionHealth = "failed"
)

// Snapshot is the full JSON payload served at /api/status.
type Snapshot struct {
	GeneratedAt      time.Time           `json:"generatedAt"`
	LastSuccessfulAt *time.Time          `json:"lastSuccessfulAt,omitempty"`
	CollectionHealth CollectionHealth    `json:"collectionHealth"`
	PollDurationMs   int64               `json:"pollDurationMs"`
	FailedBuilds     int                 `json:"failedBuilds"`
	Environments     []EnvironmentStatus `json:"environments"`
}

// EnvironmentStatus is one environment's aggregated build statuses, in the
// same order as config.toml's [[environments]].
type EnvironmentStatus struct {
	Name         string        `json:"name"`
	Emoji        string        `json:"emoji"`
	SuccessCount int           `json:"successCount"`
	TotalCount   int           `json:"totalCount"`
	Groups       []RegionGroup `json:"groups"`
}

// RegionGroup groups monitored builds sharing the same monitored_builds.name
// within an environment, in first-seen config order.
type RegionGroup struct {
	Name   string               `json:"name"`
	Builds []ProjectBuildStatus `json:"builds"`
}

// ProjectBuildStatus is a single monitored build's latest known status.
type ProjectBuildStatus struct {
	ProjectID   string      `json:"projectId"`
	BuildID     string      `json:"buildId"`
	BuildName   string      `json:"buildName"`
	ProjectName string      `json:"projectName"`
	Status      BuildStatus `json:"status"`
	Branch      string      `json:"branch,omitempty"`
	BuildNumber string      `json:"buildNumber,omitempty"`
	StatusText  string      `json:"statusText,omitempty"`
	StartedAt   *time.Time  `json:"startedAt,omitempty"`
	FinishedAt  *time.Time  `json:"finishedAt,omitempty"`
	TriggeredBy string      `json:"triggeredBy,omitempty"`
	WebURL      string      `json:"webUrl,omitempty"`
	// BranchChangedBy is the username of whoever most recently changed this
	// project's branch parameter for this environment, per TeamCity's audit
	// log. Best-effort: TeamCity's audit API doesn't expose a parameter's
	// old/new value, only that "Value of the parameter X changed" — so this
	// is empty if that exact event isn't found in the recent audit history.
	BranchChangedBy string `json:"branchChangedBy,omitempty"`
	// AttributionStatus is pending, found, not_found, or error. Audit history
	// is a bounded best-effort scan, not evidence of who deployed a build.
	AttributionStatus string `json:"attributionStatus"`
	// Error is set when the latest fetch for this build failed.
	Error string `json:"error,omitempty"`
}
