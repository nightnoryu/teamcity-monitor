# Deployment guide

TeamCity Monitor ships as a single container image containing the Go service and
compiled dashboard. It needs outbound HTTPS access to TeamCity, a readable
configuration file, and an internal audience. It does not need a database or
persistent volume.

## Before you start

- A TeamCity URL reachable from the container.
- A TeamCity access token with read access to the monitored projects. Audit-log
  read access is also needed to show branch-parameter editor attribution.
- A `config.toml` describing environments, projects, and build configurations.
  Start with [config.example.toml](../config.example.toml).
- A private network or an authenticated reverse proxy. The application has no
  built-in authentication or authorization.

## Monitor configuration

`config.toml` is loaded at startup. Unknown fields and invalid configuration
are rejected, so a malformed file prevents the service from starting.

```toml
teamcity_url = "https://teamcity.your-org.lan"
access_token = "abc..."

[[projects]]
name = "Alpha"
id = "Alpha_Testing"
environment_branch_param = "alpha_%s_branch"
monitored_builds = [
    { environment = "dev", name = "ru", id = "Alpha_Testing_Dev_Ru" },
    { environment = "dev", name = "eu", id = "Alpha_Testing_Dev_Eu" },
]

[[environments]]
name = "dev"
emoji = "🥭"
```

- `teamcity_url` must be an absolute HTTP(S) URL with a host; context paths are
  supported, but user info, query strings, and fragments are not.
- `access_token` is sent as a Bearer token and must be kept out of source
  control. The repository ignores `config.toml`.
- `[[environments]]` defines the dashboard columns and their order.
- Each `[[projects]]` entry names a TeamCity project. `monitored_builds` maps
  its build configuration IDs to environments and to a free-form display group
  such as a region.
- `environment_branch_param` has exactly one `%s`. It is filled with the
  environment name to find the TeamCity project parameter used for best-effort
  audit attribution.

Project IDs, build IDs, and environment names must be unique; every monitored
build must reference a declared environment.

## Runtime settings

All runtime variables use the `TEAMCITY_MONITOR_` prefix.

| Variable | Default | Purpose |
|---|---|---|
| `SERVE_REST_ADDRESS` | `:8080` | HTTP listen address. |
| `METRICS_ADDRESS` | `127.0.0.1:9090` | Private Prometheus listener; set to an empty value to disable. |
| `LOG_LEVEL` | `info` | JSON log level: `debug`, `info`, `warn`, `error`, or `fatal`. |
| `CONFIG_PATH` | `/app/config.toml` | Path to the monitor configuration. |
| `POLL_INTERVAL` | `20s` | Time between collection cycles; must be greater than zero. |
| `INSECURE_SKIP_TLS_VERIFY` | `false` | Disables TeamCity TLS certificate verification. Use only for a trusted local self-signed setup. |

## Docker

Mount the configuration read-only and bind the service to loopback unless a
trusted internal load balancer or reverse proxy will publish it.

```shell
docker run -d --name teamcity-monitor \
  --restart unless-stopped \
  -p 127.0.0.1:8080:8080 \
  -v "$(pwd)/config.toml:/app/config.toml:ro" \
  ghcr.io/nightnoryu/teamcity-monitor:latest
```

Use a released image tag instead of `:latest` when reproducible deployments
matter.

## Docker Compose

```yaml
services:
  teamcity-monitor:
    image: ghcr.io/nightnoryu/teamcity-monitor:latest
    container_name: teamcity-monitor
    restart: unless-stopped
    environment:
      TEAMCITY_MONITOR_CONFIG_PATH: /app/config.toml
      TEAMCITY_MONITOR_POLL_INTERVAL: 20s
      TEAMCITY_MONITOR_METRICS_ADDRESS: ":9090"
      TEAMCITY_MONITOR_LOG_LEVEL: info
      TEAMCITY_MONITOR_INSECURE_SKIP_TLS_VERIFY: "false"
    volumes:
      - "./config.toml:/app/config.toml:ro"
    ports:
      - "127.0.0.1:8080:8080"
    healthcheck:
      test: ["CMD", "curl", "-fsS", "http://localhost:8080/healthz"]
      interval: 20s
      timeout: 5s
      retries: 3
```

Run it with `docker compose up -d`. The image runs as a non-root user. Ensure
the host configuration file is readable by that user inside the container.

## Health checks and operations

| Endpoint | Meaning |
|---|---|
| `/livez` | Liveness: the HTTP process is running. |
| `/healthz` | Readiness: at least the initial collection completed and the latest cycle was not a total failure. |
| `/api/status` | Detailed snapshot metadata, collection health, failed-build count, and fetch errors. |

Metrics are served at `/metrics` on a **separate listener**. Its default bind
address is loopback inside the container. For private-network scraping, set
`TEAMCITY_MONITOR_METRICS_ADDRESS=:9090` and scrape the container on port 9090;
do not publish that port or route it through the public ingress. An explicitly
empty value disables the listener. The endpoint exposes Go and process metrics,
`teamcity_monitor_http_requests_total` and
`teamcity_monitor_http_request_duration_seconds` (fixed route labels), plus
`teamcity_monitor_polls_total`, `teamcity_monitor_poll_duration_seconds`, and
`teamcity_monitor_last_successful_collection_timestamp_seconds`. The last
timestamp is zero until a fully healthy poll completes. Poll metrics describe
published snapshots; `partial` and `failed` collections still count as attempts.

Logs are JSON. Each poll emits a summary with health, duration, failed builds,
and total builds. Fetch and audit errors include the affected identifiers.
HTTP request summaries include a fixed route name, status, duration, and bytes;
routine health checks are omitted. Request summaries omit URLs, query strings,
cookies, and bodies. TeamCity transport failures are logged as safe error
categories without request URLs or query strings. Access tokens are never logged.

Use `/livez` for a liveness probe and `/healthz` for readiness or load-balancer
health checks. A partial TeamCity outage still leaves `/healthz` ready, because
the dashboard can serve statuses for the builds that were collected.

The service has no persistent state. To upgrade, replace the container image
and restart it against the same `config.toml`; it performs a fresh collection
on startup. Retain the configuration and access token in your normal secret
management and backup process.

## Access model

The dashboard and `/api/status` expose deployment metadata, build URLs, branch
names, and trigger identities. Do not expose the application port directly to
the public internet. Keep it bound to loopback, a private network, or behind a
TLS reverse proxy that enforces your organization's authentication policy.

The monitor validates TeamCity certificates by default. Keep verification on;
if a local TeamCity uses a self-signed certificate, install the issuing CA in
the image or use `TEAMCITY_MONITOR_INSECURE_SKIP_TLS_VERIFY=true` only when that
risk is understood.
