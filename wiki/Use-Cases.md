# Use Cases

Real-world scenarios where local-persist simplifies Docker volume management. Each example includes a working `docker-compose.yml` you can adapt.

---

## 1. Homelab Media Servers

A typical media stack (Plex, Sonarr, Radarr) stores configuration in per-service directories. With local-persist, each service's config lives at a predictable path like `/opt/appdata/plex/`, making it easy to back up, inspect, or move between machines.

```yaml
services:
  local-persist:
    image: smashingtags/local-persist
    restart: always
    volumes:
      - /run/docker/plugins/:/run/docker/plugins/

  plex:
    image: linuxserver/plex
    depends_on:
      - local-persist
    volumes:
      - plex-config:/config
      - /mnt/media:/media
    environment:
      - PUID=1000
      - PGID=1000
    ports:
      - "32400:32400"

  sonarr:
    image: linuxserver/sonarr
    depends_on:
      - local-persist
    volumes:
      - sonarr-config:/config
      - /mnt/media:/media
    environment:
      - PUID=1000
      - PGID=1000
    ports:
      - "8989:8989"

  radarr:
    image: linuxserver/radarr
    depends_on:
      - local-persist
    volumes:
      - radarr-config:/config
      - /mnt/media:/media
    environment:
      - PUID=1000
      - PGID=1000
    ports:
      - "7878:7878"

volumes:
  plex-config:
    driver: local-persist
    driver_opts:
      mountpoint: /opt/appdata/plex
  sonarr-config:
    driver: local-persist
    driver_opts:
      mountpoint: /opt/appdata/sonarr
  radarr-config:
    driver: local-persist
    driver_opts:
      mountpoint: /opt/appdata/radarr
```

After `docker compose up`, the host filesystem looks like:

```
/opt/appdata/
  plex/
    Library/
    Preferences.xml
  sonarr/
    config.xml
    sonarr.db
  radarr/
    config.xml
    radarr.db
```

Every file is right where you expect it.

---

## 2. Docker Compose Template Libraries

Template authors who distribute `docker-compose.yml` files benefit from local-persist's simpler syntax. Compare three ways to put data at `/opt/appdata/nginx`:

**Bind mount (long syntax):** 3 options (`type`, `source`, `target`)
```yaml
services:
  nginx:
    volumes:
      - type: bind
        source: /opt/appdata/nginx
        target: /etc/nginx
```

**Bind mount (short syntax):** host and container paths mixed in one string
```yaml
services:
  nginx:
    volumes:
      - /opt/appdata/nginx:/etc/nginx
```

**local-persist:** 1 option (`mountpoint`), proper named volume
```yaml
services:
  nginx:
    volumes:
      - nginx-config:/etc/nginx

volumes:
  nginx-config:
    driver: local-persist
    driver_opts:
      mountpoint: /opt/appdata/nginx
```

For template libraries with hundreds of compose files, every template follows the same pattern: service references a named volume, volume block declares driver and mountpoint, users change one path. Fewer options means fewer mistakes, and `docker volume ls` shows volumes by name instead of hiding them as anonymous bind mounts.

---

## 3. Disaster Recovery and Backups

When all your application data lives under a single directory tree, backups become trivial. No need to enumerate Docker volume IDs or guess which anonymous volume belongs to which container.

**Backup everything in one command:**

```bash
tar czf /backup/appdata-$(date +%Y%m%d).tar.gz -C /opt appdata/
```

That's it. Every container's config, database, and state is captured.

**Restore to a new machine:**

```bash
# 1. Install Docker and local-persist on the new host
docker run -d --name local-persist --restart always \
  -v /run/docker/plugins/:/run/docker/plugins/ \
  smashingtags/local-persist

# 2. Extract the backup
tar xzf /backup/appdata-20260514.tar.gz -C /opt/

# 3. Bring up the stack
docker compose up -d
```

The volumes are declared in the compose file with their mountpoints. local-persist creates the Docker volume entries pointing at the directories that already contain the restored data. No `docker volume create` commands, no copying data into opaque Docker-managed directories.

**Selective restore (single service):**

```bash
docker compose stop sonarr
tar xzf /backup/appdata-20260514.tar.gz -C /opt/ appdata/sonarr/
docker compose start sonarr
```

Because each service maps to a distinct directory, you can restore individual services without touching the rest of the stack. Add a cron job for nightly backups and the entire strategy works because local-persist gives you one known directory tree.

