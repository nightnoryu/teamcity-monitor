package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfigValidatePollInterval(t *testing.T) {
	for _, tc := range []struct {
		name     string
		interval time.Duration
		valid    bool
	}{
		{"zero", 0, false}, {"negative", -time.Second, false}, {"positive", time.Second, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := (&config{PollInterval: tc.interval}).validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, "POLL_INTERVAL")
			}
		})
	}
}
