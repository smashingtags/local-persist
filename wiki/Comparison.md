# local-persist vs Native Docker Bind Mount Volumes

Docker offers several ways to give containers access to host directories. The three most common approaches are **local-persist** (a volume plugin), **native bind mount volumes** (using the built-in `local` driver with bind options), and **raw bind mounts** (inline host-path mappings). Each has different trade-offs around discoverability, lifecycle management, portability, and simplicity. This page explains when and why you would choose each one.

---

## Syntax Comparison

### local-persist

**Docker Compose:**

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

**CLI:**

```bash
docker volume create -d local-persist -o mountpoint=/opt/appdata/myapp --name myapp-data
docker run -v myapp-data:/data myapp:latest
```

The `mountpoint` option is the only required driver option. local-persist creates the directory on the host if it does not exist.

### Native Bind Mount Volume (local driver)

**Docker Compose:**

```yaml
services:
  myapp:
    image: myapp:latest
    volumes:
      - myapp-data:/data

volumes:
  myapp-data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/appdata/myapp
```

**CLI:**

```bash
docker volume create --driver local \
  --opt type=none --opt o=bind --opt device=/opt/appdata/myapp \
  --name myapp-data
docker run -v myapp-data:/data myapp:latest
```

This uses Docker's built-in `local` driver with three required options: `type: none`, `o: bind`, and `device` pointing to the host path. The directory **must already exist** on the host or volume creation will fail.

### Raw Bind Mount

**Docker Compose:**

```yaml
services:
  myapp:
    image: myapp:latest
    volumes:
      - /opt/appdata/myapp:/data
```

**CLI:**

```bash
docker run -v /opt/appdata/myapp:/data myapp:latest
```

Raw bind mounts use the `host_path:container_path` syntax directly in the service definition. They are **not Docker volumes** -- they do not appear in `docker volume ls`, cannot be managed with `docker volume` commands, and have no entry in the `volumes:` top-level key.

> **Note:** Raw bind mounts are not Docker volumes. Docker creates the host directory automatically (as root), but the mount has no volume name, no driver, and no lifecycle management.

---

## CLI Comparison

Creating the same volume with each approach at the command line:

```bash
# local-persist
docker volume create -d local-persist -o mountpoint=/opt/appdata/myapp --name myapp-data

# Native bind mount volume
docker volume create --driver local \
  --opt type=none --opt o=bind --opt device=/opt/appdata/myapp \
  --name myapp-data

# Raw bind mount (no volume to create -- used inline)
docker run -v /opt/appdata/myapp:/data myapp:latest
```

---

## Feature Comparison

