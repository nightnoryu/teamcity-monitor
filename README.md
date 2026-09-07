<p align="center"><img src="https://github.com/user-attachments/assets/3448061c-3640-475b-956a-aa75e943d39b" width="800" title="TeamCity Monitor Screenshot"></p>

<p align="center">
  <a href="https://github.com/nightnoryu/teamcity-monitor/releases"><img src="https://img.shields.io/github/release/nightnoryu/teamcity-monitor.svg?cache-control=no-cache"></a>
  <a href="https://github.com/nightnoryu/teamcity-monitor/blob/main/LICENSE"><img src="https://img.shields.io/github/license/nightnoryu/teamcity-monitor?cache-control=no-cache"></a>
  <a href="https://github.com/nightnoryu/teamcity-monitor/actions/workflows/ci.yml"><img src="https://github.com/nightnoryu/teamcity-monitor/actions/workflows/ci.yml/badge.svg?cache-control=no-cache"></a>
</p>

A lightweight environment dashboard for TeamCity.

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

TODO

## 🏗️ Architecture

TODO

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
