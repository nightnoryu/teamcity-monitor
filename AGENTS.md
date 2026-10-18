# Repository Guidelines

## Project Structure & Module Organization

TeamCity Monitor is a Go service with an embedded React dashboard.
- `backend/cmd/teamcity-monitor/`: entry point, HTTP API, configuration, and service lifecycle.
- `backend/internal/`: TeamCity client, polling and aggregation (`monitor`), monitor configuration, observability, and embedded frontend (`webui`).
- `web/src/`: React components, hooks, API types, utilities, and CSS; tests live beside source files.
- `docs/`: architecture, deployment instructions, and screenshots.
- `config.example.toml`, `compose.yml`, and `mise.toml`: configuration template, local stack, and development tasks.

## Build, Test, and Development Commands

Install tool versions with `mise install`. Copy `config.example.toml` to `config.toml` and configure TeamCity access before starting locally.

Prefer mise for running required development tools so repository-pinned versions are used. Use existing `mise run` tasks when available; otherwise use `mise exec --`, for example `mise exec -- pnpm install` instead of `pnpm install`.

- `mise run dev`: build the backend and start Docker Compose; open `http://localhost:3000`. Vite reloads frontend changes.
- `mise run dev:reload`: rebuild and restart the backend after Go changes.
- `mise run dev:down`: stop the local stack.
- `mise run build`: build the frontend and package it into the Go binary at `backend/bin/teamcity-monitor`.
- `mise run test`: run backend, monitor race, and frontend tests.
- `mise run lint`: run golangci-lint and ESLint.
- `mise run`: run the full build, tests, and lint checks used by CI.

## Coding Style & Naming Conventions

Format Go with gofmt, goimports, and gci as configured in `backend/.golangci.yml`; use tabs and group standard, external, and local imports. Use lowercase package names and idiomatic exported Go names.

Follow existing TypeScript files: four-space indentation, double quotes, semicolons, PascalCase component filenames, and `use`-prefixed hooks. ESLint checks TypeScript and React hooks. Keep frontend API types aligned with backend JSON responses.

## Testing Guidelines

Backend tests use Go's `testing` package and Testify; name files `*_test.go`. Frontend tests use Vitest, Testing Library, and jsdom; name files `*.test.tsx`. Cover changed behavior and failure cases. No numeric coverage threshold is configured. Run focused checks with `mise run backend:test` or `mise run web:test`; monitor concurrency changes also require `mise run backend:test:race`.

## Commit & Pull Request Guidelines

Recent commits use short imperative subjects, such as `Document observability settings` and `Sanitize TeamCity transport errors`. Follow that style. PRs should explain the behavior change, link relevant issues, report validation, and include screenshots for dashboard changes. Run `mise run` before requesting review.

## Security & Configuration

Keep TeamCity tokens in ignored local configuration and out of logs and commits. The service has no built-in authentication; deploy on a trusted network or behind an authenticated reverse proxy. See `docs/deployment.md` for operational settings.

## Definition of Done

`mise run` must pass before work is considered complete. Resolve all build, test, and lint failures before handing off changes.
