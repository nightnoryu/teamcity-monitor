# Architecture

TeamCity Monitor is a single Go HTTP service with a React dashboard embedded
into its executable. It polls the TeamCity REST API in the background, keeps
the newest complete snapshot in memory, and serves that snapshot to the browser.
No browser request calls TeamCity directly.

```
                 TeamCity REST API
                         ▲
                         │ latest build + audit history
                         │
              ┌──────────┴──────────┐
              │   Go service         │
              │                      │
              │  Poller → Aggregator │
              │             │        │
              │   in-memory Snapshot │
              │        │             │
              │  /api/status  SPA    │
              └───────┬────────┬─────┘
                      │        │
                 JSON │        │ embedded static files
                      ▼        ▼
                    React dashboard
```

The service is stateless. Its configuration and TeamCity token are read at
startup; all collected data is held in memory and is rebuilt after a restart.

## Components

```
backend/cmd/teamcity-monitor/  process wiring, HTTP server, health and API handlers
backend/internal/monitorconfig/ config.toml decoding and validation
backend/internal/teamcity/     TeamCity REST client
backend/internal/monitor/      collection planning, aggregation, and snapshot cache
backend/internal/webui/        embedded Vite production bundle
web/src/                       React dashboard, API client, and presentation components
```

`monitorconfig` validates the monitoring model before the server starts. The
`teamcity` package is a small client for the exact REST resources the monitor
needs. `monitor` depends on a narrow client interface, which keeps aggregation
tests independent of a real TeamCity instance.

## Collection flow

At startup, the poller immediately builds a snapshot; it then repeats at
`TEAMCITY_MONITOR_POLL_INTERVAL` (20 seconds by default).

1. The aggregator builds an ordered dashboard skeleton from `config.toml`.
   Environment order follows `[[environments]]`; groups follow first appearance
   in the monitored-build configuration.
2. It requests the latest attempt for every monitored build configuration, with
   at most eight requests in flight. Each request has a 10-second deadline and
   the whole build collection has a 15-second budget.
3. TeamCity results are mapped into dashboard states: queued, running, success,
   failure, error, unknown (no builds), or unavailable (fetch failure).
4. The complete result is atomically published as the current snapshot. HTTP
   readers always see either the previous snapshot or the new one, never a
   partially populated result.

A failed build fetch does not discard the rest of the cycle. The snapshot is
`partial` when one or more monitored builds are unavailable and `failed` when
all are unavailable. The frontend can therefore continue showing useful data
during a partial TeamCity outage.

## Branch and audit attribution

The displayed branch comes from the latest build. The client first uses the
build's branch, then its checked-out revision, then snapshot dependencies; this
covers re-runs and deploy-only configurations without their own VCS checkout.

For every project/environment pair, the monitor also looks for the most recent
audit event saying that the configured project parameter changed. Audit lookups
are asynchronous, cached for one minute, and scan the 100 newest relevant
events. This attribution is best effort: it identifies who last changed that
parameter in the scanned history, not necessarily who triggered or deployed the
shown build.

## HTTP interface and health

| Path | Purpose |
|---|---|
| `GET /` | Embedded dashboard and static assets; unknown frontend paths fall back to the SPA. |
| `GET /api/status` | Current snapshot, polling metadata, collection health, and per-build errors. |
| `GET /livez` | Process liveness; `200` while the HTTP server is running. |
| `GET /healthz` | Collection readiness; `503` before the first poll or when every build fetch in the newest cycle failed. |

The frontend polls `/api/status` at the configured interval. It distinguishes a
disconnected browser from a stale or partially failed collection, and continues
to show a cached snapshot when one is available.

## Lifecycle and limits

The process receives `SIGINT` and `SIGTERM`, stops the poller through its
context, and gives the HTTP server up to 10 seconds to drain. TeamCity requests
use a 15-second HTTP client timeout; response bodies are capped at 64 KiB.

The access token is used only in outbound Bearer-authenticated TeamCity REST
requests. It is never sent to the browser. The dashboard itself has no access
control, so deployment network boundaries are part of the security model; see
the [deployment guide](deployment.md#access-model).
