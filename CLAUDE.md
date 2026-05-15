# CLAUDE.md

Docker volume plugin that creates named volumes persisting at user-specified host paths.

## Commands

```bash
make build      # build binary for current arch
make binaries   # cross-compile linux/amd64 + arm64
make test       # run tests
make docker     # build Docker image
```

## Architecture

- **main.go** — entry point, starts Unix socket handler
- **driver.go** — implements Docker Volume Plugin API (Create, Remove, Mount, Unmount, Get, List, Path, Capabilities)
- **driver_test.go** — 4 tests covering the full volume lifecycle

State persisted to `/var/lib/docker/plugin-data/local-persist.json`. Plugin communicates via `/run/docker/plugins/local-persist.sock`.

## Rules

- Linux only
- Single direct dependency (docker/go-plugins-helpers)
- `go vet` must pass
- All tests must pass
