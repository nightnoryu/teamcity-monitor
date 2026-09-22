package monitor

import (
	"context"
	"sync/atomic"
	"time"
)

// Poller runs Aggregator.BuildSnapshot on a timer and caches the latest
// result so HTTP requests never call TeamCity synchronously.
type Poller struct {
	aggregator *Aggregator
	interval   time.Duration
	current    atomic.Pointer[Snapshot]
}

// NewPoller builds a Poller. Call Run to start polling.
func NewPoller(aggregator *Aggregator, interval time.Duration) *Poller {
	return &Poller{aggregator: aggregator, interval: interval}
}

// Interval is the configured time between collection cycles.
func (p *Poller) Interval() time.Duration { return p.interval }

// Snapshot returns the latest cached snapshot. ready is false until the
// first refresh (in Run) has completed.
func (p *Poller) Snapshot() (snapshot *Snapshot, ready bool) {
	snapshot = p.current.Load()
	return snapshot, snapshot != nil
}

// Run refreshes the snapshot immediately, then on every tick of interval,
// until ctx is canceled. It blocks; call it in a goroutine.
func (p *Poller) Run(ctx context.Context) {
	p.refresh(ctx)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.refresh(ctx)
		}
	}
}

func (p *Poller) refresh(ctx context.Context) {
	snapshot := p.aggregator.BuildSnapshot(ctx)
	if snapshot.CollectionHealth == CollectionHealthy {
		snapshot.LastSuccessfulAt = &snapshot.GeneratedAt
	} else if previous := p.current.Load(); previous != nil {
		snapshot.LastSuccessfulAt = previous.LastSuccessfulAt
	}
	p.current.Store(snapshot)
}