---

## 4. Multi-App Deployments (50+ Containers)

When every service uses local-persist with a consistent mountpoint convention, large deployments stay manageable. Here's a fragment showing the pattern at scale (the `local-persist` service definition is omitted for brevity -- same as previous examples):

```yaml
services:
  heimdall:
    image: linuxserver/heimdall
    volumes: [heimdall-config:/config]
    ports: ["8080:80"]

  jellyfin:
    image: jellyfin/jellyfin
    volumes: [jellyfin-config:/config, jellyfin-cache:/cache, "/mnt/media:/media:ro"]
    ports: ["8096:8096"]

  nextcloud:
    image: nextcloud
    volumes: [nextcloud-data:/var/www/html]
    ports: ["8443:443"]

  vaultwarden:
    image: vaultwarden/server
    volumes: [vaultwarden-data:/data]
    ports: ["8222:80"]

  homeassistant:
    image: homeassistant/home-assistant
    volumes: [homeassistant-config:/config]
    network_mode: host

volumes:
  heimdall-config:
    driver: local-persist
    driver_opts: { mountpoint: /opt/appdata/heimdall }
  jellyfin-config:
    driver: local-persist
    driver_opts: { mountpoint: /opt/appdata/jellyfin/config }
  jellyfin-cache:
    driver: local-persist
    driver_opts: { mountpoint: /opt/appdata/jellyfin/cache }
  nextcloud-data:
    driver: local-persist
    driver_opts: { mountpoint: /opt/appdata/nextcloud }
  vaultwarden-data:
    driver: local-persist
    driver_opts: { mountpoint: /opt/appdata/vaultwarden }
  homeassistant-config:
    driver: local-persist
    driver_opts: { mountpoint: /opt/appdata/homeassistant }
```

The resulting directory tree on the host:

```
/opt/appdata/
  heimdall/
  homeassistant/
  jellyfin/
    cache/
    config/
  nextcloud/
  vaultwarden/
```

At 50+ containers, this convention pays off:

- **`ls /opt/appdata/`** shows every service at a glance
- **`du -sh /opt/appdata/*`** shows disk usage per service
- **`rsync -a /opt/appdata/ backup:/opt/appdata/`** replicates everything
- Moving a service to another host means copying one directory

Without local-persist, the same containers create directories under `/var/lib/docker/volumes/` with generated hash names. Finding which volume belongs to which service requires `docker volume inspect` on each one.

---

## 5. Migrating from Unraid or TrueNAS

Users moving from Unraid or TrueNAS to a standard Docker host often have strong opinions about where their data lives. These platforms let you pick exact paths, and users build backup scripts, monitoring, and muscle memory around those paths.

### Unraid-style layout

Unraid stores app data under `/mnt/user/appdata/` by convention. Just use that as your mountpoint base:

```yaml
volumes:
  plex-config:
    driver: local-persist
    driver_opts:
      mountpoint: /mnt/user/appdata/plex
  nzbget-config:
    driver: local-persist
    driver_opts:
      mountpoint: /mnt/user/appdata/nzbget
  tautulli-config:
    driver: local-persist
    driver_opts:
      mountpoint: /mnt/user/appdata/tautulli
```

If you're copying appdata from an Unraid server, the data lands in the same paths it came from. No remapping required.

### TrueNAS-style layout

TrueNAS typically stores data on ZFS datasets under `/mnt/tank/` or `/mnt/pool/`:

```yaml
volumes:
  jellyfin-config:
    driver: local-persist
    driver_opts:
      mountpoint: /mnt/tank/appdata/jellyfin
  sabnzbd-config:
    driver: local-persist
    driver_opts:
      mountpoint: /mnt/tank/appdata/sabnzbd
```

### Migration steps

Whether coming from Unraid or TrueNAS:

1. **Copy appdata** from the old system to the new host, preserving the same path structure
2. **Install local-persist** on the new Docker host
3. **Write your compose files** with mountpoints matching the old paths
4. **`docker compose up -d`** and the services find their data where they left it

The key insight is that local-persist doesn't care what the path is. `/opt/appdata/`, `/mnt/user/appdata/`, `/mnt/tank/appdata/` -- any path works. Match whatever convention you're coming from, and your existing backup scripts, monitoring, and documentation all keep working.
