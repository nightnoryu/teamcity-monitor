package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/nightnoryu/go-kita/jsonlog"
	"github.com/nightnoryu/go-kita/log"
	"github.com/stretchr/testify/require"

	"teamcity-monitor/internal/teamcity"
)

type fakeFetcher struct {
	byBuildTypeID map[string]teamcity.Build
	errByID       map[string]error

	authorByParam map[string]string
	authorErr     error
	auditWait     <-chan struct{}
	auditEntered  chan<- struct{}
}

func (f *fakeFetcher) LatestBuild(_ context.Context, buildTypeID string) (teamcity.Build, error) {
	if err, ok := f.errByID[buildTypeID]; ok {
		return teamcity.Build{}, err
	}
	return f.byBuildTypeID[buildTypeID], nil
}

func (f *fakeFetcher) LastParameterChangeAuthor(_ context.Context, _, paramName string) (string, error) {
	if f.auditEntered != nil {
		select {
		case f.auditEntered <- struct{}{}:
		default:
		}
	}
	if f.auditWait != nil {
		<-f.auditWait
	}
	if f.authorErr != nil {
		return "", f.authorErr
	}
	if author, ok := f.authorByParam[paramName]; ok {
		return author, nil
	}
	return "", teamcity.ErrNoAuditRecord
}

func TestAggregator_GeneratedAtPrecedesAuditEnrichment(t *testing.T) {
	wait := make(chan struct{})
	entered := make(chan struct{}, 1)
	fetcher := &fakeFetcher{auditWait: wait, auditEntered: entered}
	aggregator := NewAggregator(sampleConfig(), fetcher, newTestLogger(t))
	done := make(chan *Snapshot, 1)
	go func() { done <- aggregator.BuildSnapshot(t.Context()) }()
	<-entered
	releasedAt := time.Now().UTC()
	close(wait)
	snapshot := <-done
	require.True(t, snapshot.GeneratedAt.Before(releasedAt), "build freshness must not include optional audit delay")
}

func newTestLogger(t *testing.T) log.MainLogger {
	t.Helper()
	logger, err := jsonlog.NewLogger(&jsonlog.Config{Level: jsonlog.ErrorLevel, AppName: "test"})
	require.NoError(t, err)
	return logger
}

func TestAggregator_BuildSnapshot_PartialFailureDoesNotBlankSnapshot(t *testing.T) {
	fetcher := &fakeFetcher{
		byBuildTypeID: map[string]teamcity.Build{
			"Alpha_Testing_Dev_Ru":   {Status: teamcity.StatusSuccess, Number: "1"},
			"Beta_Testing_Dev_Ru":    {Status: teamcity.StatusSuccess, Number: "2"},
			"Beta_Testing_Dev_Eu":    {Status: teamcity.StatusSuccess, Number: "4"},
			"Beta_Testing_Dev_Build": {Status: teamcity.StatusFailure, Number: "3"},
		},
		errByID: map[string]error{
			"Alpha_Testing_Dev_Eu": errors.New("network error"),
		},
	}

	logger := newTestLogger(t)
	aggregator := NewAggregator(sampleConfig(), fetcher, logger)

	snapshot := aggregator.BuildSnapshot(t.Context())

	dev := snapshot.Environments[0]
	require.Equal(t, "dev", dev.Name)

	ruGroup := dev.Groups[0]
	require.Equal(t, BuildSuccess, ruGroup.Builds[0].Status, "Alpha ru succeeded")
	require.Equal(t, BuildSuccess, ruGroup.Builds[1].Status, "Beta ru succeeded")

	euGroup := dev.Groups[1]
	require.Equal(t, BuildUnavailable, euGroup.Builds[0].Status, "Alpha eu fetch failed")
	require.Equal(t, CollectionPartial, snapshot.CollectionHealth)
	require.NotEmpty(t, euGroup.Builds[0].Error)
	require.Equal(t, BuildSuccess, euGroup.Builds[1].Status, "Beta eu succeeded despite Alpha eu failing")

	buildGroup := dev.Groups[2]
	require.Equal(t, BuildFailure, buildGroup.Builds[0].Status)
}

func TestAggregator_BuildSnapshot_SuccessFraction(t *testing.T) {
	fetcher := &fakeFetcher{
		byBuildTypeID: map[string]teamcity.Build{
			"Alpha_Testing_Dev_Ru":   {Status: teamcity.StatusSuccess},
			"Alpha_Testing_Dev_Eu":   {Status: teamcity.StatusFailure},
			"Beta_Testing_Dev_Ru":    {Status: teamcity.StatusSuccess},
			"Beta_Testing_Dev_Eu":    {Status: teamcity.StatusSuccess},
			"Beta_Testing_Dev_Build": {Status: teamcity.StatusSuccess},
		},
	}

	logger := newTestLogger(t)
	aggregator := NewAggregator(sampleConfig(), fetcher, logger)

	snapshot := aggregator.BuildSnapshot(t.Context())

	dev := snapshot.Environments[0]
	require.Equal(t, 4, dev.SuccessCount)
	require.Equal(t, 5, dev.TotalCount)

	orange := snapshot.Environments[1]
	require.Equal(t, 0, orange.TotalCount)
}

