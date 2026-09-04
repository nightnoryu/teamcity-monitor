# TeamCity Monitor

[![GitHub License](https://img.shields.io/github/license/nightnoryu/teamcity-monitor)](https://github.com/nightnoryu/teamcity-monitor/blob/main/LICENSE)
[![Build Status](https://github.com/nightnoryu/teamcity-monitor/actions/workflows/ci.yml/badge.svg)](https://github.com/nightnoryu/teamcity-monitor/actions/workflows/ci.yml)

Real-time TeamCity build status monitoring.

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
