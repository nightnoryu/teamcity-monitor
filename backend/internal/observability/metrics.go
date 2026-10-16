package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"teamcity-monitor/internal/monitor"
)

type Metrics struct {
	registry         *prometheus.Registry
	requests         *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	polls            *prometheus.CounterVec
	pollDuration     prometheus.Histogram
	lastSuccessfulAt prometheus.Gauge
}

func New() *Metrics {
	m := &Metrics{
		registry:         prometheus.NewRegistry(),
		requests:         prometheus.NewCounterVec(prometheus.CounterOpts{Name: "teamcity_monitor_http_requests_total", Help: "HTTP requests by bounded route and status."}, []string{"route", "status"}),
		requestDuration:  prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "teamcity_monitor_http_request_duration_seconds", Help: "HTTP request duration by bounded route."}, []string{"route"}),
		polls:            prometheus.NewCounterVec(prometheus.CounterOpts{Name: "teamcity_monitor_polls_total", Help: "Published collection cycles by health."}, []string{"health"}),
		pollDuration:     prometheus.NewHistogram(prometheus.HistogramOpts{Name: "teamcity_monitor_poll_duration_seconds", Help: "Build collection cycle duration.", Buckets: []float64{.1, .5, 1, 2, 5, 10, 15, 20}}),
		lastSuccessfulAt: prometheus.NewGauge(prometheus.GaugeOpts{Name: "teamcity_monitor_last_successful_collection_timestamp_seconds", Help: "Unix timestamp of the last fully healthy collection; zero before one completes."}),
	}
	m.registry.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}), m.requests, m.requestDuration, m.polls, m.pollDuration, m.lastSuccessfulAt)
	return m
}

func (m *Metrics) Request(route string, status int, duration time.Duration) {
	m.requests.WithLabelValues(route, strconv.Itoa(status)).Inc()
	m.requestDuration.WithLabelValues(route).Observe(duration.Seconds())
}

func (m *Metrics) Poll(snapshot *monitor.Snapshot) {
	m.polls.WithLabelValues(string(snapshot.CollectionHealth)).Inc()
	m.pollDuration.Observe(float64(snapshot.PollDurationMs) / 1000)
	if snapshot.LastSuccessfulAt != nil {
		m.lastSuccessfulAt.Set(float64(snapshot.LastSuccessfulAt.Unix()))
	}
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
