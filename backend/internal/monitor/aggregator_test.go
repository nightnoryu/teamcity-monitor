package monitor

import (
	"context"
	"testing"

	"github.com/go-faster/errors"
	"github.com/nightnoryu/go-kita/jsonlog"
	"github.com/stretchr/testify/require"

	"teamcity-monitor/internal/teamcity"
)

type fakeFetcher struct {
	byBuildTypeID map[string]teamcity.Build
	errByID       map[string]error

	authorByParam map[string]string
	authorErr     error
}

func (f *fakeFetcher) LatestBuild(_ context.Context, buildTypeID string) (teamcity.Build, error) {
	if err, ok := f.errByID[buildTypeID]; ok {
		return teamcity.Build{}, err
	}
	return f.byBuildTypeID[buildTypeID], nil
}

func (f *fakeFetcher) LastParameterChangeAuthor(_ context.Context, _, paramName string) (string, error) {
	if f.authorErr != nil {
		return "", f.authorErr
	}
	if author, ok := f.authorByParam[paramName]; ok {
		return author, nil
	}
	return "", teamcity.ErrNoAuditRecord
}

func testLogger() *jsonlog.Config {
	return &jsonlog.Config{Level: jsonlog.ErrorLevel, AppName: "test"}
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

	logger := jsonlog.NewLogger(testLogger())
	aggregator := NewAggregator(sampleConfig(), fetcher, logger)

	snapshot := aggregator.BuildSnapshot(t.Context())

	dev := snapshot.Environments[0]
	require.Equal(t, "dev", dev.Name)

	ruGroup := dev.Groups[0]
	require.Equal(t, BuildSuccess, ruGroup.Builds[0].Status, "Alpha ru succeeded")
	require.Equal(t, BuildSuccess, ruGroup.Builds[1].Status, "Beta ru succeeded")

	euGroup := dev.Groups[1]
	require.Equal(t, BuildUnknown, euGroup.Builds[0].Status, "Alpha eu fetch failed")
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

	logger := jsonlog.NewLogger(testLogger())
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

	logger := jsonlog.NewLogger(testLogger())
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

	logger := jsonlog.NewLogger(testLogger())
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

	logger := jsonlog.NewLogger(testLogger())
	aggregator := NewAggregator(sampleConfig(), fetcher, logger)

	snapshot := aggregator.BuildSnapshot(t.Context())

	row := snapshot.Environments[0].Groups[0].Builds[0]
	require.Equal(t, BuildUnknown, row.Status)
	require.Empty(t, row.Error)
}
