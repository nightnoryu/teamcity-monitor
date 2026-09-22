package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nightnoryu/go-kita/jsonlog"
	"github.com/stretchr/testify/require"

	"teamcity-monitor/internal/monitor"
	"teamcity-monitor/internal/monitorconfig"
	"teamcity-monitor/internal/teamcity"
)

type statusClient struct{ err error }

func (c statusClient) LatestBuild(context.Context, string) (teamcity.Build, error) {
	return teamcity.Build{State: teamcity.StateFinished, Status: teamcity.StatusSuccess}, c.err
}
func (statusClient) LastParameterChangeAuthors(context.Context, string, []string) (map[string]string, error) {
	return nil, nil
}

func TestStatusCompletedAndFailedSnapshots(t *testing.T) {
	for _, test := range []struct {
		name   string
		err    error
		health monitor.CollectionHealth
		failed int
	}{
		{"completed", nil, monitor.CollectionHealthy, 0},
		{"failed", teamcity.ErrUnauthorized, monitor.CollectionFailed, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			logger, err := jsonlog.NewLogger(&jsonlog.Config{Level: jsonlog.ErrorLevel, AppName: "test"})
			require.NoError(t, err)
			cfg := &monitorconfig.Config{Environments: []monitorconfig.Environment{{Name: "dev"}}, Projects: []monitorconfig.Project{{Name: "Alpha", ID: "A", EnvironmentBranchParam: "branch_%s", MonitoredBuilds: []monitorconfig.MonitoredBuild{{Environment: "dev", Name: "ru", ID: "A_Ru"}}}}}
			poller := monitor.NewPoller(monitor.NewAggregator(cfg, statusClient{err: test.err}, logger), time.Hour)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			go poller.Run(ctx)
			require.Eventually(t, func() bool { _, ready := poller.Snapshot(); return ready }, time.Second, time.Millisecond)
			recorder := httptest.NewRecorder()
			statusHandler(poller)(recorder, httptest.NewRequest(http.MethodGet, "/api/status", http.NoBody))
			var response statusResponse
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
			require.True(t, response.Ready)
			require.Equal(t, test.health, response.CollectionHealth)
			require.Equal(t, test.failed, response.FailedBuilds)
			require.NotNil(t, response.GeneratedAt)
		})
	}
}

func TestStatusIncludesConfiguredPollIntervalBeforeFirstRefresh(t *testing.T) {
	poller := monitor.NewPoller(nil, time.Minute)
	recorder := httptest.NewRecorder()
	statusHandler(poller)(recorder, httptest.NewRequest(http.MethodGet, "/api/status", http.NoBody))
	var response statusResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.False(t, response.Ready)
	require.EqualValues(t, 60_000, response.PollIntervalMs)
}
