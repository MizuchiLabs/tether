<p align="center">
<img src="./.github/logo.svg" width="80">
<br><br>
<img alt="GitHub Tag" src="https://img.shields.io/github/v/tag/MizuchiLabs/tether?label=Version">
<img alt="GitHub License" src="https://img.shields.io/github/license/MizuchiLabs/tether">
</p>
<img alt="GitHub Issues or Pull Requests" src="https://img.shields.io/github/issues/MizuchiLabs/tether">

# Tether

**Tether** is the central hub for your distributed Traefik setup. It gathers information from all your servers and tells Traefik exactly how to route traffic to your apps.

Think of it as a **central operator**: multiple servers (running [Tetherd](https://github.com/MizuchiLabs/tetherd)) tell Tether "I have these apps running," and Tether gives Traefik a single, master list of all of them.

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

## Quick Start (Server A)

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
| `TETHER_CONFIG` | `--config` | `/data/dynamic.yml` | Optional local file for manual Traefik rules.                  |
| `TETHER_DEBUG`  | `--debug`  | `false`             | Enable detailed logging.                                       |

---

**Next Step:** Install [Tetherd](https://github.com/MizuchiLabs/tetherd) on your other servers to start connecting them!

## License

Apache 2.0 License - see [LICENSE](LICENSE) for details