func TestAggregator_BuildSnapshot_InProgressBuildShowsAsRunning(t *testing.T) {
	fetcher := &fakeFetcher{
		byBuildTypeID: map[string]teamcity.Build{
			"Alpha_Testing_Dev_Ru": {Status: teamcity.StatusSuccess, State: teamcity.StateRunning},
			"Alpha_Testing_Dev_Eu": {State: teamcity.StateQueued},
		},
	}

	logger := newTestLogger(t)
	aggregator := NewAggregator(sampleConfig(), fetcher, logger)

	snapshot := aggregator.BuildSnapshot(t.Context())

	dev := snapshot.Environments[0]
	require.Equal(t, BuildRunning, dev.Groups[0].Builds[0].Status, "running build, even with a provisional SUCCESS status")
	require.Equal(t, BuildRunning, dev.Groups[1].Builds[0].Status, "queued build")
	require.Zero(t, dev.SuccessCount, "in-progress builds don't count as success")
}

func TestAggregator_BuildSnapshot_BranchChangedByAppliesToAllRowsInGroup(t *testing.T) {
	fetcher := &fakeFetcher{
		byBuildTypeID: map[string]teamcity.Build{
			"Alpha_Testing_Dev_Ru": {Status: teamcity.StatusSuccess},
			"Alpha_Testing_Dev_Eu": {Status: teamcity.StatusSuccess},
		},
		authorByParam: map[string]string{
			"alpha_dev_branch": "a.kovalev",
			// vcs_root_branch_dev (Beta) intentionally has no entry: falls
			// back to ErrNoAuditRecord, leaving BranchChangedBy empty.
		},
	}

	logger := newTestLogger(t)
	aggregator := NewAggregator(sampleConfig(), fetcher, logger)

	snapshot := aggregator.BuildSnapshot(t.Context())

	dev := snapshot.Environments[0]
	require.Equal(t, "a.kovalev", dev.Groups[0].Builds[0].BranchChangedBy, "Alpha ru")
	require.Equal(t, "a.kovalev", dev.Groups[1].Builds[0].BranchChangedBy, "Alpha eu, same project+environment param")
	require.Empty(t, dev.Groups[2].Builds[0].BranchChangedBy, "Beta build: no matching audit record")
}

func TestAggregator_BuildSnapshot_NoBuildsIsUnknownWithoutError(t *testing.T) {
	fetcher := &fakeFetcher{
		errByID: map[string]error{
			"Alpha_Testing_Dev_Ru": teamcity.ErrNoBuilds,
		},
	}

	logger := newTestLogger(t)
	aggregator := NewAggregator(sampleConfig(), fetcher, logger)

	snapshot := aggregator.BuildSnapshot(t.Context())

	row := snapshot.Environments[0].Groups[0].Builds[0]
	require.Equal(t, BuildUnknown, row.Status)
	require.Empty(t, row.Error)
	require.Equal(t, CollectionHealthy, snapshot.CollectionHealth)
}

func TestAggregator_BuildSnapshot_TotalFailure(t *testing.T) {
	cfg := sampleConfig()
	fetcher := &fakeFetcher{errByID: map[string]error{}}
	// Every configured build fails through the same fake error.
	_, tasks, _ := planTasks(cfg)
	for _, task := range tasks {
		fetcher.errByID[task.buildTypeID] = errors.New("authorization failed")
	}
	snapshot := NewAggregator(cfg, fetcher, newTestLogger(t)).BuildSnapshot(t.Context())
	require.Equal(t, CollectionFailed, snapshot.CollectionHealth)
	for _, group := range snapshot.Environments[0].Groups {
		for _, row := range group.Builds {
			require.Equal(t, BuildUnavailable, row.Status)
			require.Contains(t, row.Error, "authorization failed")
		}
	}
}

func TestPoller_PreservesLastSuccessfulCollectionThroughFailure(t *testing.T) {
	cfg := sampleConfig()
	fetcher := &fakeFetcher{byBuildTypeID: map[string]teamcity.Build{}}
	poller := NewPoller(NewAggregator(cfg, fetcher, newTestLogger(t)), time.Second)
	poller.refresh(t.Context())
	first, ready := poller.Snapshot()
	require.True(t, ready)
	require.Equal(t, CollectionHealthy, first.CollectionHealth)
	require.NotNil(t, first.LastSuccessfulAt)

	fetcher.errByID = map[string]error{}
	_, tasks, _ := planTasks(cfg)
	for _, task := range tasks {
		fetcher.errByID[task.buildTypeID] = errors.New("connection refused")
	}
	poller.refresh(t.Context())
	second, _ := poller.Snapshot()
	require.Equal(t, CollectionFailed, second.CollectionHealth)
	require.Equal(t, first.LastSuccessfulAt, second.LastSuccessfulAt)
}
