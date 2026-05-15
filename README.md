# Local Persist Volume Plugin for Docker

[![Docker Pulls](https://img.shields.io/docker/pulls/smashingtags/local-persist)](https://github.com/smashingtags/local-persist/pkgs/container/local-persist)
[![GitHub Release](https://img.shields.io/github/release/smashingtags/local-persist.svg)](https://github.com/smashingtags/local-persist/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/smashingtags/local-persist)](https://goreportcard.com/report/github.com/smashingtags/local-persist)

**Modern Go implementation of the Docker local-persist volume plugin**

Create named local volumes that persist in the location(s) you want! This is a modernized, containerized version of the original local-persist plugin, rewritten in Go with static compilation for maximum compatibility.

## 🚀 Quick Start

### Container Deployment (Recommended)

The easiest way to run local-persist is as a container using the pre-built image:

```bash
# Create required directories
sudo mkdir -p /run/docker/plugins /var/lib/docker/plugin-data /opt/appdata
sudo chmod 755 /run/docker/plugins /var/lib/docker/plugin-data /opt/appdata

# Deploy local-persist container
docker run -d \
  --name local-persist \
  --restart unless-stopped \
  --privileged \
  --user root \
  -v /var/run/docker.sock:/var/run/docker.sock:rw \
  -v /run/docker/plugins:/run/docker/plugins:rw \
  -v /var/lib/docker/plugin-data:/var/lib/docker/plugin-data:rw \
  -v /opt/appdata:/opt/appdata:shared \
  ghcr.io/smashingtags/local-persist:latest
```

### Docker Compose Deployment

```yaml
services:
  local-persist:
    container_name: local-persist
    image: ghcr.io/smashingtags/local-persist:latest
    restart: unless-stopped
    user: root
    privileged: true
    cap_add:
      - SYS_ADMIN
      - CHOWN
      - DAC_OVERRIDE
      - FOWNER
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:rw
      - /run/docker/plugins:/run/docker/plugins:rw
      - /var/lib/docker/plugin-data:/var/lib/docker/plugin-data:rw
      - /opt/appdata:/opt/appdata:shared
      - /mnt:/mnt:shared
    environment:
      - PUID=0
      - PGID=0
      - TZ=UTC
    healthcheck:
      test: ["CMD", "test", "-S", "/run/docker/plugins/local-persist.sock"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 15s
```

## 📚 Usage

Once local-persist is running, you can create persistent volumes that map to specific host directories:

### Create a Volume

```bash
# Create a volume that persists to /opt/appdata/myapp
docker volume create -d local-persist -o mountpoint=/opt/appdata/myapp --name myapp-data
```

### Use in Container

```bash
# Use the volume in a container
docker run -d --name myapp -v myapp-data:/data myapp:latest
```

### Use in Docker Compose

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

## 🔧 Configuration

### Environment Variables

- `PUID`: User ID for file ownership (default: 0)
- `PGID`: Group ID for file ownership (default: 0) 
- `TZ`: Timezone (default: UTC)

### Volume Options

- `mountpoint`: Host directory where the volume will persist (required)

## 🏗️ Building from Source

```bash
# Clone the repository
git clone https://github.com/smashingtags/local-persist.git
cd local-persist

# Build the binary
make build

# Build the Docker image
make docker-build

# Run tests
make test
```

## 🐳 Container Images

Pre-built images are available from GitHub Container Registry:

- `ghcr.io/smashingtags/local-persist:latest` - Latest stable release
- `ghcr.io/smashingtags/local-persist:v1.x.x` - Specific version tags

### Supported Architectures

- `linux/amd64`
- `linux/arm64`

## 🔒 Security Considerations

This plugin requires privileged access to:
- Docker socket (`/var/run/docker.sock`)
- Plugin socket directory (`/run/docker/plugins`)
- Host filesystem for volume mounts

**Important**: Only run this plugin on trusted Docker hosts as it has extensive filesystem access.

## 🆚 Comparison with Alternatives

### vs. Named Volumes
- ✅ **Local-persist**: Data stored in specific host locations
- ❌ **Named volumes**: Data stored in Docker's managed directories

### vs. Bind Mounts
- ✅ **Local-persist**: Volume lifecycle managed by Docker
- ❌ **Bind mounts**: Manual directory management required

### vs. Other Volume Plugins
- ✅ **Local-persist**: Simple, lightweight, no external dependencies
- ❌ **Other plugins**: Often require external storage systems

## 🛠️ Integration Examples

### HomelabARR Media Stack

This plugin was specifically modernized for use with HomelabARR (formerly DockServer):

```bash
# Install with HomelabARR
curl -fsSL https://raw.githubusercontent.com/smashingtags/homelabarr-cli/master/scripts/install-local-persist-container.sh | sudo bash
```

### Plex Media Server

```yaml
services:
  plex:
    image: plexinc/pms-docker:latest
    volumes:
      - plex-config:/config
      - plex-transcode:/transcode
      - /mnt/media:/media:ro

volumes:
  plex-config:
    driver: local-persist
    driver_opts:
      mountpoint: /opt/appdata/plex
  plex-transcode:
    driver: local-persist
    driver_opts:
      mountpoint: /tmp/plex-transcode
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Originally created by [Cameron Spear / MatchbookLab](https://github.com/MatchbookLab/local-persist) (archived Sept 2025)
- Maintained by [smashingtags](https://github.com/smashingtags) since 2025

## 📞 Support

- 🐛 **Issues**: [GitHub Issues](https://github.com/smashingtags/local-persist/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/smashingtags/local-persist/discussions)
- 📖 **Documentation**: This README