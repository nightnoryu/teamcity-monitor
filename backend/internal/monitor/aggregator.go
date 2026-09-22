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
	LastParameterChangeAuthors(ctx context.Context, projectID string, names []string) (map[string]string, error)
}

type auditResult struct {
	authors   map[string]string
	err       error
	fetchedAt time.Time
}

// Aggregator builds a full Snapshot by fanning out over every monitored
// build in the config. A single build's fetch failure does not affect the
// rest of the snapshot.
type Aggregator struct {
	cfg           *monitorconfig.Config
	client        teamcityClient
	logger        log.Logger
	auditMu       sync.Mutex
	auditCache    map[string]auditResult
	auditInFlight map[string]bool
	auditSlots    chan struct{}
}

// NewAggregator builds an Aggregator.
func NewAggregator(cfg *monitorconfig.Config, client teamcityClient, logger log.Logger) *Aggregator {
	return &Aggregator{cfg: cfg, client: client, logger: logger, auditCache: make(map[string]auditResult), auditInFlight: make(map[string]bool), auditSlots: make(chan struct{}, fetchConcurrency)}
}

// BuildSnapshot fetches the latest build of every monitored build within a
// cycle budget, applies cached audit attribution, and starts any due audit
// refreshes independently of publication.
func (a *Aggregator) BuildSnapshot(ctx context.Context) *Snapshot {
	started := time.Now()
	skeleton, tasks, auditTasks := planTasks(a.cfg)
	for _, task := range tasks {
		row := &skeleton[task.envIndex].Groups[task.groupIndex].Builds[task.rowIndex]
		row.Status = BuildUnavailable
		row.Error = "collection ended before this build could be fetched"
	}
	cycleCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	forEachConcurrent(cycleCtx, tasks, fetchConcurrency, func(task fetchTask) {
		a.fetchInto(cycleCtx, skeleton, task)
	})
	buildCollectedAt := time.Now().UTC()
	a.applyCachedAudit(skeleton, auditTasks)
	a.refreshAudit(ctx, auditTasks)

	for i := range skeleton {
		fillCounts(&skeleton[i])
	}
	var total, failed int
	for _, env := range skeleton {
		for _, group := range env.Groups {
			for _, row := range group.Builds {
				total++
				if row.Status == BuildUnavailable {
					failed++
				}
			}
		}
	}
	health := CollectionHealthy
	if failed > 0 {
		health = CollectionPartial
	}
	if total > 0 && failed == total {
		health = CollectionFailed
	}
	return &Snapshot{GeneratedAt: buildCollectedAt, CollectionHealth: health, PollDurationMs: time.Since(started).Milliseconds(), FailedBuilds: failed, Environments: skeleton}
}

// forEachConcurrent runs fn over items with at most concurrency in flight,
// waiting for all to finish.
func forEachConcurrent[T any](ctx context.Context, items []T, concurrency int, fn func(T)) {
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

loop:
	for _, item := range items {
		if ctx.Err() != nil {
			break
		}
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			break loop
		}
		if ctx.Err() != nil {
			<-sem
			break
		}
		wg.Add(1)

		go func(item T) {
			defer wg.Done()
			defer func() { <-sem }()

			fn(item)
		}(item)
	}

	wg.Wait()
}

// Audit history is optional and cached across cycles. A refresh runs at most
// once per project every minute, with one page request for all environments.
func (a *Aggregator) refreshAudit(ctx context.Context, tasks []auditTask) {
	byProject := make(map[string][]auditTask)
	for _, task := range tasks {
		byProject[task.projectID] = append(byProject[task.projectID], task)
	}
	for projectID, projectTasks := range byProject {
		a.auditMu.Lock()
		cached := a.auditCache[projectID]
		if a.auditInFlight[projectID] || time.Since(cached.fetchedAt) < time.Minute {
			a.auditMu.Unlock()
			continue
		}
		a.auditInFlight[projectID] = true
		a.auditMu.Unlock()
		go func(projectID string, projectTasks []auditTask) {
			select {
			case a.auditSlots <- struct{}{}:
				defer func() { <-a.auditSlots }()
			case <-ctx.Done():
				a.auditMu.Lock()
				delete(a.auditInFlight, projectID)
				a.auditMu.Unlock()
				return
			}
			requestCtx, cancel := context.WithTimeout(ctx, fetchTimeout)
			defer cancel()
			names := make([]string, len(projectTasks))
			for i, task := range projectTasks {
				names[i] = task.paramName
			}
			authors, err := a.client.LastParameterChangeAuthors(requestCtx, projectID, names)
			if err != nil {
				a.logger.Error(err, "fetch parameter change authors failed for ", projectID)
			}
			a.auditMu.Lock()
			a.auditCache[projectID] = auditResult{authors: authors, err: err, fetchedAt: time.Now()}
			delete(a.auditInFlight, projectID)
			a.auditMu.Unlock()
		}(projectID, projectTasks)
	}
}

func (a *Aggregator) applyCachedAudit(skeleton []EnvironmentStatus, tasks []auditTask) {
	a.auditMu.Lock()
	defer a.auditMu.Unlock()
	for _, task := range tasks {
		cached, ok := a.auditCache[task.projectID]
		if !ok {
			continue
		}
		for _, ref := range task.targets {
			row := &skeleton[ref.envIndex].Groups[ref.groupIndex].Builds[ref.rowIndex]
			if cached.err != nil {
				row.AttributionStatus = "error"
			} else if author, found := cached.authors[task.paramName]; found {
				row.AttributionStatus = "found"
				row.BranchChangedBy = author
			} else {
				row.AttributionStatus = "not_found"
			}
		}
	}
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
		row.Error = ""
	case err != nil:
		row.Status = BuildUnavailable
		row.Error = err.Error()
		a.logger.Error(err, "fetch latest build failed for ", task.buildTypeID, " (", task.projectName, ")")
	default:
		row.Error = ""
		fillRow(row, build)
	}
}

func fillRow(row *ProjectBuildStatus, build teamcity.Build) {
	row.Status = mapStatus(build)
	row.Branch = build.Branch
	row.BuildNumber = build.Number
	row.StatusText = build.StatusText
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
	if build.Canceled || build.FailedToStart {
		return BuildError
	}
	if build.State == teamcity.StateQueued {
		return BuildQueued
	}
	if build.State == teamcity.StateRunning {
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
