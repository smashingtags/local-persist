# local-persist Wiki

local-persist is a Docker volume plugin that creates named volumes backed by specific directories on the host filesystem. Instead of Docker choosing an opaque path under `/var/lib/docker/volumes/`, local-persist lets you specify exactly where each volume's data lives. This makes backups predictable, migrations straightforward, and multi-container setups easier to reason about.

## Table of Contents

| Page | Description |
|------|-------------|
| [Use Cases](Use-Cases) | Five real-world scenarios with full docker-compose examples |
| [Comparison](Comparison) | Side-by-side comparison with Docker bind mounts and named volumes |
| [Migration](Migration) | How to migrate existing volumes or bind mounts to local-persist |

## Quick Reference

### Start local-persist

Run the plugin as a container. It registers itself with the Docker daemon through the plugin socket:

```bash
docker run -d \
  --name local-persist \
  --restart always \
  -v /run/docker/plugins/:/run/docker/plugins/ \
  smashingtags/local-persist
```

### Create a volume at a specific host path

The `-o mountpoint=` option tells local-persist where to store the volume's data on the host:

```bash
docker volume create -d local-persist \
  -o mountpoint=/opt/appdata/myapp \
  myapp-data
```

The directory will be created automatically if it doesn't exist.

### Verify the volume

```bash
docker volume inspect myapp-data
```

The output will show `"Mountpoint": "/opt/appdata/myapp"` and the driver as `local-persist`:

```json
[
    {
        "CreatedAt": "2026-05-15T00:00:00Z",
        "Driver": "local-persist",
        "Labels": {},
        "Mountpoint": "/opt/appdata/myapp",
        "Name": "myapp-data",
        "Options": {
            "mountpoint": "/opt/appdata/myapp"
        },
        "Scope": "local"
    }
]
```

### Use the volume in a container

```bash
docker run -d \
  --name myapp \
  -v myapp-data:/data \
  myimage:latest
```

### Use in docker-compose.yml

```yaml
services:
  myapp:
    image: myimage:latest
    volumes:
      - myapp-data:/data

volumes:
  myapp-data:
    driver: local-persist
    driver_opts:
      mountpoint: /opt/appdata/myapp
```

## Requirements

- Docker Engine 19.03 or later (Linux only)
- The `/run/docker/plugins/` socket directory must be bind-mounted into the local-persist container
- local-persist runs as a single container alongside your application containers
- No external dependencies, no database, no configuration files

## How It Works

When Docker needs to mount a `local-persist` volume, it asks the plugin for the path. The plugin returns the mountpoint you specified, and Docker bind-mounts that directory into the container. The plugin also creates the directory on the host if it doesn't already exist.

Volume metadata (name-to-path mappings) is stored in a JSON file at `/var/lib/docker/plugin-data/local-persist.json` inside the plugin container. This file is persisted through container restarts via the plugin socket mount.

## About

local-persist is MIT-licensed, originally created by [MatchbookLab](https://github.com/MatchbookLab/local-persist), now maintained by Imogen Labs as part of the Eight.ly ecosystem for self-hosted infrastructure.

If you manage many Docker containers, you may also be interested in [Eight.ly Container Edition](https://github.com/smashingtags/eightly) (Docker app store with 3,400+ templates) or [Eight.ly OS](https://github.com/smashingtags/eightly-os) (NAS operating system with Docker, VMs, and storage management).
