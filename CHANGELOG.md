# Changelog

## Unreleased

- cache built web assets for one year and revalidate the dashboard HTML on each load

## v0.2.0

- dashboard now shows collection health, stale snapshots, failed build fetches,
  polling duration, and the time of the last successful collection
- failed build fetches are shown per build without hiding data that was collected
  successfully; queued and unavailable builds are distinguished from builds that have never run
- build rows now include TeamCity status details, start time, stable project/build
  identifiers where needed, and clearer branch-editor lookup states
- latest builds are selected across all branches, including canceled and
  failed-to-start attempts; personal builds remain excluded
- configuration is validated at startup, including poll intervals, monitored-build
  references, and TeamCity URLs (including installations hosted under a path)
- added `GET /livez` and `GET /healthz` for container and load-balancer checks -
  readiness reflects whether TeamCity data is currently available.

## v0.1.0

Initial release
