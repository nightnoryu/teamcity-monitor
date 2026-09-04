// Package monitor aggregates TeamCity build statuses for all monitored
// builds into a cached, environment-grouped snapshot served over HTTP.
package monitor

import "time"

// BuildStatus is the dashboard-facing status of a monitored build.
type BuildStatus string

// Dashboard build statuses. Unknown covers both "never run" and "fetch
// failed" — the frontend does not need to distinguish those. Running covers
// both TeamCity's "queued" and "running" states.
const (
	BuildSuccess BuildStatus = "success"
	BuildFailure BuildStatus = "failure"
	BuildError   BuildStatus = "error"
	BuildRunning BuildStatus = "running"
	BuildUnknown BuildStatus = "unknown"
)

// Snapshot is the full JSON payload served at /api/status.
type Snapshot struct {
	GeneratedAt  time.Time           `json:"generatedAt"`
	Environments []EnvironmentStatus `json:"environments"`
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
	ProjectName string      `json:"projectName"`
	Status      BuildStatus `json:"status"`
	Branch      string      `json:"branch,omitempty"`
	BuildNumber string      `json:"buildNumber,omitempty"`
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
	// Error is set only when the latest fetch for this build failed for a
	// reason other than "never run".
	Error string `json:"error,omitempty"`
}
