<p align="center"><img src="https://github.com/user-attachments/assets/3448061c-3640-475b-956a-aa75e943d39b" width="800" alt="TeamCity Monitor dashboard"></p>

<p align="center">
  <a href="https://github.com/nightnoryu/teamcity-monitor/releases"><img src="https://img.shields.io/github/release/nightnoryu/teamcity-monitor.svg?cache-control=no-cache" alt="Latest release"></a>
  <a href="https://github.com/nightnoryu/teamcity-monitor/blob/main/LICENSE"><img src="https://img.shields.io/github/license/nightnoryu/teamcity-monitor?cache-control=no-cache" alt="License"></a>
  <a href="https://github.com/nightnoryu/teamcity-monitor/actions/workflows/ci.yml"><img src="https://github.com/nightnoryu/teamcity-monitor/actions/workflows/ci.yml/badge.svg?cache-control=no-cache" alt="CI status"></a>
</p>

**TeamCity Monitor** is a lightweight, self-hosted dashboard that turns the
latest TeamCity build attempts into an environment-oriented view. See what is
running, queued, successful, failed, or unavailable across the projects and
regions you care about—without digging through individual build configurations.

> [!NOTE]
> This is a high-level operational view, not a replacement for TeamCity. A
> successful build means a deployment only when that build configuration
> performs deployment; use the linked TeamCity build to investigate details.

## ✨ Features

- Environment and project-oriented dashboard
- Multiple build configurations per environment and grouping
- Continuous TeamCity polling with partial-failure reporting
- Latest branch, build number, trigger, timing, and TeamCity build link
- Best-effort attribution of the last environment-branch parameter change
- One static Go binary with an embedded React frontend

## 🚀 Run your own

Create a `config.toml` from [the example](config.example.toml), set your
TeamCity URL and access token, then run the published image:

```shell
docker run -d --name teamcity-monitor \
  --restart unless-stopped \
  -p 127.0.0.1:8080:8080 \
  -v "$(pwd)/config.toml:/app/config.toml:ro" \
  ghcr.io/nightnoryu/teamcity-monitor:latest
```

The service has no built-in authentication. Keep it on a trusted internal
network, or put an authenticated reverse proxy in front of it.

See the [deployment guide](docs/deployment.md) for Compose, configuration,
health checks, and operational guidance.

## 📚 Documentation

- [Architecture](docs/architecture.md) — components, data flow, polling, and API behavior
- [Deployment guide](docs/deployment.md) — configuration, Docker, health checks, and access model
- [Example monitor configuration](config.example.toml)
- [Changelog](CHANGELOG.md)

## ⚒️ Local development

Install [mise](https://mise.jdx.dev) and Docker with the Compose plugin, copy
`config.example.toml` to `config.toml`, then run:

```shell
mise run dev
```

Vite reloads frontend changes. After Go changes, run `mise run dev:reload`.
`mise run` performs the full build, test, and lint pass.

## 📜 License

Distributed under the MIT License. See [License](/LICENSE) for more information.
