# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Docker volume plugin called `local-persist` that creates named local volumes persisting in user-specified locations. It's written in Go and implements the Docker Volume Plugin API to allow containers to use persistent storage outside of the default Docker volume driver behavior.

## Common Development Commands

### Building
- `make binary` - Build binary for current architecture
- `make binaries` - Build all supported architecture binaries (Linux amd64, arm64)
- `make docker-build` - Build using Docker container
- `go build -o bin/local-persist -v` - Direct Go build

### Testing
- `make test` - Run all tests with verbose output
- `go test -v .` - Run tests directly
- `make coverage` - Generate test coverage report and open in browser

### Running
- `make run` - Run the plugin locally with sudo (requires root for Docker socket access)
- `make docker-run` - Run using Docker container via script

### Development Scripts
- `./scripts/install.sh` - Install plugin as systemd service
- `./scripts/docker-run.sh` - Run plugin in Docker container
- `./scripts/release.sh` - Build and release binaries

## Architecture

### Core Components

**main.go**: Entry point that creates a `localPersistDriver` instance and starts the Docker volume plugin handler using Unix socket communication.

**driver.go**: Contains the main `localPersistDriver` struct and implements the Docker Volume Plugin API interface with methods:
- `Create()` - Creates volume with required `mountpoint` option
- `Remove()` - Removes volume from state (but preserves data)
- `Mount()/Unmount()` - Handle container mount/unmount operations
- `Get()/List()` - Retrieve volume information
- `Path()` - Return volume mountpoint path
- `Capabilities()` - Declare plugin scope as "local"

### Key Design Patterns

- **State Persistence**: Volume state is saved to `/var/lib/docker/plugin-data/local-persist.json` to survive plugin restarts
- **Thread Safety**: Uses mutex for concurrent volume operations
- **Colored Logging**: Extensive console output with color-coded messages for debugging
- **Docker API Integration**: Can query Docker daemon to rediscover existing volumes from containers

### Volume Lifecycle

1. Volume creation requires `mountpoint` option specifying host directory
2. Plugin ensures target directory exists with `os.MkdirAll()`
3. Volume state stored in both memory map and JSON file
4. Data persists even after volume removal (unlike default Docker local driver)

### Plugin Communication

The plugin creates a Unix socket at `/run/docker/plugins/local-persist.sock` that Docker daemon uses to communicate volume operations. When running in container, this socket must be bind-mounted to the host.

## Testing

Tests are located in `driver_test.go` and cover:
- Volume creation with and without required options
- Volume retrieval and listing
- Mount/unmount/path operations
- Proper cleanup of test volumes

Test environment variable `GO_ENV=test` is used during testing.

## Docker Integration

The plugin can run either:
1. **Natively on host** - Installed as systemd service, communicates via Unix socket
2. **In container** - Requires bind mounts for Docker socket and plugin directory

For container deployment, requires these volume mounts:
- `/run/docker/plugins/:/run/docker/plugins/` - Plugin socket communication
- `/var/lib/docker/plugin-data/:/var/lib/docker/plugin-data/` - State persistence
- Host directories where volumes will be created