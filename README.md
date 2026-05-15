# Local Persist Volume Plugin for Docker

[![GitHub Release](https://img.shields.io/github/release/smashingtags/local-persist.svg)](https://github.com/smashingtags/local-persist/releases)

Create named Docker volumes that persist at specific host paths.

> **Note:** Modern Docker supports similar functionality natively with `driver: local` and `driver_opts: {type: none, device: /path, o: bind}`. This plugin remains useful for legacy setups and workflows that depend on the `local-persist` driver name.

## Install

### Binary (systemd)

```bash
ARCH=$(uname -m); [ "$ARCH" = "x86_64" ] && ARCH="amd64"; [ "$ARCH" = "aarch64" ] && ARCH="arm64"
curl -fsSL "https://github.com/smashingtags/local-persist/releases/latest/download/local-persist-linux-${ARCH}" \
  -o /usr/local/bin/local-persist
chmod +x /usr/local/bin/local-persist

# Install systemd service
cp init/systemd.service /etc/systemd/system/local-persist.service
systemctl daemon-reload
systemctl enable --now local-persist
```

### Docker

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

Or use the included `docker-compose.yml`.

## Usage

```bash
# Create a volume
docker volume create -d local-persist -o mountpoint=/opt/appdata/myapp --name myapp-data

# Use it
docker run -d -v myapp-data:/data myapp:latest
```

In Docker Compose:

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

## Build

```bash
make build          # build for current arch
make binaries       # cross-compile linux/amd64 + arm64
make test           # run tests
make docker         # build Docker image
```

## How It Works

The plugin implements the Docker Volume Plugin API via a Unix socket at `/run/docker/plugins/local-persist.sock`. When you create a volume with a `mountpoint` option, it ensures the directory exists on the host and maps it as the volume's backing store. Volume state is persisted to `/var/lib/docker/plugin-data/local-persist.json`.

Key difference from bind mounts: volumes created with this plugin are managed by Docker (show up in `docker volume ls`, can be referenced by name, cleaned up with `docker volume rm`) while still living at a user-specified host path.

## Supported Architectures

- `linux/amd64`
- `linux/arm64`

## License

MIT. See [LICENSE](LICENSE).

## Acknowledgments

Originally created by [Cameron Spear / MatchbookLab](https://github.com/MatchbookLab/local-persist) (archived Sept 2025). Maintained by [smashingtags](https://github.com/smashingtags).
