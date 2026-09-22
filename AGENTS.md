# Repository Guidelines

## Project Structure & Module Organization

`backend/` contains the Go service: `cmd/teamcity-monitor/` wires up the server, while `internal/teamcity/`, `internal/monitor/`, and `internal/monitorconfig/` handle API access, polling and aggregation, and configuration. `web/src/` contains the React dashboard, with components, hooks, API types, styles, and nearby tests. Root files include the example configuration, Dockerfiles, Compose setup, and `mise.toml` task definitions. CI lives in `.github/workflows/`.

## Build, Test, and Development Commands

Install the tool versions declared in `mise.toml` (Go, Node, pnpm, and golangci-lint). Copy `config.example.toml` to `config.toml` and set a TeamCity URL and access token before starting the stack.

- `mise run` builds the backend and starts the local Docker Compose stack.
- `mise run dev:reload` rebuilds and restarts the backend after Go changes; Vite reloads frontend changes automatically.
- `mise run build` builds both the Go binary and production web assets.
- `mise run check` runs all tests and linters. Use `mise run test` or `mise run lint` for either group alone.
- `mise run dev:down` stops the local stack.

## Coding Style & Naming Conventions

Format Go with `gofmt`; keep package names lowercase and tests in adjacent `*_test.go` files. Follow the existing TypeScript/React style: four-space indentation in TSX, double-quoted strings, PascalCase component files, and `use` prefixes for hooks. Keep shared API shapes in `web/src/api/`. The frontend uses ESLint with TypeScript, React Hooks, and React Refresh rules; the backend uses `golangci-lint`.

## Testing Guidelines

Go tests use the standard `testing` package and `testify`; frontend tests use Vitest, jsdom, and Testing Library. Place frontend tests beside their components as `*.test.tsx`. Add focused tests for polling, aggregation, configuration, or UI behavior when changing those paths. Run `mise run check` before opening a pull request. CI also builds both parts; no coverage threshold is configured.

## Commit & Pull Request Guidelines

Recent commits use short, imperative subjects such as `Fix import` and `Add tls setting`; follow that pattern and keep each commit focused. In pull requests, describe the behavior changed, link a relevant issue when one exists, and include screenshots for visible dashboard changes. Ensure the CI build and check jobs pass.

## Security & Configuration

Keep TeamCity access tokens in local `config.toml`, which is gitignored. Update `config.example.toml` with safe placeholders when adding configuration. Leave TLS verification enabled unless a local self-signed setup requires otherwise.
