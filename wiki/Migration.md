# Migrating To and From local-persist

This guide covers migrating Docker volumes between local-persist, native bind mount volumes (`driver: local` with `type: none, o: bind`), and default named volumes. In most cases your data does not move -- you are changing how Docker references an existing directory on the host.

---

## Prerequisites

- local-persist is running and healthy (`docker inspect local-persist --format '{{.State.Health.Status}}'` returns `healthy`).
- The plugin socket exists at `/run/docker/plugins/local-persist.sock`.
- You have a backup of any data you are migrating.
- You have the host paths for all volumes you plan to migrate (`docker volume inspect <name>`).

---

## From Native Bind Mount Volumes to local-persist

Native bind mount volumes use three driver options (`type`, `o`, `device`). local-persist uses one (`mountpoint`). Your data stays in the same directory on disk.

### Step 1: Identify volumes to migrate

```bash
docker volume ls --filter driver=local
docker volume inspect myapp-data
```

Look for volumes with `Options` containing `type: none` and `o: bind`. The `device` field is the host path.

### Step 2: Stop containers and remove the old volume

```bash
docker compose down
docker volume rm myapp-data
```

This removes the Docker volume metadata. **It does not delete the directory or its contents on the host.**

### Step 3: Update your Compose file

**Before (native bind mount volume):**
```yaml
volumes:
  myapp-data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/appdata/myapp
```

**After (local-persist):**
```yaml
volumes:
  myapp-data:
    driver: local-persist
    driver_opts:
      mountpoint: /opt/appdata/myapp
```

### Step 4: Recreate and verify

```bash
docker compose up -d
docker volume inspect myapp-data
docker exec myapp ls /data
```

local-persist will see the directory already exists and register it. **Your data does not move** -- you are only changing which driver manages the volume reference.

---

## From local-persist to Native Bind Mounts

Switch to native bind mount volumes when moving to Docker Swarm, Docker Desktop, or a system where you cannot run the plugin.

### Step 1: Identify local-persist volumes and note mountpoints

```bash
docker volume ls --filter driver=local-persist
docker volume inspect myapp-data --format '{{.Mountpoint}}'
# Example output: /opt/appdata/myapp
```

### Step 2: Stop containers and remove the volume

```bash
docker compose down
docker volume rm myapp-data
ls -la /opt/appdata/myapp   # verify data is still on disk
```

This removes the volume from local-persist's state file. **The directory and its contents remain on disk.**

### Step 3: Update your Compose file

**Before (local-persist):**
```yaml
volumes:
  myapp-data:
    driver: local-persist
    driver_opts:
      mountpoint: /opt/appdata/myapp
```

**After (native bind mount volume):**
```yaml
volumes:
  myapp-data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/appdata/myapp
```

> **Note:** The host directory must exist before running `docker compose up`. Native bind mount volumes do not auto-create directories.

### Step 4: Recreate and verify

```bash
docker compose up -d
docker volume inspect myapp-data
docker exec myapp ls /data
```

---

## From Default Named Volumes to local-persist

Default named volumes store data under `/var/lib/docker/volumes/<name>/_data/`. Unlike the previous migrations, this one requires **copying data** because the source and destination are different paths.

### Step 1: Find the data path and copy it

```bash
docker volume inspect myapp-data --format '{{.Mountpoint}}'
# Output: /var/lib/docker/volumes/myapp-data/_data

sudo mkdir -p /opt/appdata/myapp
sudo cp -a /var/lib/docker/volumes/myapp-data/_data/. /opt/appdata/myapp/
```

The `-a` flag preserves permissions, ownership, and timestamps. The trailing `/.` copies contents without creating a nested `_data` directory.

### Step 2: Stop containers and remove the old volume

```bash
docker compose down
docker volume rm myapp-data
```

This deletes the volume metadata **and** the data under `/var/lib/docker/volumes/myapp-data/`. That is why you copied it first.

### Step 3: Update your Compose file

**Before** (default named volume -- either form):
```yaml
volumes:
  myapp-data:          # implicit local driver
```

**After** (local-persist):
```yaml
volumes:
  myapp-data:
    driver: local-persist
    driver_opts:
      mountpoint: /opt/appdata/myapp
```

### Step 4: Recreate and verify

```bash
docker compose up -d
docker volume inspect myapp-data
docker exec myapp ls /data
```

Confirm the data you copied in Step 2 is accessible inside the container.

---

## Backing Up the State File

local-persist tracks all volume-to-mountpoint mappings in a JSON file:

```
/var/lib/docker/plugin-data/local-persist.json
```

```json
{
  "state": {
    "myapp-data": "/opt/appdata/myapp",
    "plex-config": "/opt/appdata/plex",
    "radarr-config": "/opt/appdata/radarr"
  }
}
```

Each key is a Docker volume name, each value is the host path.

**Why back it up:** If this file is lost, local-persist starts with an empty volume map. Running containers keep working (mounts are active in the kernel), but the plugin will not recognize any volumes. You would need to recreate each with `docker volume create -d local-persist -o mountpoint=/path`.

**How to back it up:**

```bash
cp /var/lib/docker/plugin-data/local-persist.json /path/to/backups/
```

If you run local-persist as a container, this file lives on the host at the bind-mounted path, not inside the container. Include the host path in your backup tool or cron job.

**Restoring from backup:** Stop local-persist, copy the backup file to `/var/lib/docker/plugin-data/local-persist.json`, and restart the plugin. It will load the state on startup and recognize all previously registered volumes.
