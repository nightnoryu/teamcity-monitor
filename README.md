# TeamCity Monitor

[![GitHub License](https://img.shields.io/github/license/nightnoryu/teamcity-monitor)](https://github.com/nightnoryu/teamcity-monitor/blob/main/LICENSE)
[![Build Status](https://github.com/nightnoryu/teamcity-monitor/actions/workflows/ci.yml/badge.svg)](https://github.com/nightnoryu/teamcity-monitor/actions/workflows/ci.yml)

Real-time TeamCity environments monitoring.

## Installation

### Config

Copy `config.example.toml` and fill it in according to the template.

### Running

Run with docker-compose:

```yaml
services:
  teamcity-monitor:
    image: ghcr.io/nightnoryu/teamcity-monitor:latest
    container_name: teamcity-monitor
    restart: unless-stopped
    environment:
      TEAMCITY_MONITOR_CONFIG_PATH: /app/config.toml  # Config location
      TEAMCITY_MONITOR_POLL_INTERVAL: 20s             # Polling interval
      TEAMCITY_MONITOR_INSECURE_SKIP_TLS_VERIFY: false        # Set to true if having issues with self-signed certs
    volumes:
      - "./config.toml:/app/config.toml" # Map your config
    ports:
      - "8080:8080"
```

## Local Development

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

## License

Distributed under the MIT License. See [License](/LICENSE) for more information.