| Feature | local-persist | Native bind mount volume | Raw bind mount |
|---|---|---|---|
| Syntax options | 1 (`mountpoint`) | 3 (`type`, `o`, `device`) | Inline path only |
| Auto-creates host directory | Yes | No (must pre-exist) | Yes (as root) |
| Appears in `docker volume ls` | Yes | Yes | No |
| Appears in `docker volume inspect` | Yes | Yes | No |
| Works with `docker volume prune` | Yes | No ([moby/moby#36907][prune-bug]) | N/A |
| SELinux `:z` / `:Z` labels | No | Yes | Yes |
| Requires plugin installation | Yes | No | No |
| Data survives `docker volume rm` | Yes | Yes | N/A |
| Docker Desktop (Mac / Windows) | No | Yes | Yes |
| Docker Swarm mode | No | Yes | Yes |
| Named volume (referenceable) | Yes | Yes | No |
| Can share across services | Yes (by volume name) | Yes (by volume name) | Yes (by path) |

[prune-bug]: https://github.com/moby/moby/issues/36907

---

## When to Use local-persist

- You want named volumes that point to specific host directories with minimal syntax -- one driver option instead of three.
- You rely on `docker volume prune` to clean up unused volumes and need bind-backed volumes to be included in the prune list. Native bind mount volumes are silently skipped by prune (see below).
- You want the host directory to be created automatically if it does not exist. Native bind mount volumes fail if the directory is missing.
- You manage a homelab or self-hosted stack where all services run on a single Linux host and you do not need Swarm or Docker Desktop support.
- You use a convention like `/opt/appdata/<service>/` for all container config and want volumes to follow that pattern declaratively.
- You want `docker volume rm` to deregister the volume but leave the data on disk -- local-persist never deletes host directories.
- You are deploying media stacks (Plex, Sonarr, Radarr, etc.) where each service needs its config directory mapped to a predictable host path.
- You want a single source of truth for all volume-to-path mappings (the state file at `/var/lib/docker/plugin-data/local-persist.json`).

---

## When to Use Native Bind Mount Volumes

- You need SELinux `:z` or `:Z` relabeling support (Fedora, RHEL, CentOS, Rocky Linux).
- You run Docker Desktop on macOS or Windows where Unix socket plugins are not available.
- You need Docker Swarm compatibility for multi-node deployments.
- You do not want to install or maintain a third-party plugin and prefer using only built-in Docker functionality.
- You are comfortable creating host directories before running `docker compose up`, or you use a provisioning tool (Ansible, Terraform) that handles directory creation.
- You do not rely on `docker volume prune` to clean up bind-backed volumes (see the prune bug below).
- You need to mount NFS, CIFS, or tmpfs filesystems as volumes -- the native `local` driver supports these via `type` and `o` options, while local-persist only supports local directories.

---

## When to Use Raw Bind Mounts

- You need quick, simple host-path-to-container-path mapping with no volume abstraction.
- You are mounting read-only media directories (e.g., `/mnt/media/movies:/movies:ro`) where volume lifecycle management is unnecessary.
- You do not need the volume to appear in `docker volume ls` or be manageable with `docker volume` commands.
- You want Docker to auto-create the host directory (it will be owned by root).
- You are writing one-off `docker run` commands or quick prototypes rather than long-lived Compose stacks.
- You need to mount individual files (e.g., `/etc/localtime:/etc/localtime:ro`) -- volume drivers only support directories.
- You are sharing host tool binaries or sockets into containers (e.g., `/var/run/docker.sock`).

---

## The `docker volume prune` Bug

Native bind mount volumes (using `driver: local` with `type: none, o: bind`) are **never removed** by `docker volume prune`, even when the volume is completely unused. This is a known Docker bug tracked at [moby/moby#36907](https://github.com/moby/moby/issues/36907), open since 2018.

The root cause is that Docker's prune logic does not inspect driver options to determine whether a local volume is backed by a bind mount. Docker treats all `local` driver volumes with options as potentially containing important data and skips them during prune. As a result, these volumes accumulate over time and must be removed manually with `docker volume rm <name>`.

This is particularly problematic for users who run `docker system prune --volumes` as part of regular maintenance, expecting it to clean up all unused volumes. Bind mount volumes silently survive the prune, cluttering `docker volume ls` output and making it harder to identify truly orphaned volumes.

local-persist volumes do not have this problem. Because they use a separate driver (`local-persist` rather than `local`), Docker's prune logic correctly identifies them as unused and removes them. Note that removing a local-persist volume only deletes the volume reference -- the host directory and its contents are never deleted by `docker volume rm` or `docker volume prune`.

---

## Limitations of local-persist

- **Privileged container required.** local-persist needs access to the Docker socket, the plugin socket directory, and arbitrary host filesystem paths. It runs as a privileged container (or as a root-level systemd service).
- **Linux only.** The plugin communicates via a Unix socket at `/run/docker/plugins/local-persist.sock`. It does not work on Docker Desktop for macOS or Windows.
- **No SELinux support.** local-persist does not apply `:z` or `:Z` relabeling to mounted directories. On SELinux-enforcing systems, containers may get "permission denied" errors unless you configure SELinux policies manually.
- **No Docker Swarm support.** The plugin has `local` scope and does not participate in Swarm's volume scheduling or cross-node volume management.
- **State file must be backed up.** Volume-to-mountpoint mappings are stored in `/var/lib/docker/plugin-data/local-persist.json`. If this file is lost, the plugin loses track of all existing volumes on next restart. Include this path in your backup strategy.
- **Single `mountpoint` option only.** Unlike the native `local` driver, local-persist does not support NFS, CIFS, tmpfs, or other filesystem types. It only creates simple directories on the local filesystem.
- **Plugin must start before dependent containers.** If local-persist is deployed as a container in the same Compose file as your services, use `depends_on` with a healthcheck to ensure the plugin socket exists before services attempt to mount volumes.
- **No volume size limits.** local-persist does not support quotas or size constraints on volumes. The volume can grow to fill the entire filesystem.
