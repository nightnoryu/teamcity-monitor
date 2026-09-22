# Repository Guidelines

## Project Structure & Module Organization

`backend/cmd/teamcity-monitor/` starts the Go HTTP service. Code in `backend/internal/` handles TeamCity requests, configuration, polling, aggregation, and web UI delivery. `web/src/` contains the React dashboard: `components/`, `hooks/`, `api/`, `utils/`, and `styles/`. Tests sit beside the code they cover. Root files provide the example configuration, Docker Compose setup, and `mise.toml` tasks; CI is in `.github/workflows/`.

## Build, Test, and Development Commands

Install the Go, Node, pnpm, and golangci-lint versions declared in `mise.toml`. Copy `config.example.toml` to `config.toml` and supply your TeamCity URL and access token for local runs.

- `mise run`: build both parts, then run tests and linters; this is the CI command.
- `mise run dev`: build the backend and start the Docker Compose stack. Frontend edits reload through Vite.
- `mise run dev:reload`: rebuild and restart the backend after Go changes. `mise run dev:down` stops the stack.
- `mise run build`, `mise run test`, and `mise run lint`: run those stages separately. Use `mise run backend:test` or `mise run web:test` for a focused test pass.

## Coding Style & Naming Conventions

Format Go with `gofmt`; use lowercase package names and adjacent `*_test.go` files. Follow existing frontend style: four-space TSX indentation, double-quoted strings, PascalCase component files, and `use` prefixes for hooks. Put shared API types in `web/src/api/`. ESLint checks TypeScript and React rules; `golangci-lint` checks Go.

## Testing Guidelines

Go tests use `testing` and `testify`. Frontend tests use Vitest, jsdom, and Testing Library; name them `*.test.tsx` beside the component. Add focused behavior tests when changing configuration, TeamCity fetching, polling, aggregation, or UI states. Run `mise run` before a pull request. No coverage threshold is configured.

## Commit & Pull Request Guidelines

Recent commits use short, imperative subjects, such as `Reject invalid poll intervals`. Keep commits focused. In pull requests, describe the behavior change, link a relevant issue when available, and add screenshots for visible dashboard changes. Confirm CI passes.

## Security & Configuration

Keep access tokens in the gitignored `config.toml`. Add safe placeholders to `config.example.toml` when introducing settings. Keep TLS verification enabled except when a local self-signed setup requires otherwise.
