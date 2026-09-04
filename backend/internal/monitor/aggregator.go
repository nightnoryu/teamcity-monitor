package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/go-faster/errors"
	"github.com/nightnoryu/go-kita/log"

	"teamcity-monitor/internal/monitorconfig"
	"teamcity-monitor/internal/teamcity"
)

const (
	// fetchConcurrency bounds how many TeamCity requests run at once.
	fetchConcurrency = 8
	// fetchTimeout bounds a single TeamCity request.
	fetchTimeout = 10 * time.Second
)

// teamcityClient is the narrow teamcity.Client surface the aggregator
// needs, so tests can fake it without a real HTTP server.
type teamcityClient interface {
	LatestBuild(ctx context.Context, buildTypeID string) (teamcity.Build, error)
	LastParameterChangeAuthor(ctx context.Context, projectID, paramName string) (string, error)
}

// Aggregator builds a full Snapshot by fanning out over every monitored
// build in the config. A single build's fetch failure does not affect the
// rest of the snapshot.
type Aggregator struct {
	cfg    *monitorconfig.Config
	client teamcityClient
	logger log.Logger
}

// NewAggregator builds an Aggregator.
func NewAggregator(cfg *monitorconfig.Config, client teamcityClient, logger log.Logger) *Aggregator {
	return &Aggregator{cfg: cfg, client: client, logger: logger}
}

// BuildSnapshot fetches the latest build of every monitored build, and the
// branch-parameter audit info for every (project, environment) pair,
// concurrently, and assembles the result into a Snapshot.
func (a *Aggregator) BuildSnapshot(ctx context.Context) *Snapshot {
	skeleton, tasks, auditTasks := planTasks(a.cfg)

	forEachConcurrent(tasks, fetchConcurrency, func(task fetchTask) {
		a.fetchInto(ctx, skeleton, task)
	})
	forEachConcurrent(auditTasks, fetchConcurrency, func(task auditTask) {
		a.fetchAuditInto(ctx, skeleton, task)
	})

	for i := range skeleton {
		fillCounts(&skeleton[i])
	}

	return &Snapshot{GeneratedAt: time.Now().UTC(), Environments: skeleton}
}

// forEachConcurrent runs fn over items with at most concurrency in flight,
// waiting for all to finish.
func forEachConcurrent[T any](items []T, concurrency int, fn func(T)) {
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, item := range items {
		wg.Add(1)
		sem <- struct{}{}

		go func(item T) {
			defer wg.Done()
			defer func() { <-sem }()

			fn(item)
		}(item)
	}

	wg.Wait()
}

func (a *Aggregator) fetchInto(ctx context.Context, skeleton []EnvironmentStatus, task fetchTask) {
	reqCtx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	build, err := a.client.LatestBuild(reqCtx, task.buildTypeID)

	row := &skeleton[task.envIndex].Groups[task.groupIndex].Builds[task.rowIndex]
	row.ProjectName = task.projectName

	switch {
	case errors.Is(err, teamcity.ErrNoBuilds):
		row.Status = BuildUnknown
	case err != nil:
		row.Status = BuildUnknown
		row.Error = err.Error()
		a.logger.Error(err, "fetch latest build failed for ", task.buildTypeID, " (", task.projectName, ")")
	default:
		fillRow(row, build)
	}
}

func (a *Aggregator) fetchAuditInto(ctx context.Context, skeleton []EnvironmentStatus, task auditTask) {
	reqCtx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	username, err := a.client.LastParameterChangeAuthor(reqCtx, task.projectID, task.paramName)

	switch {
	case errors.Is(err, teamcity.ErrNoAuditRecord):
		return
	case err != nil:
		a.logger.Error(err, "fetch last parameter change author failed for ", task.paramName, " (", task.projectID, ")")
		return
	}

	for _, ref := range task.targets {
		skeleton[ref.envIndex].Groups[ref.groupIndex].Builds[ref.rowIndex].BranchChangedBy = username
	}
}

func fillRow(row *ProjectBuildStatus, build teamcity.Build) {
	row.Status = mapStatus(build)
	row.Branch = build.Branch
	row.BuildNumber = build.Number
	row.TriggeredBy = build.TriggeredBy
	row.WebURL = build.WebURL

	if !build.StartedAt.IsZero() {
		row.StartedAt = new(build.StartedAt)
	}
	if !build.FinishedAt.IsZero() {
		row.FinishedAt = new(build.FinishedAt)
	}
}

func mapStatus(build teamcity.Build) BuildStatus {
	if build.State == teamcity.StateQueued || build.State == teamcity.StateRunning {
		return BuildRunning
	}

	switch build.Status {
	case teamcity.StatusSuccess:
		return BuildSuccess
	case teamcity.StatusFailure:
		return BuildFailure
	case teamcity.StatusError:
		return BuildError
	default:
		return BuildUnknown
	}
}

func fillCounts(env *EnvironmentStatus) {
	var success, total int

	for _, group := range env.Groups {
		for _, build := range group.Builds {
			total++
			if build.Status == BuildSuccess {
				success++
			}
		}
	}

	env.SuccessCount = success
	env.TotalCount = total
}
