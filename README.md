# Local Persist Volume Plugin for Docker

[![GitHub Release](https://img.shields.io/github/release/smashingtags/local-persist.svg)](https://github.com/smashingtags/local-persist/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Create named Docker volumes that persist at specific host paths.

## Why local-persist

Docker's native `local` driver with bind-mount options works, but has friction:

- **Auto-creates directories.** `docker volume create` with local-persist calls `os.MkdirAll` on the mountpoint. Native bind-mount volumes fail at container start if the host path does not exist.
- **Simpler syntax.** One option (`mountpoint`) instead of three (`type`, `device`, `o`). Less YAML, fewer typos.
- **Works with `docker volume prune`.** Native bind-mount volumes are silently skipped by `docker volume prune` due to a long-standing Docker bug ([moby/moby#36907](https://github.com/moby/moby/issues/36907), open since 2019). local-persist volumes prune normally.

### Side-by-side comparison

```yaml
# local-persist
volumes:
  appdata:
    driver: local-persist
    driver_opts:
      mountpoint: /opt/appdata/myapp

# native bind-mount volume (equivalent)
volumes:
  appdata:
    driver: local
    driver_opts:
      type: none
      device: /opt/appdata/myapp
      o: bind
```

## Quick Start

### Docker Compose (recommended)

```yaml
services:
  local-persist:
    image: ghcr.io/smashingtags/local-persist:latest
    container_name: local-persist
    restart: unless-stopped
    privileged: true
    volumes:
      - /run/docker/plugins/:/run/docker/plugins/
      - /var/lib/docker/plugin-data/:/var/lib/docker/plugin-data/
      - /var/run/docker.sock:/var/run/docker.sock:ro
    healthcheck:
      test: ["CMD", "test", "-S", "/run/docker/plugins/local-persist.sock"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
    logging:
      driver: json-file
      options:
        max-size: 10m
        max-file: "3"
```

A ready-to-use `docker-compose.yml` is included in the repo.

### Docker CLI

```bash
docker run -d \
  --name local-persist \
  --restart unless-stopped \
  --privileged \
  -v /run/docker/plugins/:/run/docker/plugins/ \
  -v /var/lib/docker/plugin-data/:/var/lib/docker/plugin-data/ \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  ghcr.io/smashingtags/local-persist:latest
```

### Binary (systemd)

```bash
ARCH=$(uname -m); [ "$ARCH" = "x86_64" ] && ARCH="amd64"; [ "$ARCH" = "aarch64" ] && ARCH="arm64"
curl -fsSL "https://github.com/smashingtags/local-persist/releases/latest/download/local-persist-linux-${ARCH}" \
  -o /usr/local/bin/local-persist
chmod +x /usr/local/bin/local-persist

cp init/systemd.service /etc/systemd/system/local-persist.service
systemctl daemon-reload
systemctl enable --now local-persist
```

## Usage

### Create a volume (CLI)

```bash
docker volume create -d local-persist -o mountpoint=/opt/appdata/myapp --name myapp-data

# Use it in a container
docker run -d -v myapp-data:/data myapp:latest
```

### Docker Compose service

```yaml
services:
  myapp:
    image: myapp:latest
    volumes:
      - myapp-data:/data

volumes:
  myapp-data:
    driver: local-persist
    driver_opts:
      mountpoint: /opt/appdata/myapp
```

### Multi-volume example (media server)

```yaml
services:
  jellyfin:
    image: jellyfin/jellyfin:latest
    volumes:
      - jellyfin-config:/config
      - jellyfin-cache:/cache
      - media-library:/media:ro

volumes:
  jellyfin-config:
    driver: local-persist
    driver_opts:
      mountpoint: /opt/appdata/jellyfin/config
  jellyfin-cache:
    driver: local-persist
    driver_opts:
      mountpoint: /opt/appdata/jellyfin/cache
  media-library:
    driver: local-persist
    driver_opts:
      mountpoint: /mnt/storage/media
```

## How It Works

The plugin registers a Unix socket at `/run/docker/plugins/local-persist.sock` and implements the [Docker Volume Plugin API](https://docs.docker.com/engine/extend/plugins_volume/). When a volume is created with a `mountpoint` option, the driver calls `os.MkdirAll` to ensure the host directory exists, then records the name-to-path mapping. Volume state is persisted to `/var/lib/docker/plugin-data/local-persist.json` so volumes survive plugin restarts. On mount, Docker bind-mounts the host path into the container at the target specified in the volume mapping.

## Project Structure

```
cmd/local-persist/
  main.go                      entry point, serves Unix socket
internal/driver/
  driver.go                    volume plugin implementation (Create, Remove, Mount, List, Get, Path)
  driver_test.go               unit tests
init/
  systemd.service              systemd unit file
docker-compose.yml             run the plugin as a container
Dockerfile                     multi-stage build (golang:alpine -> alpine)
Makefile                       build, test, cross-compile targets
```

## Build

```bash
make build          # build for current platform -> bin/local-persist
make binaries       # cross-compile linux/amd64 + linux/arm64
make test           # go vet + go test -v ./...
make docker         # build Docker image
make run            # sudo go run (for development)
make clean          # remove bin/
```

## Comparison

| Feature | local-persist | Native bind-mount volume | Raw bind mount |
|---|---|---|---|
| Syntax | `driver: local-persist` + 1 option | `driver: local` + 3 options | `volumes: - /host:/container` |
| Auto-creates host dir | Yes (`os.MkdirAll`) | No (fails at runtime) | No (Docker creates it as root) |
| Shows in `docker volume ls` | Yes | Yes | No |
| Works with `docker volume prune` | Yes | No ([moby/moby#36907](https://github.com/moby/moby/issues/36907)) | N/A (not a volume) |
| Named volume references | Yes | Yes | No |
| Survives `docker compose down` | Yes (unless `-v` flag) | Yes (unless `-v` flag) | N/A |
| Requires plugin running | Yes | No | No |
| SELinux `:z`/`:Z` support | No | No | Yes |
| Works in Docker Swarm | No (local scope only) | Yes | Yes |
| Cross-platform | Linux only | Linux, macOS, Windows | Linux, macOS, Windows |

## Supported Architectures

- `linux/amd64`
- `linux/arm64`

## Limitations

- **Linux only.** The plugin uses a Unix domain socket at `/run/docker/plugins/`. It does not work on macOS or Windows (including Docker Desktop with WSL2).
- **No SELinux relabeling.** The `:z` and `:Z` mount options are not supported. If your host enforces SELinux, you will need to set labels manually or use `semanage fcontext`.
- **Plugin must be running.** If the local-persist container or service stops, volumes using the `local-persist` driver cannot be mounted until it restarts. Containers with those volumes will fail to start.
- **Local scope only.** Volumes are not replicated across Swarm nodes. In a Swarm cluster, the volume exists only on the node where it was created.
- **No volume content deletion on remove.** `docker volume rm` removes the volume from local-persist's state file but does not delete the host directory or its contents. This is intentional — data preservation is the default.
- **Privileged mode required.** The container needs privileged access to create the Unix socket in `/run/docker/plugins/` and to create arbitrary host directories.

## See Also

- [Eight.ly Container Edition](https://github.com/smashingtags/eightly) -- Docker app store with 3,400+ templates for self-hosted applications.
- [Eight.ly OS](https://github.com/smashingtags/eightly-os) -- NAS operating system with Docker, VMs, and storage management.

## License

MIT. See [LICENSE](LICENSE).

Originally created by [Cameron Spear / MatchbookLab](https://github.com/MatchbookLab/local-persist) (archived Sept 2025). Maintained by [Imogen Labs](https://eight.ly).

## Links

- [Wiki](https://github.com/smashingtags/local-persist/wiki)
- [GitHub Issues](https://github.com/smashingtags/local-persist/issues)
