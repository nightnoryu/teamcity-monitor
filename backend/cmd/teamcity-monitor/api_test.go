package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"teamcity-monitor/internal/monitor"
)

func TestStatusIncludesConfiguredPollIntervalBeforeFirstRefresh(t *testing.T) {
	poller := monitor.NewPoller(nil, time.Minute)
	recorder := httptest.NewRecorder()
	statusHandler(poller)(recorder, httptest.NewRequest(http.MethodGet, "/api/status", http.NoBody))
	var response statusResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.False(t, response.Ready)
	require.EqualValues(t, 60_000, response.PollIntervalMs)
}
