# TeamCity Monitor

Real-time TeamCity build status monitoring.

## Local Development

### Prerequisites

- [mise](https://mise.jdx.dev)
- Docker with docker-compose-plugin

### First launch

```shell
git clone https://github.com/nightnoryu/teamcity-monitor
cd cadence

# Local env domain
echo "127.0.0.1 teamcity-monitor.lan" | sudo tee -a /etc/hosts

mise run
docker compose up -d
```

`teamcity-monitor-web` picks up changes automatically via `vite`.

Build and restart `teamcity-monitor-backend` container to pick up backend changes:

```shell
mise run backend:build
docker compose restart teamcity-monitor-backend
```

## License

Distributed under the MIT License. See [License](/LICENSE) for more information.
