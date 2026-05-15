# CLAUDE.md

Docker volume plugin that creates named volumes persisting at user-specified host paths.

## Commands

```bash
make build      # build binary for current arch
make binaries   # cross-compile linux/amd64 + arm64
make test       # vet + run tests
make docker     # build Docker image
```

## Structure

```
cmd/local-persist/main.go          — entry point
internal/driver/driver.go          — volume driver implementation
internal/driver/driver_test.go     — tests
```

State persisted to `/var/lib/docker/plugin-data/local-persist.json`. Plugin communicates via `/run/docker/plugins/local-persist.sock`.

## Rules

- Linux only
- Single direct dependency (docker/go-plugins-helpers)
- `go vet` must pass
- All tests must pass
