package monitor

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"teamcity-monitor/internal/teamcity"
)

type countingClient struct{ builds atomic.Int64 }

func (c *countingClient) LatestBuild(context.Context, string) (teamcity.Build, error) {
	c.builds.Add(1)
	return teamcity.Build{Status: teamcity.StatusSuccess}, nil
}

func (*countingClient) LastParameterChangeAuthors(context.Context, string, []string) (map[string]string, error) {
	return nil, nil
}

func TestPoller_InitialPeriodicAndCancel(t *testing.T) {
	client := &countingClient{}
	poller := NewPoller(NewAggregator(sampleConfig(), client, newTestLogger(t)), 10*time.Millisecond)
	_, ready := poller.Snapshot()
	require.False(t, ready)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() { poller.Run(ctx); close(done) }()
	require.Eventually(t, func() bool { _, isReady := poller.Snapshot(); return isReady }, time.Second, time.Millisecond)
	require.Eventually(t, func() bool { return client.builds.Load() >= 10 }, time.Second, time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("poller did not stop")
	}
	snapshot, ready := poller.Snapshot()
	require.True(t, ready)
	require.Equal(t, CollectionHealthy, snapshot.CollectionHealth)
}
