<p align="center">
<img src="./.github/logo.svg" width="80">
<br><br>
<img alt="GitHub Tag" src="https://img.shields.io/github/v/tag/MizuchiLabs/tether?label=Version">
<img alt="GitHub License" src="https://img.shields.io/github/license/MizuchiLabs/tether">
<img alt="GitHub Issues or Pull Requests" src="https://img.shields.io/github/issues/MizuchiLabs/tether">
</p>

# Tether

**Tether** is the central hub for your distributed Traefik setup. It gathers information from all your servers and tells Traefik exactly how to route traffic to your apps.

Think of it as a **central operator**: multiple servers (running [Tetherd](https://github.com/MizuchiLabs/tetherd)) tell Tether which apps are running, and Tether gives Traefik a single, master list of all of them.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./.github/screenshots/dashboard-dark.png">
  <img alt="Tether dashboard with four connected agents and the merged Traefik config" src="./.github/screenshots/dashboard-light.png">
</picture>

## Why use Tether?

If you have multiple physical servers or VPS instances but don't want the complexity of Kubernetes or Docker Swarm, Tether is for you.

- **One Public IP:** Point your router/firewall (Port 443) to just **one** server running Traefik.
- **Auto-Discovery:** Apps on other servers are automatically found and added to Traefik.
- **Simple:** No complex networking, no KV stores (Consul/Redis), just simple HTTP heartbeats.

## The "One IP, Many Servers" Setup

Imagine you have 3 servers, but only one public WAN connection.

1. **Server A (The Gateway):** Runs **Traefik** and **Tether**. Your router forwards port 443 here.
2. **Server B & C (The Workers):** Run your apps (Websites, APIs, etc.) and a small agent called [Tetherd](https://github.com/MizuchiLabs/tetherd).

Tetherd on Server B/C tells Tether (on Server A) what is running. Traefik asks Tether for the config and magically knows to send `app1.com` to Server B and `app2.com` to Server C.

## Quick Start

Run Tether using Docker Compose:

```yaml
services:
  tether:
    image: ghcr.io/mizuchilabs/tether:latest
    ports:
      - 3000:3000
    environment:
      - TETHER_TOKEN=your-secret-password # Shared with agents
    restart: unless-stopped
```

### Configure Traefik

Tell your Traefik instance to get its routing rules from Tether:

```yaml
providers:
  http:
    endpoint: "http://tether:3000/config"
    pollInterval: "5s"
    headers:
      Authorization: "Bearer your-secret-password"
```

## Configuration

| Env Var         | Flag       | Default             | Description                                                    |
| --------------- | ---------- | ------------------- | -------------------------------------------------------------- |
| `TETHER_TOKEN`  | `--token`  |                     | **Strongly recommended**: Shared secret for agents to connect. |
| `TETHER_PORT`   | `--port`   | `3000`              | Port Tether listens on.                                        |
| `TETHER_NO_WEB` | `--no-web` | `false`             | Disable serving the web UI.                                    |
| `TETHER_CONFIG` | `--config` | `/data/dynamic.yml` | Optional local file for manual Traefik rules.                  |
| `TETHER_DEBUG`  | `--debug`  | `false`             | Enable detailed logging.                                       |
| `TETHER_TRUSTED_PROXIES` | `--trusted-proxies` | `direct` | Where to read client IPs for rate limiting: `direct`, `cloudflare`, `traefik`, or CIDRs. |

## How it behaves

- **Stateless.** Tether keeps everything in memory and never writes files. Agents push their config again when they reconnect.
- **Restarts.** For the first 15s after start, `/config` returns `503`. Traefik keeps its last config during that window while agents reconnect, so routes don't flap.
- **Offline agents.** When an agent disconnects, its routes stay for 30s. If it doesn't come back, they are removed so Traefik stops sending traffic to a dead host.
- **Load balancing.** Run the same app on several machines with the same labels and Tether load balances across them. An HTTP service is shared when every agent routes to it with identical routers (same rule, entrypoints, middlewares, TLS) and the service settings match apart from the servers. Unrelated apps that happen to share a name have different routers, so they are never merged. TCP and UDP are never shared. An offline agent leaves a shared service right away since the others still serve it.
- **Name collisions.** If two sources define the same router, service or middleware name differently, the first one wins: the local file first, then agents sorted by name. Skipped entries are shown in the UI.
- **Local file.** Mount the directory instead of the single file if you want edits picked up live. Editors replace files on save and a single-file bind mount doesn't see that.

---

**Next Step:** Install [Tetherd](https://github.com/MizuchiLabs/tetherd) on your other servers to start connecting them!

## License

Apache 2.0 License - see [LICENSE](LICENSE) for details
