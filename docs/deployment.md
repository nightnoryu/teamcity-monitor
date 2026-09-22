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
