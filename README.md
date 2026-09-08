<p align="center"><img src="https://github.com/user-attachments/assets/3448061c-3640-475b-956a-aa75e943d39b" width="800" title="TeamCity Monitor Screenshot"></p>

<p align="center">
  <a href="https://github.com/nightnoryu/teamcity-monitor/releases"><img src="https://img.shields.io/github/release/nightnoryu/teamcity-monitor.svg?cache-control=no-cache"></a>
  <a href="https://github.com/nightnoryu/teamcity-monitor/blob/main/LICENSE"><img src="https://img.shields.io/github/license/nightnoryu/teamcity-monitor?cache-control=no-cache"></a>
  <a href="https://github.com/nightnoryu/teamcity-monitor/actions/workflows/ci.yml"><img src="https://github.com/nightnoryu/teamcity-monitor/actions/workflows/ci.yml/badge.svg?cache-control=no-cache"></a>
</p>

Lightweight self-hosted dashboard for monitoring TeamCity environments and deployments.

> [!NOTE]
> This is not a replacement for TeamCity UI or CLI. This project provides a high-level environment view for teams that need to see the state of multiple environments at a glance.

## ✅ Features

- Environment and project-oriented dashboard
- Multiple builds per environment
- Real-time build status polling
- Lightweight Go backend
- Packaged in a single docker container

## 🚀 Quick Start

1. Copy `config.example.toml` and fill it in according to the template.
2. Run with docker-compose with the following config

    ```yaml
    services:
      teamcity-monitor:
        image: ghcr.io/nightnoryu/teamcity-monitor:latest
        container_name: teamcity-monitor
        restart: unless-stopped
        environment:
          TEAMCITY_MONITOR_CONFIG_PATH: /app/config.toml   # Config location
          TEAMCITY_MONITOR_POLL_INTERVAL: 20s              # Polling interval
          TEAMCITY_MONITOR_INSECURE_SKIP_TLS_VERIFY: false # Set to true if having issues with self-signed certs
        volumes:
          - "./config.toml:/app/config.toml" # Map your config
        ports:
          - "8080:8080"
    ```

## ⚙️ Configuration

There are two layers of configuration: runtime settings passed as environment
variables, and the monitoring domain model described in `config.toml`.

### Environment variables

All variables are prefixed with `TEAMCITY_MONITOR_`.

| Variable                    | Default            | Description                                                       |
|-----------------------------|--------------------|-----------------------------------------------------------------|
| `SERVE_REST_ADDRESS`        | `:8080`            | Address the HTTP server listens on.                              |
| `CONFIG_PATH`               | `/app/config.toml` | Path to the `config.toml` file.                                  |
| `POLL_INTERVAL`             | `20s`              | How often TeamCity is polled for fresh build statuses.           |
| `INSECURE_SKIP_TLS_VERIFY`  | `false`            | Skip TLS verification for TeamCity requests (self-signed certs). |

### `config.toml`

```toml
teamcity_url = "https://teamcity.your-org.lan"
access_token = "abc..."

[[projects]]
name = "Alpha"                               # Display name
id = "Alpha_Testing"                         # TeamCity project ID
environment_branch_param = "alpha_%s_branch" # See below
monitored_builds = [
    { environment = "dev", name = "ru", id = "Alpha_Testing_Dev_Ru" },
    { environment = "dev", name = "eu", id = "Alpha_Testing_Dev_Eu" },
]

[[environments]]
name = "dev"
emoji = "🥭"

[[environments]]
name = "stage"
emoji = "☢️"
```

- `access_token` is a TeamCity access token; it needs read access to the
  monitored projects and their audit log.
- `[[projects]]` lists the TeamCity projects to monitor. Each `monitored_builds`
  entry pins one build configuration (`id`) to an environment (`environment`)
  and a grouping/display key (`name`, e.g. a region - free-form, not an enum).
- `environment_branch_param` is a template with exactly one `%s`, substituted
  with the environment name to produce the name of a TeamCity project
  parameter. Its edit history in the audit log is used to show who last changed
  the deployed branch for that environment.
- `[[environments]]` declares the deployment tiers shown on the dashboard, in
  the order columns appear.

The config is validated on load: every `monitored_builds.environment` must
reference a declared environment, project and build IDs must be unique, and
`environment_branch_param` must contain exactly one `%s`.

## 🏗️ Architecture

The application is a single Go binary with an embedded React frontend.

```
                 ┌───────────────────────────── teamcity-monitor ─────────────────────────────┐
  TeamCity  ◄────┤  Poller ──► Aggregator ──► in-memory Snapshot ──► GET /api/status           │
  REST API       │   (every POLL_INTERVAL)         cache            embedded SPA (/, static)   │
                 └───────────────────────────────────────▲───────────────────────────────────┘
                                                         │ polls /api/status
                                                    React + Vite SPA
```

## ⚒️ Local Development

### Prerequisites

- [mise](https://mise.jdx.dev)
- Docker with docker-compose-plugin

### First launch

```shell
git clone https://github.com/nightnoryu/teamcity-monitor
cd teamcity-monitor

# Set up local env domain
echo "127.0.0.1 teamcity-monitor.lan" | sudo tee -a /etc/hosts

# Copy the config template
cp config.example.toml config.toml

# Builds backend binary and spins up docker containers
mise run
```

Web picks up changes automatically via `vite`. Backend needs to be rebuilt and restarted in order to pick up changes,
use `mise run dev:reload` shorthand for this.

## 📜 License

Distributed under the MIT License. See [License](/LICENSE) for more information.
